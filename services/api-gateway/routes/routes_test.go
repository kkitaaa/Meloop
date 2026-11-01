package routes_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/meloop/api-gateway/routes"
)

func TestVersionedSocialRouteProxiesToSocialService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	type forwardedRequest struct {
		path   string
		userID string
	}
	forwarded := make(chan forwardedRequest, 1)
	socialService := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		forwarded <- forwardedRequest{
			path:   request.URL.Path,
			userID: request.Header.Get("X-User-ID"),
		}
		writer.WriteHeader(http.StatusNoContent)
	}))
	defer socialService.Close()
	t.Setenv("SOCIAL_SERVICE_URL", socialService.URL)

	router := gin.New()
	routes.SetupRoutes(router)
	gateway := httptest.NewServer(router)
	defer gateway.Close()
	request, err := http.NewRequest(
		http.MethodGet,
		gateway.URL+"/v1/social/users/target/mutual-friends?limit=3",
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("X-User-ID", "viewer")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", response.StatusCode)
	}
	requestReceived := <-forwarded
	if requestReceived.path != "/users/target/mutual-friends" {
		t.Fatalf("unexpected proxied path: %q", requestReceived.path)
	}
	if requestReceived.userID != "viewer" {
		t.Fatalf("identity header was not forwarded: %q", requestReceived.userID)
	}
}
