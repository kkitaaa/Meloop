package controllers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestFeedReturnsPaginatedCategoryItems(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/feed", Feed)

	response := httptest.NewRecorder()
	router.ServeHTTP(
		response,
		httptest.NewRequest(http.MethodGet, "/api/feed?category=friends&page=2&pageSize=4", nil),
	)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}

	var body struct {
		Success bool `json:"success"`
		Data    struct {
			Items    []feedPost `json:"items"`
			Page     int        `json:"page"`
			PageSize int        `json:"pageSize"`
			HasMore  bool       `json:"hasMore"`
		} `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !body.Success || body.Data.Page != 2 || body.Data.PageSize != 4 ||
		len(body.Data.Items) != 4 || !body.Data.HasMore {
		t.Fatalf("unexpected response: %s", response.Body.String())
	}
	if body.Data.Items[0].ID != 5 || body.Data.Items[0].Content[:len("Publicación de tus amigos")] != "Publicación de tus amigos" {
		t.Fatalf("unexpected category data: %+v", body.Data.Items[0])
	}
}

func TestFeedReturnsLastPageAndEmptyPage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/feed", Feed)

	lastPage := httptest.NewRecorder()
	router.ServeHTTP(
		lastPage,
		httptest.NewRequest(http.MethodGet, "/api/feed?page=3&pageSize=5", nil),
	)
	if lastPage.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", lastPage.Code, http.StatusOK)
	}

	var lastBody struct {
		Data feedPage `json:"data"`
	}
	if err := json.Unmarshal(lastPage.Body.Bytes(), &lastBody); err != nil {
		t.Fatalf("decode last page: %v", err)
	}
	if len(lastBody.Data.Items) != 5 || lastBody.Data.HasMore {
		t.Fatalf("unexpected last page: %+v", lastBody.Data)
	}

	emptyPage := httptest.NewRecorder()
	router.ServeHTTP(
		emptyPage,
		httptest.NewRequest(http.MethodGet, "/api/feed?page=4&pageSize=5", nil),
	)
	var emptyBody struct {
		Data feedPage `json:"data"`
	}
	if err := json.Unmarshal(emptyPage.Body.Bytes(), &emptyBody); err != nil {
		t.Fatalf("decode empty page: %v", err)
	}
	if len(emptyBody.Data.Items) != 0 || emptyBody.Data.HasMore {
		t.Fatalf("unexpected empty page: %+v", emptyBody.Data)
	}
}

func TestFeedRejectsInvalidQueries(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/feed", Feed)

	for _, path := range []string{
		"/api/feed?category=unknown",
		"/api/feed?page=0",
		"/api/feed?page=invalid",
		"/api/feed?pageSize=21",
	} {
		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
		if response.Code != http.StatusBadRequest {
			t.Errorf("%s status = %d, want %d", path, response.Code, http.StatusBadRequest)
		}
	}
}
