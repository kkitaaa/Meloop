package controllers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type adminRoleRepositoryFake struct {
	admin bool
	err   error
}

func (repository adminRoleRepositoryFake) IsAdmin(context.Context, string) (bool, error) {
	return repository.admin, repository.err
}

func TestAdminRequiredRequiresValidatedAdministrator(t *testing.T) {
	gin.SetMode(gin.TestMode)
	authServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		if request.Header.Get("Authorization") != "Bearer valid-token" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"data":{"id":"admin-1"}}`))
	}))
	defer authServer.Close()

	tests := []struct {
		name       string
		authority  string
		isAdmin    bool
		wantStatus int
	}{
		{name: "missing token", wantStatus: http.StatusUnauthorized},
		{name: "valid non-admin", authority: "Bearer valid-token", wantStatus: http.StatusForbidden},
		{name: "administrator", authority: "Bearer valid-token", isAdmin: true, wantStatus: http.StatusOK},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			router := gin.New()
			router.GET("/admin", AdminRequired(authServer.URL, adminRoleRepositoryFake{admin: test.isAdmin}), func(c *gin.Context) {
				if c.GetString(ContextUserIDKey) != "admin-1" {
					t.Errorf("user ID = %q, want admin-1", c.GetString(ContextUserIDKey))
				}
				c.Status(http.StatusOK)
			})
			request := httptest.NewRequest(http.MethodGet, "/admin", nil)
			if test.authority != "" {
				request.Header.Set("Authorization", test.authority)
			}
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			if response.Code != test.wantStatus {
				t.Fatalf("status = %d, want %d", response.Code, test.wantStatus)
			}
		})
	}
}
