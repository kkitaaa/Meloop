package controllers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
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

	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", response.Code)
	}
	if got := response.Header().Get("Retry-After"); got != "7" {
		t.Fatalf("Retry-After = %q, want 7", got)
	}
	if !strings.Contains(response.Body.String(), "MUSIC_RATE_LIMITED") {
		t.Fatalf("unexpected error body: %s", response.Body.String())
	}
}

func TestSearchValidatesUnicodeLengthAndRequestedLimits(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	routes.SetupRoutes(router, controllers.NewMusicController(fakeCatalog{
		searchResult: map[string]any{"results": []any{}},
	}))
	cases := []struct {
		query  string
		status int
	}{
		{"/v1/music/search?q=" + url.QueryEscape(strings.Repeat("á", 100)) + "&limit=20", 200},
		{"/v1/music/search?q=" + url.QueryEscape(strings.Repeat("á", 101)), 400},
		{"/v1/music/search?q=ok&limit=21", 400},
		{"/v1/music/search?q=ok&offset=1000&limit=11", 400},
	}
	for _, tc := range cases {
		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, tc.query, nil))
		if response.Code != tc.status {
			t.Errorf("%s: status = %d, want %d", tc.query, response.Code, tc.status)
		}
		if tc.status == 400 && !strings.Contains(response.Body.String(), "INVALID_PARAMETER") {
			t.Errorf("%s: unexpected error: %s", tc.query, response.Body.String())
		}
	}
}

func TestDetailReturnsContractErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cases := []struct {
		err    error
		status int
		code   string
	}{
		{&services.SpotifyAPIError{StatusCode: http.StatusNotFound}, http.StatusNotFound, "MUSIC_NOT_FOUND"},
		{&services.SpotifyAPIError{StatusCode: http.StatusBadGateway}, http.StatusServiceUnavailable, "MUSIC_UNAVAILABLE"},
		{&services.SpotifyAPIError{StatusCode: http.StatusTooManyRequests, RetryAfter: "7"}, http.StatusServiceUnavailable, "MUSIC_RATE_LIMITED"},
		{context.DeadlineExceeded, http.StatusServiceUnavailable, "MUSIC_UNAVAILABLE"},
	}
	for _, tc := range cases {
		router := gin.New()
		routes.SetupRoutes(router, controllers.NewMusicController(fakeCatalog{err: tc.err}))
		response := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, "/v1/music/tracks/4iV5W9uYEdYUVa79Axb7Rh", nil)
		router.ServeHTTP(response, request)
		if response.Code != tc.status || !strings.Contains(response.Body.String(), tc.code) {
			t.Errorf("error %v: got %d %s", tc.err, response.Code, response.Body.String())
		}
		if tc.code == "MUSIC_RATE_LIMITED" && response.Header().Get("Retry-After") != "7" {
			t.Errorf("missing Retry-After: %v", response.Header())
		}
	}
}
