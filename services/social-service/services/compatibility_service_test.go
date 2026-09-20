package services

import (
	"context"
	"errors"
	"math"
	"testing"

	"github.com/meloop/social-service/models"
)

type compatibilityRepositoryFake struct {
	dataA *models.UserMusicalData
	dataB *models.UserMusicalData
	err   error
}

func (f *compatibilityRepositoryFake) GetMusicalProfiles(ctx context.Context, userAID, userBID string) (*models.UserMusicalData, *models.UserMusicalData, error) {
	if f.err != nil {
		return nil, nil, f.err
	}
	return f.dataA, f.dataB, nil
}

func TestCalculateCompatibilityScore_Scenarios(t *testing.T) {
	t.Run("1. Identical preferences across all 4 categories returns 100%", func(t *testing.T) {
		userA := &models.UserMusicalData{
			UserID:           "user-a",
			Genres:           []string{"rock", "pop", "jazz"},
			Artists:          []string{"artist-1", "artist-2"},
			Tracks:           []string{"track-1", "track-2", "track-3"},
			InteractedTracks: []string{"track-1", "track-4"},
		}
		userB := &models.UserMusicalData{
			UserID:           "user-b",
			Genres:           []string{"rock", "pop", "jazz"},
			Artists:          []string{"artist-1", "artist-2"},
			Tracks:           []string{"track-1", "track-2", "track-3"},
			InteractedTracks: []string{"track-1", "track-4"},
		}

		result := CalculateCompatibilityScore(userA, userB)

		if result.MatchScore != 100 {
			t.Fatalf("expected MatchScore = 100, got %d", result.MatchScore)
		}
		if result.GenreSimilarity != 1.0 || result.ArtistSimilarity != 1.0 || result.TrackSimilarity != 1.0 || result.InteractionSimilarity != 1.0 {
			t.Fatalf("expected all similarities to be 1.0, got %+v", result)
		}
		if len(result.CommonGenres) != 3 || len(result.CommonArtists) != 2 || len(result.CommonTracks) != 3 || len(result.CommonInteractedTracks) != 2 {
			t.Fatalf("unexpected common elements length in result: %+v", result)
		}
	})

	t.Run("2. Completely disjoint preferences returns 0%", func(t *testing.T) {
		userA := &models.UserMusicalData{
			UserID:           "user-a",
			Genres:           []string{"rock", "metal"},
			Artists:          []string{"artist-1"},
			Tracks:           []string{"track-1"},
			InteractedTracks: []string{"track-10"},
		}
		userB := &models.UserMusicalData{
			UserID:           "user-b",
			Genres:           []string{"pop", "jazz"},
			Artists:          []string{"artist-2"},
			Tracks:           []string{"track-2"},
			InteractedTracks: []string{"track-20"},
		}

		result := CalculateCompatibilityScore(userA, userB)

		if result.MatchScore != 0 {
			t.Fatalf("expected MatchScore = 0, got %d", result.MatchScore)
		}
		if result.GenreSimilarity != 0.0 || result.ArtistSimilarity != 0.0 || result.TrackSimilarity != 0.0 || result.InteractionSimilarity != 0.0 {
			t.Fatalf("expected all similarities to be 0.0, got %+v", result)
		}
		if len(result.CommonGenres) != 0 || len(result.CommonArtists) != 0 || len(result.CommonTracks) != 0 || len(result.CommonInteractedTracks) != 0 {
			t.Fatalf("expected no common elements, got %+v", result)
		}
	})

	t.Run("3. Partial matches calculate accurate weighted percentage", func(t *testing.T) {
		// Genres: A={Rock, Metal, Jazz}, B={Rock, Metal, Pop} => Inter=2, Union=4 => sim=0.5 (weight 0.30 => 0.15)
		// Artists: A={A1, A2}, B={A1, A3} => Inter=1, Union=3 => sim=1/3 (weight 0.30 => 0.10)
		// Tracks: A={T1, T2}, B={T1, T2} => Inter=2, Union=2 => sim=1.0 (weight 0.30 => 0.30)
		// Interactions: A={T1}, B={T2} => Inter=0, Union=2 => sim=0.0 (weight 0.10 => 0.0)
		// Total score = 0.15 + 0.10 + 0.30 + 0.00 = 0.55 => 55%
		userA := &models.UserMusicalData{
			UserID:           "user-a",
			Genres:           []string{"Rock", "Metal", "Jazz"},
			Artists:          []string{"A1", "A2"},
			Tracks:           []string{"T1", "T2"},
			InteractedTracks: []string{"T1"},
		}
		userB := &models.UserMusicalData{
			UserID:           "user-b",
			Genres:           []string{"Rock", "Metal", "Pop"},
			Artists:          []string{"A1", "A3"},
			Tracks:           []string{"T1", "T2"},
			InteractedTracks: []string{"T2"},
		}

		result := CalculateCompatibilityScore(userA, userB)

		if math.Abs(result.GenreSimilarity-0.5) > 1e-9 {
			t.Fatalf("expected GenreSimilarity 0.5, got %f", result.GenreSimilarity)
		}
		if math.Abs(result.ArtistSimilarity-(1.0/3.0)) > 1e-9 {
			t.Fatalf("expected ArtistSimilarity 1/3, got %f", result.ArtistSimilarity)
		}
		if result.TrackSimilarity != 1.0 {
			t.Fatalf("expected TrackSimilarity 1.0, got %f", result.TrackSimilarity)
		}
		if result.InteractionSimilarity != 0.0 {
			t.Fatalf("expected InteractionSimilarity 0.0, got %f", result.InteractionSimilarity)
		}
		if result.MatchScore != 55 {
			t.Fatalf("expected MatchScore = 55, got %d", result.MatchScore)
		}
	})

	t.Run("4. Match only in genres yields 30%", func(t *testing.T) {
		userA := &models.UserMusicalData{
			UserID:           "user-a",
			Genres:           []string{"Rock"},
			Artists:          []string{"A1"},
			Tracks:           []string{"T1"},
			InteractedTracks: []string{"IT1"},
		}
		userB := &models.UserMusicalData{
			UserID:           "user-b",
			Genres:           []string{"Rock"},
			Artists:          []string{"A2"},
			Tracks:           []string{"T2"},
			InteractedTracks: []string{"IT2"},
		}

		result := CalculateCompatibilityScore(userA, userB)

		if result.MatchScore != 30 {
			t.Fatalf("expected MatchScore = 30, got %d", result.MatchScore)
		}
		if result.GenreSimilarity != 1.0 || result.ArtistSimilarity != 0.0 || result.TrackSimilarity != 0.0 || result.InteractionSimilarity != 0.0 {
			t.Fatalf("unexpected similarities: %+v", result)
		}
	})

	t.Run("5. Match only in artists yields 30%", func(t *testing.T) {
		userA := &models.UserMusicalData{
			UserID:           "user-a",
			Genres:           []string{"G1"},
			Artists:          []string{"ArtistMatch"},
			Tracks:           []string{"T1"},
			InteractedTracks: []string{"IT1"},
		}
		userB := &models.UserMusicalData{
			UserID:           "user-b",
			Genres:           []string{"G2"},
			Artists:          []string{"ArtistMatch"},
			Tracks:           []string{"T2"},
			InteractedTracks: []string{"IT2"},
		}

		result := CalculateCompatibilityScore(userA, userB)

		if result.MatchScore != 30 {
			t.Fatalf("expected MatchScore = 30, got %d", result.MatchScore)
		}
		if result.ArtistSimilarity != 1.0 || result.GenreSimilarity != 0.0 {
			t.Fatalf("unexpected similarities: %+v", result)
		}
	})

	t.Run("6. Match only in tracks yields 30%", func(t *testing.T) {
		userA := &models.UserMusicalData{
			UserID:           "user-a",
			Genres:           []string{"G1"},
			Artists:          []string{"A1"},
			Tracks:           []string{"TrackMatch"},
			InteractedTracks: []string{"IT1"},
		}
		userB := &models.UserMusicalData{
			UserID:           "user-b",
			Genres:           []string{"G2"},
			Artists:          []string{"A2"},
			Tracks:           []string{"TrackMatch"},
			InteractedTracks: []string{"IT2"},
		}

		result := CalculateCompatibilityScore(userA, userB)

		if result.MatchScore != 30 {
			t.Fatalf("expected MatchScore = 30, got %d", result.MatchScore)
		}
		if result.TrackSimilarity != 1.0 || result.ArtistSimilarity != 0.0 {
			t.Fatalf("unexpected similarities: %+v", result)
		}
	})

	t.Run("7. Match only in interactions yields 10%", func(t *testing.T) {
		userA := &models.UserMusicalData{
			UserID:           "user-a",
			Genres:           []string{"G1"},
			Artists:          []string{"A1"},
			Tracks:           []string{"T1"},
			InteractedTracks: []string{"InteractedMatch"},
		}
		userB := &models.UserMusicalData{
			UserID:           "user-b",
			Genres:           []string{"G2"},
			Artists:          []string{"A2"},
			Tracks:           []string{"T2"},
			InteractedTracks: []string{"InteractedMatch"},
		}

		result := CalculateCompatibilityScore(userA, userB)

		if result.MatchScore != 10 {
			t.Fatalf("expected MatchScore = 10, got %d", result.MatchScore)
		}
		if result.InteractionSimilarity != 1.0 || result.TrackSimilarity != 0.0 {
			t.Fatalf("unexpected similarities: %+v", result)
		}
	})

	t.Run("8. Both users without musical data returns 0% without errors or NaN", func(t *testing.T) {
		userA := &models.UserMusicalData{
			UserID:           "user-a",
			Genres:           []string{},
			Artists:          []string{},
			Tracks:           []string{},
			InteractedTracks: []string{},
		}
		userB := &models.UserMusicalData{
			UserID:           "user-b",
			Genres:           []string{},
			Artists:          []string{},
			Tracks:           []string{},
			InteractedTracks: []string{},
		}

		result := CalculateCompatibilityScore(userA, userB)

		if result.MatchScore != 0 {
			t.Fatalf("expected MatchScore = 0, got %d", result.MatchScore)
		}
		if math.IsNaN(result.GenreSimilarity) || math.IsNaN(result.ArtistSimilarity) ||
			math.IsNaN(result.TrackSimilarity) || math.IsNaN(result.InteractionSimilarity) {
			t.Fatal("similarties must not be NaN")
		}
	})

	t.Run("9. Nil user musical data handled gracefully as empty profiles", func(t *testing.T) {
		result := CalculateCompatibilityScore(nil, nil)
		if result.MatchScore != 0 {
			t.Fatalf("expected MatchScore = 0 for nil profiles, got %d", result.MatchScore)
		}
	})

	t.Run("10. One user with data and other user empty returns 0%", func(t *testing.T) {
		userA := &models.UserMusicalData{
			UserID:           "user-a",
			Genres:           []string{"Rock", "Pop"},
			Artists:          []string{"A1", "A2"},
			Tracks:           []string{"T1"},
			InteractedTracks: []string{"IT1"},
		}
		userB := &models.UserMusicalData{
			UserID:           "user-b",
			Genres:           []string{},
			Artists:          []string{},
			Tracks:           []string{},
			InteractedTracks: []string{},
		}

		result := CalculateCompatibilityScore(userA, userB)

		if result.MatchScore != 0 {
			t.Fatalf("expected MatchScore = 0, got %d", result.MatchScore)
		}
		if result.GenreSimilarity != 0.0 || result.ArtistSimilarity != 0.0 || result.TrackSimilarity != 0.0 || result.InteractionSimilarity != 0.0 {
			t.Fatalf("expected all similarities to be 0.0, got %+v", result)
		}
	})

	t.Run("11. Partial empty categories in both users (only genres populated and matching)", func(t *testing.T) {
		userA := &models.UserMusicalData{
			UserID:           "user-a",
			Genres:           []string{"Indie"},
			Artists:          []string{},
			Tracks:           []string{},
			InteractedTracks: []string{},
		}
		userB := &models.UserMusicalData{
			UserID:           "user-b",
			Genres:           []string{"Indie"},
			Artists:          []string{},
			Tracks:           []string{},
			InteractedTracks: []string{},
		}

		result := CalculateCompatibilityScore(userA, userB)

		if result.MatchScore != 30 {
			t.Fatalf("expected MatchScore = 30, got %d", result.MatchScore)
		}
		if result.GenreSimilarity != 1.0 {
			t.Fatalf("expected GenreSimilarity 1.0, got %f", result.GenreSimilarity)
		}
	})

	t.Run("12. Determinism across 1000 iterations", func(t *testing.T) {
		userA := &models.UserMusicalData{
			UserID:           "user-a",
			Genres:           []string{"Jazz", "Blues", "Rock"},
			Artists:          []string{"Miles Davis", "B.B. King"},
			Tracks:           []string{"So What", "The Thrill Is Gone"},
			InteractedTracks: []string{"So What", "Blue in Green"},
		}
		userB := &models.UserMusicalData{
			UserID:           "user-b",
			Genres:           []string{"Jazz", "Soul", "Blues"},
			Artists:          []string{"Miles Davis", "Ray Charles"},
			Tracks:           []string{"So What", "Hit the Road Jack"},
			InteractedTracks: []string{"So What"},
		}

		firstResult := CalculateCompatibilityScore(userA, userB)

		for i := 0; i < 1000; i++ {
			res := CalculateCompatibilityScore(userA, userB)
			if res.MatchScore != firstResult.MatchScore {
				t.Fatalf("non-deterministic MatchScore on iteration %d: got %d, expected %d", i, res.MatchScore, firstResult.MatchScore)
			}
			if res.GenreSimilarity != firstResult.GenreSimilarity ||
				res.ArtistSimilarity != firstResult.ArtistSimilarity ||
				res.TrackSimilarity != firstResult.TrackSimilarity ||
				res.InteractionSimilarity != firstResult.InteractionSimilarity {
				t.Fatalf("non-deterministic float similarities on iteration %d", i)
			}
		}
	})
}

func TestCompatibilityService_CalculateCompatibility(t *testing.T) {
	t.Run("successful calculation through service and repo", func(t *testing.T) {
		repo := &compatibilityRepositoryFake{
			dataA: &models.UserMusicalData{
				UserID: "user-1",
				Genres: []string{"Rock"},
			},
			dataB: &models.UserMusicalData{
				UserID: "user-2",
				Genres: []string{"Rock"},
			},
		}

		service := NewCompatibilityService(repo)
		result, err := service.CalculateCompatibility(context.Background(), "user-1", "user-2")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result == nil || result.MatchScore != 30 {
			t.Fatalf("expected MatchScore = 30, got %+v", result)
		}
	})

	t.Run("empty user ID returns ErrInvalidUser", func(t *testing.T) {
		service := NewCompatibilityService(&compatibilityRepositoryFake{})

		_, err := service.CalculateCompatibility(context.Background(), "", "user-2")
		if !errors.Is(err, ErrInvalidUser) {
			t.Fatalf("expected ErrInvalidUser, got %v", err)
		}

		_, err = service.CalculateCompatibility(context.Background(), "user-1", "   ")
		if !errors.Is(err, ErrInvalidUser) {
			t.Fatalf("expected ErrInvalidUser, got %v", err)
		}
	})

	t.Run("same user IDs returns ErrSelfCompatibility", func(t *testing.T) {
		service := NewCompatibilityService(&compatibilityRepositoryFake{})

		_, err := service.CalculateCompatibility(context.Background(), "user-1", "user-1")
		if !errors.Is(err, ErrSelfCompatibility) {
			t.Fatalf("expected ErrSelfCompatibility, got %v", err)
		}
	})

	t.Run("repository error propagates properly", func(t *testing.T) {
		dbErr := errors.New("db connection failure")
		repo := &compatibilityRepositoryFake{
			err: dbErr,
		}

		service := NewCompatibilityService(repo)
		_, err := service.CalculateCompatibility(context.Background(), "user-1", "user-2")
		if !errors.Is(err, dbErr) {
			t.Fatalf("expected %v, got %v", dbErr, err)
		}
	})
}
