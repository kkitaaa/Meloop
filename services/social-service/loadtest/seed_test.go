package main

import (
	"fmt"
	"testing"
)

func TestCycleFixtureHasThirtyUniqueFriendsPerUser(t *testing.T) {
	userIDs := make([]string, userCount)
	for i := range userIDs {
		userIDs[i] = fmt.Sprintf("user-%03d", i)
	}

	edges := buildEdges(userIDs)
	if len(edges) != 7500 {
		t.Fatalf("expected 7500 undirected edges, got %d", len(edges))
	}

	degrees := make(map[string]int, len(userIDs))
	pairs := make(map[[2]string]struct{}, len(edges))
	for _, edge := range edges {
		if edge.userID1 == edge.userID2 {
			t.Fatalf("self friendship generated for %q", edge.userID1)
		}
		degrees[edge.userID1]++
		degrees[edge.userID2]++
		pair := [2]string{edge.userID1, edge.userID2}
		if pair[0] > pair[1] {
			pair[0], pair[1] = pair[1], pair[0]
		}
		if _, exists := pairs[pair]; exists {
			t.Fatalf("duplicate undirected friendship generated: %v", pair)
		}
		pairs[pair] = struct{}{}
	}
	for _, userID := range userIDs {
		if degrees[userID] != friendsPerUser {
			t.Fatalf("expected degree %d for user %q, got %d", friendsPerUser, userID, degrees[userID])
		}
	}
}

func TestSixteenStepCandidateHasFifteenMutualFriends(t *testing.T) {
	userIDs := make([]string, userCount)
	for i := range userIDs {
		userIDs[i] = fmt.Sprintf("user-%03d", i)
	}
	edges := buildEdges(userIDs)
	adjacency := make(map[string]map[string]struct{}, len(userIDs))
	for _, edge := range edges {
		if adjacency[edge.userID1] == nil {
			adjacency[edge.userID1] = make(map[string]struct{})
		}
		if adjacency[edge.userID2] == nil {
			adjacency[edge.userID2] = make(map[string]struct{})
		}
		adjacency[edge.userID1][edge.userID2] = struct{}{}
		adjacency[edge.userID2][edge.userID1] = struct{}{}
	}

	for i, userID := range userIDs {
		candidateID := userIDs[(i+neighborsPerSide+1)%userCount]
		if _, isFriend := adjacency[userID][candidateID]; isFriend {
			t.Fatalf("candidate at distance 16 unexpectedly is a direct friend of %q", userID)
		}
		mutualCount := 0
		for friendID := range adjacency[userID] {
			if _, isMutual := adjacency[candidateID][friendID]; isMutual {
				mutualCount++
			}
		}
		if mutualCount != neighborsPerSide {
			t.Fatalf("expected %d mutual friends for user %q, got %d",
				neighborsPerSide, userID, mutualCount)
		}
	}
}

func TestBuildPreferencesCreatesTenUniqueRowsPerUser(t *testing.T) {
	userIDs := make([]string, userCount)
	for i := range userIDs {
		userIDs[i] = fmt.Sprintf("user-%03d", i)
	}

	preferences := buildPreferences(userIDs)
	if len(preferences) != userCount*10 {
		t.Fatalf("expected %d preferences, got %d", userCount*10, len(preferences))
	}
	ids := make(map[string]struct{}, len(preferences))
	for _, preference := range preferences {
		if _, exists := ids[preference.id]; exists {
			t.Fatalf("duplicate preference ID %q", preference.id)
		}
		ids[preference.id] = struct{}{}
	}
}
