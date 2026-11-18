package controllers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/meloop/post-service/models"
	"github.com/meloop/post-service/services"
)

// mockFeedRepository simula el acceso a datos para pruebas unitarias aisladas
type mockFeedRepository struct {
	lastUserID    string
	lastLimit     int
	lastOffset    int
	lastAlgorithm string
	mockData      []models.FeedPost
	mockErr       error
}

func (m *mockFeedRepository) GetMainFeed(userID string, limit, offset int, algorithm string) ([]models.FeedPost, error) {
	m.lastUserID = userID
	m.lastLimit = limit
	m.lastOffset = offset
	m.lastAlgorithm = algorithm
	return m.mockData, m.mockErr
}

func setupFeedRouter(repo *mockFeedRepository) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	svc := services.NewFeedService(repo)
	ctrl := NewFeedController(svc)
	r.GET("/v1/feed", ctrl.GetMainFeed)
	return r
}

func TestGetMainFeed_UnauthorizedWithoutHeader(t *testing.T) {
	repo := &mockFeedRepository{}
	r := setupFeedRouter(repo)

	req, _ := http.NewRequest(http.MethodGet, "/v1/feed", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("esperado status 401, obtenido %d", w.Code)
	}
}

func TestGetMainFeed_InvalidParameters(t *testing.T) {
	repo := &mockFeedRepository{}
	r := setupFeedRouter(repo)

	cases := []struct {
		name string
		url  string
	}{
		{"algoritmo invalido", "/v1/feed?algorithm=aleatorio"},
		{"page no numerico", "/v1/feed?page=abc"},
		{"page cero", "/v1/feed?page=0"},
		{"limit negativo", "/v1/feed?limit=-5"},
		{"limit mayor a 100", "/v1/feed?limit=150"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodGet, tc.url, nil)
			req.Header.Set("X-User-ID", "1")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("caso '%s': esperado status 400, obtenido %d", tc.name, w.Code)
			}
		})
	}
}

func TestGetMainFeed_ValidRequestAndPagination(t *testing.T) {
	repo := &mockFeedRepository{
		mockData: []models.FeedPost{},
	}
	r := setupFeedRouter(repo)

	req, _ := http.NewRequest(http.MethodGet, "/v1/feed?algorithm=recent&page=3&limit=15", nil)
	req.Header.Set("X-User-ID", "42")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("esperado status 200, obtenido %d", w.Code)
	}

	if repo.lastUserID != "42" || repo.lastAlgorithm != "recent" || repo.lastLimit != 15 || repo.lastOffset != 30 {
		t.Errorf("parámetros enviados al repo incorrectos: user=%s, algo=%s, limit=%d, offset=%d",
			repo.lastUserID, repo.lastAlgorithm, repo.lastLimit, repo.lastOffset)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("error parseando JSON de respuesta: %v", err)
	}

	if resp["algorithm"] != "recent" {
		t.Errorf("esperado algorithm 'recent' en respuesta, obtenido %v", resp["algorithm"])
	}
}
