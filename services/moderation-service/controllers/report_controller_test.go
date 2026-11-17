package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/meloop/moderation-service/models"
	"github.com/meloop/moderation-service/services"
)

type reportServiceStub struct {
	result models.CreatedReport
	err    error
	userID string
}

func (s *reportServiceStub) Create(_ context.Context, userID string, _ models.CreateReportRequest) (models.CreatedReport, error) {
	s.userID = userID
	return s.result, s.err
}

type identityResolverStub struct {
	userID string
	err    error
}

func (s identityResolverStub) Resolve(_ context.Context, authorization string) (string, error) {
	if authorization == "" {
		return "", services.ErrUnauthenticated
	}
	return s.userID, s.err
}

func TestCreateReportEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name          string
		userID        string
		authorization string
		spoofedUserID string
		body          string
		serviceError  error
		identityError error
		serviceResult models.CreatedReport
		wantStatus    int
		wantErrorCode string
	}{
		{
			name:          "created response schema",
			userID:        "user-1",
			authorization: "Bearer valid",
			body:          `{"target_type":"post","target_id":"post-1","reason":"spam","details":""}`,
			serviceResult: models.CreatedReport{ID: "report-1", Status: "pendiente", DueAt: "2026-11-18T00:00:00Z"},
			wantStatus:    http.StatusCreated,
		},
		{
			name:          "ignores untrusted user id header",
			userID:        "authenticated-user",
			authorization: "Bearer valid",
			spoofedUserID: "spoofed-user",
			body:          `{"target_type":"post","target_id":"post-1","reason":"spam","details":""}`,
			serviceResult: models.CreatedReport{ID: "report-1", Status: "pendiente", DueAt: "2026-11-18T00:00:00Z"},
			wantStatus:    http.StatusCreated,
		},
		{
			name:       "requires authenticated user",
			body:       `{"target_type":"post","target_id":"post-1","reason":"spam"}`,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:          "auth service unavailable",
			userID:        "user-1",
			authorization: "Bearer valid",
			identityError: errors.New("auth service unavailable"),
			wantStatus:    http.StatusServiceUnavailable,
			wantErrorCode: "AUTH_UNAVAILABLE",
		},
		{
			name:          "self report",
			userID:        "user-1",
			authorization: "Bearer valid",
			body:          `{"target_type":"post","target_id":"post-1","reason":"spam"}`,
			serviceError:  services.ErrCannotReportSelf,
			wantStatus:    http.StatusUnprocessableEntity,
			wantErrorCode: "CANNOT_REPORT_SELF",
		},
		{
			name:          "target not found or visible",
			userID:        "user-1",
			authorization: "Bearer valid",
			body:          `{"target_type":"comment","target_id":"comment-1","reason":"spam"}`,
			serviceError:  services.ErrTargetNotFound,
			wantStatus:    http.StatusNotFound,
			wantErrorCode: "TARGET_NOT_FOUND",
		},
		{
			name:          "duplicate target",
			userID:        "user-1",
			authorization: "Bearer valid",
			body:          `{"target_type":"user","target_id":"user-2","reason":"acoso"}`,
			serviceError:  services.ErrAlreadyReported,
			wantStatus:    http.StatusConflict,
			wantErrorCode: "ALREADY_REPORTED",
		},
		{
			name:          "rate limited",
			userID:        "user-1",
			authorization: "Bearer valid",
			body:          `{"target_type":"user","target_id":"user-2","reason":"acoso"}`,
			serviceError:  services.ErrRateLimit,
			wantStatus:    http.StatusTooManyRequests,
			wantErrorCode: "RATE_LIMIT_EXCEEDED",
		},
		{
			name:          "invalid body",
			userID:        "user-1",
			authorization: "Bearer valid",
			body:          `{"target_type":"post","target_id":"","reason":"otro"}`,
			serviceError:  services.ErrInvalidReport,
			wantStatus:    http.StatusUnprocessableEntity,
			wantErrorCode: "INVALID_REPORT",
		},
		{
			name:          "dependency unavailable",
			userID:        "user-1",
			authorization: "Bearer valid",
			body:          `{"target_type":"post","target_id":"post-1","reason":"spam"}`,
			serviceError:  errors.New("database unavailable"),
			wantStatus:    http.StatusServiceUnavailable,
			wantErrorCode: "MODERATION_UNAVAILABLE",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			router := gin.New()
			reportService := &reportServiceStub{result: test.serviceResult, err: test.serviceError}
			router.POST("/v1/moderation/reports", NewReportController(
				reportService, identityResolverStub{userID: test.userID, err: test.identityError},
			).Create)
			request := httptest.NewRequest(http.MethodPost, "/v1/moderation/reports", strings.NewReader(test.body))
			request.Header.Set("Content-Type", "application/json")
			if test.authorization != "" {
				request.Header.Set("Authorization", test.authorization)
			}
			if test.spoofedUserID != "" {
				request.Header.Set("X-User-ID", test.spoofedUserID)
			}
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)

			if response.Code != test.wantStatus {
				t.Fatalf("status = %d, want %d; body: %s", response.Code, test.wantStatus, response.Body.String())
			}
			if test.wantStatus == http.StatusCreated {
				var body map[string]any
				if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
					t.Fatalf("invalid JSON response: %v", err)
				}
				if len(body) != 3 || body["id"] != "report-1" || body["status"] != "pendiente" || body["due_at"] == nil {
					t.Fatalf("unexpected response schema: %#v", body)
				}
				if reportService.userID != test.userID {
					t.Fatalf("reported user = %q, want %q", reportService.userID, test.userID)
				}
				return
			}
			if test.wantErrorCode != "" {
				var body struct {
					Error struct {
						Code string `json:"code"`
					} `json:"error"`
				}
				if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
					t.Fatalf("invalid error response: %v", err)
				}
				if body.Error.Code != test.wantErrorCode {
					t.Fatalf("error code = %q, want %q", body.Error.Code, test.wantErrorCode)
				}
			}
		})
	}
}
