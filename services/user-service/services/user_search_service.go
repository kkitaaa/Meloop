package services

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/meloop/user-service/models"
	"github.com/meloop/user-service/repositories"
)

const (
	DefaultUserSearchLimit = 20
	MaxUserSearchLimit     = 50
	MaxUserSearchOffset    = 1000000
)

type UserSearchService interface {
	SearchPublicUsers(ctx context.Context, requesterID, query string, limit, offset int) (*models.UserSearchPage, error)
}

type userSearchService struct {
	repo repositories.UserSearchRepository
}

func NewUserSearchService(repo repositories.UserSearchRepository) UserSearchService {
	return &userSearchService{repo: repo}
}

func (s *userSearchService) SearchPublicUsers(
	ctx context.Context,
	requesterID, query string,
	limit, offset int,
) (*models.UserSearchPage, error) {
	requesterID = strings.TrimSpace(requesterID)
	query = strings.TrimSpace(query)

	if requesterID == "" {
		return nil, ErrUserNotFound
	}
	queryLength := utf8.RuneCountInString(query)
	if queryLength < 2 {
		return nil, &ValidationError{Field: "q", Issue: "too_short", Message: "La búsqueda debe tener al menos 2 caracteres"}
	}
	if queryLength > 100 {
		return nil, &ValidationError{Field: "q", Issue: "too_long", Message: "La búsqueda no puede superar los 100 caracteres"}
	}
	if limit < 1 || limit > MaxUserSearchLimit {
		return nil, &ValidationError{Field: "limit", Issue: "out_of_range", Message: fmt.Sprintf("El límite debe estar entre 1 y %d", MaxUserSearchLimit)}
	}
	if offset < 0 || offset > MaxUserSearchOffset {
		return nil, &ValidationError{Field: "offset", Issue: "out_of_range", Message: fmt.Sprintf("El offset debe estar entre 0 y %d", MaxUserSearchOffset)}
	}

	results, err := s.repo.SearchPublicUsers(ctx, requesterID, query, limit+1, offset)
	if err != nil {
		return nil, err
	}

	hasMore := len(results) > limit
	if hasMore {
		results = results[:limit]
	}

	page := &models.UserSearchPage{
		Query:   query,
		Users:   results,
		Limit:   limit,
		Offset:  offset,
		HasMore: hasMore,
	}
	if hasMore {
		nextOffset := offset + len(results)
		page.NextOffset = &nextOffset
	}
	return page, nil
}
