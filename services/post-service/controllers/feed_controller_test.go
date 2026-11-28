package controllers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestFeedControllerRequiresAuthenticatedContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	controller := NewFeedController(nil, 0)
	router := gin.New()
	router.GET("/v1/feed", controller.GetMainFeed)
	request := httptest.NewRequest(http.MethodGet, "/v1/feed", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusUnauthorized)
	}
}
