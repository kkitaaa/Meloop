package services

import (
	"errors"
	"testing"

	"github.com/meloop/post-service/models"
)

type stubFeedRepo struct {
	called bool
	offset int
	limit  int
}

func (s *stubFeedRepo) GetMainFeed(userID string, limit, offset int, algorithm string) ([]models.FeedPost, error) {
	s.called = true
	s.limit = limit
	s.offset = offset
	return []models.FeedPost{}, nil
}

func TestFeedService_GetFeedValidation(t *testing.T) {
	repo := &stubFeedRepo{}
	svc := NewFeedService(repo)

	if _, err := svc.GetFeed("1", 0, 10, "relevant"); !errors.Is(err, ErrInvalidPagination) {
		t.Fatalf("esperado ErrInvalidPagination para page=0, obtenido %v", err)
	}

	if _, err := svc.GetFeed("1", 1, 101, "relevant"); !errors.Is(err, ErrInvalidPagination) {
		t.Fatalf("esperado ErrInvalidPagination para limit=101, obtenido %v", err)
	}

	if _, err := svc.GetFeed("1", 1, 10, "popular"); !errors.Is(err, ErrInvalidAlgorithm) {
		t.Fatalf("esperado ErrInvalidAlgorithm para algorithm='popular', obtenido %v", err)
	}

	if _, err := svc.GetFeed("1", 2, 20, "relevant"); err != nil {
		t.Fatalf("no se esperaba error con parámetros válidos, obtenido %v", err)
	}

	if !repo.called || repo.offset != 20 || repo.limit != 20 {
		t.Errorf("offset/limit mal calculados: called=%v, offset=%d, limit=%d", repo.called, repo.offset, repo.limit)
	}
}
