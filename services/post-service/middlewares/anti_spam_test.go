package middlewares

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

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
	validator := &stubTokenValidator{
		tokens: map[string]string{
			"tok-spam":    "user-spam",
			"tok-sliding": "user-sliding",
			"tok-crud":    "user-crud",
		},
	}

	r := gin.New()
	r.Use(OptionalAuthMiddleware(validator))
	r.Use(AntiSpamMiddlewareWithLimiter(limiter, nowFn))

	okHandler := func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) }
	r.GET("/v1/feed", okHandler)
	r.POST("/api/v1/posts", okHandler)
	r.PUT("/api/v1/posts/:id", okHandler)
	r.DELETE("/api/v1/posts/:id", okHandler)
	r.POST("/api/v1/posts/media", okHandler)
	r.POST("/v1/posts/:id/like", okHandler)
	r.DELETE("/v1/posts/:id/like", okHandler)
	r.POST("/posts/:postId/likes", okHandler)
	r.DELETE("/posts/:postId/likes", okHandler)
	r.POST("/api/v1/posts/:id/comments", okHandler)
	r.POST("/api/v1/posts/:id/comments/:commentId/replies", okHandler)
	r.POST("/api/v1/posts/:id/save", okHandler)
	r.DELETE("/api/v1/posts/:id/save", okHandler)
	r.POST("/api/v1/posts/:id/share", okHandler)
	return r
}

func TestAntiSpam_RejectsUnauthenticatedInteractionsWith401(t *testing.T) {
	limiter := newMemorySlidingLimiter()
	r := setupAntiSpamTestRouter(limiter, time.Now)

	endpoints := []struct {
		method string
		path   string
	}{
		{http.MethodPost, "/v1/posts/post-1/like"},
		{http.MethodDelete, "/v1/posts/post-1/like"},
		{http.MethodPost, "/posts/post-1/likes"},
		{http.MethodDelete, "/posts/post-1/likes"},
		{http.MethodPost, "/api/v1/posts/post-1/comments"},
		{http.MethodPost, "/api/v1/posts/post-1/comments/c-1/replies"},
		{http.MethodPost, "/api/v1/posts/post-1/save"},
		{http.MethodDelete, "/api/v1/posts/post-1/save"},
		{http.MethodPost, "/api/v1/posts/post-1/share"},
	}

	for _, ep := range endpoints {
		req, _ := http.NewRequest(ep.method, ep.path, nil)
		// Intentar suplantación mediante X-User-ID sin token válido
		req.Header.Set("X-User-ID", "spoofed-user")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("%s %s sin autenticación: esperado 401, obtenido %d", ep.method, ep.path, w.Code)
		}
	}
}

func TestAntiSpam_Allows50Blocks51AndEnforces15MinBlock(t *testing.T) {
	limiter := newMemorySlidingLimiter()
	currentTime := time.Date(2026, 11, 15, 10, 0, 0, 0, time.UTC)
	r := setupAntiSpamTestRouter(limiter, func() time.Time { return currentTime })

	for i := 1; i <= 50; i++ {
		path := "/api/v1/posts/post-1/comments"
		if i%3 == 0 {
			path = "/api/v1/posts/post-1/share"
		} else if i%2 == 0 {
			path = "/api/v1/posts/post-1/save"
		}
		req, _ := http.NewRequest(http.MethodPost, path, nil)
		req.Header.Set("Authorization", "Bearer tok-spam")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("interacción %d (%s) falló: esperado 200, obtenido %d", i, path, w.Code)
		}
		currentTime = currentTime.Add(10 * time.Second)
	}

	req51, _ := http.NewRequest(http.MethodPost, "/api/v1/posts/post-1/share", nil)
	req51.Header.Set("Authorization", "Bearer tok-spam")
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
}

func TestAntiSpam_ConcurrentBurstEnforcesExact50Limit(t *testing.T) {
	limiter := newMemorySlidingLimiter()
	fixedTime := time.Date(2026, 11, 15, 10, 0, 0, 0, time.UTC)
	r := setupAntiSpamTestRouter(limiter, func() time.Time { return fixedTime })

	const totalRequests = 80
	var allowedCount int32
	var blockedCount int32
	var wg sync.WaitGroup

	for i := 0; i < totalRequests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req, _ := http.NewRequest(http.MethodPost, "/v1/posts/post-1/like", nil)
			req.Header.Set("Authorization", "Bearer tok-spam")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			switch w.Code {
			case http.StatusOK:
				atomic.AddInt32(&allowedCount, 1)
			case http.StatusTooManyRequests:
				atomic.AddInt32(&blockedCount, 1)
			}
		}()
	}
	wg.Wait()

	if allowedCount != MaxInteractions || blockedCount != totalRequests-MaxInteractions {
		t.Fatalf("bajo concurrencia se esperaban %d permitidas y %d bloqueadas; obtenidas %d y %d",
			MaxInteractions, totalRequests-MaxInteractions, allowedCount, blockedCount)
	}
}

func TestAntiSpam_SlidingWindowBehavior(t *testing.T) {
	limiter := newMemorySlidingLimiter()
	currentTime := time.Date(2026, 11, 15, 10, 0, 0, 0, time.UTC)
	r := setupAntiSpamTestRouter(limiter, func() time.Time { return currentTime })

	for i := 0; i < 30; i++ {
		req, _ := http.NewRequest(http.MethodPost, "/v1/posts/post-1/like", nil)
		req.Header.Set("Authorization", "Bearer tok-sliding")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("esperado 200, obtenido %d", w.Code)
		}
	}

	currentTime = currentTime.Add(40 * time.Minute)
	for i := 0; i < 20; i++ {
		req, _ := http.NewRequest(http.MethodPost, "/v1/posts/post-1/like", nil)
		req.Header.Set("Authorization", "Bearer tok-sliding")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("esperado 200, obtenido %d", w.Code)
		}
	}

	currentTime = currentTime.Add(21 * time.Minute)
	for i := 0; i < 30; i++ {
		req, _ := http.NewRequest(http.MethodPost, "/v1/posts/post-1/like", nil)
		req.Header.Set("Authorization", "Bearer tok-sliding")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("interacción %d tras deslizar ventana falló con %d", i+1, w.Code)
		}
	}
}

func TestAntiSpam_IgnoresNonInteractionsAndDegradesGracefullyOnRedisError(t *testing.T) {
	primary := newMemorySlidingLimiter()
	primary.failErr = errors.New("connection refused")
	resilient := NewResilientSlidingWindowLimiter(primary, NewMemorySlidingWindowLimiter())
	r := setupAntiSpamTestRouter(resilient, time.Now)

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
		req.Header.Set("Authorization", "Bearer tok-crud")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("ruta no-interacción %s %s no debió bloquearse: %d", item.method, item.path, w.Code)
		}
	}

	// Con Redis caído, el fallback en memoria mantiene operativo el límite de 50 interacciones sin devolver 500
	for i := 1; i <= MaxInteractions; i++ {
		req, _ := http.NewRequest(http.MethodPost, "/v1/posts/123/like", nil)
		req.Header.Set("Authorization", "Bearer tok-crud")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("interacción %d en modo fallback: esperado 200, obtenido %d", i, w.Code)
		}
	}

	reqBlocked, _ := http.NewRequest(http.MethodPost, "/v1/posts/123/like", nil)
	reqBlocked.Header.Set("Authorization", "Bearer tok-crud")
	wBlocked := httptest.NewRecorder()
	r.ServeHTTP(wBlocked, reqBlocked)
	if wBlocked.Code != http.StatusTooManyRequests {
		t.Fatalf("esperado 429 en modo fallback al superar cuota, obtenido %d", wBlocked.Code)
	}
}
