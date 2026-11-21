package controllers_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/meloop/social-service/controllers"
	"github.com/meloop/social-service/models"
	"github.com/meloop/social-service/repositories"
	"github.com/meloop/social-service/routes"
)

type mutualFriendsRepositoryStub struct {
	data models.MutualFriendsData
	err  error
}

func (s mutualFriendsRepositoryStub) ListMutualFriends(
	_ context.Context,
	_, _, _ string,
	_ int,
) (models.MutualFriendsData, error) {
	return s.data, s.err
}

func newMutualFriendsRouter(repository repositories.MutualFriendsReader) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	routes.SetupMutualFriendsRoutes(router, controllers.NewMutualFriendsController(repository))
	return router
}

func TestListMutualFriends_ReturnsCursorAndUniquePage(t *testing.T) {
	router := newMutualFriendsRouter(mutualFriendsRepositoryStub{
		data: models.MutualFriendsData{
			TotalCount: 3,
			Items: []models.MutualFriend{
				{ID: "friend-a", Username: "Ana"},
				{ID: "friend-b", Username: "Beto"},
				{ID: "friend-c", Username: "Caro"},
			},
		},
	})
	request := httptest.NewRequest(http.MethodGet, "/users/target/mutual-friends?limit=2", nil)
	request.Header.Set("X-User-ID", "viewer")
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", response.Code, response.Body.String())
	}
	if !containsJSON(response.Body.String(), `"totalCount":3`) ||
		!containsJSON(response.Body.String(), `"hasMore":true`) ||
		!containsJSON(response.Body.String(), `"nextCursor":"ZnJpZW5kLWI"`) {
		t.Fatalf("response is missing pagination metadata: %s", response.Body.String())
	}
	if containsJSON(response.Body.String(), `"id_usuario":"friend-c"`) {
		t.Fatalf("response returned an item beyond the requested page: %s", response.Body.String())
	}
}

func TestListMutualFriends_HiddenReturnsCountWithoutItems(t *testing.T) {
	router := newMutualFriendsRouter(mutualFriendsRepositoryStub{
		data: models.MutualFriendsData{TotalCount: 4, Hidden: true},
	})
	request := httptest.NewRequest(http.MethodGet, "/users/target/mutual-friends", nil)
	request.Header.Set("X-User-ID", "viewer")
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK ||
		!containsJSON(response.Body.String(), `"hidden":true`) ||
		!containsJSON(response.Body.String(), `"totalCount":4`) ||
		!containsJSON(response.Body.String(), `"items":[]`) {
		t.Fatalf("hidden response does not respect privacy: %d %s", response.Code, response.Body.String())
	}
}

func TestListMutualFriends_NotFoundDoesNotRevealBlocking(t *testing.T) {
	router := newMutualFriendsRouter(mutualFriendsRepositoryStub{err: repositories.ErrUserNotFound})
	request := httptest.NewRequest(http.MethodGet, "/users/target/mutual-friends", nil)
	request.Header.Set("X-User-ID", "viewer")
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusNotFound ||
		containsJSON(response.Body.String(), "bloque") {
		t.Fatalf("expected a neutral 404, got %d %s", response.Code, response.Body.String())
	}
}

func TestListMutualFriends_RejectsInvalidCursor(t *testing.T) {
	router := newMutualFriendsRouter(mutualFriendsRepositoryStub{})
	request := httptest.NewRequest(http.MethodGet, "/users/target/mutual-friends?cursor=!!!", nil)
	request.Header.Set("X-User-ID", "viewer")
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", response.Code, response.Body.String())
	}
}

func TestListMutualFriends_RepositoryFailureReturnsInternalError(t *testing.T) {
	router := newMutualFriendsRouter(mutualFriendsRepositoryStub{err: errors.New("db unavailable")})
	request := httptest.NewRequest(http.MethodGet, "/users/target/mutual-friends", nil)
	request.Header.Set("X-User-ID", "viewer")
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", response.Code, response.Body.String())
	}
}

func containsJSON(body, value string) bool {
	return strings.Contains(body, value)
}
