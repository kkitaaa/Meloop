package routes

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/meloop/post-service/controllers"
)

func TestRegisteredRoutesRequireAuthenticationExceptHealth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	authMiddleware := controllers.AuthRequired(nil)
	SetupRoutes(router, authMiddleware, nil, nil, nil, nil, nil)

	routes := router.Routes()
	if len(routes) < 2 {
		t.Fatalf("registered routes = %d, expected service routes", len(routes))
	}

	for _, route := range routes {
		t.Run(route.Method+" "+route.Path, func(t *testing.T) {
			if strings.Contains(route.Path, ":userId") {
				t.Fatalf("route must not accept a user ID supplied by the caller: %s", route.Path)
			}
			requestPath := strings.NewReplacer(
				":postId", "post-1",
				":authorId", "author-1",
				":commentId", "comment-1",
				":id", "item-1",
			).Replace(route.Path)
			request := httptest.NewRequest(route.Method, requestPath, nil)
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, request)

			if route.Path == "/health" {
				if recorder.Code != http.StatusOK {
					t.Fatalf("public health status = %d, want %d", recorder.Code, http.StatusOK)
				}
				return
			}
			if recorder.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, want %d; body = %s", recorder.Code, http.StatusUnauthorized, recorder.Body.String())
			}
		})
	}
}

func TestAdditionalRouteRegistrarsRequireAuthentication(t *testing.T) {
	gin.SetMode(gin.TestMode)
	authMiddleware := controllers.AuthRequired(nil)

	t.Run("post read routes", func(t *testing.T) {
		router := gin.New()
		SetupPostRoutes(router.Group(""), nil, authMiddleware)
		assertAllRoutesRequireAuthentication(t, router)
	})

	t.Run("interaction routes", func(t *testing.T) {
		router := gin.New()
		SetupInteractionRoutes(router, nil, authMiddleware)
		assertAllRoutesRequireAuthentication(t, router)
	})
}

func assertAllRoutesRequireAuthentication(t *testing.T, router *gin.Engine) {
	t.Helper()
	for _, route := range router.Routes() {
		t.Run(route.Method+" "+route.Path, func(t *testing.T) {
			requestPath := strings.NewReplacer(":id", "item-1", ":authorId", "author-1").Replace(route.Path)
			request := httptest.NewRequest(route.Method, requestPath, nil)
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, request)
			if recorder.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, want %d; body = %s", recorder.Code, http.StatusUnauthorized, recorder.Body.String())
			}
		})
	}
}
