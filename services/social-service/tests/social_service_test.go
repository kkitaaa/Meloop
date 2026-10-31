package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/meloop/social-service/controllers"
	"github.com/meloop/social-service/models"
	"github.com/meloop/social-service/routes"
	"github.com/meloop/social-service/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBlockedUsersCannotRequestOrAppearInSuggestions(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name        string
		blocker     int
		blocked     int
		requestFrom int
		requestTo   int
	}{
		{name: "A blocks B", blocker: 0, blocked: 1, requestFrom: 1, requestTo: 0},
		{name: "B blocks A", blocker: 1, blocked: 0, requestFrom: 0, requestTo: 1},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			f := setupFixture(t)
			service := f.databaseService()
			require.NoError(t, service.BlockUser(context.Background(), f.userIDs[test.blocker], f.userIDs[test.blocked]))

			_, err := service.SendRequest(context.Background(), f.userIDs[test.requestFrom], f.userIDs[test.requestTo])
			require.ErrorIs(t, err, services.ErrBlocked)

			suggestions, err := service.GetFriendSuggestions(context.Background(), f.userIDs[test.requestFrom])
			require.NoError(t, err)
			for _, suggestion := range suggestions {
				assert.NotEqual(t, f.userIDs[test.requestTo], suggestion.IDUsuario)
			}
		})
	}
}

func TestBlockingRemovesExistingFriendship(t *testing.T) {
	f := setupFixture(t)
	service := f.databaseService()
	f.createFriendship(t, 0, 1, models.FriendshipAccepted)

	require.NoError(t, service.BlockUser(context.Background(), f.userIDs[0], f.userIDs[1]))
	friends, err := service.ListFriends(context.Background(), f.userIDs[0])
	require.NoError(t, err)
	for _, friend := range friends {
		assert.NotEqual(t, f.userIDs[1], friend.IDUsuario)
	}

	var status string
	err = f.pool.QueryRow(context.Background(), `
		SELECT estado FROM amistad
		WHERE (id_usuario_1 = $1 AND id_usuario_2 = $2)
		   OR (id_usuario_1 = $2 AND id_usuario_2 = $1)`,
		f.userIDs[0], f.userIDs[1]).Scan(&status)
	require.NoError(t, err)
	assert.Equal(t, models.FriendshipDeleted, status)
}

func TestInvalidFriendRequestsReturnContractHTTPStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name       string
		wantStatus int
		prepare    func(*testing.T, *fixture, services.FriendshipService)
		sender     int
		receiver   int
	}{
		{name: "self request", wantStatus: http.StatusBadRequest, sender: 0, receiver: 0},
		{
			name: "duplicate request", wantStatus: http.StatusConflict, sender: 0, receiver: 1,
			prepare: func(t *testing.T, f *fixture, _ services.FriendshipService) {
				f.createFriendship(t, 0, 1, models.FriendshipPending)
			},
		},
		{
			name: "blocked user", wantStatus: http.StatusForbidden, sender: 1, receiver: 0,
			prepare: func(t *testing.T, f *fixture, service services.FriendshipService) {
				require.NoError(t, service.BlockUser(context.Background(), f.userIDs[0], f.userIDs[1]))
			},
		},
		{
			name: "existing friendship", wantStatus: http.StatusConflict, sender: 0, receiver: 1,
			prepare: func(t *testing.T, f *fixture, _ services.FriendshipService) {
				f.createFriendship(t, 0, 1, models.FriendshipAccepted)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			f := setupFixture(t)
			service := f.databaseService()
			if test.prepare != nil {
				test.prepare(t, f, service)
			}

			recorder := httptest.NewRecorder()
			payload, err := json.Marshal(models.SendFriendRequest{ReceiverID: f.userIDs[test.receiver]})
			require.NoError(t, err)
			request := httptest.NewRequest(http.MethodPost, "/friends/requests", strings.NewReader(string(payload)))
			request.Header.Set("X-User-ID", f.userIDs[test.sender])
			router := gin.New()
			routes.SetupRoutes(router, controllers.NewFriendshipController(service))
			router.ServeHTTP(recorder, request)
			assert.Equal(t, test.wantStatus, recorder.Code)
		})
	}
}

func TestMutualFriendCounts(t *testing.T) {
	tests := []struct {
		name  string
		count int
	}{
		{name: "zero mutual friends", count: 0},
		{name: "one mutual friend", count: 1},
		{name: "multiple mutual friends", count: 2},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			f := setupFixture(t)
			service := f.databaseService()
			for i := 0; i < test.count; i++ {
				commonFriend := i + 2
				f.createFriendship(t, 0, commonFriend, models.FriendshipAccepted)
				f.createFriendship(t, 1, commonFriend, models.FriendshipAccepted)
			}

			suggestions, err := service.GetFriendSuggestions(context.Background(), f.userIDs[0])
			require.NoError(t, err)
			var candidate *models.FriendSuggestion
			for i := range suggestions {
				if suggestions[i].IDUsuario == f.userIDs[1] {
					candidate = &suggestions[i]
					break
				}
			}
			require.NotNil(t, candidate)
			assert.Equal(t, test.count, candidate.MutualFriends)
		})
	}
}

func TestBlockedFriendProfileReturnsNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	f := setupFixture(t)
	service := f.databaseService()
	f.createFriendship(t, 0, 1, models.FriendshipAccepted)
	require.NoError(t, service.BlockUser(context.Background(), f.userIDs[0], f.userIDs[1]))

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/friends/%s/profile", f.userIDs[1]), nil)
	request.Header.Set("X-User-ID", f.userIDs[0])
	router := gin.New()
	routes.SetupRoutes(router, controllers.NewFriendshipController(service))
	router.ServeHTTP(recorder, request)
	assert.Equal(t, http.StatusNotFound, recorder.Code)
}
