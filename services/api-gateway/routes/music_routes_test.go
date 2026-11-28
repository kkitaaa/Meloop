package routes_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/meloop/api-gateway/routes"
)

func TestVersionedMusicRouteProxiesPathQueryAndAuthorization(t *testing.T) {
	gin.SetMode(gin.TestMode)
	type forwardedRequest struct {
		path          string
		query         string
		authorization string
	}

	forwarded := make(chan forwardedRequest, 1)
	musicService := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		forwarded <- forwardedRequest{
			path:          request.URL.Path,
			query:         request.URL.RawQuery,
			authorization: request.Header.Get("Authorization"),
		}
		writer.WriteHeader(http.StatusNoContent)
	}))
	defer musicService.Close()
	t.Setenv("MUSIC_SERVICE_URL", musicService.URL)

	router := gin.New()
	routes.SetupRoutes(router)
	gateway := httptest.NewServer(router)
	defer gateway.Close()

	request, err := http.NewRequest(
		http.MethodGet,
		gateway.URL+"/v1/music/search?q=radiohead&type=track&limit=10",
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Authorization", "Bearer meloop-session")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", response.StatusCode)
	}
	got := <-forwarded
	if got.path != "/v1/music/search" {
		t.Fatalf("proxied path = %q", got.path)
	}
	if got.query != "q=radiohead&type=track&limit=10" {
		t.Fatalf("proxied query = %q", got.query)
	}
	if got.authorization != "Bearer meloop-session" {
		t.Fatalf("authorization header = %q", got.authorization)
	}
}

func TestMusicPreferencesRouteProxiesToUserService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	forwarded := make(chan string, 1)
	userService := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		forwarded <- request.URL.Path + "?" + request.URL.RawQuery
		writer.WriteHeader(http.StatusNoContent)
	}))
	defer userService.Close()
	musicService := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		t.Errorf("preferences request was sent to music-service: %s", request.URL.Path)
		writer.WriteHeader(http.StatusInternalServerError)
	}))
	defer musicService.Close()
	t.Setenv("USER_SERVICE_URL", userService.URL)
	t.Setenv("MUSIC_SERVICE_URL", musicService.URL)

	router := gin.New()
	routes.SetupRoutes(router)
	gateway := httptest.NewServer(router)
	defer gateway.Close()

	response, err := http.Get(gateway.URL + "/v1/music/preferences?source=flutter")
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", response.StatusCode)
	}
	if got := <-forwarded; got != "/v1/music/preferences?source=flutter" {
		t.Fatalf("user-service received %q", got)
	}
}
