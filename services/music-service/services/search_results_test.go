package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/meloop/music-service/config"
)

type memorySearchCache struct {
	values map[string]map[string]any
	ttl    time.Duration
}

type failingSearchCache struct{}

func (failingSearchCache) Get(context.Context, string) (map[string]any, bool, error) {
	return nil, false, errors.New("cache unavailable")
}

func (failingSearchCache) Set(context.Context, string, map[string]any, time.Duration) error {
	return errors.New("cache unavailable")
}

func (c *memorySearchCache) Get(_ context.Context, key string) (map[string]any, bool, error) {
	value, ok := c.values[key]
	return value, ok, nil
}

func (c *memorySearchCache) Set(_ context.Context, key string, value map[string]any, ttl time.Duration) error {
	c.values[key] = value
	c.ttl = ttl
	return nil
}

func TestSearchNormalizesPaginatesAndCaches(t *testing.T) {
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/api/token" {
			_, _ = w.Write([]byte(`{"access_token":"fake-token","expires_in":3600}`))
			return
		}
		requests++
		offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		if limit > 10 || r.URL.Query().Get("q") != "Rosalía" {
			t.Errorf("unexpected query %s", r.URL.RawQuery)
		}
		items := make([]map[string]any, 0, limit)
		for i := 0; i < limit; i++ {
			items = append(items, map[string]any{
				"id": fmt.Sprintf("track-%d", offset+i), "name": "Canción",
				"artists":       []any{map[string]any{"name": "Rosalía"}},
				"album":         map[string]any{"name": "Álbum", "images": nil},
				"external_urls": map[string]any{"spotify": "https://open.spotify.com/track/example"},
			})
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"tracks": map[string]any{"items": items, "total": 100},
		})
	}))
	defer server.Close()
	catalog, err := NewSpotifyCatalog(config.Config{
		SpotifyClientID: "test-client", SpotifyClientSecret: "test-secret", SpotifyMarket: "CL",
	})
	if err != nil {
		t.Fatal(err)
	}
	catalog.client.apiURL = server.URL + "/v1"
	catalog.client.tokenURL = server.URL + "/api/token"
	cache := &memorySearchCache{values: make(map[string]map[string]any)}
	catalog.SetSearchCache(cache)
	for range 2 {
		result, err := catalog.Search(context.Background(), "Rosalía", "track", 0, 20)
		if err != nil {
			t.Fatal(err)
		}
		items := result["results"].([]musicSearchItem)
		if len(items) != 20 || items[0].Artists[0] != "Rosalía" || items[0].Album != "Álbum" || items[0].ImageURL != "" || items[19].SpotifyID != "track-19" {
			t.Fatalf("unexpected normalized results: %+v", items)
		}
		if result["limit"] != 20 || result["offset"] != 0 || result["total"] != 100 {
			t.Fatalf("unexpected pagination: %+v", result)
		}
	}
	if requests != 2 || cache.ttl != 5*time.Minute {
		t.Fatalf("Spotify requests = %d, cache TTL = %v", requests, cache.ttl)
	}
}

func TestSearchAtSpotifyOffsetBoundary(t *testing.T) {
	var offsets []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/api/token" {
			_, _ = w.Write([]byte(`{"access_token":"fake-token","expires_in":3600}`))
			return
		}

		offsets = append(offsets, r.URL.Query().Get("offset"))
		offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		items := make([]map[string]string, 0, limit)
		for i := 0; i < limit; i++ {
			items = append(items, map[string]string{"id": fmt.Sprint(offset + i), "name": "song"})
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"tracks": map[string]any{"items": items, "total": 2000}})
	}))
	defer server.Close()
	catalog, err := NewSpotifyCatalog(config.Config{SpotifyClientID: "id", SpotifyClientSecret: "secret", SpotifyMarket: "CL"})
	if err != nil {
		t.Fatal(err)
	}
	catalog.client.apiURL, catalog.client.tokenURL = server.URL+"/v1", server.URL+"/api/token"
	result, err := catalog.Search(context.Background(), "test", "track", 995, 15)
	if err != nil {
		t.Fatal(err)
	}
	items := result["results"].([]musicSearchItem)
	if len(items) != 15 || items[10].SpotifyID != "1005" || len(offsets) != 2 || offsets[1] != "1000" {
		t.Fatalf("offsets = %v, items = %+v", offsets, items)
	}
}

func TestSearchContinuesWhenCacheFailsAndReturnsEmptyResults(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/api/token" {
			_, _ = w.Write([]byte(`{"access_token":"fake-token","expires_in":3600}`))
			return
		}
		_, _ = w.Write([]byte(`{"artists":{"items":[],"total":0}}`))
	}))
	defer server.Close()
	catalog, err := NewSpotifyCatalog(config.Config{SpotifyClientID: "id", SpotifyClientSecret: "secret", SpotifyMarket: "CL"})
	if err != nil {
		t.Fatal(err)
	}
	catalog.client.apiURL, catalog.client.tokenURL = server.URL+"/v1", server.URL+"/api/token"
	catalog.SetSearchCache(failingSearchCache{})
	result, err := catalog.Search(context.Background(), "none", "artist", 0, 10)
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	var response struct {
		Results []json.RawMessage `json:"results"`
		Total   int               `json:"total"`
	}
	if err := json.Unmarshal(encoded, &response); err != nil || response.Results == nil || len(response.Results) != 0 || response.Total != 0 {
		t.Fatalf("unexpected empty response: %s (%v)", encoded, err)
	}
}

func TestNormalizeSearchItemAcceptsNullFields(t *testing.T) {
	var item spotifySearchItem
	if err := json.Unmarshal([]byte(`{"id":"id","name":"song","artists":null,"album":null,"images":null,"external_urls":null}`), &item); err != nil {
		t.Fatal(err)
	}
	result := normalizeSearchItem(item, "track")
	if result.SpotifyID != "id" || result.Artists == nil || len(result.Artists) != 0 || result.ImageURL != "" || result.ExternalURL != "" || result.Album != "" {
		t.Fatalf("unexpected result: %+v", result)
	}
}
