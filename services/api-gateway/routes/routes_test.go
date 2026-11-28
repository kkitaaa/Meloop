package routes

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestVersionedSocialRouteForwardsPathAndQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)
	forwarded := make(chan string, 1)
	socialService := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		forwarded <- r.Method + " " + r.URL.RequestURI()
		w.WriteHeader(http.StatusNoContent)
	}))
	defer socialService.Close()

	t.Setenv("SOCIAL_SERVICE_URL", socialService.URL)
	gateway := httptest.NewServer(funcRouter())
	defer gateway.Close()

	response, err := http.Get(gateway.URL + "/v1/social/users/u2/mutual-friends?limit=1&cursor=u1")
	require.NoError(t, err)
	defer response.Body.Close()
	require.Equal(t, http.StatusNoContent, response.StatusCode)
	require.Equal(t, "GET /v1/social/users/u2/mutual-friends?limit=1&cursor=u1", <-forwarded)
}

func TestSavedPostsRouteForwardsToPostService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	forwarded := make(chan string, 1)
	postService := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		forwarded <- r.Method + " " + r.URL.RequestURI()
		w.WriteHeader(http.StatusNoContent)
	}))
	defer postService.Close()
	authService := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/auth/validate" || r.Header.Get("Authorization") != "Bearer test-session" {
			t.Errorf("auth request = %s with Authorization %q", r.URL.Path, r.Header.Get("Authorization"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"data":{"id":"user-1","role":"user"},"error":null}`))
	}))
	defer authService.Close()

	t.Setenv("POST_SERVICE_URL", postService.URL)
	t.Setenv("AUTH_SERVICE_URL", authService.URL)
	gateway := httptest.NewServer(funcRouter())
	defer gateway.Close()

	request, err := http.NewRequest(http.MethodGet, gateway.URL+"/v1/me/saved?offset=20&limit=20", nil)
	require.NoError(t, err)
	request.Header.Set("Authorization", "Bearer test-session")
	response, err := http.DefaultClient.Do(request)
	require.NoError(t, err)
	defer response.Body.Close()
	require.Equal(t, http.StatusNoContent, response.StatusCode)
	require.Equal(t, "GET /v1/me/saved?offset=20&limit=20", <-forwarded)
}

func funcRouter() http.Handler {
	router := gin.New()
	SetupRoutes(router)
	return router
}
