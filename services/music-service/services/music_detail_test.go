package services

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/meloop/music-service/config"
)

type memorySongStore struct {
	values map[string]map[string]any
	err    error
}

func (s *memorySongStore) Get(_ context.Context, id, market string) (map[string]any, bool, error) {
	value, found := s.values[id+":"+market]
	return value, found, s.err
}

func (s *memorySongStore) Set(_ context.Context, id, market string, value map[string]any) error {
	if s.err != nil {
		return s.err
	}
	s.values[id+":"+market] = value
	return nil
}

func testDetailCatalog(t *testing.T, handler http.HandlerFunc) *SpotifyCatalog {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/api/token" {
			_, _ = w.Write([]byte(`{"access_token":"fake-token","expires_in":3600}`))
			return
		}
		handler(w, r)
	}))
	t.Cleanup(server.Close)
	catalog, err := NewSpotifyCatalog(config.Config{SpotifyClientID: "id", SpotifyClientSecret: "secret", SpotifyMarket: "CL"})
	if err != nil {
		t.Fatal(err)
	}
	catalog.client.apiURL, catalog.client.tokenURL = server.URL+"/v1", server.URL+"/api/token"
	return catalog
}

func TestTrackDetailUsesPersistentStoreAfterFirstRequest(t *testing.T) {
	requests := 0
	catalog := testDetailCatalog(t, func(w http.ResponseWriter, r *http.Request) {
		requests++
		if r.URL.Query().Get("market") != "CL" {
			t.Errorf("market = %q", r.URL.Query().Get("market"))
		}
		_, _ = w.Write([]byte(`{"id":"track-id","name":"Song","artists":[{"name":"Artist"}],"album":{"name":"Album","images":[{"url":"https://example.test/cover"}]},"external_urls":{"spotify":"https://open.spotify.com/track/example"}}`))
	})
	store := &memorySongStore{values: make(map[string]map[string]any)}
	catalog.SetSongDetailStore(store)
	for range 2 {
		detail, err := catalog.Detail(context.Background(), "tracks", "track-id")
		if err != nil {
			t.Fatal(err)
		}
		if detail["spotify_id"] != "track-id" || detail["name"] != "Song" ||
			detail["album"] != "Album" || detail["image_url"] != "https://example.test/cover" {
			t.Fatalf("unexpected normalized detail: %+v", detail)
		}
	}
	if requests != 1 || store.values["track-id:CL"] == nil {
		t.Fatalf("Spotify requests = %d, store = %+v", requests, store.values)
	}
}

func TestArtistDetailCachesOptionalGenresAndImage(t *testing.T) {
	requests := 0
	catalog := testDetailCatalog(t, func(w http.ResponseWriter, r *http.Request) {
		requests++
		_, _ = w.Write([]byte(`{"id":"artist-id","name":"Artist","genres":null,"images":null,"external_urls":null}`))
	})
	cache := &memorySearchCache{values: make(map[string]map[string]any)}
	catalog.SetSearchCache(cache)
	for range 2 {
		detail, err := catalog.Detail(context.Background(), "artists", "artist-id")
		if err != nil {
			t.Fatal(err)
		}
		genres, ok := detail["genres"].([]string)
		if !ok || len(genres) != 0 || detail["image_url"] != nil {
			t.Fatalf("unexpected artist detail: %+v", detail)
		}
	}
	if requests != 1 || cache.ttl != 24*time.Hour {
		t.Fatalf("requests = %d, TTL = %v", requests, cache.ttl)
	}
}

func TestAlbumDetailFollowsTrackPages(t *testing.T) {
	catalog := testDetailCatalog(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/albums/album-id":
			_, _ = w.Write([]byte(`{"id":"album-id","name":"Album","tracks":{"items":[{"id":"first","name":"First"}],"offset":0,"next":"https://example.test/next"}}`))
		case "/v1/albums/album-id/tracks":
			if r.URL.Query().Get("offset") != "1" || r.URL.Query().Get("market") != "CL" {
				t.Errorf("unexpected pagination: %s", r.URL.RawQuery)
			}
			_, _ = w.Write([]byte(`{"items":[{"id":"second","name":"Second"}],"offset":1,"next":null}`))
		default:
			http.NotFound(w, r)
		}
	})
	detail, err := catalog.Detail(context.Background(), "albums", "album-id")
	if err != nil {
		t.Fatal(err)
	}
	tracks := detail["tracks"].([]musicSearchItem)
	if len(tracks) != 2 || tracks[1].SpotifyID != "second" {
		t.Fatalf("tracks = %+v", tracks)
	}
}

func TestTrackDetailFallsBackWhenDatabaseUnavailable(t *testing.T) {
	catalog := testDetailCatalog(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"id":"track-id","name":"Song"}`))
	})
	catalog.SetSongDetailStore(&memorySongStore{err: errors.New("database unavailable")})
	if _, err := catalog.Detail(context.Background(), "tracks", "track-id"); err != nil {
		t.Fatalf("Spotify fallback failed: %v", err)
	}
}

func TestDetailCacheSeparatesMarketAndDoesNotCacheErrors(t *testing.T) {
	requests := 0
	catalog := testDetailCatalog(t, func(w http.ResponseWriter, r *http.Request) {
		requests++
		if requests == 1 {
			http.Error(w, "unavailable", http.StatusServiceUnavailable)
			return
		}
		_, _ = w.Write([]byte(`{"id":"artist-id","name":"Artist"}`))
	})
	cache := &memorySearchCache{values: make(map[string]map[string]any)}
	catalog.SetSearchCache(cache)
	if _, err := catalog.Detail(context.Background(), "artists", "artist-id"); err == nil {
		t.Fatal("expected Spotify error")
	}
	if len(cache.values) != 0 {
		t.Fatal("Spotify error must not be cached")
	}
	if _, err := catalog.Detail(context.Background(), "artists", "artist-id"); err != nil {
		t.Fatal(err)
	}
	catalog.client.market = "US"
	if _, err := catalog.Detail(context.Background(), "artists", "artist-id"); err != nil {
		t.Fatal(err)
	}
	if requests != 3 || len(cache.values) != 2 {
		t.Fatalf("requests = %d, cache entries = %d", requests, len(cache.values))
	}
}
