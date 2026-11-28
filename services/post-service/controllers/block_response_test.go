package controllers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestBlockedResponseUsesRequiredErrorCode(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodPost, "/", nil)

	writeBlockedError(context)

	var response struct {
		Success bool `json:"success"`
		Error   struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode blocked response: %v", err)
	}
	if recorder.Code != http.StatusForbidden || response.Success || response.Error.Code != "BLOCKED" {
		t.Fatalf("response status=%d body=%s, want 403 with error code BLOCKED", recorder.Code, recorder.Body.String())
	}
}
