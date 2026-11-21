package controllers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sort"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/meloop/post-service/models"
	"github.com/meloop/post-service/repositories"
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

func setupFeedRouter(repo repositories.FeedRepository) *gin.Engine {
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

type stubSortedFeedRepository struct {
	posts []models.FeedPost
}

func (s *stubSortedFeedRepository) GetMainFeed(userID string, limit, offset int, algorithm string) ([]models.FeedPost, error) {
	cloned := make([]models.FeedPost, len(s.posts))
	copy(cloned, s.posts)

	sort.Slice(cloned, func(i, j int) bool {
		if algorithm == "relevant" && cloned[i].RelevanceScore != cloned[j].RelevanceScore {
			return cloned[i].RelevanceScore > cloned[j].RelevanceScore
		}
		return cloned[i].CreatedAt.After(cloned[j].CreatedAt)
	})

	if offset >= len(cloned) {
		return []models.FeedPost{}, nil
	}
	end := offset + limit
	if end > len(cloned) {
		end = len(cloned)
	}
	return cloned[offset:end], nil
}

func TestGetMainFeed_RecentVsRelevantAndSerialization(t *testing.T) {
	olderHighScore := models.FeedPost{
		ID:             "post-relevante",
		Content:        "Post con afinidad rock",
		CreatedAt:      time.Date(2026, 11, 20, 10, 0, 0, 0, time.UTC),
		AuthorID:       "user-2",
		AuthorUsername: "ana",
		LikesCount:     10,
		CommentsCount:  4,
		RelevanceScore: 47,
	}
	newerLowScore := models.FeedPost{
		ID:             "post-reciente",
		Content:        "Post nuevo sin interacciones",
		CreatedAt:      time.Date(2026, 11, 21, 12, 0, 0, 0, time.UTC),
		AuthorID:       "user-3",
		AuthorUsername: "luis",
		LikesCount:     0,
		CommentsCount:  0,
		RelevanceScore: 0,
	}

	repo := &stubSortedFeedRepository{posts: []models.FeedPost{olderHighScore, newerLowScore}}
	r := setupFeedRouter(repo)

	// 1. Verificar algoritmo 'relevant' (por defecto) y entrega de relevance_score en JSON
	reqRel, _ := http.NewRequest(http.MethodGet, "/v1/feed?page=1&limit=10", nil)
	reqRel.Header.Set("X-User-ID", "user-1")
	wRel := httptest.NewRecorder()
	r.ServeHTTP(wRel, reqRel)

	if wRel.Code != http.StatusOK {
		t.Fatalf("esperado 200, obtenido %d: %s", wRel.Code, wRel.Body.String())
	}

	var respRel struct {
		Algorithm string            `json:"algorithm"`
		Data      []models.FeedPost `json:"data"`
	}
	if err := json.Unmarshal(wRel.Body.Bytes(), &respRel); err != nil {
		t.Fatalf("error deserializando respuesta relevant: %v", err)
	}
	if respRel.Algorithm != "relevant" {
		t.Fatalf("esperado algorithm='relevant', obtenido '%s'", respRel.Algorithm)
	}
	if len(respRel.Data) != 2 || respRel.Data[0].ID != "post-relevante" || respRel.Data[0].RelevanceScore != 47 {
		t.Fatalf("orden o relevance_score incorrecto en relevant: %+v", respRel.Data)
	}

	// 2. Verificar algoritmo 'recent'
	reqRec, _ := http.NewRequest(http.MethodGet, "/v1/feed?algorithm=recent&page=1&limit=10", nil)
	reqRec.Header.Set("X-User-ID", "user-1")
	wRec := httptest.NewRecorder()
	r.ServeHTTP(wRec, reqRec)

	if wRec.Code != http.StatusOK {
		t.Fatalf("esperado 200, obtenido %d", wRec.Code)
	}

	var respRec struct {
		Algorithm string            `json:"algorithm"`
		Data      []models.FeedPost `json:"data"`
	}
	if err := json.Unmarshal(wRec.Body.Bytes(), &respRec); err != nil {
		t.Fatalf("error deserializando respuesta recent: %v", err)
	}
	if respRec.Data[0].ID != "post-reciente" {
		t.Fatalf("esperado 'post-reciente' primero en recent, obtenido '%s'", respRec.Data[0].ID)
	}
}