package controllers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/meloop/user-service/models"
)

type mockUserSearchService struct {
	requesterID string
	query       string
	limit       int
	offset      int
	err         error
	page        *models.UserSearchPage
}

func (m *mockUserSearchService) SearchPublicUsers(
	_ context.Context,
	requesterID, query string,
	limit, offset int,
) (*models.UserSearchPage, error) {
	m.requesterID = requesterID
	m.query = query
	m.limit = limit
	m.offset = offset
	return m.page, m.err
}

func TestUserSearchControllerReturnsBadRequestForInvalidPagination(t *testing.T) {
	gin.SetMode(gin.TestMode)
	searchService := &mockUserSearchService{}
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(ContextUserIDKey, "requester")
		c.Next()
	})
	router.GET("/users/search", NewUserSearchController(searchService).Search)

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/users/search?q=ana&limit=oops", nil))

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body: %s", response.Code, response.Body.String())
	}
	if searchService.query != "" {
		t.Fatal("search service should not be called with invalid pagination")
	}
}

func TestUserSearchControllerPassesAuthenticatedUserAndPagination(t *testing.T) {
	gin.SetMode(gin.TestMode)
	searchService := &mockUserSearchService{
		page: &models.UserSearchPage{
			Query:  "ana",
			Users:  []models.UserSearchResult{{IDUsuario: "user-2", Username: "ana"}},
			Limit:  10,
			Offset: 2,
		},
	}
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(ContextUserIDKey, "requester-1")
		c.Next()
	})
	router.GET("/users/search", NewUserSearchController(searchService).Search)

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/users/search?q=ana&limit=10&offset=2", nil))

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body: %s", response.Code, response.Body.String())
	}
	if searchService.requesterID != "requester-1" || searchService.query != "ana" || searchService.limit != 10 || searchService.offset != 2 {
		t.Fatalf("unexpected search arguments: %+v", searchService)
	}
	if !strings.Contains(response.Body.String(), `"id_usuario":"user-2"`) {
		t.Fatalf("response does not contain search result: %s", response.Body.String())
	}
}

func TestUserSearchControllerRequiresAuthentication(t *testing.T) {
	gin.SetMode(gin.TestMode)
	searchService := &mockUserSearchService{}
	router := gin.New()
	router.GET("/users/search", NewUserSearchController(searchService).Search)

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/users/search?q=ana", nil))

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", response.Code)
	}
}
