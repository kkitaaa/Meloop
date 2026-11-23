package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/meloop/social-service/client"
	"github.com/meloop/social-service/controllers"
	"github.com/meloop/social-service/models"
	"github.com/meloop/social-service/routes"
	"github.com/meloop/social-service/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testSessionValidator struct{}

func (testSessionValidator) ValidateToken(_ context.Context, token string) (string, error) {
	if token == "invalid" {
		return "", client.ErrInvalidSession
	}
	return token, nil
}

func setAuthenticatedUser(request *http.Request, userID string) {
	request.Header.Set("Authorization", "Bearer "+userID)
}

type recordingBlockPublisher struct {
	requests     []models.FriendRequest
	err          error
	afterPublish func(models.FriendRequest)
}

func (p *recordingBlockPublisher) PublishFriendAccepted(context.Context, *models.FriendRequest) error {
	return nil
}

func (p *recordingBlockPublisher) PublishFriendRequestCancelled(_ context.Context, request *models.FriendRequest) error {
	p.requests = append(p.requests, *request)
	if p.afterPublish != nil {
		p.afterPublish(*request)
	}
	return p.err
}

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

func TestBlockingCancelsPendingRequestsInBothDirectionsAndIsIdempotent(t *testing.T) {
	f := setupFixture(t)
	requestAB := f.createFriendship(t, 0, 1, models.FriendshipPending)
	requestBA := f.createFriendship(t, 1, 0, models.FriendshipPending)
	accepted := f.createFriendship(t, 0, 1, models.FriendshipAccepted)
	publisher := &recordingBlockPublisher{
		afterPublish: func(request models.FriendRequest) {
			var status string
			require.NoError(t, f.pool.QueryRow(context.Background(), `
				SELECT estado FROM amistad WHERE id_amistad = $1`, string(request.ID)).Scan(&status))
			require.Equal(t, models.FriendshipCanceled, status, "event must be published after commit")
		},
	}
	service := f.databaseService(publisher)

	require.NoError(t, service.BlockUser(context.Background(), f.userIDs[0], f.userIDs[1]))
	require.Len(t, publisher.requests, 2)
	publishedIDs := map[models.FriendshipID]bool{}
	for _, request := range publisher.requests {
		publishedIDs[request.ID] = true
		require.Equal(t, models.FriendshipCanceled, request.Status)
	}
	require.True(t, publishedIDs[requestAB])
	require.True(t, publishedIDs[requestBA])

	for _, userIndex := range []int{0, 1} {
		sent, err := service.SentRequests(context.Background(), f.userIDs[userIndex])
		require.NoError(t, err)
		received, err := service.ReceivedRequests(context.Background(), f.userIDs[userIndex])
		require.NoError(t, err)
		require.Empty(t, sent)
		require.Empty(t, received)
	}

	var acceptedStatus string
	require.NoError(t, f.pool.QueryRow(context.Background(), `
		SELECT estado FROM amistad WHERE id_amistad = $1`, string(accepted)).Scan(&acceptedStatus))
	require.Equal(t, models.FriendshipDeleted, acceptedStatus)

	require.NoError(t, service.BlockUser(context.Background(), f.userIDs[0], f.userIDs[1]))
	require.Len(t, publisher.requests, 2, "repeat block must not publish duplicate cancellation events")
	var blockCount int
	require.NoError(t, f.pool.QueryRow(context.Background(), `
		SELECT COUNT(*) FROM bloqueo
		WHERE id_usuario_bloqueador = $1 AND id_usuario_bloqueado = $2`,
		f.userIDs[0], f.userIDs[1]).Scan(&blockCount))
	require.Equal(t, 1, blockCount)
}

func TestBlockingRemainsDirectionalForIdempotency(t *testing.T) {
	f := setupFixture(t)
	service := f.databaseService()

	require.NoError(t, service.BlockUser(context.Background(), f.userIDs[1], f.userIDs[0]))
	require.NoError(t, service.BlockUser(context.Background(), f.userIDs[0], f.userIDs[1]))

	var blockCount int
	require.NoError(t, f.pool.QueryRow(context.Background(), `
		SELECT COUNT(*) FROM bloqueo
		WHERE (id_usuario_bloqueador = $1 AND id_usuario_bloqueado = $2)
		   OR (id_usuario_bloqueador = $2 AND id_usuario_bloqueado = $1)`,
		f.userIDs[0], f.userIDs[1]).Scan(&blockCount))
	require.Equal(t, 2, blockCount)
}

func TestBlockUserPublisherFailureDoesNotRollbackCommittedBlock(t *testing.T) {
	f := setupFixture(t)
	requestID := f.createFriendship(t, 0, 1, models.FriendshipPending)
	publisher := &recordingBlockPublisher{err: fmt.Errorf("broker unavailable")}
	service := f.databaseService(publisher)

	require.NoError(t, service.BlockUser(context.Background(), f.userIDs[0], f.userIDs[1]))
	require.Len(t, publisher.requests, 1)

	var blockCount int
	var status string
	require.NoError(t, f.pool.QueryRow(context.Background(), `
		SELECT COUNT(*) FROM bloqueo
		WHERE id_usuario_bloqueador = $1 AND id_usuario_bloqueado = $2`,
		f.userIDs[0], f.userIDs[1]).Scan(&blockCount))
	require.Equal(t, 1, blockCount)
	require.NoError(t, f.pool.QueryRow(context.Background(), `
		SELECT estado FROM amistad WHERE id_amistad = $1`, string(requestID)).Scan(&status))
	require.Equal(t, models.FriendshipCanceled, status)
}

func TestBlockUserRollsBackWhenCancellationUpdateFails(t *testing.T) {
	f := setupFixture(t)
	pendingID := f.createFriendship(t, 0, 1, models.FriendshipPending)
	acceptedID := f.createFriendship(t, 0, 1, models.FriendshipAccepted)
	suffix := strings.ReplaceAll(f.userIDs[0], "-", "_")
	functionName := "social_test_fail_cancel_" + suffix
	triggerName := "social_test_fail_cancel_trigger_" + suffix

	_, err := f.pool.Exec(context.Background(), fmt.Sprintf(`
		CREATE FUNCTION %s() RETURNS trigger LANGUAGE plpgsql AS $$
		BEGIN
			IF NEW.estado = 'CANCELADA' AND TG_ARGV[0] IN (NEW.id_usuario_1, NEW.id_usuario_2) THEN
				RAISE EXCEPTION 'injected cancellation failure';
			END IF;
			RETURN NEW;
		END;
		$$`, functionName))
	require.NoError(t, err)
	t.Cleanup(func() {
		_, dropTriggerErr := f.pool.Exec(context.Background(), fmt.Sprintf(`DROP TRIGGER IF EXISTS %s ON amistad`, triggerName))
		_, dropFunctionErr := f.pool.Exec(context.Background(), fmt.Sprintf(`DROP FUNCTION IF EXISTS %s()`, functionName))
		require.NoError(t, dropTriggerErr)
		require.NoError(t, dropFunctionErr)
	})
	_, err = f.pool.Exec(context.Background(), fmt.Sprintf(`
		CREATE TRIGGER %s BEFORE UPDATE OF estado ON amistad
		FOR EACH ROW EXECUTE FUNCTION %s('%s')`, triggerName, functionName, f.userIDs[0]))
	require.NoError(t, err)

	err = f.databaseService().BlockUser(context.Background(), f.userIDs[0], f.userIDs[1])
	require.Error(t, err)

	var blockCount int
	require.NoError(t, f.pool.QueryRow(context.Background(), `
		SELECT COUNT(*) FROM bloqueo
		WHERE id_usuario_bloqueador = $1 AND id_usuario_bloqueado = $2`,
		f.userIDs[0], f.userIDs[1]).Scan(&blockCount))
	require.Zero(t, blockCount)
	for id, expected := range map[models.FriendshipID]string{
		pendingID:  models.FriendshipPending,
		acceptedID: models.FriendshipAccepted,
	} {
		var status string
		require.NoError(t, f.pool.QueryRow(context.Background(), `
			SELECT estado FROM amistad WHERE id_amistad = $1`, string(id)).Scan(&status))
		require.Equal(t, expected, status)
	}
}

func TestConcurrentRequestAndBlockLeaveNoPendingRelationship(t *testing.T) {
	f := setupFixture(t)
	service := f.databaseService()
	start := make(chan struct{})
	sendResult := make(chan error, 1)
	blockResult := make(chan error, 1)
	var wait sync.WaitGroup
	wait.Add(2)

	go func() {
		defer wait.Done()
		<-start
		_, err := service.SendRequest(context.Background(), f.userIDs[0], f.userIDs[1])
		sendResult <- err
	}()
	go func() {
		defer wait.Done()
		<-start
		blockResult <- service.BlockUser(context.Background(), f.userIDs[0], f.userIDs[1])
	}()

	close(start)
	wait.Wait()
	sendErr := <-sendResult
	require.NoError(t, <-blockResult)
	if sendErr != nil {
		require.ErrorIs(t, sendErr, services.ErrBlocked)
	}

	var pendingCount int
	require.NoError(t, f.pool.QueryRow(context.Background(), `
		SELECT COUNT(*) FROM amistad
		WHERE ((id_usuario_1 = $1 AND id_usuario_2 = $2)
		    OR (id_usuario_1 = $2 AND id_usuario_2 = $1))
		  AND estado = 'PENDIENTE'`, f.userIDs[0], f.userIDs[1]).Scan(&pendingCount))
	require.Zero(t, pendingCount)
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
			setAuthenticatedUser(request, f.userIDs[test.sender])
			router := gin.New()
			routes.SetupRoutes(router, controllers.NewFriendshipController(service), testSessionValidator{})
			router.ServeHTTP(recorder, request)
			assert.Equal(t, test.wantStatus, recorder.Code)
		})
	}
}

func TestFriendRequestsPersistAndProcessTextIDsThroughHTTP(t *testing.T) {
	gin.SetMode(gin.TestMode)
	f := setupFixture(t)
	service := f.databaseService()
	router := gin.New()
	routes.SetupRoutes(router, controllers.NewFriendshipController(service), testSessionValidator{})

	doRequest := func(method, path, userID, body string) *httptest.ResponseRecorder {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(method, path, strings.NewReader(body))
		setAuthenticatedUser(request, userID)
		if body != "" {
			request.Header.Set("Content-Type", "application/json")
		}
		router.ServeHTTP(recorder, request)
		return recorder
	}
	readRequest := func(t *testing.T, response *httptest.ResponseRecorder) models.FriendRequest {
		t.Helper()
		require.Equal(t, http.StatusOK, response.Code, response.Body.String())
		var envelope struct {
			Data models.FriendRequest `json:"data"`
		}
		require.NoError(t, json.Unmarshal(response.Body.Bytes(), &envelope))
		return envelope.Data
	}
	sendRequest := func(t *testing.T, sender, receiver int) models.FriendRequest {
		t.Helper()
		body, err := json.Marshal(models.SendFriendRequest{ReceiverID: f.userIDs[receiver]})
		require.NoError(t, err)
		response := doRequest(http.MethodPost, "/friends/requests", f.userIDs[sender], string(body))
		require.Equal(t, http.StatusCreated, response.Code, response.Body.String())
		var envelope struct {
			Data models.FriendRequest `json:"data"`
		}
		require.NoError(t, json.Unmarshal(response.Body.Bytes(), &envelope))
		require.Len(t, string(envelope.Data.ID), 36)
		return envelope.Data
	}
	requireStatus := func(response *httptest.ResponseRecorder, status int) {
		t.Helper()
		require.Equal(t, status, response.Code, response.Body.String())
	}

	accepted := sendRequest(t, 0, 1)
	var storedID string
	var storedStatus string
	require.NoError(t, f.pool.QueryRow(context.Background(), `
		SELECT id_amistad, estado FROM amistad WHERE id_amistad = $1`, string(accepted.ID)).Scan(
		&storedID, &storedStatus,
	))
	require.Equal(t, string(accepted.ID), storedID)
	require.Equal(t, models.FriendshipPending, storedStatus)

	var sent []models.FriendRequest
	response := doRequest(http.MethodGet, "/friends/requests/sent", f.userIDs[0], "")
	requireStatus(response, http.StatusOK)
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &struct {
		Data *[]models.FriendRequest `json:"data"`
	}{Data: &sent}))
	require.Len(t, sent, 1)
	require.Equal(t, accepted.ID, sent[0].ID)

	var received []models.FriendRequest
	response = doRequest(http.MethodGet, "/friends/requests/received", f.userIDs[1], "")
	requireStatus(response, http.StatusOK)
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &struct {
		Data *[]models.FriendRequest `json:"data"`
	}{Data: &received}))
	require.Len(t, received, 1)
	require.Equal(t, accepted.ID, received[0].ID)

	response = doRequest(http.MethodPost, "/friends/requests/"+string(accepted.ID)+"/accept", f.userIDs[1], "")
	result := readRequest(t, response)
	require.Equal(t, accepted.ID, result.ID)
	require.Equal(t, models.FriendshipAccepted, result.Status)
	require.NoError(t, f.pool.QueryRow(context.Background(), `
		SELECT estado FROM amistad WHERE id_amistad = $1`, string(accepted.ID)).Scan(&storedStatus))
	require.Equal(t, models.FriendshipAccepted, storedStatus)

	for _, user := range []int{0, 1} {
		response = doRequest(http.MethodGet, "/friends", f.userIDs[user], "")
		requireStatus(response, http.StatusOK)
		var envelope struct {
			Data []models.Friend `json:"data"`
		}
		require.NoError(t, json.Unmarshal(response.Body.Bytes(), &envelope))
		require.Len(t, envelope.Data, 1)
		require.Equal(t, accepted.ID, envelope.Data[0].IDAmistad)
		expectedFriend := f.userIDs[1-user]
		require.Equal(t, expectedFriend, envelope.Data[0].IDUsuario)

		response = doRequest(http.MethodGet, "/friends/"+string(accepted.ID)+"/profile", f.userIDs[user], "")
		requireStatus(response, http.StatusOK)
		var profileEnvelope struct {
			Data models.FriendProfile `json:"data"`
		}
		require.NoError(t, json.Unmarshal(response.Body.Bytes(), &profileEnvelope))
		require.Equal(t, expectedFriend, profileEnvelope.Data.IDUsuario)
		require.Equal(t, accepted.ID, profileEnvelope.Data.IDAmistad)
	}
	var acceptedRows int
	require.NoError(t, f.pool.QueryRow(context.Background(), `
		SELECT COUNT(*) FROM amistad
		WHERE ((id_usuario_1 = $1 AND id_usuario_2 = $2)
		    OR (id_usuario_1 = $2 AND id_usuario_2 = $1))
		  AND estado = 'ACEPTADA'`, f.userIDs[0], f.userIDs[1]).Scan(&acceptedRows))
	require.Equal(t, 1, acceptedRows)

	rejected := sendRequest(t, 1, 2)
	response = doRequest(http.MethodPost, "/friends/requests/"+string(rejected.ID)+"/reject", f.userIDs[2], "")
	result = readRequest(t, response)
	require.Equal(t, models.FriendshipRejected, result.Status)

	canceled := sendRequest(t, 2, 3)
	response = doRequest(http.MethodPost, "/friends/requests/"+string(canceled.ID)+"/cancel", f.userIDs[2], "")
	result = readRequest(t, response)
	require.Equal(t, models.FriendshipCanceled, result.Status)
	for _, test := range []struct {
		request models.FriendRequest
		status  string
	}{
		{request: rejected, status: models.FriendshipRejected},
		{request: canceled, status: models.FriendshipCanceled},
	} {
		require.NoError(t, f.pool.QueryRow(context.Background(), `
			SELECT estado FROM amistad WHERE id_amistad = $1`, string(test.request.ID)).Scan(&storedStatus))
		require.Equal(t, test.status, storedStatus)
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
	setAuthenticatedUser(request, f.userIDs[0])
	router := gin.New()
	routes.SetupRoutes(router, controllers.NewFriendshipController(service), testSessionValidator{})
	router.ServeHTTP(recorder, request)
	assert.Equal(t, http.StatusNotFound, recorder.Code)
}
