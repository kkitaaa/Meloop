package repositories

import (
	"context"
	"testing"
)

func TestCompatibilityRepository_NilPool(t *testing.T) {
	repo := NewCompatibilityRepository(nil)
	_, _, err := repo.GetMusicalProfiles(context.Background(), "user-1", "user-2")
	if err == nil {
		t.Fatal("expected error when connection pool is nil")
	}
}

func TestClassifyPreference(t *testing.T) {
	genres := make(map[string]struct{})
	artists := make(map[string]struct{})
	tracks := make(map[string]struct{})

	classifyPreference("GENERO", "genre-1", genres, artists, tracks)
	classifyPreference("GENRE", "genre-2", genres, artists, tracks)
	classifyPreference("ARTISTA", "artist-1", genres, artists, tracks)
	classifyPreference("ARTIST", "artist-2", genres, artists, tracks)
	classifyPreference("CANCION", "track-1", genres, artists, tracks)
	classifyPreference("TRACK", "track-2", genres, artists, tracks)
	classifyPreference("SONG", "track-3", genres, artists, tracks)
	classifyPreference("UNKNOWN", "unknown-1", genres, artists, tracks)

	if len(genres) != 2 || len(artists) != 2 || len(tracks) != 3 {
		t.Fatalf("unexpected classification counts: genres=%d, artists=%d, tracks=%d", len(genres), len(artists), len(tracks))
	}
}

func TestMapKeysToSlice(t *testing.T) {
	m := map[string]struct{}{
		"item-1": {},
		"item-2": {},
	}
	slice := mapKeysToSlice(m)
	if len(slice) != 2 {
		t.Fatalf("expected 2 items, got %d", len(slice))
	}
}
