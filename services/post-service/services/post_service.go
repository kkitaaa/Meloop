package services

import (
	"context"
	"errors"

	// Cambia esto si el nombre del módulo en tu go.mod es distinto

	"github.com/meloop/post-service/repositories"
)

type PostService struct {
	repo *repositories.PostRepository
}

func NewPostService(repo *repositories.PostRepository) *PostService {
	return &PostService{repo: repo}
}

// EditPost valida los datos antes de enviarlos a la base de datos (RF-17)
func (s *PostService) EditPost(ctx context.Context, postID string, authorID string, newContent string) error {
	if newContent == "" {
		return errors.New("el contenido de la publicación no puede estar vacío")
	}

	return s.repo.UpdatePost(ctx, postID, authorID, newContent)
}

// DeletePost procesa la solicitud de eliminación (RF-18)
func (s *PostService) DeletePost(ctx context.Context, postID string, authorID string) error {
	return s.repo.DeletePost(ctx, postID, authorID)
func (s *PostService) GetPost(ctx context.Context, postID string, requesterID string) (*repositories.Post, error) {
	return s.repo.GetPostByID(ctx, postID, requesterID)
}

func (s *PostService) GetAuthorPosts(ctx context.Context, authorID string, requesterID string, limit int, offset int) ([]repositories.Post, error) {
	if limit <= 0 || limit > 50 {
		limit = 10 // Máximo por defecto seguro
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
