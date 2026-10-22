package services

import (
	"context"
	"errors"
	"testing"

	"github.com/meloop/social-service/models"
	"github.com/meloop/social-service/repositories"
)

type friendshipRepositoryFake struct {
	createErr         error
	acceptErr         error
	rejectErr         error
	cancelErr         error
	listFriendsErr    error
	removeFriendErr   error
	getProfileErr     error
	blockUserErr      error
	validateErr       error
	request           *models.FriendRequest
	friends           []models.Friend
	profile           *models.FriendProfile
	created           bool
	removedFriend     bool
	blockedUser       bool
	removedFriendship bool
	excludeBlockedErr error
	blockedPairs      map[string]map[string]bool
}

func (f *friendshipRepositoryFake) CreatePending(context.Context, string, string) (*models.FriendRequest, error) {
	if f.createErr != nil {
		return nil, f.createErr
	}
	f.created = true
	return f.request, nil
}

func (f *friendshipRepositoryFake) ListReceived(context.Context, string) ([]models.FriendRequest, error) {
	return []models.FriendRequest{}, nil
}

func (f *friendshipRepositoryFake) ListSent(context.Context, string) ([]models.FriendRequest, error) {
	return []models.FriendRequest{}, nil
}

func (f *friendshipRepositoryFake) Accept(context.Context, int, string) (*models.FriendRequest, error) {
	if f.acceptErr != nil {
		return nil, f.acceptErr
	}
	return f.request, nil
}

func (f *friendshipRepositoryFake) Reject(context.Context, int, string) (*models.FriendRequest, error) {
	if f.rejectErr != nil {
		return nil, f.rejectErr
	}
	return f.request, nil
}

func (f *friendshipRepositoryFake) Cancel(context.Context, int, string) (*models.FriendRequest, error) {
	if f.cancelErr != nil {
		return nil, f.cancelErr
	}
	return f.request, nil
}

func (f *friendshipRepositoryFake) ListFriends(ctx context.Context, userID string) ([]models.Friend, error) {
	if f.listFriendsErr != nil {
		return nil, f.listFriendsErr
	}
	if f.friends == nil {
		return []models.Friend{}, nil
	}
	return f.friends, nil
}

func (f *friendshipRepositoryFake) RemoveFriend(ctx context.Context, userID, targetID string) error {
	if f.removeFriendErr != nil {
		return f.removeFriendErr
	}
	f.removedFriend = true
	return nil
}

func (f *friendshipRepositoryFake) GetFriendProfile(ctx context.Context, userID, friendID string) (*models.FriendProfile, error) {
	if f.getProfileErr != nil {
		return nil, f.getProfileErr
	}
	return f.profile, nil
}

func (f *friendshipRepositoryFake) BlockUser(ctx context.Context, blockerID, blockedID string) error {
	if f.blockUserErr != nil {
		return f.blockUserErr
	}
	f.blockedUser = true
	f.removedFriendship = true
	return nil
}

func (f *friendshipRepositoryFake) IsBlocked(ctx context.Context, user1ID, user2ID string) (bool, error) {
	if f.validateErr != nil {
		return true, nil
	}
	return false, nil
}

func (f *friendshipRepositoryFake) ValidateInteraction(ctx context.Context, user1ID, user2ID string) error {
	if f.validateErr != nil {
		return f.validateErr
	}
	return nil
}

func (f *friendshipRepositoryFake) GetEligibleCandidates(ctx context.Context, userID string) ([]models.CandidateUser, error) {
	return []models.CandidateUser{}, nil
}

func (f *friendshipRepositoryFake) GetMutualFriendsCount(ctx context.Context, userID string, candidateIDs []string) (map[string]int, error) {
	return map[string]int{}, nil
}

func (f *friendshipRepositoryFake) ExcludeBlocked(ctx context.Context, userID string, candidateIDs []string) ([]string, error) {
	if f.excludeBlockedErr != nil {
		return nil, f.excludeBlockedErr
	}
	result := make([]string, 0, len(candidateIDs))
	for _, cand := range candidateIDs {
		blocked := (f.blockedPairs != nil && f.blockedPairs[userID] != nil && f.blockedPairs[userID][cand]) ||
			(f.blockedPairs != nil && f.blockedPairs[cand] != nil && f.blockedPairs[cand][userID])
		if !blocked {
			result = append(result, cand)
		}
	}
	return result, nil
}

func (f *friendshipRepositoryFake) GetBlockedCandidateIDs(ctx context.Context, userID string, candidateIDs []string) ([]string, error) {
	if f.excludeBlockedErr != nil {
		return nil, f.excludeBlockedErr
	}
	result := make([]string, 0)
	for _, cand := range candidateIDs {
		blocked := (f.blockedPairs != nil && f.blockedPairs[userID] != nil && f.blockedPairs[userID][cand]) ||
			(f.blockedPairs != nil && f.blockedPairs[cand] != nil && f.blockedPairs[cand][userID])
		if blocked {
			result = append(result, cand)
		}
	}
	return result, nil
}

type eventPublisherFake struct {
	called bool
}

func (f *eventPublisherFake) PublishFriendAccepted(context.Context, *models.FriendRequest) error {
	f.called = true
	return nil
}

func TestFriendshipServiceSendRequest(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want error
	}{
		{name: "valid request"},
		{name: "self request", err: ErrSelfRequest, want: ErrSelfRequest},
		{name: "duplicate request", err: repositories.ErrRequestConflict, want: ErrRequestConflict},
		{name: "inverse pending request", err: repositories.ErrRequestConflict, want: ErrRequestConflict},
		{name: "blocked users", err: repositories.ErrBlocked, want: ErrBlocked},
		{name: "accepted friendship", err: repositories.ErrRequestConflict, want: ErrRequestConflict},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := &friendshipRepositoryFake{
				createErr: test.err,
				request:   &models.FriendRequest{ID: 1, SenderID: "a", ReceiverID: "b", Status: models.FriendshipPending},
			}
			service := NewFriendshipService(repo, nil)
			_, err := service.SendRequest(context.Background(), "a", map[bool]string{true: "a", false: "b"}[test.name == "self request"])
			if !errors.Is(err, test.want) {
				t.Fatalf("error = %v, want %v", err, test.want)
			}
			if test.want == nil && !repo.created {
				t.Fatal("expected repository to create the request")
			}
		})
	}
}

func TestFriendshipServiceActionsEnforceAuthorizationAndState(t *testing.T) {
	tests := []struct {
		name string
		err  error
		call func(FriendshipService) error
	}{
		{name: "accept valid", call: func(s FriendshipService) error {
			_, err := s.AcceptRequest(context.Background(), 1, "receiver")
			return err
		}},
		{name: "reject valid", call: func(s FriendshipService) error {
			_, err := s.RejectRequest(context.Background(), 1, "receiver")
			return err
		}},
		{name: "cancel valid", call: func(s FriendshipService) error {
			_, err := s.CancelRequest(context.Background(), 1, "sender")
			return err
		}},
		{name: "accept non receiver", err: ErrNotReceiver, call: func(s FriendshipService) error {
			_, err := s.AcceptRequest(context.Background(), 1, "other")
			return err
		}},
		{name: "reject non receiver", err: ErrNotReceiver, call: func(s FriendshipService) error {
			_, err := s.RejectRequest(context.Background(), 1, "other")
			return err
		}},
		{name: "cancel non sender", err: ErrNotSender, call: func(s FriendshipService) error {
			_, err := s.CancelRequest(context.Background(), 1, "other")
			return err
		}},
		{name: "missing request", err: ErrRequestNotFound, call: func(s FriendshipService) error {
			_, err := s.AcceptRequest(context.Background(), 99, "receiver")
			return err
		}},
		{name: "already processed", err: ErrRequestProcessed, call: func(s FriendshipService) error {
			_, err := s.RejectRequest(context.Background(), 1, "receiver")
			return err
		}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := &friendshipRepositoryFake{
				request: &models.FriendRequest{ID: 1, SenderID: "sender", ReceiverID: "receiver", Status: models.FriendshipAccepted},
			}
			switch test.name {
			case "accept non receiver", "reject non receiver":
				repo.acceptErr = repositories.ErrNotReceiver
				repo.rejectErr = repositories.ErrNotReceiver
			case "cancel non sender":
				repo.cancelErr = repositories.ErrNotSender
			case "missing request":
				repo.acceptErr = repositories.ErrRequestNotFound
			case "already processed":
				repo.rejectErr = repositories.ErrRequestProcessed
			}
			if err := test.call(NewFriendshipService(repo, nil)); !errors.Is(err, test.err) {
				t.Fatalf("error = %v, want %v", err, test.err)
			}
		})
	}
}

func TestFriendshipServiceAcceptPublishesEventAfterRepositoryAction(t *testing.T) {
	repo := &friendshipRepositoryFake{
		request: &models.FriendRequest{ID: 1, SenderID: "sender", ReceiverID: "receiver", Status: models.FriendshipAccepted},
	}
	publisher := &eventPublisherFake{}
	service := NewFriendshipService(repo, publisher)

	request, err := service.AcceptRequest(context.Background(), 1, "receiver")
	if err != nil {
		t.Fatalf("AcceptRequest() error = %v", err)
	}
	if request.Status != models.FriendshipAccepted {
		t.Fatalf("status = %q, want %q", request.Status, models.FriendshipAccepted)
	}
	if !publisher.called {
		t.Fatal("expected friend.accepted event to be published")
	}
}

// =========================================================================
// Tests para RF-12: Gestión de Amigos
// =========================================================================

func TestFriendshipServiceListFriends(t *testing.T) {
	t.Run("successful list with friends", func(t *testing.T) {
		repo := &friendshipRepositoryFake{
			friends: []models.Friend{
				{IDAmistad: 1, IDUsuario: "user-2", Username: "carlos", Estado: models.FriendshipAccepted},
				{IDAmistad: 2, IDUsuario: "user-3", Username: "maria", Estado: models.FriendshipAccepted},
			},
		}
		service := NewFriendshipService(repo, nil)

		friends, err := service.ListFriends(context.Background(), "user-1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(friends) != 2 {
			t.Fatalf("expected 2 friends, got %d", len(friends))
		}
		if friends[0].Username != "carlos" || friends[1].Username != "maria" {
			t.Fatalf("unexpected friend data: %+v", friends)
		}
	})

	t.Run("empty friend list", func(t *testing.T) {
		repo := &friendshipRepositoryFake{
			friends: []models.Friend{},
		}
		service := NewFriendshipService(repo, nil)

		friends, err := service.ListFriends(context.Background(), "user-1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(friends) != 0 {
			t.Fatalf("expected 0 friends, got %d", len(friends))
		}
	})

	t.Run("empty user id returns ErrInvalidUser", func(t *testing.T) {
		repo := &friendshipRepositoryFake{}
		service := NewFriendshipService(repo, nil)

		_, err := service.ListFriends(context.Background(), "")
		if !errors.Is(err, ErrInvalidUser) {
			t.Fatalf("expected ErrInvalidUser, got %v", err)
		}
	})
}

func TestFriendshipServiceRemoveFriend(t *testing.T) {
	t.Run("successful remove friend", func(t *testing.T) {
		repo := &friendshipRepositoryFake{}
		service := NewFriendshipService(repo, nil)

		err := service.RemoveFriend(context.Background(), "user-1", "user-2")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !repo.removedFriend {
			t.Fatal("expected repository RemoveFriend to be called")
		}
	})

	t.Run("remove non-existent friendship", func(t *testing.T) {
		repo := &friendshipRepositoryFake{
			removeFriendErr: repositories.ErrFriendshipNotFound,
		}
		service := NewFriendshipService(repo, nil)

		err := service.RemoveFriend(context.Background(), "user-1", "999")
		if !errors.Is(err, ErrFriendshipNotFound) {
			t.Fatalf("expected ErrFriendshipNotFound, got %v", err)
		}
	})

	t.Run("remove unauthorized friendship", func(t *testing.T) {
		repo := &friendshipRepositoryFake{
			removeFriendErr: repositories.ErrNotAuthorized,
		}
		service := NewFriendshipService(repo, nil)

		err := service.RemoveFriend(context.Background(), "user-1", "12")
		if !errors.Is(err, ErrNotAuthorized) {
			t.Fatalf("expected ErrNotAuthorized, got %v", err)
		}
	})

	t.Run("invalid user or target parameters", func(t *testing.T) {
		repo := &friendshipRepositoryFake{}
		service := NewFriendshipService(repo, nil)

		if err := service.RemoveFriend(context.Background(), "", "user-2"); !errors.Is(err, ErrInvalidUser) {
			t.Fatalf("expected ErrInvalidUser, got %v", err)
		}
		if err := service.RemoveFriend(context.Background(), "user-1", ""); !errors.Is(err, ErrInvalidUser) {
			t.Fatalf("expected ErrInvalidUser, got %v", err)
		}
	})
}

func TestFriendshipServiceGetFriendProfile(t *testing.T) {
	t.Run("successful friend profile retrieval", func(t *testing.T) {
		nivel := 5
		repo := &friendshipRepositoryFake{
			profile: &models.FriendProfile{
				IDUsuario:   "user-2",
				Username:    "carlos",
				Correo:      "carlos@example.com",
				IDNivel:     &nivel,
				Experiencia: 2500,
				IDAmistad:   1,
			},
		}
		service := NewFriendshipService(repo, nil)

		profile, err := service.GetFriendProfile(context.Background(), "user-1", "user-2")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if profile == nil || profile.Username != "carlos" || profile.Experiencia != 2500 {
			t.Fatalf("unexpected profile: %+v", profile)
		}
	})

	t.Run("profile retrieval blocked between users", func(t *testing.T) {
		repo := &friendshipRepositoryFake{
			getProfileErr: repositories.ErrBlocked,
		}
		service := NewFriendshipService(repo, nil)

		_, err := service.GetFriendProfile(context.Background(), "user-1", "user-2")
		if !errors.Is(err, ErrBlocked) {
			t.Fatalf("expected ErrBlocked, got %v", err)
		}
	})

	t.Run("profile retrieval when not friends", func(t *testing.T) {
		repo := &friendshipRepositoryFake{
			getProfileErr: repositories.ErrNotFriends,
		}
		service := NewFriendshipService(repo, nil)

		_, err := service.GetFriendProfile(context.Background(), "user-1", "user-stranger")
		if !errors.Is(err, ErrNotFriends) {
			t.Fatalf("expected ErrNotFriends, got %v", err)
		}
	})

	t.Run("profile retrieval when user not found", func(t *testing.T) {
		repo := &friendshipRepositoryFake{
			getProfileErr: repositories.ErrUserNotFound,
		}
		service := NewFriendshipService(repo, nil)

		_, err := service.GetFriendProfile(context.Background(), "user-1", "non-existent")
		if !errors.Is(err, ErrUserNotFound) {
			t.Fatalf("expected ErrUserNotFound, got %v", err)
		}
	})
}

// =========================================================================
// Tests para RF-13: Bloqueo de Usuarios
// =========================================================================

func TestFriendshipServiceBlockUser(t *testing.T) {
	t.Run("successful block and friendship removal", func(t *testing.T) {
		repo := &friendshipRepositoryFake{}
		service := NewFriendshipService(repo, nil)

		err := service.BlockUser(context.Background(), "user-1", "user-2")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !repo.blockedUser {
			t.Fatal("expected user to be blocked")
		}
		if !repo.removedFriendship {
			t.Fatal("expected existing friendship to be removed upon block")
		}
	})

	t.Run("self block attempt is rejected", func(t *testing.T) {
		repo := &friendshipRepositoryFake{}
		service := NewFriendshipService(repo, nil)

		err := service.BlockUser(context.Background(), "user-1", "user-1")
		if !errors.Is(err, ErrSelfBlock) {
			t.Fatalf("expected ErrSelfBlock, got %v", err)
		}
	})

	t.Run("block when already blocked", func(t *testing.T) {
		repo := &friendshipRepositoryFake{
			blockUserErr: repositories.ErrAlreadyBlocked,
		}
		service := NewFriendshipService(repo, nil)

		err := service.BlockUser(context.Background(), "user-1", "user-2")
		if !errors.Is(err, ErrAlreadyBlocked) {
			t.Fatalf("expected ErrAlreadyBlocked, got %v", err)
		}
	})

	t.Run("empty blocker or blocked id", func(t *testing.T) {
		repo := &friendshipRepositoryFake{}
		service := NewFriendshipService(repo, nil)

		if err := service.BlockUser(context.Background(), "", "user-2"); !errors.Is(err, ErrInvalidUser) {
			t.Fatalf("expected ErrInvalidUser, got %v", err)
		}
		if err := service.BlockUser(context.Background(), "user-1", ""); !errors.Is(err, ErrInvalidUser) {
			t.Fatalf("expected ErrInvalidUser, got %v", err)
		}
	})
}

func TestFriendshipServiceValidateInteraction(t *testing.T) {
	t.Run("interaction allowed when not blocked", func(t *testing.T) {
		repo := &friendshipRepositoryFake{}
		service := NewFriendshipService(repo, nil)

		err := service.ValidateInteraction(context.Background(), "user-1", "user-2")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("interaction rejected when blocked (bidirectional A to B)", func(t *testing.T) {
		repo := &friendshipRepositoryFake{
			validateErr: repositories.ErrBlocked,
		}
		service := NewFriendshipService(repo, nil)

		err := service.ValidateInteraction(context.Background(), "user-1", "user-2")
		if !errors.Is(err, ErrBlocked) {
			t.Fatalf("expected ErrBlocked, got %v", err)
		}
	})

	t.Run("interaction rejected when blocked (bidirectional B to A)", func(t *testing.T) {
		repo := &friendshipRepositoryFake{
			validateErr: repositories.ErrBlocked,
		}
		service := NewFriendshipService(repo, nil)

		err := service.ValidateInteraction(context.Background(), "user-2", "user-1")
		if !errors.Is(err, ErrBlocked) {
			t.Fatalf("expected ErrBlocked, got %v", err)
		}
	})
}

func TestFriendshipService_ExcludeBlocked(t *testing.T) {
	ctx := context.Background()
	refUser := "usr-ref"

	// Configurar pares de bloqueo:
	// 1. usr-ref bloqueó a cand-blocked-by-ref
	// 2. cand-blocked-ref bloqueó a usr-ref
	// 3. usr-ref y cand-mutual se bloquearon mutuamente
	// 4. cand-third bloqueó a third-party-user (no debe afectar a usr-ref)
	blockedPairs := map[string]map[string]bool{
		"usr-ref": {
			"cand-blocked-by-ref": true,
			"cand-mutual":         true,
		},
		"cand-blocked-ref": {
			"usr-ref": true,
		},
		"cand-mutual": {
			"usr-ref": true,
		},
		"cand-third": {
			"unrelated-user": true,
		},
	}

	repo := &friendshipRepositoryFake{
		blockedPairs: blockedPairs,
	}
	service := NewFriendshipService(repo, nil)

	t.Run("Excluye candidato cuando el usuario de referencia lo bloqueó (unilateral A -> B)", func(t *testing.T) {
		candidates := []string{"cand-blocked-by-ref", "cand-clean-1"}
		filtered, err := service.ExcludeBlocked(ctx, refUser, candidates)
		if err != nil {
			t.Fatalf("error inesperado: %v", err)
		}

		for _, id := range filtered {
			if id == "cand-blocked-by-ref" {
				t.Errorf("se esperaba que 'cand-blocked-by-ref' fuera excluido de la lista")
			}
		}
		if len(filtered) != 1 || filtered[0] != "cand-clean-1" {
			t.Errorf("se esperaba conservar únicamente 'cand-clean-1', se obtuvo: %v", filtered)
		}
	})

	t.Run("Excluye candidato cuando este bloqueó al usuario de referencia (unilateral B -> A)", func(t *testing.T) {
		candidates := []string{"cand-blocked-ref", "cand-clean-1"}
		filtered, err := service.ExcludeBlocked(ctx, refUser, candidates)
		if err != nil {
			t.Fatalf("error inesperado: %v", err)
		}

		for _, id := range filtered {
			if id == "cand-blocked-ref" {
				t.Errorf("se esperaba que 'cand-blocked-ref' fuera excluido de la lista")
			}
		}
		if len(filtered) != 1 || filtered[0] != "cand-clean-1" {
			t.Errorf("se esperaba conservar únicamente 'cand-clean-1', se obtuvo: %v", filtered)
		}
	})

	t.Run("Conserva los candidatos sin bloqueos mutuos entre ambos usuarios", func(t *testing.T) {
		candidates := []string{"cand-clean-1", "cand-clean-2", "cand-clean-3"}
		filtered, err := service.ExcludeBlocked(ctx, refUser, candidates)
		if err != nil {
			t.Fatalf("error inesperado: %v", err)
		}

		if len(filtered) != 3 {
			t.Fatalf("se esperaban 3 candidatos conservados, se obtuvieron %d", len(filtered))
		}
		expected := []string{"cand-clean-1", "cand-clean-2", "cand-clean-3"}
		for i, id := range filtered {
			if id != expected[i] {
				t.Errorf("en posición %d se esperaba %s, se obtuvo %s", i, expected[i], id)
			}
		}
	})

	t.Run("Gestiona correctamente una lista vacía", func(t *testing.T) {
		filtered, err := service.ExcludeBlocked(ctx, refUser, []string{})
		if err != nil {
			t.Fatalf("error inesperado con lista vacía: %v", err)
		}
		if filtered == nil || len(filtered) != 0 {
			t.Errorf("se esperaba slice vacío no nulo, se obtuvo: %v", filtered)
		}
	})

	t.Run("No altera la lista original de candidatos (inmutabilidad)", func(t *testing.T) {
		original := []string{"cand-blocked-by-ref", "cand-clean-1", "cand-blocked-ref"}
		// Copia para verificación posterior
		originalBackup := make([]string, len(original))
		copy(originalBackup, original)

		filtered, err := service.ExcludeBlocked(ctx, refUser, original)
		if err != nil {
			t.Fatalf("error inesperado: %v", err)
		}

		// Verificar que el slice original no haya cambiado en tamaño ni contenido
		if len(original) != len(originalBackup) {
			t.Fatalf("la longitud de la lista original fue alterada: antes %d, después %d", len(originalBackup), len(original))
		}
		for i := range original {
			if original[i] != originalBackup[i] {
				t.Errorf("elemento %d de la lista original fue alterado: antes '%s', después '%s'", i, originalBackup[i], original[i])
			}
		}

		// Y que el resultado contenga solo el candidato no bloqueado
		if len(filtered) != 1 || filtered[0] != "cand-clean-1" {
			t.Errorf("resultado inesperado: %v", filtered)
		}
	})

	t.Run("Ausencia de falsos positivos por bloqueos con terceros o sólo en una dirección", func(t *testing.T) {
		// cand-third tiene un bloqueo registrado, pero con 'unrelated-user', no con refUser
		candidates := []string{"cand-third", "cand-clean-1"}
		filtered, err := service.ExcludeBlocked(ctx, refUser, candidates)
		if err != nil {
			t.Fatalf("error inesperado: %v", err)
		}

		if len(filtered) != 2 {
			t.Fatalf("cand-third no debió ser excluido por falsos positivos con terceros. Obtenido: %v", filtered)
		}
	})

	t.Run("Identifica correctamente los IDs excluidos por bloqueo", func(t *testing.T) {
		candidates := []string{"cand-blocked-by-ref", "cand-clean-1", "cand-blocked-ref", "cand-mutual"}
		blockedIDs, err := service.GetBlockedCandidateIDs(ctx, refUser, candidates)
		if err != nil {
			t.Fatalf("error inesperado: %v", err)
		}

		if len(blockedIDs) != 3 {
			t.Fatalf("se esperaban 3 IDs bloqueados, se obtuvieron %d: %v", len(blockedIDs), blockedIDs)
		}
	})

	t.Run("Error cuando el usuario de referencia está vacío", func(t *testing.T) {
		_, err := service.ExcludeBlocked(ctx, "   ", []string{"cand-1"})
		if !errors.Is(err, ErrInvalidUser) {
			t.Errorf("se esperaba ErrInvalidUser cuando userID está vacío, se obtuvo: %v", err)
		}
	})
}

func TestFilterBlockedCandidates(t *testing.T) {
	candidates := []string{"u1", "u2", "u3", "u4"}
	blocked := []string{"u2", "u4"}

	result := FilterBlockedCandidates(candidates, blocked)

	if len(result) != 2 || result[0] != "u1" || result[1] != "u3" {
		t.Errorf("resultado inesperado en FilterBlockedCandidates: %v", result)
	}

	// Inmutabilidad de la lista original
	if len(candidates) != 4 || candidates[1] != "u2" {
		t.Errorf("la lista original fue modificada: %v", candidates)
	}

	// Lista vacía
	emptyRes := FilterBlockedCandidates([]string{}, blocked)
	if emptyRes == nil || len(emptyRes) != 0 {
		t.Errorf("se esperaba slice vacío")
	}
}
