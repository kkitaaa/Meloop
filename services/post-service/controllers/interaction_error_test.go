package controllers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/meloop/post-service/repositories"
)

func TestMissingPostInteractionReturnsNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodPost, "/", nil)

	writeInteractionError(context, errors.Join(errors.New("lookup failed"), repositories.ErrPostNotFound))

	var response struct {
		Success bool `json:"success"`
		Error   struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode interaction error response: %v", err)
	}
	if recorder.Code != http.StatusNotFound || response.Success || response.Error.Code != "NOT_FOUND" {
		t.Fatalf("response status=%d body=%s, want 404 with NOT_FOUND", recorder.Code, recorder.Body.String())
	}
}
