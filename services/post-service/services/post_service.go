package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"mime/multipart"
	"regexp"
	"strings"

	"github.com/meloop/post-service/models"
	"github.com/meloop/post-service/repositories"
	"github.com/minio/minio-go/v7"
)

var (
	ErrInvalidPostInput     = errors.New("datos de publicación inválidos")
	ErrInternalPostCreation = errors.New("no se pudo guardar la publicación")
)

// spotifyIDRegex valida que el ID sea alfanumérico y de exactamente 22 caracteres
var spotifyIDRegex = regexp.MustCompile(`^[a-zA-Z0-9]{22}$`)

// ValidateMusicReferences encapsula la lógica de negocio para validar los metadatos musicales (RN-04, MU-07)
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

		if strings.TrimSpace(ref.ReferenceID) == "" {
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

// CreatePostInput representa la entrada validada desde el controlador
type CreatePostInput struct {
	AuthorID       string
	Content        string
	MediaFile      *multipart.FileHeader
	MediaURLs      []string
	MusicReference *models.MusicReference
}

type PostService struct {
	repo    repositories.PostRepository
	storage MediaStorage
}

func NewPostService(repo repositories.PostRepository, minioClient *minio.Client) *PostService {
	var storage MediaStorage
	if minioClient != nil {
		storage = NewMinioMediaStorage(minioClient)
	}
	return &PostService{
		repo:    repo,
		storage: storage,
	}
}

func NewPostServiceWithStorage(repo repositories.PostRepository, storage MediaStorage) *PostService {
	return &PostService{
		repo:    repo,
		storage: storage,
	}
}

// CreatePost ejecuta: validar -> validar música -> subir a MinIO -> insertar en transacción -> compensar si falla (RF-16)
func (s *PostService) CreatePost(ctx context.Context, input CreatePostInput) (*repositories.Post, error) {
	// 1. Validar contenido / adjuntos
	content := strings.TrimSpace(input.Content)
	hasMediaURL := len(input.MediaURLs) > 0 && strings.TrimSpace(input.MediaURLs[0]) != ""
	hasMediaFile := input.MediaFile != nil
	hasMusic := input.MusicReference != nil && strings.TrimSpace(input.MusicReference.ReferenceID) != ""

	if content == "" && !hasMediaURL && !hasMediaFile && !hasMusic {
		return nil, fmt.Errorf("%w: la publicación debe contener texto, multimedia o una referencia musical", ErrInvalidPostInput)
	}

	// 2. Validar referencia musical (RN-04, MU-07) antes de subir archivos o abrir transacción
	if input.MusicReference != nil {
		if err := ValidateMusicReferences([]models.MusicReference{*input.MusicReference}); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrInvalidPostInput, err)
		}
	}

	// 3. Subir archivo a MinIO si viene en la petición
	var finalMediaURL *string
	var uploadedInThisRequest string

	if hasMediaFile {
		if s.storage == nil {
			log.Printf("[CreatePost] Almacenamiento multimedia no configurado")
			return nil, ErrInternalPostCreation
		}
		uploadedURL, err := s.storage.Upload(ctx, input.MediaFile)
		if err != nil {
			// Si el error es de formato o tamaño inválido, se clasifica como 400
			errMsg := err.Error()
			if strings.Contains(errMsg, "formato no soportado") || strings.Contains(errMsg, "supera el tamaño") {
				return nil, fmt.Errorf("%w: %v", ErrInvalidPostInput, err)
			}
			log.Printf("[CreatePost] Error subiendo archivo a MinIO: %v", err)
			return nil, ErrInternalPostCreation
		}
		uploadedInThisRequest = uploadedURL
		finalMediaURL = &uploadedURL
	} else if hasMediaURL {
		url := strings.TrimSpace(input.MediaURLs[0])
		finalMediaURL = &url
	}

	// 4. Insertar en PostgreSQL dentro de una transacción
	post, err := s.repo.CreatePost(ctx, repositories.CreatePostParams{
		AuthorID:      input.AuthorID,
		Content:       content,
		UrlMultimedia: finalMediaURL,
		Music:         input.MusicReference,
	})
	if err != nil {
		log.Printf("[CreatePost] Falló inserción transaccional en PostgreSQL: %v", err)

		// 5. Compensación: eliminar de MinIO el objeto recién subido para no dejar archivos huérfanos
		if uploadedInThisRequest != "" && s.storage != nil {
			if compErr := s.storage.Delete(ctx, uploadedInThisRequest); compErr != nil {
				log.Printf("[CreatePost] Error en compensación MinIO (%s): %v", uploadedInThisRequest, compErr)
			}
		}

		return nil, ErrInternalPostCreation
	}

	return post, nil
}

// EditPost valida los datos antes de enviarlos a la base de datos (RF-17)
func (s *PostService) EditPost(ctx context.Context, postID string, authorID string, newContent string) error {
	if strings.TrimSpace(newContent) == "" {
		return errors.New("el contenido de la publicación no puede estar vacío")
	}
	return s.repo.UpdatePost(ctx, postID, authorID, newContent)
}

// DeletePost elimina la publicación y limpia su archivo en MinIO (RF-18)
func (s *PostService) DeletePost(ctx context.Context, postID string, authorID string) error {
	mediaURL, err := s.repo.DeletePost(ctx, postID, authorID)
	if err != nil {
		return err
	}

	if mediaURL != "" && s.storage != nil {
		if delErr := s.storage.Delete(ctx, mediaURL); delErr != nil {
			log.Printf("[DeletePost] Post %s borrado en BD, fallo limpieza MinIO (%s): %v", postID, mediaURL, delErr)
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
