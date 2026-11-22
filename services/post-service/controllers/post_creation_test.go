package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/meloop/post-service/repositories"
	"github.com/meloop/post-service/services"
)

type fakeMediaStorage struct {
	uploaded      []string
	deleted       []string
	deletedOwners []string
}

func (f *fakeMediaStorage) Upload(ctx context.Context, ownerID string, file *multipart.FileHeader) (string, error) {
	url := "http://localhost:9000/meloop-media/users/" + ownerID + "/550e8400-e29b-41d4-a716-446655440000-" + file.Filename
	f.uploaded = append(f.uploaded, url)
	return url, nil
}

func (f *fakeMediaStorage) Delete(ctx context.Context, ownerID string, mediaURL string) error {
	f.deleted = append(f.deleted, mediaURL)
	f.deletedOwners = append(f.deletedOwners, ownerID)
	return nil
}

func (f *fakeMediaStorage) EndpointHost() string {
	return "localhost:9000"
}

type fakePostRepo struct {
	posts      map[string]*repositories.Post
	failInsert bool
}

func newFakePostRepo() *fakePostRepo {
	return &fakePostRepo{posts: make(map[string]*repositories.Post)}
}

func (r *fakePostRepo) CreatePost(ctx context.Context, params repositories.CreatePostParams) (*repositories.Post, error) {
	if r.failInsert {
		return nil, errors.New("pq: duplicate key or internal database failure details")
	}
	var musicID *string
	if params.Music != nil {
		id := params.Music.ReferenceID
		musicID = &id
	}
	post := &repositories.Post{
		ID:            "550e8400-e29b-41d4-a716-446655440000",
		AuthorID:      params.AuthorID,
		Content:       params.Content,
		MusicID:       musicID,
		UrlMultimedia: params.UrlMultimedia,
		CreatedAt:     time.Now().UTC(),
	}
	r.posts[post.ID] = post
	return post, nil
}

func (r *fakePostRepo) UpdatePost(ctx context.Context, postID, authorID, newContent string) error {
	return nil
}

func (r *fakePostRepo) DeletePost(ctx context.Context, postID, authorID string) (string, error) {
	delete(r.posts, postID)
	return "", nil
}

func (r *fakePostRepo) GetPostByID(ctx context.Context, postID, requesterID string) (*repositories.Post, error) {
	p, ok := r.posts[postID]
	if !ok {
		return nil, repositories.ErrPostNotFound
	}
	return p, nil
}

func (r *fakePostRepo) GetPostsByAuthor(ctx context.Context, authorID, requesterID string, limit, offset int) ([]repositories.Post, error) {
	return []repositories.Post{}, nil
}

func (r *fakePostRepo) SearchPosts(ctx context.Context, searchTerm, requesterID string, limit, offset int) ([]repositories.Post, error) {
	return []repositories.Post{}, nil
}

func (r *fakePostRepo) AddLike(ctx context.Context, postID, userID string) error    { return nil }
func (r *fakePostRepo) RemoveLike(ctx context.Context, postID, userID string) error { return nil }

func setupCreatePostRouter(repo *fakePostRepo, storage *fakeMediaStorage) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		if uid := c.GetHeader("X-User-ID"); uid != "" {
			c.Set("userID", uid)
		}
		c.Next()
	})
	svc := services.NewPostServiceWithStorage(repo, storage)
	ctrl := NewPostController(svc)
	r.POST("/api/v1/posts", ctrl.CreatePost)
	r.GET("/api/v1/posts/:id", ctrl.GetPostByID)
	return r
}

func TestCreatePost_LifecycleAndCompensation(t *testing.T) {
	repo := newFakePostRepo()
	storage := &fakeMediaStorage{}
	r := setupCreatePostRouter(repo, storage)

	// 1. Texto solo + verifica que author_id del body sea ignorado (PST-02)
	bodyText := `{"author_id":"attacker-id","content":"Mi primer post real en Meloop"}`
	req1, _ := http.NewRequest(http.MethodPost, "/api/v1/posts", bytes.NewBufferString(bodyText))
	req1.Header.Set("Content-Type", "application/json")
	req1.Header.Set("X-User-ID", "real-user-1")
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)

	if w1.Code != http.StatusCreated {
		t.Fatalf("esperado 201, obtenido %d: %s", w1.Code, w1.Body.String())
	}
	var created1 repositories.Post
	_ = json.Unmarshal(w1.Body.Bytes(), &created1)
	if created1.AuthorID != "real-user-1" {
		t.Fatalf("PST-02 violado: esperado author_id='real-user-1', obtenido '%s'", created1.AuthorID)
	}
	if created1.CreatedAt.IsZero() {
		t.Fatal("esperado fecha_creacion del servidor no vacía")
	}

	// Verificar lectura posterior vía GET
	reqGet, _ := http.NewRequest(http.MethodGet, "/api/v1/posts/"+created1.ID, nil)
	wGet := httptest.NewRecorder()
	r.ServeHTTP(wGet, reqGet)
	if wGet.Code != http.StatusOK {
		t.Fatalf("esperado 200 al leer post persistido, obtenido %d", wGet.Code)
	}

	// 2. Texto + canción válida (22 chars Spotify) -> 201
	bodyMusic := `{"content":"Temazo recomendado","music_id":"4cOdK2wGLETKBW3PvgPWqT"}`
	req2, _ := http.NewRequest(http.MethodPost, "/api/v1/posts", bytes.NewBufferString(bodyMusic))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("X-User-ID", "real-user-1")
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusCreated {
		t.Fatalf("esperado 201 para texto + canción válida, obtenido %d", w2.Code)
	}

	// Música inválida -> 400 sin tocar MinIO
	bodyBadMusic := `{"content":"Canción inválida","music_id":"id-invalido"}`
	reqBad, _ := http.NewRequest(http.MethodPost, "/api/v1/posts", bytes.NewBufferString(bodyBadMusic))
	reqBad.Header.Set("Content-Type", "application/json")
	reqBad.Header.Set("X-User-ID", "real-user-1")
	wBad := httptest.NewRecorder()
	r.ServeHTTP(wBad, reqBad)
	if wBad.Code != http.StatusBadRequest {
		t.Fatalf("esperado 400 para música inválida, obtenido %d", wBad.Code)
	}

	// 3. Texto + imagen multipart -> 201
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	_ = mw.WriteField("content", "Post con portada")
	partHeader := make(textproto.MIMEHeader)
	partHeader.Set("Content-Disposition", `form-data; name="file"; filename="cover.png"`)
	partHeader.Set("Content-Type", "image/png")
	part, _ := mw.CreatePart(partHeader)
	_, _ = part.Write([]byte("fake-png-binary"))
	_ = mw.Close()

	req3, _ := http.NewRequest(http.MethodPost, "/api/v1/posts", &buf)
	req3.Header.Set("Content-Type", mw.FormDataContentType())
	req3.Header.Set("X-User-ID", "real-user-1")
	w3 := httptest.NewRecorder()
	r.ServeHTTP(w3, req3)
	if w3.Code != http.StatusCreated {
		t.Fatalf("esperado 201 para texto + imagen, obtenido %d: %s", w3.Code, w3.Body.String())
	}
	if len(storage.uploaded) != 1 {
		t.Fatalf("esperado 1 archivo subido a MinIO, obtenidos %d", len(storage.uploaded))
	}

	// 4. Fallo de inserción con imagen -> compensación elimina archivo en MinIO y responde 500 sin detalles internos
	repo.failInsert = true
	var bufFail bytes.Buffer
	mwFail := multipart.NewWriter(&bufFail)
	_ = mwFail.WriteField("content", "Post que fallará en BD")
	partHeaderFail := make(textproto.MIMEHeader)
	partHeaderFail.Set("Content-Disposition", `form-data; name="file"; filename="orphan.png"`)
	partHeaderFail.Set("Content-Type", "image/png")
	partFail, _ := mwFail.CreatePart(partHeaderFail)
	_, _ = partFail.Write([]byte("fake-png-binary"))
	_ = mwFail.Close()

	req4, _ := http.NewRequest(http.MethodPost, "/api/v1/posts", &bufFail)
	req4.Header.Set("Content-Type", mwFail.FormDataContentType())
	req4.Header.Set("X-User-ID", "real-user-1")
	w4 := httptest.NewRecorder()
	r.ServeHTTP(w4, req4)

	if w4.Code != http.StatusInternalServerError {
		t.Fatalf("esperado 500 ante fallo de BD, obtenido %d", w4.Code)
	}
	if strings.Contains(w4.Body.String(), "pq:") || strings.Contains(w4.Body.String(), "duplicate key") {
		t.Fatalf("se filtraron detalles internos en el error 500: %s", w4.Body.String())
	}
	if len(storage.deleted) != 1 || storage.deleted[0] != storage.uploaded[1] {
		t.Fatalf("compensación fallida: esperado borrado de %s, borrados=%v", storage.uploaded[1], storage.deleted)
	}
	if len(storage.deletedOwners) != 1 || storage.deletedOwners[0] != "real-user-1" {
		t.Fatalf("compensación fallida: esperado ownerID 'real-user-1', obtenido=%v", storage.deletedOwners)
	}
}
