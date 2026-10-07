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
}
