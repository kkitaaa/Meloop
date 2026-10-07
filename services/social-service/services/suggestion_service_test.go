package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/meloop/social-service/models"
	"github.com/meloop/social-service/services"
)

type mockFriendshipRepoForSuggestions struct {
	candidates []models.CandidateUser
	mutual     map[string]int
	candErr    error
	mutualErr  error
}

func (m *mockFriendshipRepoForSuggestions) CreatePending(ctx context.Context, senderID, receiverID string) (*models.FriendRequest, error) {
	return nil, nil
}
func (m *mockFriendshipRepoForSuggestions) ListReceived(ctx context.Context, userID string) ([]models.FriendRequest, error) {
	return nil, nil
}
func (m *mockFriendshipRepoForSuggestions) ListSent(ctx context.Context, userID string) ([]models.FriendRequest, error) {
	return nil, nil
}
func (m *mockFriendshipRepoForSuggestions) Accept(ctx context.Context, requestID int, receiverID string) (*models.FriendRequest, error) {
	return nil, nil
}
func (m *mockFriendshipRepoForSuggestions) Reject(ctx context.Context, requestID int, receiverID string) (*models.FriendRequest, error) {
	return nil, nil
}
func (m *mockFriendshipRepoForSuggestions) Cancel(ctx context.Context, requestID int, senderID string) (*models.FriendRequest, error) {
	return nil, nil
}
func (m *mockFriendshipRepoForSuggestions) ListFriends(ctx context.Context, userID string) ([]models.Friend, error) {
	return nil, nil
}
func (m *mockFriendshipRepoForSuggestions) RemoveFriend(ctx context.Context, userID, targetID string) error {
	return nil
}
func (m *mockFriendshipRepoForSuggestions) GetFriendProfile(ctx context.Context, userID, friendID string) (*models.FriendProfile, error) {
	return nil, nil
}
func (m *mockFriendshipRepoForSuggestions) BlockUser(ctx context.Context, blockerID, blockedID string) error {
	return nil
}
func (m *mockFriendshipRepoForSuggestions) IsBlocked(ctx context.Context, user1ID, user2ID string) (bool, error) {
	return false, nil
}
func (m *mockFriendshipRepoForSuggestions) ValidateInteraction(ctx context.Context, user1ID, user2ID string) error {
	return nil
}
func (m *mockFriendshipRepoForSuggestions) GetEligibleCandidates(ctx context.Context, userID string) ([]models.CandidateUser, error) {
	if m.candErr != nil {
		return nil, m.candErr
	}
	return m.candidates, nil
}
func (m *mockFriendshipRepoForSuggestions) GetMutualFriendsCount(ctx context.Context, userID string, candidateIDs []string) (map[string]int, error) {
	if m.mutualErr != nil {
		return nil, m.mutualErr
	}
	if m.mutual == nil {
		return map[string]int{}, nil
	}
	return m.mutual, nil
}

type mockCompatibilityRepoForSuggestions struct {
	profiles map[string]*models.UserMusicalData
	err      error
}

func (m *mockCompatibilityRepoForSuggestions) GetMusicalProfiles(ctx context.Context, userAID, userBID string) (*models.UserMusicalData, *models.UserMusicalData, error) {
	if m.err != nil {
		return nil, nil, m.err
	}
	return m.profiles[userAID], m.profiles[userBID], nil
}

func (m *mockCompatibilityRepoForSuggestions) GetBatchMusicalProfiles(ctx context.Context, userIDs []string) (map[string]*models.UserMusicalData, error) {
	if m.err != nil {
		return nil, m.err
	}
	res := make(map[string]*models.UserMusicalData, len(userIDs))
	for _, id := range userIDs {
		if data, ok := m.profiles[id]; ok {
			res[id] = data
		} else {
			res[id] = &models.UserMusicalData{UserID: id}
		}
	}
	return res, nil
}

// =========================================================================
// 1. Tests de Exclusiones
// =========================================================================

func TestSuggestions_Exclusions(t *testing.T) {
	t.Run("Self user is strictly excluded", func(t *testing.T) {
		repo := &mockFriendshipRepoForSuggestions{
			candidates: []models.CandidateUser{
				{IDUsuario: "user-me", Username: "me", IsActive: true},
				{IDUsuario: "user-cand-1", Username: "carlos", IsActive: true},
			},
		}
		service := services.NewFriendshipService(repo, nil)

		suggestions, err := service.GetFriendSuggestions(context.Background(), "user-me")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		for _, s := range suggestions {
			if s.IDUsuario == "user-me" {
				t.Fatalf("own user must not appear in suggestions")
			}
		}
		if len(suggestions) != 1 || suggestions[0].IDUsuario != "user-cand-1" {
			t.Fatalf("expected 1 valid candidate, got %+v", suggestions)
		}
	})

	t.Run("Inactive accounts are strictly excluded", func(t *testing.T) {
		repo := &mockFriendshipRepoForSuggestions{
			candidates: []models.CandidateUser{
				{IDUsuario: "user-inactive", Username: "banned_user", IsActive: false},
				{IDUsuario: "user-cand-1", Username: "active_user", IsActive: true},
			},
		}
		service := services.NewFriendshipService(repo, nil)

		suggestions, err := service.GetFriendSuggestions(context.Background(), "user-me")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		for _, s := range suggestions {
			if s.IDUsuario == "user-inactive" {
				t.Fatalf("inactive account must not appear in suggestions")
			}
		}
		if len(suggestions) != 1 || suggestions[0].IDUsuario != "user-cand-1" {
			t.Fatalf("expected only active candidate, got %+v", suggestions)
		}
	})
}

// =========================================================================
// 2. Tests de Motivos (DetermineSuggestionReason)
// =========================================================================

func TestDetermineSuggestionReason(t *testing.T) {
	t.Run("Shared Artist reason generation", func(t *testing.T) {
		score := &models.CompatibilityScore{
			CommonArtists: []string{"The Strokes"},
			CommonTracks:  []string{"Reptilia"},
		}
		reason := services.DetermineSuggestionReason(score, 0)
		expected := "Ambos escuchan a The Strokes"
		if reason != expected {
			t.Fatalf("expected %q, got %q", expected, reason)
		}
	})

	t.Run("Shared Track reason generation when no shared artist", func(t *testing.T) {
		score := &models.CompatibilityScore{
			CommonArtists: []string{},
			CommonTracks:  []string{"Bohemian Rhapsody"},
		}
		reason := services.DetermineSuggestionReason(score, 0)
		expected := "Ambos tienen interés en Bohemian Rhapsody"
		if reason != expected {
			t.Fatalf("expected %q, got %q", expected, reason)
		}
	})

	t.Run("Single Mutual Friend reason generation", func(t *testing.T) {
		score := &models.CompatibilityScore{}
		reason := services.DetermineSuggestionReason(score, 1)
		expected := "Tienen 1 amigo en común"
		if reason != expected {
			t.Fatalf("expected %q, got %q", expected, reason)
		}
	})

	t.Run("Multiple Mutual Friends reason generation", func(t *testing.T) {
		score := &models.CompatibilityScore{}
		reason := services.DetermineSuggestionReason(score, 4)
		expected := "Tienen 4 amigos en común"
		if reason != expected {
			t.Fatalf("expected %q, got %q", expected, reason)
		}
	})

	t.Run("Shared Interaction Activity reason generation", func(t *testing.T) {
		score := &models.CompatibilityScore{
			CommonInteractedTracks: []string{"Track 123"},
		}
		reason := services.DetermineSuggestionReason(score, 0)
		expected := "Ambos interactuaron con la canción Track 123"
		if reason != expected {
			t.Fatalf("expected %q, got %q", expected, reason)
		}
	})

	t.Run("Shared Genre reason generation", func(t *testing.T) {
		score := &models.CompatibilityScore{
			CommonGenres: []string{"Indie Rock"},
		}
		reason := services.DetermineSuggestionReason(score, 0)
		expected := "Ambos disfrutan del género Indie Rock"
		if reason != expected {
			t.Fatalf("expected %q, got %q", expected, reason)
		}
	})

	t.Run("Fallback generic motive when no signals exist", func(t *testing.T) {
		score := &models.CompatibilityScore{}
		reason := services.DetermineSuggestionReason(score, 0)
		expected := "Sugerencia de la comunidad"
		if reason != expected {
			t.Fatalf("expected %q, got %q", expected, reason)
		}
	})

	t.Run("Nil score handled safely without panic", func(t *testing.T) {
		reason := services.DetermineSuggestionReason(nil, 2)
		if reason != "Tienen 2 amigos en común" {
			t.Fatalf("expected 'Tienen 2 amigos en común', got %q", reason)
		}

		reasonNoSignals := services.DetermineSuggestionReason(nil, 0)
		if reasonNoSignals != "Sugerencia de la comunidad" {
			t.Fatalf("expected 'Sugerencia de la comunidad', got %q", reasonNoSignals)
		}
	})
}

// =========================================================================
// 3. Tests de Ranking y Afinidad
// =========================================================================

func TestRankSuggestions_AffinityAndSignals(t *testing.T) {
	t.Run("Higher RF-15 compatibility ranks first", func(t *testing.T) {
		candidates := []models.CandidateUser{
			{IDUsuario: "cand-low", Username: "low_compat", IsActive: true},
			{IDUsuario: "cand-high", Username: "high_compat", IsActive: true},
		}
		scores := map[string]*models.CompatibilityScore{
			"cand-low": {
				UserAID:    "user-me",
				UserBID:    "cand-low",
				MatchScore: 20,
			},
			"cand-high": {
				UserAID:       "user-me",
				UserBID:       "cand-high",
				MatchScore:    85,
				CommonArtists: []string{"Daft Punk"},
			},
		}
		mutual := map[string]int{
			"cand-low":  0,
			"cand-high": 0,
		}

		ranked := services.RankSuggestions(candidates, scores, mutual)

		if len(ranked) != 2 {
			t.Fatalf("expected 2 suggestions, got %d", len(ranked))
		}
		if ranked[0].IDUsuario != "cand-high" || ranked[1].IDUsuario != "cand-low" {
			t.Fatalf("expected cand-high first, got order: %s, %s", ranked[0].IDUsuario, ranked[1].IDUsuario)
		}
		if ranked[0].MatchScore != 85 || ranked[0].Motivo != "Ambos escuchan a Daft Punk" {
			t.Fatalf("unexpected candidate data: %+v", ranked[0])
		}
	})

	t.Run("Mutual friends increase affinity ranking", func(t *testing.T) {
		// cand-1: MatchScore = 50, MutualFriends = 0 => Affinity = 50
		// cand-2: MatchScore = 50, MutualFriends = 3 => Affinity = 80
		candidates := []models.CandidateUser{
			{IDUsuario: "cand-1", Username: "cand_no_mutual", IsActive: true},
			{IDUsuario: "cand-2", Username: "cand_with_mutual", IsActive: true},
		}
		scores := map[string]*models.CompatibilityScore{
			"cand-1": {MatchScore: 50},
			"cand-2": {MatchScore: 50},
		}
		mutual := map[string]int{
			"cand-1": 0,
			"cand-2": 3,
		}

		ranked := services.RankSuggestions(candidates, scores, mutual)

		if ranked[0].IDUsuario != "cand-2" {
			t.Fatalf("expected candidate with mutual friends to rank first, got %s", ranked[0].IDUsuario)
		}
		if ranked[0].MutualFriends != 3 || ranked[0].Motivo != "Tienen 3 amigos en común" {
			t.Fatalf("unexpected data: %+v", ranked[0])
		}
	})

	t.Run("Activity history interacciones participes en ranking", func(t *testing.T) {
		candidates := []models.CandidateUser{
			{IDUsuario: "cand-a", Username: "cand_a", IsActive: true},
			{IDUsuario: "cand-b", Username: "cand_b", IsActive: true},
		}
		scores := map[string]*models.CompatibilityScore{
			"cand-a": {
				MatchScore:             40,
				CommonInteractedTracks: []string{"Song 1"},
			},
			"cand-b": {
				MatchScore:             40,
				CommonInteractedTracks: []string{},
			},
		}
		mutual := map[string]int{
			"cand-a": 0,
			"cand-b": 0,
		}

		ranked := services.RankSuggestions(candidates, scores, mutual)

		if ranked[0].IDUsuario != "cand-a" {
			t.Fatalf("expected candidate with interacted track to rank first on tie-break, got %s", ranked[0].IDUsuario)
		}
	})

	t.Run("Strict determinism on multiple executions", func(t *testing.T) {
		candidates := []models.CandidateUser{
			{IDUsuario: "user-z", Username: "z", IsActive: true},
			{IDUsuario: "user-a", Username: "a", IsActive: true},
			{IDUsuario: "user-m", Username: "m", IsActive: true},
		}
		scores := map[string]*models.CompatibilityScore{
			"user-z": {MatchScore: 50},
			"user-a": {MatchScore: 50},
			"user-m": {MatchScore: 50},
		}
		mutual := map[string]int{
			"user-z": 1,
			"user-a": 1,
			"user-m": 1,
		}

		firstRun := services.RankSuggestions(candidates, scores, mutual)

		for i := 0; i < 500; i++ {
			run := services.RankSuggestions(candidates, scores, mutual)
			for j := range firstRun {
				if run[j].IDUsuario != firstRun[j].IDUsuario {
					t.Fatalf("non-deterministic ranking at index %d on iteration %d", j, i)
				}
			}
		}
	})
}

// =========================================================================
// 4. Tests de Casos Límite y Flujo Completo
// =========================================================================

func TestGetFriendSuggestions_FullFlowAndEdgeCases(t *testing.T) {
	t.Run("Full flow with RF-15 compatibility engine integration", func(t *testing.T) {
		userMe := "user-me"
		userCand1 := "user-cand-1"
		userCand2 := "user-cand-2"

		friendRepo := &mockFriendshipRepoForSuggestions{
			candidates: []models.CandidateUser{
				{IDUsuario: userCand1, Username: "carlos", IsActive: true, Experiencia: 1200},
				{IDUsuario: userCand2, Username: "ana", IsActive: true, Experiencia: 800},
			},
			mutual: map[string]int{
				userCand1: 2,
				userCand2: 0,
			},
		}

		compatRepo := &mockCompatibilityRepoForSuggestions{
			profiles: map[string]*models.UserMusicalData{
				userMe: {
					UserID:  userMe,
					Genres:  []string{"Rock", "Indie"},
					Artists: []string{"Arctic Monkeys", "The Strokes"},
					Tracks:  []string{"Do I Wanna Know"},
				},
				userCand1: {
					UserID:  userCand1,
					Genres:  []string{"Rock"},
					Artists: []string{"The Strokes"},
					Tracks:  []string{"Someday"},
				},
				userCand2: {
					UserID:  userCand2,
					Genres:  []string{"Pop", "Electronic"},
					Artists: []string{"Dua Lipa"},
				},
			},
		}

		service := services.NewFriendshipService(friendRepo, nil, compatRepo)

		suggestions, err := service.GetFriendSuggestions(context.Background(), userMe)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(suggestions) != 2 {
			t.Fatalf("expected 2 suggestions, got %d", len(suggestions))
		}

		// userCand1 has shared artist "The Strokes" and Rock + 2 mutual friends => high score & rank 1
		if suggestions[0].IDUsuario != userCand1 {
			t.Fatalf("expected userCand1 to rank 1, got %s", suggestions[0].IDUsuario)
		}
		if suggestions[0].MatchScore <= 0 {
			t.Fatalf("expected positive MatchScore from RF-15, got %d", suggestions[0].MatchScore)
		}
		if suggestions[0].MutualFriends != 2 {
			t.Fatalf("expected 2 mutual friends, got %d", suggestions[0].MutualFriends)
		}
		if suggestions[0].Motivo != "Ambos escuchan a The Strokes" {
			t.Fatalf("expected reason 'Ambos escuchan a The Strokes', got %q", suggestions[0].Motivo)
		}

		// userCand2 has 0 compatibility and 0 mutual friends => rank 2, fallback reason
		if suggestions[1].IDUsuario != userCand2 {
			t.Fatalf("expected userCand2 to rank 2, got %s", suggestions[1].IDUsuario)
		}
		if suggestions[1].Motivo != "Sugerencia de la comunidad" {
			t.Fatalf("expected fallback reason, got %q", suggestions[1].Motivo)
		}
	})

	t.Run("No eligible candidates returns empty slice without error", func(t *testing.T) {
		friendRepo := &mockFriendshipRepoForSuggestions{
			candidates: []models.CandidateUser{},
		}
		service := services.NewFriendshipService(friendRepo, nil)

		suggestions, err := service.GetFriendSuggestions(context.Background(), "user-me")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if suggestions == nil || len(suggestions) != 0 {
			t.Fatalf("expected empty non-nil slice, got %+v", suggestions)
		}
	})

	t.Run("Empty user ID returns ErrInvalidUser", func(t *testing.T) {
		service := services.NewFriendshipService(&mockFriendshipRepoForSuggestions{}, nil)

		_, err := service.GetFriendSuggestions(context.Background(), "")
		if !errors.Is(err, services.ErrInvalidUser) {
			t.Fatalf("expected ErrInvalidUser, got %v", err)
		}

		_, err = service.GetFriendSuggestions(context.Background(), "   ")
		if !errors.Is(err, services.ErrInvalidUser) {
			t.Fatalf("expected ErrInvalidUser, got %v", err)
		}
	})

	t.Run("Repository candidate query error propagates", func(t *testing.T) {
		dbErr := errors.New("db error on candidates")
		friendRepo := &mockFriendshipRepoForSuggestions{
			candErr: dbErr,
		}
		service := services.NewFriendshipService(friendRepo, nil)

		_, err := service.GetFriendSuggestions(context.Background(), "user-me")
		if !errors.Is(err, dbErr) {
			t.Fatalf("expected %v, got %v", dbErr, err)
		}
	})

	t.Run("Repository mutual friends query error propagates", func(t *testing.T) {
		dbErr := errors.New("db error on mutual friends")
		friendRepo := &mockFriendshipRepoForSuggestions{
			candidates: []models.CandidateUser{
				{IDUsuario: "cand-1", Username: "carlos", IsActive: true},
			},
			mutualErr: dbErr,
		}
		service := services.NewFriendshipService(friendRepo, nil)

		_, err := service.GetFriendSuggestions(context.Background(), "user-me")
		if !errors.Is(err, dbErr) {
			t.Fatalf("expected %v, got %v", dbErr, err)
		}
	})

	t.Run("Compatibility repository error propagates", func(t *testing.T) {
		compatErr := errors.New("error fetching musical data")
		friendRepo := &mockFriendshipRepoForSuggestions{
			candidates: []models.CandidateUser{
				{IDUsuario: "cand-1", Username: "carlos", IsActive: true},
			},
		}
		compatRepo := &mockCompatibilityRepoForSuggestions{
			err: compatErr,
		}
		service := services.NewFriendshipService(friendRepo, nil, compatRepo)

		_, err := service.GetFriendSuggestions(context.Background(), "user-me")
		if !errors.Is(err, compatErr) {
			t.Fatalf("expected %v, got %v", compatErr, err)
		}
	})
}
