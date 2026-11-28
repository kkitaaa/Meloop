package controllers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/meloop/music-service/controllers"
	"github.com/meloop/music-service/models"
	"github.com/meloop/music-service/routes"
	"github.com/meloop/music-service/services"
	"github.com/meloop/music-service/spotify"
)

type memoryCache struct {
	mu    sync.Mutex
	items map[string]string
}

func newMemoryCache() *memoryCache {
	return &memoryCache{items: make(map[string]string)}
}

func (m *memoryCache) Get(_ context.Context, key string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	v, ok := m.items[key]
	if !ok {
		return "", context.Canceled
	}
	return v, nil
}

func (m *memoryCache) Set(_ context.Context, key string, value string, _ time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.items[key] = value
	return nil
}

func setupSearchRouter(t *testing.T, searchHandler http.HandlerFunc) (*gin.Engine, *int32) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	var spotifySearchCalls int32

	tokenSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"mock-bearer-token","token_type":"Bearer","expires_in":3600}`))
	}))
	t.Cleanup(tokenSrv.Close)

	searchSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&spotifySearchCalls, 1)
		searchHandler(w, r)
	}))
	t.Cleanup(searchSrv.Close)

	tp, err := spotify.NewTokenProviderWithConfig("id", "secret", tokenSrv.URL, nil)
	if err != nil {
		t.Fatalf("error creando TokenProvider: %v", err)
	}

	sc := spotify.NewSearchClientWithConfig(tp, searchSrv.URL, nil)
	cache := newMemoryCache()
	svc := services.NewMusicSearchService(sc, cache, nil)
	ctrl := controllers.NewMusicController(svc, nil)

	r := gin.New()
	routes.RegisterMusicRoutes(r, ctrl, true)
	return r, &spotifySearchCalls
}

func TestSearchMusic_NormalizedAndCached(t *testing.T) {
	router, calls := setupSearchRouter(t, func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer mock-bearer-token" {
			t.Errorf("Authorization inesperado: %s", got)
		}
		if gotQ := r.URL.Query().Get("q"); gotQ != "Rosalía" {
			t.Errorf("q esperado 'Rosalía', obtenido %q", gotQ)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"tracks": {
				"total": 134,
				"items": [{
					"id": "4uLU6hMCjMI75M1A2tKUQC",
					"name": "DESPECHÁ",
					"artists": [{"name": "ROSALÍA"}],
					"album": {
						"name": "MOTOMAMI +",
						"images": null
					},
					"external_urls": {"spotify": "https://open.spotify.com/track/4uLU6hMCjMI75M1A2tKUQC"},
					"preview_url": null
				}]
			}
		}`))
	})

	// 1. Primera búsqueda: consulta Spotify y normaliza campos null sin romper
	req1 := httptest.NewRequest(http.MethodGet, "/v1/music/search?q=Rosal%C3%ADa&type=track&limit=10&offset=0", nil)
	rec1 := httptest.NewRecorder()
	router.ServeHTTP(rec1, req1)

	if rec1.Code != http.StatusOK {
		t.Fatalf("esperado 200 OK, recibido %d: %s", rec1.Code, rec1.Body.String())
	}

	var resp1 models.MusicSearchResponse
	if err := json.Unmarshal(rec1.Body.Bytes(), &resp1); err != nil {
		t.Fatalf("JSON inválido: %v", err)
	}
	if len(resp1.Results) != 1 || resp1.Results[0].Name != "DESPECHÁ" || resp1.Results[0].ImageURL != "" || resp1.Total != 134 {
		t.Fatalf("respuesta normalizada inesperada: %+v", resp1)
	}

	// 2. Segunda búsqueda idéntica: sale de caché sin volver a llamar a Spotify
	req2 := httptest.NewRequest(http.MethodGet, "/v1/music/search?q=Rosal%C3%ADa&type=track&limit=10&offset=0", nil)
	rec2 := httptest.NewRecorder()
	router.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusOK {
		t.Fatalf("esperado 200 OK desde caché, recibido %d", rec2.Code)
	}
	if got := atomic.LoadInt32(calls); got != 1 {
		t.Fatalf("la segunda búsqueda debió salir de caché (1 llamada a Spotify), hubo %d", got)
	}
}

func TestSearchMusic_GenreFilterAndEmptyResults(t *testing.T) {
	router, _ := setupSearchRouter(t, func(w http.ResponseWriter, r *http.Request) {
		if gotQ := r.URL.Query().Get("q"); gotQ != `genre:"indie rock"` {
			t.Errorf("filtro genre esperado 'genre:\"indie rock\"', recibido %q", gotQ)
		}
		if gotType := r.URL.Query().Get("type"); gotType != "track,artist" {
			t.Errorf("type esperado 'track,artist', recibido %q", gotType)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"tracks":{"total":0,"items":[]},"artists":{"total":0,"items":[]}}`))
	})

	req := httptest.NewRequest(http.MethodGet, "/v1/music/search?q=indie%20rock&type=genre", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("esperado 200 OK sin resultados, recibido %d", rec.Code)
	}

	var resp models.MusicSearchResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("error decodificando respuesta vacía: %v", err)
	}
	if resp.Results == nil || len(resp.Results) != 0 || resp.Limit != 10 {
		t.Fatalf("se esperaba lista vacía [] y limit=10, recibido: %+v", resp)
	}
}

func TestSearchMusic_ValidationsAndSpotifyErrors(t *testing.T) {
	longQ := strings.Repeat("a", 101)

	// Casos de validación -> 400 INVALID_PARAMETER
	routerOK, _ := setupSearchRouter(t, func(w http.ResponseWriter, r *http.Request) {})
	invalidURLs := []string{
		"/v1/music/search?q=&type=track",
		"/v1/music/search?q=" + longQ + "&type=track",
		"/v1/music/search?q=test&type=podcast",
		"/v1/music/search?q=test&type=track&limit=25",
		"/v1/music/search?q=test&type=track&limit=0",
		"/v1/music/search?q=test&type=track&offset=-1",
	}

	for _, u := range invalidURLs {
		rec := httptest.NewRecorder()
		routerOK.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, u, nil))
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("URL %s debió retornar 400, retornó %d", u, rec.Code)
		}
		var errResp models.ErrorResponse
		_ = json.Unmarshal(rec.Body.Bytes(), &errResp)
		if errResp.Error != "INVALID_PARAMETER" {
			t.Fatalf("URL %s debió retornar INVALID_PARAMETER, retornó %s", u, errResp.Error)
		}
	}

	// Caso Spotify 429 -> 503 MUSIC_RATE_LIMITED con Retry-After
	router429, _ := setupSearchRouter(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "30")
		w.WriteHeader(http.StatusTooManyRequests)
	})
	rec429 := httptest.NewRecorder()
	router429.ServeHTTP(rec429, httptest.NewRequest(http.MethodGet, "/v1/music/search?q=Daft%20Punk&type=artist", nil))

	if rec429.Code != http.StatusServiceUnavailable {
		t.Fatalf("esperado 503 ante 429 de Spotify, obtenido %d", rec429.Code)
	}
	if rec429.Header().Get("Retry-After") != "30" {
		t.Fatalf("esperado Retry-After=30, obtenido %q", rec429.Header().Get("Retry-After"))
	}
	var err429 models.ErrorResponse
	_ = json.Unmarshal(rec429.Body.Bytes(), &err429)
	if err429.Error != "MUSIC_RATE_LIMITED" {
		t.Fatalf("esperado MUSIC_RATE_LIMITED, obtenido %s", err429.Error)
	}

	// Caso Spotify 500 -> nunca debe devolver 500
	router500, _ := setupSearchRouter(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	rec500 := httptest.NewRecorder()
	router500.ServeHTTP(rec500, httptest.NewRequest(http.MethodGet, "/v1/music/search?q=Radiohead&type=album", nil))

	if rec500.Code == http.StatusInternalServerError || rec500.Code != http.StatusServiceUnavailable {
		t.Fatalf("errores de Spotify no deben generar 500 (esperado 503), obtenido %d", rec500.Code)
	}
}
