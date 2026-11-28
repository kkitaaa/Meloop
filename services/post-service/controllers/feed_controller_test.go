package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/meloop/post-service/client"
	"github.com/meloop/post-service/models"
	"github.com/meloop/post-service/repositories"
	"github.com/meloop/post-service/services"
)

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

type stubSessionValidator struct {
	userID string
	err    error
}

func (s stubSessionValidator) ValidateToken(_ context.Context, _ string) (string, error) {
	return s.userID, s.err
}

func setupFeedRouter(repo repositories.FeedRepository, validator SessionValidator) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	svc := services.NewFeedService(repo)
	ctrl := NewFeedController(svc)

	v1 := r.Group("/v1")
	v1.Use(AuthRequired(validator))
	v1.GET("/feed", ctrl.GetMainFeed)
	return r
}

func TestGetMainFeed_UnauthorizedWithoutBearerTokenEvenWithSpoofedHeader(t *testing.T) {
	repo := &mockFeedRepository{}
	r := setupFeedRouter(repo, stubSessionValidator{userID: "user-real"})

	req, _ := http.NewRequest(http.MethodGet, "/v1/feed", nil)
	req.Header.Set("X-User-ID", "attacker-spoofed-id")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("esperado status 401 sin token Bearer, obtenido %d", w.Code)
	}
	if repo.lastUserID != "" {
		t.Fatalf("no se debía consultar el repositorio sin autenticación, userID=%s", repo.lastUserID)
	}
}

func TestGetMainFeed_RejectsInvalidSessionAndHandlesAuthOutage(t *testing.T) {
	repo := &mockFeedRepository{}

	t.Run("sesion invalida o expirada retorna 401", func(t *testing.T) {
		r := setupFeedRouter(repo, stubSessionValidator{err: client.ErrInvalidSession})
		req, _ := http.NewRequest(http.MethodGet, "/v1/feed", nil)
		req.Header.Set("Authorization", "Bearer expired-token")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Fatalf("esperado status 401, obtenido %d", w.Code)
		}
	})

	t.Run("auth-service caido falla cerrado con 503", func(t *testing.T) {
		r := setupFeedRouter(repo, stubSessionValidator{err: client.ErrAuthServiceUnavailable})
		req, _ := http.NewRequest(http.MethodGet, "/v1/feed", nil)
		req.Header.Set("Authorization", "Bearer valid-token")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusServiceUnavailable {
			t.Fatalf("esperado status 503, obtenido %d", w.Code)
		}
	})
}

func TestGetMainFeed_UsesAuthenticatedIdentityAndIgnoresSpoofedHeader(t *testing.T) {
	repo := &mockFeedRepository{
		mockData: []models.FeedPost{},
	}
	r := setupFeedRouter(repo, stubSessionValidator{userID: "authenticated-user-99"})

	req, _ := http.NewRequest(http.MethodGet, "/v1/feed", nil)
	req.Header.Set("Authorization", "Bearer valid-session-token")
	req.Header.Set("X-User-ID", "spoofed-user-1")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("esperado status 200, obtenido %d: %s", w.Code, w.Body.String())
	}
	if repo.lastUserID != "authenticated-user-99" {
		t.Fatalf("esperado userID autenticado 'authenticated-user-99', obtenido '%s'", repo.lastUserID)
	}
}

func TestGetMainFeed_InternalErrorDoesNotLeakDetails(t *testing.T) {
	repo := &mockFeedRepository{
		mockErr: errors.New("pq: syntax error at or near SELECT * FROM pg_shadow"),
	}
	r := setupFeedRouter(repo, stubSessionValidator{userID: "user-1"})

	req, _ := http.NewRequest(http.MethodGet, "/v1/feed", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("esperado status 500, obtenido %d", w.Code)
	}

	body := w.Body.String()
	if strings.Contains(body, "detalle") || strings.Contains(body, "pg_shadow") || strings.Contains(body, "pq:") {
		t.Fatalf("la respuesta 500 filtró detalles internos al cliente: %s", body)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("error parseando JSON: %v", err)
	}
	if resp["error"] != "no se pudo cargar el feed" {
		t.Fatalf("mensaje de error inesperado: %v", resp["error"])
	}
}

func TestGetMainFeed_InvalidParameters(t *testing.T) {
	repo := &mockFeedRepository{}
	r := setupFeedRouter(repo, stubSessionValidator{userID: "1"})

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
			req.Header.Set("Authorization", "Bearer valid-token")
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
	r := setupFeedRouter(repo, stubSessionValidator{userID: "42"})

	req, _ := http.NewRequest(http.MethodGet, "/v1/feed?algorithm=recent&page=3&limit=15", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
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
	r := setupFeedRouter(repo, stubSessionValidator{userID: "user-1"})

	// 1. Verificar algoritmo 'relevant' (por defecto) y entrega de relevance_score en JSON
	reqRel, _ := http.NewRequest(http.MethodGet, "/v1/feed?page=1&limit=10", nil)
	reqRel.Header.Set("Authorization", "Bearer valid-token")
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
	reqRec.Header.Set("Authorization", "Bearer valid-token")
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
