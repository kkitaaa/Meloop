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
func (s *FeedService) GetFeed(userID string, page, limit int, algorithm string) ([]models.FeedPost, error) {
	offset := (page - 1) * limit
	return s.repo.GetMainFeed(userID, limit, offset, algorithm)
}
