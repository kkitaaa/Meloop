package controllers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/meloop/user-service/models"
	"github.com/meloop/user-service/services"
)

type trackingUserService struct {
	listCalls        int
	suspendCalls     int
	reactivateCalls  int
	moderationCalls  int
	lastModerationID string
	lastModerator    bool
}

func (s *trackingUserService) GetAccount(ctx context.Context, userID string) (*models.AccountResponse, error) {
	return &models.AccountResponse{IDUsuario: userID}, nil
}

func (s *trackingUserService) UpdateUsername(
	ctx context.Context,
	userID string,
	req *models.UpdateUsernameRequest,
) (*models.AccountResponse, error) {
	return &models.AccountResponse{IDUsuario: userID, Username: req.Username}, nil
}

func (s *trackingUserService) RequestEmailChange(
	ctx context.Context,
	userID string,
	req *models.RequestEmailChangeRequest,
) (*models.EmailChangeResponse, error) {
	return &models.EmailChangeResponse{}, nil
}

func (s *trackingUserService) ListUsers(ctx context.Context) ([]models.AdminUserSummary, error) {
	s.listCalls++
	return []models.AdminUserSummary{}, nil
}

func (s *trackingUserService) SuspendUser(
	ctx context.Context,
	userID string,
	reason string,
) (*models.AdminUserSummary, error) {
	s.suspendCalls++
	return &models.AdminUserSummary{IDUsuario: userID, Suspendido: true, Motivo: reason}, nil
}

func (s *trackingUserService) ReactivateUser(ctx context.Context, userID string) (*models.AdminUserSummary, error) {
	s.reactivateCalls++
	return &models.AdminUserSummary{IDUsuario: userID}, nil
}

func (s *trackingUserService) SetModeratorRole(
	ctx context.Context,
	userID string,
	moderator bool,
) (*models.AdminUserSummary, error) {
	s.moderationCalls++
	s.lastModerationID = userID
	s.lastModerator = moderator
	return &models.AdminUserSummary{IDUsuario: userID, EsModerador: moderator}, nil
}

func newAdminPermissionsRouter(
	validator testSessionValidator,
	role string,
	srv *trackingUserService,
) *gin.Engine {
	gin.SetMode(gin.TestMode)
	roleProvider := testRoleProvider{role: role}
	ctrl := NewAdminController(srv)

	router := gin.New()
	protected := router.Group("/users")
	protected.Use(AuthRequired(validator))
	adminRoutes := protected.Group("/admin")
	adminRoutes.GET("/users", AdminRequired(roleProvider), ctrl.ListUsers)
	adminRoutes.PATCH("/users/:id/suspend", AdminRequired(roleProvider), ctrl.SuspendUser)
	adminRoutes.PATCH("/users/:id/reactivate", AdminRequired(roleProvider), ctrl.ReactivateUser)
	adminRoutes.PATCH("/users/:id/moderation", AdminOnly(roleProvider), ctrl.UpdateModeratorRole)
	return router
}

func adminPermissionRequest(
	router *gin.Engine,
	method,
	path,
	body string,
	headers map[string]string,
) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if _, exists := headers["Authorization"]; !exists {
		req.Header.Set("Authorization", "Bearer valid-session")
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	return recorder
}

func TestModeratorCanListSuspendAndReactivateButCannotChangeRoles(t *testing.T) {
	srv := &trackingUserService{}
	router := newAdminPermissionsRouter(
		testSessionValidator{userID: "moderator-user"},
		"moderador",
		srv,
	)

	cases := []struct {
		method string
		path   string
		body   string
	}{
		{method: http.MethodGet, path: "/users/admin/users"},
		{method: http.MethodPatch, path: "/users/admin/users/u2/suspend", body: `{"reason":"spam"}`},
		{method: http.MethodPatch, path: "/users/admin/users/u2/reactivate", body: `{}`},
	}

	for _, tc := range cases {
		recorder := adminPermissionRequest(router, tc.method, tc.path, tc.body, nil)
		if recorder.Code != http.StatusOK {
			t.Errorf("%s %s: status = %d, want 200; body = %s", tc.method, tc.path, recorder.Code, recorder.Body.String())
		}
	}
	if srv.listCalls != 1 || srv.suspendCalls != 1 || srv.reactivateCalls != 1 {
		t.Errorf("llamadas = list:%d suspend:%d reactivate:%d", srv.listCalls, srv.suspendCalls, srv.reactivateCalls)
	}

	recorder := adminPermissionRequest(
		router,
		http.MethodPatch,
		"/users/admin/users/u2/moderation",
		`{"moderador":true}`,
		nil,
	)
	if recorder.Code != http.StatusForbidden {
		t.Fatalf("moderación: status = %d, want 403; body = %s", recorder.Code, recorder.Body.String())
	}
	if srv.moderationCalls != 0 {
		t.Fatal("un moderador no debe poder cambiar roles")
	}
}

func TestAdminCanPerformAllAdministrativeActions(t *testing.T) {
	srv := &trackingUserService{}
	router := newAdminPermissionsRouter(testSessionValidator{userID: "admin-user"}, "admin", srv)

	for _, tc := range []struct {
		method string
		path   string
		body   string
	}{
		{method: http.MethodGet, path: "/users/admin/users"},
		{method: http.MethodPatch, path: "/users/admin/users/u2/suspend", body: `{"reason":"spam"}`},
		{method: http.MethodPatch, path: "/users/admin/users/u2/reactivate", body: `{}`},
		{method: http.MethodPatch, path: "/users/admin/users/u2/moderation", body: `{"moderador":true}`},
	} {
		recorder := adminPermissionRequest(router, tc.method, tc.path, tc.body, nil)
		if recorder.Code != http.StatusOK {
			t.Errorf("%s %s: status = %d, want 200; body = %s", tc.method, tc.path, recorder.Code, recorder.Body.String())
		}
	}
	if srv.moderationCalls != 1 || srv.lastModerationID != "u2" || !srv.lastModerator {
		t.Errorf("moderación no propagada correctamente: calls=%d id=%q value=%t",
			srv.moderationCalls, srv.lastModerationID, srv.lastModerator)
	}
}

func TestAdminEndpointsReturn401WithoutValidSession(t *testing.T) {
	srv := &trackingUserService{}
	router := newAdminPermissionsRouter(
		testSessionValidator{err: services.ErrInvalidSession},
		"admin",
		srv,
	)

	for _, headers := range []map[string]string{
		{},
		{"Authorization": "Bearer invalid-session"},
	} {
		for _, tc := range []struct{ method, path string }{
			{method: http.MethodGet, path: "/users/admin/users"},
			{method: http.MethodPatch, path: "/users/admin/users/u2/suspend"},
			{method: http.MethodPatch, path: "/users/admin/users/u2/reactivate"},
			{method: http.MethodPatch, path: "/users/admin/users/u2/moderation"},
		} {
			recorder := adminPermissionRequest(router, tc.method, tc.path, `{}`, headers)
			if recorder.Code != http.StatusUnauthorized {
				t.Errorf("path %s headers %v: status = %d, want 401", tc.path, headers, recorder.Code)
			}
		}
	}

	if srv.listCalls+srv.suspendCalls+srv.reactivateCalls+srv.moderationCalls != 0 {
		t.Fatal("una sesión inválida no debe ejecutar operaciones administrativas")
	}
}

func TestAdminEndpointsReturn403ForRegularUsersDespiteSpoofedHeaders(t *testing.T) {
	srv := &trackingUserService{}
	router := newAdminPermissionsRouter(testSessionValidator{userID: "regular-user"}, "usuario", srv)
	headers := map[string]string{
		"X-User-ID":   "attacker",
		"X-User-Role": "admin",
		"X-Role":      "administrator",
	}

	for _, tc := range []struct{ method, path string }{
		{method: http.MethodGet, path: "/users/admin/users"},
		{method: http.MethodPatch, path: "/users/admin/users/u2/suspend"},
		{method: http.MethodPatch, path: "/users/admin/users/u2/reactivate"},
		{method: http.MethodPatch, path: "/users/admin/users/u2/moderation"},
	} {
		recorder := adminPermissionRequest(router, tc.method, tc.path, `{}`, headers)
		if recorder.Code != http.StatusForbidden {
			t.Errorf("path %s: status = %d, want 403", tc.path, recorder.Code)
		}
	}
}

func TestAdminCannotModifyOwnRole(t *testing.T) {
	srv := &trackingUserService{}
	router := newAdminPermissionsRouter(testSessionValidator{userID: "admin-user"}, "admin", srv)

	recorder := adminPermissionRequest(
		router,
		http.MethodPatch,
		"/users/admin/users/admin-user/moderation",
		`{"moderador":false}`,
		nil,
	)
	if recorder.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403; body = %s", recorder.Code, recorder.Body.String())
	}
	if srv.moderationCalls != 0 {
		t.Fatal("no se debe modificar el propio rol")
	}
}

func TestAdminCanAssignAndRemoveModeratorRoleForOtherUsers(t *testing.T) {
	srv := &trackingUserService{}
	router := newAdminPermissionsRouter(testSessionValidator{userID: "admin-user"}, "admin", srv)

	for _, moderator := range []bool{true, false} {
		body := `{"moderador":false}`
		if moderator {
			body = `{"moderador":true}`
		}
		recorder := adminPermissionRequest(
			router,
			http.MethodPatch,
			"/users/admin/users/u2/moderation",
			body,
			nil,
		)
		if recorder.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200; body = %s", recorder.Code, recorder.Body.String())
		}
	}
	if srv.moderationCalls != 2 || srv.lastModerator {
		t.Errorf("llamadas = %d, último valor = %t; want 2, false", srv.moderationCalls, srv.lastModerator)
	}
}
