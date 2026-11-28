package services

import (
	"errors"

	"github.com/meloop/post-service/models"
	"github.com/meloop/post-service/repositories"
)

var (
	ErrInvalidAlgorithm  = errors.New("algoritmo inválido: solo se permite 'recent' o 'relevant'")
	ErrInvalidPagination = errors.New("paginación inválida: 'page' debe ser >= 1 y 'limit' entre 1 y 100")
)

type FeedService struct {
	repo repositories.FeedRepository
}

func NewFeedService(repo repositories.FeedRepository) *FeedService {
	return &FeedService{repo: repo}
}

// GetFeed valida parámetros, calcula el offset y solicita los datos al repositorio
func (s *FeedService) GetFeed(userID string, page, limit int, algorithm string) ([]models.FeedPost, error) {
	if page < 1 || limit < 1 || limit > 100 {
		return nil, ErrInvalidPagination
	}

	if algorithm != "recent" && algorithm != "relevant" {
		return nil, ErrInvalidAlgorithm
	}

	offset := (page - 1) * limit
	return s.repo.GetMainFeed(userID, limit, offset, algorithm)
}
