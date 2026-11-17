package spotify

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestTokenProvider_InitialFetchAndReuse(t *testing.T) {
	var calls int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)

		if r.Method != http.MethodPost {
			t.Errorf("método esperado POST, recibido %s", r.Method)
		}
		expectedAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("test-id:test-secret"))
		if got := r.Header.Get("Authorization"); got != expectedAuth {
			t.Errorf("Authorization esperado %q, recibido %q", expectedAuth, got)
		}
		body, _ := io.ReadAll(r.Body)
		if string(body) != "grant_type=client_credentials" {
			t.Errorf("body inesperado: %s", string(body))
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"token-inicial-123","token_type":"Bearer","expires_in":3600}`))
	}))
	defer srv.Close()

	tp, err := NewTokenProviderWithConfig("test-id", "test-secret", srv.URL, nil)
	if err != nil {
		t.Fatalf("error inesperado creando TokenProvider: %v", err)
	}

	if tp.httpClient.Timeout != 5*time.Second {
		t.Fatalf("timeout esperado 5s, obtenido %v", tp.httpClient.Timeout)
	}

	ctx := context.Background()
	for i := 0; i < 5; i++ {
		tok, err := tp.Token(ctx)
		if err != nil {
			t.Fatalf("llamada %d falló: %v", i, err)
		}
		if tok != "token-inicial-123" {
			t.Fatalf("token esperado 'token-inicial-123', obtenido %q", tok)
		}
	}

	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Fatalf("se esperaba 1 única petición a Spotify y reutilización en memoria, hubo %d", got)
	}
}

func TestTokenProvider_AutoRefreshAt80PercentAndOn401(t *testing.T) {
	var calls int32
	currentTime := time.Date(2026, 11, 17, 16, 0, 0, 0, time.UTC)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&calls, 1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintf(w, `{"access_token":"token-v%d","token_type":"Bearer","expires_in":1000}`, n)
	}))
	defer srv.Close()

	tp, _ := NewTokenProviderWithConfig("id", "secret", srv.URL, nil)
	tp.nowFunc = func() time.Time { return currentTime }

	ctx := context.Background()

	// t = 0s -> obtiene token-v1 (expires_in=1000s, renueva a partir de 800s)
	tok1, err := tp.Token(ctx)
	if err != nil || tok1 != "token-v1" {
		t.Fatalf("esperado token-v1, obtenido %q (err=%v)", tok1, err)
	}

	// t = 799s (falta 20.1%) -> sigue reutilizando token-v1
	currentTime = currentTime.Add(799 * time.Second)
	tokStill1, _ := tp.Token(ctx)
	if tokStill1 != "token-v1" || atomic.LoadInt32(&calls) != 1 {
		t.Fatalf("no debió renovar antes del 80%% (calls=%d)", atomic.LoadInt32(&calls))
	}

	// t = 800s (falta exactamente 20%) -> renueva automáticamente a token-v2
	currentTime = currentTime.Add(1 * time.Second)
	tok2, err := tp.Token(ctx)
	if err != nil || tok2 != "token-v2" {
		t.Fatalf("esperado renovación automática a token-v2, obtenido %q (err=%v)", tok2, err)
	}

	// Simular respuesta 401 desde Spotify -> ForceRefresh obtiene token-v3 inmediatamente
	tok3, err := tp.ForceRefresh(ctx)
	if err != nil || tok3 != "token-v3" {
		t.Fatalf("esperado ForceRefresh a token-v3, obtenido %q (err=%v)", tok3, err)
	}
	if got := atomic.LoadInt32(&calls); got != 3 {
		t.Fatalf("esperadas 3 llamadas totales, hubo %d", got)
	}
}

func TestTokenProvider_ConcurrentRefreshDeduplication(t *testing.T) {
	var calls int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		time.Sleep(40 * time.Millisecond) // Simula latencia de red
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"token-concurrente","token_type":"Bearer","expires_in":3600}`))
	}))
	defer srv.Close()

	tp, _ := NewTokenProviderWithConfig("id", "secret", srv.URL, nil)

	const workers = 20
	var wg sync.WaitGroup
	wg.Add(workers)

	results := make([]string, workers)
	errs := make([]error, workers)

	for i := 0; i < workers; i++ {
		go func(idx int) {
			defer wg.Done()
			results[idx], errs[idx] = tp.Token(context.Background())
		}(i)
	}
	wg.Wait()

	for i := 0; i < workers; i++ {
		if errs[i] != nil {
			t.Fatalf("worker %d recibió error: %v", i, errs[i])
		}
		if results[i] != "token-concurrente" {
			t.Fatalf("worker %d recibió token inesperado: %q", i, results[i])
		}
	}

	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Fatalf("se esperaba 1 única llamada HTTP concurrente, pero hubo %d", got)
	}
}

func TestTokenProvider_InvalidCredentialsAndSpotifyDownFallback(t *testing.T) {
	// 1. Credenciales vacías
	if _, err := NewTokenProvider("", "secret"); !errors.Is(err, ErrMissingCredentials) {
		t.Fatalf("esperado ErrMissingCredentials, obtenido %v", err)
	}

	// 2. Credenciales rechazadas con 401 por Spotify
	srvAuthFail := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"invalid_client"}`))
	}))
	defer srvAuthFail.Close()

	tpFail, _ := NewTokenProviderWithConfig("bad-id", "bad-secret", srvAuthFail.URL, nil)
	if _, err := tpFail.Token(context.Background()); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("esperado ErrInvalidCredentials, obtenido %v", err)
	}

	// 3. Spotify caído (503) cuando toca renovar (85% TTL) pero el token aún no expira -> devuelve último token
	var failMode int32
	currentTime := time.Date(2026, 11, 17, 17, 0, 0, 0, time.UTC)

	srvFlaky := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.LoadInt32(&failMode) == 1 {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte("service unavailable"))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"token-respaldo","token_type":"Bearer","expires_in":1000}`))
	}))
	defer srvFlaky.Close()

	tpFlaky, _ := NewTokenProviderWithConfig("id", "secret", srvFlaky.URL, nil)
	tpFlaky.nowFunc = func() time.Time { return currentTime }

	tok, err := tpFlaky.Token(context.Background())
	if err != nil || tok != "token-respaldo" {
		t.Fatalf("falló obtención inicial: %v", err)
	}

	// Spotify cae y avanzamos al 85% (850s < 1000s): aún no expiró, debe devolver token-respaldo
	atomic.StoreInt32(&failMode, 1)
	currentTime = currentTime.Add(850 * time.Second)

	fallbackTok, err := tpFlaky.Token(context.Background())
	if err != nil || fallbackTok != "token-respaldo" {
		t.Fatalf("esperado reutilizar 'token-respaldo' mientras no expire, got=%q err=%v", fallbackTok, err)
	}

	// Avanzamos a 1001s (ya expiró por completo y Spotify sigue caído): ahora sí debe devolver error
	currentTime = currentTime.Add(151 * time.Second)
	if _, err := tpFlaky.Token(context.Background()); err == nil {
		t.Fatal("esperado error cuando el token expiró completamente y Spotify está caído")
	}
}
