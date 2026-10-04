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

func TestAdminRoutesProxyToTheirServices(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var userPath, rewardPath string
	userTarget := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userPath = r.URL.RequestURI()
		w.WriteHeader(http.StatusOK)
	}))
	defer userTarget.Close()
	rewardTarget := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rewardPath = r.URL.RequestURI()
		w.WriteHeader(http.StatusOK)
	}))
	defer rewardTarget.Close()
	t.Setenv("USER_SERVICE_URL", userTarget.URL)
	t.Setenv("GAMIFICATION_SERVICE_URL", rewardTarget.URL)

	router := gin.New()
	SetupRoutes(router)
	gateway := httptest.NewServer(router)
	defer gateway.Close()

	for path := range map[string]*string{
		"/admin/users?scope=all":   &userPath,
		"/admin/rewards?scope=all": &rewardPath,
	} {
		request, err := http.NewRequest(http.MethodGet, gateway.URL+path, nil)
		if err != nil {
			t.Fatalf("create request: %v", err)
		}
		request.Header.Set("Authorization", "Bearer admin-token")
		response, err := http.DefaultClient.Do(request)
		if err != nil {
			t.Fatalf("send request: %v", err)
		}
		response.Body.Close()
		if response.StatusCode != http.StatusOK {
			t.Fatalf("GET %s status = %d, want %d", path, response.StatusCode, http.StatusOK)
		}
	}
	if userPath != "/admin/users?scope=all" || rewardPath != "/admin/rewards?scope=all" {
		t.Fatalf("admin paths not proxied correctly: users=%q rewards=%q", userPath, rewardPath)
	}
}
