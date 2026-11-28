package services

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/meloop/music-service/config"
)

func TestSearchUsesCachedClientCredentialsTokenAndMarket(t *testing.T) {
	var tokenRequests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/api/token":
			tokenRequests.Add(1)
			username, password, ok := request.BasicAuth()
			if !ok || username != "test-client" || password != "test-secret" {
				t.Errorf("unexpected client credentials authorization")
			}
			if err := request.ParseForm(); err != nil {
				t.Errorf("parse token form: %v", err)
			}
			if request.Form.Get("grant_type") != "client_credentials" {
				t.Errorf("grant_type = %q", request.Form.Get("grant_type"))
			}
			writer.Header().Set("Content-Type", "application/json")
			_, _ = writer.Write([]byte(`{"access_token":"fake-access-token","expires_in":3600}`))
		case "/v1/search":
			if request.Header.Get("Authorization") != "Bearer fake-access-token" {
				t.Errorf("Spotify access token was not sent")
			}
			if request.URL.Query().Get("q") != "daft punk" ||
				request.URL.Query().Get("type") != "track" ||
				request.URL.Query().Get("market") != "CL" ||
				request.URL.Query().Get("offset") != "20" ||
				request.URL.Query().Get("limit") != "10" {
				t.Errorf("unexpected search query: %s", request.URL.RawQuery)
			}
			writer.Header().Set("Content-Type", "application/json")
			_, _ = writer.Write([]byte(`{"tracks":{"items":[],"offset":20,"limit":10,"total":0,"next":null}}`))
		default:
			http.NotFound(writer, request)
		}
	}))
	defer server.Close()

	catalog, err := NewSpotifyCatalog(config.Config{
		SpotifyClientID:     "test-client",
		SpotifyClientSecret: "test-secret",
		SpotifyMarket:       "CL",
	})
	if err != nil {
		t.Fatal(err)
	}
	catalog.client.apiURL = server.URL + "/v1"
	catalog.client.tokenURL = server.URL + "/api/token"

	for range 2 {
		result, err := catalog.Search(context.Background(), "daft punk", "track", 20, 10)
		if err != nil {
			t.Fatal(err)
		}
		tracks, ok := result["tracks"].(map[string]any)
		if !ok || tracks["offset"] != float64(20) {
			t.Fatalf("unexpected Spotify search payload: %#v", result)
		}
	}
	if got := tokenRequests.Load(); got != 1 {
		t.Fatalf("token requests = %d, want 1", got)
	}
}

func TestDetailRequestsSpotifyTrackWithMarket(t *testing.T) {
	var tokenRequests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path == "/api/token" {
			tokenRequests.Add(1)
			writer.Header().Set("Content-Type", "application/json")
			_, _ = writer.Write([]byte(`{"access_token":"fake-access-token","expires_in":3600}`))
			return
		}
		if request.URL.Path != "/v1/tracks/4iV5W9uYEdYUVa79Axb7Rh" ||
			request.URL.Query().Get("market") != "CL" {
			t.Errorf("unexpected Spotify detail request: %s", request.URL.RequestURI())
		}
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"id":"4iV5W9uYEdYUVa79Axb7Rh","name":"Test Track"}`))
	}))
	defer server.Close()

	catalog, err := NewSpotifyCatalog(config.Config{
		SpotifyClientID:     "test-client",
		SpotifyClientSecret: "test-secret",
		SpotifyMarket:       "CL",
	})
	if err != nil {
		t.Fatal(err)
	}
	catalog.client.apiURL = server.URL + "/v1"
	catalog.client.tokenURL = server.URL + "/api/token"

	result, err := catalog.Detail(context.Background(), "tracks", "4iV5W9uYEdYUVa79Axb7Rh")
	if err != nil {
		t.Fatal(err)
	}
	if result["name"] != "Test Track" {
		t.Fatalf("detail name = %v", result["name"])
	}
	if tokenRequests.Load() != 1 {
		t.Fatal("expected one client-credentials token request")
	}
}

func TestGenreSearchReturnsFlutterCompatiblePage(t *testing.T) {
	catalog := &SpotifyCatalog{}
	result, err := catalog.Search(context.Background(), "hip", "genre", 0, 10)
	if err != nil {
		t.Fatal(err)
	}

	genres, ok := result["genres"].(map[string]any)
	if !ok {
		t.Fatalf("genres page has type %T", result["genres"])
	}
	items, ok := genres["items"].([]genre)
	if !ok {
		t.Fatalf("genre items have type %T", genres["items"])
	}
	if len(items) != 1 || items[0].ID != "hip-hop" {
		t.Fatalf("unexpected genre items: %+v", items)
	}
	if genres["has_more"] != false {
		t.Fatalf("has_more = %v, want false", genres["has_more"])
	}
}
