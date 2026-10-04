package routes

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestNotificationsRouteProxiesAuthorizationAndPath(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var gotPath, gotAuthorization string
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.RequestURI()
		gotAuthorization = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
	}))
	defer target.Close()
	t.Setenv("NOTIFICATION_SERVICE_URL", target.URL)

	router := gin.New()
	SetupRoutes(router)
	gateway := httptest.NewServer(router)
	defer gateway.Close()
	request, err := http.NewRequest(http.MethodGet, gateway.URL+"/notifications/preferences?scope=all", nil)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	request.Header.Set("Authorization", "Bearer test-token")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatalf("send request: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusOK)
	}
	if gotPath != "/notifications/preferences?scope=all" {
		t.Fatalf("proxied path = %q", gotPath)
	}
	if gotAuthorization != "Bearer test-token" {
		t.Fatalf("proxied Authorization = %q", gotAuthorization)
	}
}
