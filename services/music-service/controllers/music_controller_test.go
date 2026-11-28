package controllers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/meloop/music-service/controllers"
	"github.com/meloop/music-service/routes"
	"github.com/meloop/music-service/services"
)

type fakeCatalog struct {
	searchResult map[string]any
	detailResult map[string]any
	err          error
}

func (f fakeCatalog) Search(context.Context, string, string, int, int) (map[string]any, error) {
	return f.searchResult, f.err
}

func (f fakeCatalog) Detail(context.Context, string, string) (map[string]any, error) {
	return f.detailResult, f.err
}

func TestSearchReturnsMeloopEnvelopeWithCatalogData(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	routes.SetupRoutes(router, controllers.NewMusicController(fakeCatalog{
		searchResult: map[string]any{"tracks": map[string]any{"items": []any{}}},
	}))

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/v1/music/search?q=hello&type=track", nil)
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", response.Code)
	}
	var body struct {
		Success bool `json:"success"`
		Data    struct {
			Tracks map[string]any `json:"tracks"`
		} `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if !body.Success || body.Data.Tracks == nil {
		t.Fatalf("unexpected response envelope: %s", response.Body.String())
	}
}

func TestSearchRejectsInvalidPagination(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	routes.SetupRoutes(router, controllers.NewMusicController(fakeCatalog{}))

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/v1/music/search?q=hello&offset=1001", nil)
	router.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", response.Code)
	}
}

func TestDetailRejectsInvalidSpotifyID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	routes.SetupRoutes(router, controllers.NewMusicController(fakeCatalog{}))

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/v1/music/tracks/not-a-spotify-id", nil)
	router.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", response.Code)
	}
}

func TestRateLimitPropagatesRetryAfterWithoutProviderDetails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	routes.SetupRoutes(router, controllers.NewMusicController(fakeCatalog{
		err: &services.SpotifyAPIError{
			StatusCode: http.StatusTooManyRequests,
			RetryAfter: "7",
		},
	}))

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/v1/music/search?q=hello", nil)
	router.ServeHTTP(response, request)

	if response.Code != http.StatusTooManyRequests {
		t.Fatalf("status = %d, want 429", response.Code)
	}
	if got := response.Header().Get("Retry-After"); got != "7" {
		t.Fatalf("Retry-After = %q, want 7", got)
	}
}
