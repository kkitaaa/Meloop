package services

import (
	"github.com/meloop/post-service/models"
	"github.com/meloop/post-service/repositories"
)

type FeedService struct {
	repo repositories.FeedRepository
}

func NewFeedService(repo repositories.FeedRepository) *FeedService {
	return &FeedService{repo: repo}
}

// GetFeed calcula el paginado y solicita los datos al repositorio
func (s *FeedService) GetFeed(userID string, page, limit int) ([]models.FeedPost, error) {
	// Valores por defecto de seguridad
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 10 // Máximo 50 posts por petición, por defecto 10
	}
	
	offset := (page - 1) * limit
	return s.repo.GetMainFeed(userID, limit, offset)
}