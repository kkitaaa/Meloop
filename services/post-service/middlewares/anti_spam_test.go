package middlewares

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

// memorySlidingLimiter replica exactamente la lógica del script Lua de Redis para tests deterministas
type memorySlidingLimiter struct {
	mu      sync.Mutex
	events  map[string][]time.Time
	blocked map[string]time.Time
	failErr error
}

func newMemorySlidingLimiter() *memorySlidingLimiter {
	return &memorySlidingLimiter{
		events:  make(map[string][]time.Time),
		blocked: make(map[string]time.Time),
	}
}

func (m *memorySlidingLimiter) AllowInteraction(ctx context.Context, userID string, now time.Time) (bool, time.Duration, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.failErr != nil {
		return false, 0, m.failErr
	}

	if until, ok := m.blocked[userID]; ok && until.After(now) {
		return false, until.Sub(now), nil
	}
	delete(m.blocked, userID)

	cutoff := now.Add(-WindowDuration)
	var active []time.Time
	for _, ts := range m.events[userID] {
		if ts.After(cutoff) {
			active = append(active, ts)
		}
	}

	if len(active) >= MaxInteractions {
		m.events[userID] = active
		m.blocked[userID] = now.Add(BlockDuration)
		return false, BlockDuration, nil
	}

	m.events[userID] = append(active, now)
	return true, 0, nil
}

func setupAntiSpamTestRouter(limiter RateLimiter, nowFn func() time.Time) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		if uid := c.GetHeader("X-User-ID"); uid != "" {
			c.Set("userID", uid)
		}
		c.Next()
	})
	r.Use(AntiSpamMiddlewareWithLimiter(limiter, nowFn))

	okHandler := func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) }
	r.GET("/v1/feed", okHandler)
	r.POST("/api/v1/posts", okHandler)
	r.PUT("/api/v1/posts/:id", okHandler)
	r.DELETE("/api/v1/posts/:id", okHandler)
	r.POST("/api/v1/posts/media", okHandler)
	r.POST("/v1/posts/:id/like", okHandler)
	r.DELETE("/v1/posts/:id/like", okHandler)
	r.POST("/api/v1/posts/:id/comments", okHandler)
	r.POST("/api/v1/posts/:id/comments/:commentId/replies", okHandler)
	return r
}

func TestAntiSpam_Allows50Blocks51AndEnforces15MinBlock(t *testing.T) {
	limiter := newMemorySlidingLimiter()
	currentTime := time.Date(2026, 11, 15, 10, 0, 0, 0, time.UTC)
	r := setupAntiSpamTestRouter(limiter, func() time.Time { return currentTime })

	// 1. Las primeras 50 interacciones deben pasar (200)
	for i := 1; i <= 50; i++ {
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/posts/post-1/comments", nil)
		req.Header.Set("X-User-ID", "user-spam")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("interacción %d falló: esperado 200, obtenido %d", i, w.Code)
		}
		currentTime = currentTime.Add(10 * time.Second)
	}

	// 2. La interacción 51 debe bloquearse con 429 y retry_after_ms = 900000 (15 min)
	req51, _ := http.NewRequest(http.MethodPost, "/v1/posts/post-1/like", nil)
	req51.Header.Set("X-User-ID", "user-spam")
	w51 := httptest.NewRecorder()
	r.ServeHTTP(w51, req51)

	if w51.Code != http.StatusTooManyRequests {
		t.Fatalf("interacción 51: esperado 429, obtenido %d", w51.Code)
	}
	var resp51 map[string]any
	_ = json.Unmarshal(w51.Body.Bytes(), &resp51)
	if int64(resp51["retry_after_ms"].(float64)) != BlockDuration.Milliseconds() {
		t.Fatalf("esperado retry_after_ms=%d, obtenido %v", BlockDuration.Milliseconds(), resp51["retry_after_ms"])
	}

	// 3. A los 10 minutos sigue bloqueado con 5 minutos restantes
	currentTime = currentTime.Add(10 * time.Minute)
	reqBlocked, _ := http.NewRequest(http.MethodPost, "/api/v1/posts/post-1/comments", nil)
	reqBlocked.Header.Set("X-User-ID", "user-spam")
	wBlocked := httptest.NewRecorder()
	r.ServeHTTP(wBlocked, reqBlocked)
	if wBlocked.Code != http.StatusTooManyRequests {
		t.Fatalf("durante el bloqueo de 15m: esperado 429, obtenido %d", wBlocked.Code)
	}
}

func TestAntiSpam_SlidingWindowBehavior(t *testing.T) {
	limiter := newMemorySlidingLimiter()
	currentTime := time.Date(2026, 11, 15, 10, 0, 0, 0, time.UTC)
	r := setupAntiSpamTestRouter(limiter, func() time.Time { return currentTime })

	// 30 interacciones en t = 0
	for i := 0; i < 30; i++ {
		req, _ := http.NewRequest(http.MethodPost, "/v1/posts/post-1/like", nil)
		req.Header.Set("X-User-ID", "user-sliding")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("esperado 200, obtenido %d", w.Code)
		}
	}

	// 20 interacciones en t = +40 min (total = 50 en la última hora)
	currentTime = currentTime.Add(40 * time.Minute)
	for i := 0; i < 20; i++ {
		req, _ := http.NewRequest(http.MethodPost, "/v1/posts/post-1/like", nil)
		req.Header.Set("X-User-ID", "user-sliding")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("esperado 200, obtenido %d", w.Code)
		}
	}

	// Avanzamos a t = +61 min: las 30 primeras ya salieron de la ventana móvil de 60 min,
	// solo quedan 20 activas -> deben permitirse 30 nuevas interacciones sin bloqueo
	currentTime = currentTime.Add(21 * time.Minute)
	for i := 0; i < 30; i++ {
		req, _ := http.NewRequest(http.MethodPost, "/v1/posts/post-1/like", nil)
		req.Header.Set("X-User-ID", "user-sliding")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("interacción %d tras deslizar ventana falló con %d", i+1, w.Code)
		}
	}
}

func TestAntiSpam_IgnoresNonInteractionsAndHandlesRedisErrors(t *testing.T) {
	limiter := newMemorySlidingLimiter()
	r := setupAntiSpamTestRouter(limiter, time.Now)

	// Rutas que NO son interacciones no deben consumir cuota
	nonInteractions := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/v1/feed"},
		{http.MethodPost, "/api/v1/posts"},
		{http.MethodPut, "/api/v1/posts/123"},
		{http.MethodDelete, "/api/v1/posts/123"},
		{http.MethodPost, "/api/v1/posts/media"},
	}

	for i := 0; i < 60; i++ {
		item := nonInteractions[i%len(nonInteractions)]
		req, _ := http.NewRequest(item.method, item.path, nil)
		req.Header.Set("X-User-ID", "user-crud")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("ruta no-interacción %s %s no debió bloquearse: %d", item.method, item.path, w.Code)
		}
	}

	// Si Redis falla en una interacción, debe responder 500
	limiter.failErr = errors.New("connection refused")
	reqFail, _ := http.NewRequest(http.MethodPost, "/v1/posts/123/like", nil)
	reqFail.Header.Set("X-User-ID", "user-crud")
	wFail := httptest.NewRecorder()
	r.ServeHTTP(wFail, reqFail)
	if wFail.Code != http.StatusInternalServerError {
		t.Fatalf("esperado 500 ante fallo de Redis, obtenido %d", wFail.Code)
	}
}
