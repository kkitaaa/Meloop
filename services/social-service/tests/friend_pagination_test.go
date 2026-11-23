package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sort"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/meloop/social-service/controllers"
	"github.com/meloop/social-service/models"
	"github.com/meloop/social-service/routes"
	"github.com/stretchr/testify/require"
)

func TestListFriendsCursorPaginationIntegration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	f := setupFixture(t)
	service := f.databaseService()

	const owner = 0
	acceptedAt := time.Date(2026, time.January, 2, 3, 4, 5, 0, time.UTC)
	expectedIDs := make([]string, 0, 5)
	for i := 0; i < 5; i++ {
		friendID := f.createUser(t, fmt.Sprintf("cursor-match-%d", i))
		friendshipID := f.createFriendshipIDs(t, f.userIDs[owner], friendID, models.FriendshipAccepted)
		_, err := f.pool.Exec(context.Background(),
			`UPDATE amistad SET fecha_amistad = $1 WHERE id_amistad = $2`, acceptedAt, friendshipID)
		require.NoError(t, err)
		expectedIDs = append(expectedIDs, friendshipID)
	}

	unmatchedFriendID := f.createUser(t, "not-in-search")
	unmatchedFriendshipID := f.createFriendshipIDs(t, f.userIDs[owner], unmatchedFriendID, models.FriendshipAccepted)
	_, err := f.pool.Exec(context.Background(),
		`UPDATE amistad SET fecha_amistad = $1 WHERE id_amistad = $2`, acceptedAt, unmatchedFriendshipID)
	require.NoError(t, err)

	pendingFriendID := f.createUser(t, "cursor-match-pending")
	f.createFriendshipIDs(t, f.userIDs[owner], pendingFriendID, models.FriendshipPending)

	blockedFriendID := f.createUser(t, "cursor-match-blocked")
	f.createFriendshipIDs(t, f.userIDs[owner], blockedFriendID, models.FriendshipAccepted)
	require.NoError(t, service.BlockUser(context.Background(), f.userIDs[owner], blockedFriendID))

	sort.Slice(expectedIDs, func(i, j int) bool { return expectedIDs[i] > expectedIDs[j] })
	_, err = service.ListFriends(context.Background(), f.userIDs[owner], 2, "cursor-match", nil)
	require.NoError(t, err)

	router := gin.New()
	routes.SetupRoutes(router, controllers.NewFriendshipController(service))
	var allFriends []models.Friend
	cursor := ""
	for pageNumber := 0; pageNumber < 10; pageNumber++ {
		params := url.Values{
			"limit": []string{"2"},
			"q":     []string{"cursor-match"},
		}
		if cursor != "" {
			params.Set("cursor", cursor)
		}
		request := httptest.NewRequest(http.MethodGet, "/friends?"+params.Encode(), nil)
		request.Header.Set("X-User-ID", f.userIDs[owner])
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, request)
		require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())

		var response struct {
			Data       []models.Friend `json:"data"`
			NextCursor string          `json:"next_cursor"`
		}
		require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
		require.LessOrEqual(t, len(response.Data), 2)
		allFriends = append(allFriends, response.Data...)
		if response.NextCursor == "" {
			cursor = ""
			break
		}
		cursor = response.NextCursor
	}

	require.Empty(t, cursor, "la última página debe devolver next_cursor vacío")
	require.Len(t, allFriends, len(expectedIDs))
	actualIDs := make([]string, 0, len(allFriends))
	for _, friend := range allFriends {
		actualIDs = append(actualIDs, string(friend.IDAmistad))
		require.Contains(t, friend.Username, "cursor-match")
		require.Equal(t, models.FriendshipAccepted, friend.Estado)
	}
	require.Equal(t, expectedIDs, actualIDs, "el recorrido debe mantener orden estable, sin saltos ni duplicados")

	emptyParams := url.Values{"limit": []string{"2"}, "q": []string{"no-such-friend"}}
	emptyRequest := httptest.NewRequest(http.MethodGet, "/friends?"+emptyParams.Encode(), nil)
	emptyRequest.Header.Set("X-User-ID", f.userIDs[owner])
	emptyResponse := httptest.NewRecorder()
	router.ServeHTTP(emptyResponse, emptyRequest)
	require.Equal(t, http.StatusOK, emptyResponse.Code)
	var emptyPage struct {
		Data       []models.Friend `json:"data"`
		NextCursor string          `json:"next_cursor"`
	}
	require.NoError(t, json.Unmarshal(emptyResponse.Body.Bytes(), &emptyPage))
	require.Empty(t, emptyPage.Data)
	require.Empty(t, emptyPage.NextCursor)

	noFriendsRequest := httptest.NewRequest(http.MethodGet, "/friends", nil)
	noFriendsRequest.Header.Set("X-User-ID", f.userIDs[3])
	noFriendsResponse := httptest.NewRecorder()
	router.ServeHTTP(noFriendsResponse, noFriendsRequest)
	require.Equal(t, http.StatusOK, noFriendsResponse.Code)
	var noFriendsPage struct {
		Data       []models.Friend `json:"data"`
		NextCursor string          `json:"next_cursor"`
	}
	require.NoError(t, json.Unmarshal(noFriendsResponse.Body.Bytes(), &noFriendsPage))
	require.Empty(t, noFriendsPage.Data)
	require.Empty(t, noFriendsPage.NextCursor)
}
