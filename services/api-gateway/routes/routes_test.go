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

func funcRouter() http.Handler {
	router := gin.New()
	SetupRoutes(router)
	return router
}
