package controllers

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestProxyToServiceForwardsRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/users" || r.URL.RawQuery != "source=flutter" {
			t.Errorf("request URL = %s, want /users?source=flutter", r.URL.RequestURI())
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read body: %v", err)
		}
		if string(body) != `{"name":"Ada"}` {
			t.Errorf("body = %q, want JSON payload", body)
		}
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"created":true}`))
	}))
	defer service.Close()

	const envVar = "API_GATEWAY_TEST_SERVICE_URL"
	t.Setenv(envVar, service.URL)
	router := gin.New()
	router.POST("/users", ProxyToService(envVar, "", ""))

	request := httptest.NewRequest(http.MethodPost, "/users?source=flutter", strings.NewReader(`{"name":"Ada"}`))
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusCreated)
	}
	if response.Body.String() != `{"created":true}` {
		t.Errorf("response = %q, want proxied response", response.Body.String())
	}
}

func TestProxyToServiceReturnsBadGatewayWhenUnavailable(t *testing.T) {
	gin.SetMode(gin.TestMode)
	const envVar = "API_GATEWAY_TEST_UNAVAILABLE_URL"
	t.Setenv(envVar, "http://127.0.0.1:1")
	router := gin.New()
	router.GET("/users", ProxyToService(envVar, "", ""))

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/users", nil))

	if response.Code != http.StatusBadGateway {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadGateway)
	}
}

func TestProxyToServiceUsesDefaultTarget(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer service.Close()

	const envVar = "API_GATEWAY_TEST_DEFAULT_URL"
	t.Setenv(envVar, "")
	router := gin.New()
	router.GET("/health", ProxyToService(envVar, service.URL, ""))

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/health", nil))

	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNoContent)
	}
}
