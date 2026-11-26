package controllers_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/meloop/notification-service/client"
	"github.com/meloop/notification-service/controllers"
	"github.com/meloop/notification-service/models"
	"github.com/meloop/notification-service/repositories"
	"github.com/meloop/notification-service/routes"
	"github.com/meloop/notification-service/services"
)

type fakeInboxRepository struct {
	userID         string
	notificationID string
	limit          int
	offset         int
	markAllCount   int64
	listErr        error
	markReadErr    error
}

func (r *fakeInboxRepository) ListByUserID(_ context.Context, userID string, limit, offset int) ([]models.NotificationResponse, error) {
	r.userID, r.limit, r.offset = userID, limit, offset
	if r.listErr != nil {
		return nil, r.listErr
	}
	return []models.NotificationResponse{{ID: "notification-1", Recipient: userID}}, nil
}

func (r *fakeInboxRepository) MarkRead(_ context.Context, userID, notificationID string) error {
	r.userID, r.notificationID = userID, notificationID
	return r.markReadErr
}

func (r *fakeInboxRepository) MarkAllRead(_ context.Context, userID string) (int64, error) {
	r.userID = userID
	return r.markAllCount, nil
}

type fakeSessionValidator struct{}

func (fakeSessionValidator) ValidateToken(_ context.Context, token string) (string, error) {
	if token != "valid-session" {
		return "", client.ErrInvalidSession
	}
	return "authenticated-user", nil
}

func newTestRouter(repository *fakeInboxRepository) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	routes.SetupRoutes(router, controllers.NewNotificationController(services.NewInboxService(repository)), fakeSessionValidator{})
	return router
}

func TestListNotificationsUsesAuthenticatedUserAndPagination(t *testing.T) {
	repository := &fakeInboxRepository{}
	router := newTestRouter(repository)
	request := httptest.NewRequest(http.MethodGet, "/notifications?limit=5&offset=10", nil)
	request.Header.Set("Authorization", "Bearer valid-session")
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", response.Code, http.StatusOK, response.Body.String())
	}
	if repository.userID != "authenticated-user" || repository.limit != 5 || repository.offset != 10 {
		t.Fatalf("repository received user=%q limit=%d offset=%d", repository.userID, repository.limit, repository.offset)
	}
}

func TestMarkReadIsScopedToAuthenticatedUser(t *testing.T) {
	repository := &fakeInboxRepository{}
	router := newTestRouter(repository)
	request := httptest.NewRequest(http.MethodPatch, "/notifications/notification-1/read", nil)
	request.Header.Set("Authorization", "Bearer valid-session")
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", response.Code, http.StatusOK, response.Body.String())
	}
	if repository.userID != "authenticated-user" || repository.notificationID != "notification-1" {
		t.Fatalf("repository received user=%q notification=%q", repository.userID, repository.notificationID)
	}
}

func TestMarkAllReadUsesAuthenticatedUser(t *testing.T) {
	repository := &fakeInboxRepository{markAllCount: 3}
	router := newTestRouter(repository)
	request := httptest.NewRequest(http.MethodPatch, "/notifications/read", nil)
	request.Header.Set("Authorization", "Bearer valid-session")
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", response.Code, http.StatusOK, response.Body.String())
	}
	if repository.userID != "authenticated-user" {
		t.Fatalf("repository received user %q, want authenticated-user", repository.userID)
	}
}

func TestListNotificationsRejectsInvalidPaginationAndUnauthorizedRequests(t *testing.T) {
	for _, testCase := range []struct {
		name       string
		path       string
		token      string
		wantStatus int
	}{
		{name: "invalid limit", path: "/notifications?limit=101", token: "valid-session", wantStatus: http.StatusBadRequest},
		{name: "invalid session", path: "/notifications", token: "invalid-session", wantStatus: http.StatusUnauthorized},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			repository := &fakeInboxRepository{}
			router := newTestRouter(repository)
			request := httptest.NewRequest(http.MethodGet, testCase.path, nil)
			if testCase.token != "" {
				request.Header.Set("Authorization", "Bearer "+testCase.token)
			}
			response := httptest.NewRecorder()

			router.ServeHTTP(response, request)

			if response.Code != testCase.wantStatus {
				t.Fatalf("status = %d, want %d: %s", response.Code, testCase.wantStatus, response.Body.String())
			}
			if repository.userID != "" {
				t.Fatalf("repository unexpectedly called for user %q", repository.userID)
			}
		})
	}
}

func TestMarkReadReturnsNotFoundForOtherUsersNotifications(t *testing.T) {
	repository := &fakeInboxRepository{markReadErr: repositories.ErrNotificationNotFound}
	router := newTestRouter(repository)
	request := httptest.NewRequest(http.MethodPatch, "/notifications/private-id/read", nil)
	request.Header.Set("Authorization", "Bearer valid-session")
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNotFound)
	}
}
