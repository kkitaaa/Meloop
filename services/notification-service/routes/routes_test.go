package routes

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/meloop/notification-service/models"
)

type fakeStore struct {
	userID           string
	limit            int
	offset           int
	markedID         string
	markedAll        string
	preferenceUserID string
	preferenceType   string
	preferenceValue  bool
}

func (store *fakeStore) List(_ context.Context, userID string, limit, offset int) ([]models.NotificationRecord, error) {
	store.userID, store.limit, store.offset = userID, limit, offset
	return []models.NotificationRecord{{ID: "notification-1", Type: models.NotificationTypePostLiked}}, nil
}

func (store *fakeStore) MarkRead(_ context.Context, userID, notificationID string) (bool, error) {
	store.userID, store.markedID = userID, notificationID
	return true, nil
}

func (store *fakeStore) MarkAllRead(_ context.Context, userID string) (int64, error) {
	store.markedAll = userID
	return 3, nil
}

func (store *fakeStore) ListPreferences(context.Context, string) ([]models.NotificationPreference, error) {
	return []models.NotificationPreference{{Type: models.NotificationTypePostLiked, Enabled: true}}, nil
}

func (store *fakeStore) SetPreference(_ context.Context, userID, notificationType string, enabled bool) error {
	store.preferenceUserID, store.preferenceType, store.preferenceValue = userID, notificationType, enabled
	return nil
}

func TestNotificationRoutesUseAuthenticatedUser(t *testing.T) {
	store := &fakeStore{}
	authServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/auth/validate" || r.Header.Get("Authorization") != "Bearer valid-token" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"data":{"id":"user-42"}}`))
	}))
	defer authServer.Close()

	handler := Setup(store, authServer.URL)
	request := httptest.NewRequest(http.MethodGet, "/notifications?limit=10&offset=20", nil)
	request.Header.Set("Authorization", "Bearer valid-token")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("GET status = %d, want %d: %s", response.Code, http.StatusOK, response.Body.String())
	}
	if store.userID != "user-42" || store.limit != 10 || store.offset != 20 {
		t.Fatalf("List called with user=%q limit=%d offset=%d", store.userID, store.limit, store.offset)
	}
}

func TestNotificationRoutesReadAndPreferences(t *testing.T) {
	store := &fakeStore{}
	authServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"data":{"id":"user-7"}}`))
	}))
	defer authServer.Close()
	handler := Setup(store, authServer.URL)

	tests := []struct {
		method string
		path   string
		body   string
	}{
		{http.MethodPatch, "/notifications/notification-9/read", ""},
		{http.MethodPatch, "/notifications/read", ""},
		{http.MethodPut, "/notifications/preferences/post.liked", `{"enabled":false}`},
	}
	for _, test := range tests {
		request := httptest.NewRequest(test.method, test.path, strings.NewReader(test.body))
		request.Header.Set("Authorization", "Bearer token")
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if response.Code != http.StatusOK {
			t.Fatalf("%s %s status = %d, want %d: %s", test.method, test.path, response.Code, http.StatusOK, response.Body.String())
		}
	}
	if store.userID != "user-7" || store.markedID != "notification-9" || store.markedAll != "user-7" {
		t.Fatalf("read operations did not use the authenticated user: %+v", store)
	}
	if store.preferenceUserID != "user-7" || store.preferenceType != models.NotificationTypePostLiked || store.preferenceValue {
		t.Fatalf("preference update did not use request values: %+v", store)
	}
}

func TestNotificationRoutesRequireValidSession(t *testing.T) {
	store := &fakeStore{}
	authServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	}))
	defer authServer.Close()
	handler := Setup(store, authServer.URL)

	for _, authorization := range []string{"", "Bearer expired-token"} {
		request := httptest.NewRequest(http.MethodGet, "/notifications", nil)
		request.Header.Set("Authorization", authorization)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if response.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want %d", response.Code, http.StatusUnauthorized)
		}
	}
}

func TestPreferenceUpdateRejectsUnknownType(t *testing.T) {
	store := &fakeStore{}
	authServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"data":{"id":"user-1"}}`))
	}))
	defer authServer.Close()
	handler := Setup(store, authServer.URL)

	request := httptest.NewRequest(http.MethodPut, "/notifications/preferences/unknown.type", strings.NewReader(`{"enabled":false}`))
	request.Header.Set("Authorization", "Bearer valid-token")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
}

func TestNotificationListResponseIsJSON(t *testing.T) {
	store := &fakeStore{}
	authServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"data":{"id":"user-1"}}`))
	}))
	defer authServer.Close()
	handler := Setup(store, authServer.URL)
	request := httptest.NewRequest(http.MethodGet, "/notifications", nil)
	request.Header.Set("Authorization", "Bearer valid-token")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	var body map[string]json.RawMessage
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}
}
