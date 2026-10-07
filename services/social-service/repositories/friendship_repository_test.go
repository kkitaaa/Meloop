package repositories

import (
	"context"
	"testing"
)

func TestFriendshipRepository_NilPool(t *testing.T) {
	repo := NewFriendshipRepository(nil)
	ctx := context.Background()

	if _, err := repo.CreatePending(ctx, "u1", "u2"); err == nil {
		t.Error("expected error with nil pool on CreatePending")
	}
	if _, err := repo.ListReceived(ctx, "u1"); err == nil {
		t.Error("expected error with nil pool on ListReceived")
	}
	if _, err := repo.ListSent(ctx, "u1"); err == nil {
		t.Error("expected error with nil pool on ListSent")
	}
	if _, err := repo.Accept(ctx, 1, "u1"); err == nil {
		t.Error("expected error with nil pool on Accept")
	}
	if _, err := repo.Reject(ctx, 1, "u1"); err == nil {
		t.Error("expected error with nil pool on Reject")
	}
	if _, err := repo.Cancel(ctx, 1, "u1"); err == nil {
		t.Error("expected error with nil pool on Cancel")
	}
	if _, err := repo.ListFriends(ctx, "u1"); err == nil {
		t.Error("expected error with nil pool on ListFriends")
	}
	if err := repo.RemoveFriend(ctx, "u1", "u2"); err == nil {
		t.Error("expected error with nil pool on RemoveFriend")
	}
	if _, err := repo.GetFriendProfile(ctx, "u1", "u2"); err == nil {
		t.Error("expected error with nil pool on GetFriendProfile")
	}
	if err := repo.BlockUser(ctx, "u1", "u2"); err == nil {
		t.Error("expected error with nil pool on BlockUser")
	}
	if _, err := repo.IsBlocked(ctx, "u1", "u2"); err == nil {
		t.Error("expected error with nil pool on IsBlocked")
	}
	if _, err := repo.GetEligibleCandidates(ctx, "u1"); err == nil {
		t.Error("expected error with nil pool on GetEligibleCandidates")
	}
	if _, err := repo.GetMutualFriendsCount(ctx, "u1", []string{"u2"}); err == nil {
		t.Error("expected error with nil pool on GetMutualFriendsCount")
	}
}
