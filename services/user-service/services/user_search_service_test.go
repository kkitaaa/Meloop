package services_test

import (
	"context"
	"testing"

	"github.com/meloop/user-service/models"
	"github.com/meloop/user-service/services"
)

type mockUserSearchRepository struct {
	requesterID string
	query       string
	limit       int
	offset      int
	results     []models.UserSearchResult
}

func (m *mockUserSearchRepository) SearchPublicUsers(
	_ context.Context,
	requesterID, query string,
	limit, offset int,
) ([]models.UserSearchResult, error) {
	m.requesterID = requesterID
	m.query = query
	m.limit = limit
	m.offset = offset
	return m.results, nil
}

func TestSearchPublicUsersReturnsStablePageAndNextOffset(t *testing.T) {
	repo := &mockUserSearchRepository{
		results: []models.UserSearchResult{
			{IDUsuario: "1", Username: "ana"},
			{IDUsuario: "2", Username: "anabel"},
			{IDUsuario: "3", Username: "anita"},
		},
	}
	service := services.NewUserSearchService(repo)

	page, err := service.SearchPublicUsers(context.Background(), "requester", " ana ", 2, 4)
	if err != nil {
		t.Fatalf("SearchPublicUsers returned error: %v", err)
	}
	if repo.requesterID != "requester" || repo.query != "ana" || repo.limit != 3 || repo.offset != 4 {
		t.Fatalf("repository received requester=%q query=%q limit=%d offset=%d", repo.requesterID, repo.query, repo.limit, repo.offset)
	}
	if len(page.Users) != 2 || !page.HasMore || page.NextOffset == nil || *page.NextOffset != 6 {
		t.Fatalf("unexpected page: %+v", page)
	}
}

func TestSearchPublicUsersRejectsInvalidQueryAndPagination(t *testing.T) {
	service := services.NewUserSearchService(&mockUserSearchRepository{})
	tests := []struct {
		name   string
		query  string
		limit  int
		offset int
		field  string
	}{
		{name: "short query", query: "a", limit: 20, field: "q"},
		{name: "zero limit", query: "ana", limit: 0, field: "limit"},
		{name: "large limit", query: "ana", limit: 51, field: "limit"},
		{name: "negative offset", query: "ana", limit: 20, offset: -1, field: "offset"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := service.SearchPublicUsers(context.Background(), "requester", test.query, test.limit, test.offset)
			validationErr, ok := err.(*services.ValidationError)
			if !ok || validationErr.Field != test.field {
				t.Fatalf("error = %v, want validation error for %s", err, test.field)
			}
		})
	}
}
