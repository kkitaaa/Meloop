package services_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/meloop/user-service/models"
	"github.com/meloop/user-service/services"
)

func TestHTTPSocialFriendCounter_RequestsInternalEndpointWithToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/internal/users/usr-target/friends/count" {
			t.Errorf("request = %s %s", r.Method, r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer service-token" {
			t.Errorf("authorization = %q", r.Header.Get("Authorization"))
		}
		_, _ = fmt.Fprint(w, `{"success":true,"data":{"count":6},"error":null}`)
	}))
	defer server.Close()

	count, err := services.NewHTTPSocialFriendCounter(server.URL, "service-token").
		GetFriendsCount(context.Background(), "usr-target")
	if err != nil {
		t.Fatalf("GetFriendsCount returned error: %v", err)
	}
	if count != 6 {
		t.Fatalf("count = %d, want 6", count)
	}
}

func TestHTTPSocialFriendCounter_FailsExplicitlyWhenUnavailable(t *testing.T) {
	counter := services.NewHTTPSocialFriendCounter("http://127.0.0.1:1", "service-token")
	_, err := counter.GetFriendsCount(context.Background(), "usr-target")
	if !errors.Is(err, services.ErrSocialCounterUnavailable) {
		t.Fatalf("error = %v, want ErrSocialCounterUnavailable", err)
	}
}

func TestHTTPSocialFriendCounter_AllowsResponseAfterFiveSeconds(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(5100 * time.Millisecond)
		_, _ = fmt.Fprint(w, `{"success":true,"data":{"count":6},"error":null}`)
	}))
	defer server.Close()

	count, err := services.NewHTTPSocialFriendCounter(server.URL, "service-token").
		GetFriendsCount(context.Background(), "usr-target")
	if err != nil {
		t.Fatalf("GetFriendsCount returned error for a 5.1s response: %v", err)
	}
	if count != 6 {
		t.Fatalf("count = %d, want 6", count)
	}
}

func TestHTTPSocialFriendCounter_TimeoutIsReportedAsUnavailable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer server.Close()

	_, err := services.NewHTTPSocialFriendCounter(server.URL, "service-token").
		GetFriendsCount(context.Background(), "usr-target")
	if !errors.Is(err, services.ErrSocialCounterUnavailable) {
		t.Fatalf("error = %v, want ErrSocialCounterUnavailable after the 7s timeout", err)
	}
}

func TestProfileService_ConsecutiveProfilesUseSocialCounter(t *testing.T) {
	var requests int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if r.Header.Get("Authorization") != "Bearer service-token" {
			t.Errorf("missing service token")
		}
		_, _ = fmt.Fprint(w, `{"success":true,"data":{"count":8},"error":null}`)
	}))
	defer server.Close()

	privacyRepo := &mockPrivacyRepository{
		configs: map[string]*models.ConfiguracionPrivacidad{
			"usr-target": {
				IDUsuario:                "usr-target",
				VisibilidadPerfil:        models.VisibilidadPublico,
				VisibilidadPublicaciones: models.VisibilidadPublico,
			},
		},
	}
	service := services.NewProfileServiceWithFriendCounter(
		&mockProfileRepository{},
		&mockUserRepository{},
		&mockMediaVerifier{},
		privacyRepo,
		services.NewHTTPSocialFriendCounter(server.URL, "service-token"),
	)

	for i := 0; i < 2; i++ {
		profile, err := service.GetThirdPartyProfile(context.Background(), "usr-requester", "usr-target")
		if err != nil {
			t.Fatalf("profile request %d failed: %v", i+1, err)
		}
		if profile.Estadisticas == nil || profile.Estadisticas.Amigos != 8 {
			t.Fatalf("profile friend count = %+v, want 8", profile.Estadisticas)
		}
	}
	if requests != 2 {
		t.Fatalf("social-service requests = %d, want a request for each profile lookup", requests)
	}
}
