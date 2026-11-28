package services

import (
	"context"
	"errors"
	"mime/multipart"
	"testing"
	"time"

	"github.com/meloop/post-service/repositories"
)

type stubPostRepo struct {
	posts      map[string]repositories.Post
	failCreate bool
}

func newStubPostRepo() *stubPostRepo {
	return &stubPostRepo{posts: make(map[string]repositories.Post)}
}

func (r *stubPostRepo) CreatePost(_ context.Context, params repositories.CreatePostParams) (*repositories.Post, error) {
	if r.failCreate {
		return nil, errors.New("db insert failure after client disconnect")
	}
	post := repositories.Post{
		ID:            "post-1",
		AuthorID:      params.AuthorID,
		Content:       params.Content,
		UrlMultimedia: params.UrlMultimedia,
		CreatedAt:     time.Now().UTC(),
	}
	r.posts[post.ID] = post
	return &post, nil
}

func (r *stubPostRepo) UpdatePost(_ context.Context, _, _, _ string) error { return nil }

func (r *stubPostRepo) DeletePost(_ context.Context, postID, authorID string) (string, error) {
	post, ok := r.posts[postID]
	if !ok {
		return "", repositories.ErrPostNotFound
	}
	if post.AuthorID != authorID {
		return "", repositories.ErrUnauthorized
	}
	delete(r.posts, postID)
	if post.UrlMultimedia == nil {
		return "", nil
	}
	return *post.UrlMultimedia, nil
}

func (r *stubPostRepo) GetPostByID(_ context.Context, _, _ string) (*repositories.Post, error) {
	return nil, nil
}
func (r *stubPostRepo) GetPostsByAuthor(_ context.Context, _, _ string, _, _ int) ([]repositories.Post, error) {
	return nil, nil
}
func (r *stubPostRepo) SearchPosts(_ context.Context, _, _ string, _, _ int) ([]repositories.Post, error) {
	return nil, nil
}
func (r *stubPostRepo) AddLike(_ context.Context, _, _ string) error    { return nil }
func (r *stubPostRepo) RemoveLike(_ context.Context, _, _ string) error { return nil }

type spyMediaStorage struct {
	host                string
	deletedKeys         []string
	deleteAttempts      int
	failDeletesBeforeOK int
	onUpload            func()
}

func (s *spyMediaStorage) Upload(_ context.Context, ownerID string, _ *multipart.FileHeader) (string, error) {
	if s.onUpload != nil {
		s.onUpload()
	}
	return "http://" + s.host + "/meloop-media/users/" + ownerID + "/550e8400-e29b-41d4-a716-446655440000.jpg", nil
}

func (s *spyMediaStorage) Delete(ctx context.Context, ownerID string, mediaURL string) error {
	s.deleteAttempts++
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if s.deleteAttempts <= s.failDeletesBeforeOK {
		return errors.New("temporary minio timeout")
	}
	if key, ok := ExtractOwnedMinioObjectKey(mediaURL, ownerID, s.host); ok {
		s.deletedKeys = append(s.deletedKeys, key)
	}
	return nil
}

func (s *spyMediaStorage) EndpointHost() string {
	return s.host
}

func TestCreatePost_RejectsForeignMinioURL(t *testing.T) {
	repo := newStubPostRepo()
	storage := &spyMediaStorage{host: "minio.local:9000"}
	svc := NewPostServiceWithStorage(repo, storage)

	victimURL := "http://minio.local:9000/meloop-media/users/victim-user/550e8400-e29b-41d4-a716-446655440000.jpg"
	_, err := svc.CreatePost(context.Background(), CreatePostInput{
		AuthorID:  "attacker-user",
		Content:   "intento de adjuntar objeto ajeno",
		MediaURLs: []string{victimURL},
	})
	if !errors.Is(err, ErrInvalidPostInput) {
		t.Fatalf("esperado ErrInvalidPostInput al adjuntar URL de otro usuario, obtenido: %v", err)
	}
	if len(repo.posts) != 0 {
		t.Fatalf("no debió guardarse ningún post en el repositorio")
	}
}

func TestCreateAndDeletePost_ExternalURLDoesNotDeleteMinioObject(t *testing.T) {
	repo := newStubPostRepo()
	storage := &spyMediaStorage{host: "minio.local:9000"}
	svc := NewPostServiceWithStorage(repo, storage)

	spoofedExternal := "https://evil.example.com/meloop-media/users/victim-user/550e8400-e29b-41d4-a716-446655440000.jpg"
	post, err := svc.CreatePost(context.Background(), CreatePostInput{
		AuthorID:  "attacker-user",
		Content:   "enlace externo",
		MediaURLs: []string{spoofedExternal},
	})
	if err != nil {
		t.Fatalf("error inesperado creando post con enlace externo: %v", err)
	}

	if err := svc.DeletePost(context.Background(), post.ID, "attacker-user"); err != nil {
		t.Fatalf("error inesperado eliminando post: %v", err)
	}

	if len(storage.deletedKeys) != 0 {
		t.Fatalf("no debió eliminarse ningún objeto de MinIO para un enlace externo, eliminado: %v", storage.deletedKeys)
	}
}

func TestCreateAndDeletePost_DeletesOnlyOwnerScopedMinioKey(t *testing.T) {
	repo := newStubPostRepo()
	storage := &spyMediaStorage{host: "minio.local:9000"}
	svc := NewPostServiceWithStorage(repo, storage)

	ownedURL := "http://minio.local:9000/meloop-media/users/author-1/550e8400-e29b-41d4-a716-446655440000.jpg"
	post, err := svc.CreatePost(context.Background(), CreatePostInput{
		AuthorID:  "author-1",
		Content:   "mi imagen",
		MediaURLs: []string{ownedURL},
	})
	if err != nil {
		t.Fatalf("error inesperado creando post con objeto propio: %v", err)
	}

	if err := svc.DeletePost(context.Background(), post.ID, "author-1"); err != nil {
		t.Fatalf("error inesperado eliminando post: %v", err)
	}

	if len(storage.deletedKeys) != 1 || storage.deletedKeys[0] != "users/author-1/550e8400-e29b-41d4-a716-446655440000.jpg" {
		t.Fatalf("clave eliminada inesperada: %v", storage.deletedKeys)
	}
}

func TestCreatePost_CompensatesWithIndependentContextAndRetriesWhenClientCancels(t *testing.T) {
	repo := newStubPostRepo()
	repo.failCreate = true

	reqCtx, cancelReq := context.WithCancel(context.Background())
	storage := &spyMediaStorage{
		host:                "minio.local:9000",
		failDeletesBeforeOK: 1, // primer intento falla, segundo reintento tiene éxito
		onUpload: func() {
			// Simula que el cliente HTTP desconecta justo después de subir el archivo
			cancelReq()
		},
	}
	svc := NewPostServiceWithStorage(repo, storage)

	_, err := svc.CreatePost(reqCtx, CreatePostInput{
		AuthorID:  "author-1",
		Content:   "post con cancelación",
		MediaFile: &multipart.FileHeader{Filename: "foto.jpg", Size: 128},
	})
	if !errors.Is(err, ErrInternalPostCreation) {
		t.Fatalf("esperado ErrInternalPostCreation, obtenido: %v", err)
	}

	if storage.deleteAttempts != 2 {
		t.Fatalf("esperados 2 intentos de compensación con reintentos, obtenidos: %d", storage.deleteAttempts)
	}
	if len(storage.deletedKeys) != 1 || storage.deletedKeys[0] != "users/author-1/550e8400-e29b-41d4-a716-446655440000.jpg" {
		t.Fatalf("la compensación debió eliminar el archivo en MinIO pese a cancelarse reqCtx, deletedKeys=%v", storage.deletedKeys)
	}
}
