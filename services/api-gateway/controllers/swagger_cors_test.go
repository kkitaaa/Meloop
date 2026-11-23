package controllers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestSwaggerUICORSAllowsLocalSwaggerUI(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(SwaggerUICORS())
	router.GET("/health", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	request := httptest.NewRequest(http.MethodGet, "/health", nil)
	request.Header.Set("Origin", swaggerUIOrigin)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if got := response.Header().Get("Access-Control-Allow-Origin"); got != swaggerUIOrigin {
		t.Fatalf("Access-Control-Allow-Origin = %q, want %q", got, swaggerUIOrigin)
	}
}

func TestSwaggerUICORSHandlesPreflight(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(SwaggerUICORS())

	request := httptest.NewRequest(http.MethodOptions, "/v1/users/me/profile", nil)
	request.Header.Set("Origin", swaggerUIOrigin)
	request.Header.Set("Access-Control-Request-Method", http.MethodPut)
	request.Header.Set("Access-Control-Request-Headers", "authorization,content-type")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNoContent)
	}
	if got := response.Header().Get("Access-Control-Allow-Methods"); got == "" {
		t.Fatal("Access-Control-Allow-Methods is empty")
	}
	if got := response.Header().Get("Access-Control-Allow-Headers"); got == "" {
		t.Fatal("Access-Control-Allow-Headers is empty")
	}
}

func TestSwaggerUICORSDoesNotAllowOtherOrigins(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(SwaggerUICORS())
	router.GET("/health", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	request := httptest.NewRequest(http.MethodGet, "/health", nil)
	request.Header.Set("Origin", "https://example.com")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if got := response.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("unexpected Access-Control-Allow-Origin %q", got)
	}
}
