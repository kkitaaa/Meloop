package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"regexp"

	"github.com/meloop/post-service/models"
	"github.com/meloop/post-service/repositories"
	"github.com/minio/minio-go/v7"
)

// spotifyIDRegex valida que el ID sea alfanumérico y de exactamente 22 caracteres
var spotifyIDRegex = regexp.MustCompile(`^[a-zA-Z0-9]{22}$`)

// ValidateMusicReferences encapsula la lógica de negocio para validar los metadatos musicales
func ValidateMusicReferences(refs []models.MusicReference) error {
	for _, ref := range refs {
		if ref.ReferenceType != models.TrackReference &&
			ref.ReferenceType != models.ArtistReference &&
			ref.ReferenceType != models.AlbumReference {
			return fmt.Errorf("tipo de referencia inválido: %s", ref.ReferenceType)
		}

		if ref.Provider != "spotify" && ref.Provider != "youtube" {
			return fmt.Errorf("proveedor no soportado: %s", ref.Provider)
		}

		if ref.ReferenceID == "" {
			return errors.New("el reference_id no puede estar vacío")
		}

		if ref.Provider == "spotify" {
			if !spotifyIDRegex.MatchString(ref.ReferenceID) {
				return fmt.Errorf("formato de spotify_id inválido: %s", ref.ReferenceID)
			}
		}
	}
	return nil
}

type PostService struct {
	repo        repositories.PostRepository
	minioClient *minio.Client
}

func NewPostService(repo repositories.PostRepository, minioClient *minio.Client) *PostService {
	return &PostService{
		repo:        repo,
		minioClient: minioClient,
	}
}

// CreatePost valida y envía la nueva publicación al repositorio
func (s *PostService) CreatePost(ctx context.Context, authorID string, content string, mediaURLs []string) (*repositories.Post, error) {
	if content == "" {
		return nil, errors.New("el contenido de la publicación no puede estar vacío")
	}

	var url *string
	if len(mediaURLs) > 0 && mediaURLs[0] != "" {
		url = &mediaURLs[0]
	}

	return s.repo.CreatePost(ctx, authorID, content, url)
}

// EditPost valida los datos antes de enviarlos a la base de datos (RF-17)
func (s *PostService) EditPost(ctx context.Context, postID string, authorID string, newContent string) error {
	if newContent == "" {
		return errors.New("el contenido de la publicación no puede estar vacío")
	}

	return s.repo.UpdatePost(ctx, postID, authorID, newContent)
}

// DeletePost elimina la publicación en BD y limpia su archivo multimedia en MinIO si existe (RF-18)
func (s *PostService) DeletePost(ctx context.Context, postID string, authorID string) error {
	mediaURL, err := s.repo.DeletePost(ctx, postID, authorID)
	if err != nil {
		return err
	}

	if mediaURL != "" && s.minioClient != nil {
		if delErr := DeleteMedia(ctx, s.minioClient, mediaURL); delErr != nil {
			log.Printf("[Alerta] Post %s eliminado de BD, pero falló la limpieza en MinIO (%s): %v", postID, mediaURL, delErr)
		}
	}

	return nil
}

func (s *PostService) GetPost(ctx context.Context, postID string, requesterID string) (*repositories.Post, error) {
	return s.repo.GetPostByID(ctx, postID, requesterID)
}

func (s *PostService) GetAuthorPosts(ctx context.Context, authorID string, requesterID string, limit int, offset int) ([]repositories.Post, error) {
	if limit <= 0 || limit > 50 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}
	return s.repo.GetPostsByAuthor(ctx, authorID, requesterID, limit, offset)
}

func (s *PostService) SearchPosts(ctx context.Context, query string, requesterID string, limit int, offset int) ([]repositories.Post, error) {
	if limit <= 0 || limit > 50 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}
	return s.repo.SearchPosts(ctx, query, requesterID, limit, offset)
}

// RF-20: Agregar Me gusta
func (s *PostService) AddLike(ctx context.Context, postID string, userID string) error {
	return s.repo.AddLike(ctx, postID, userID)
}

// RF-20: Quitar Me gusta
func (s *PostService) RemoveLike(ctx context.Context, postID string, userID string) error {
	return s.repo.RemoveLike(ctx, postID, userID)
}
