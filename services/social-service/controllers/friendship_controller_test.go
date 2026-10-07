package controllers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/meloop/social-service/controllers"
	"github.com/meloop/social-service/models"
	"github.com/meloop/social-service/routes"
	"github.com/meloop/social-service/services"
)

type mockFriendshipService struct {
	sendRequestFunc          func(ctx context.Context, senderID, receiverID string) (*models.FriendRequest, error)
	receivedRequestsFunc     func(ctx context.Context, userID string) ([]models.FriendRequest, error)
	sentRequestsFunc         func(ctx context.Context, userID string) ([]models.FriendRequest, error)
	acceptRequestFunc        func(ctx context.Context, requestID int, receiverID string) (*models.FriendRequest, error)
	rejectRequestFunc        func(ctx context.Context, requestID int, receiverID string) (*models.FriendRequest, error)
	cancelRequestFunc        func(ctx context.Context, requestID int, senderID string) (*models.FriendRequest, error)
	listFriendsFunc          func(ctx context.Context, userID string) ([]models.Friend, error)
	removeFriendFunc         func(ctx context.Context, userID, targetID string) error
	getFriendProfileFunc     func(ctx context.Context, userID, friendID string) (*models.FriendProfile, error)
	blockUserFunc            func(ctx context.Context, blockerID, blockedID string) error
	validateInteractionFunc  func(ctx context.Context, user1ID, user2ID string) error
	getFriendSuggestionsFunc func(ctx context.Context, userID string) ([]models.FriendSuggestion, error)
}

func (m *mockFriendshipService) GetFriendSuggestions(ctx context.Context, userID string) ([]models.FriendSuggestion, error) {
	if m.getFriendSuggestionsFunc != nil {
		return m.getFriendSuggestionsFunc(ctx, userID)
	}
	return []models.FriendSuggestion{
		{
			IDUsuario:     "user-sugg-1",
			Username:      "suggested_user",
			MatchScore:    75,
			MutualFriends: 2,
			Motivo:        "Ambos escuchan a The Strokes",
		},
	}, nil
}

func (m *mockFriendshipService) SendRequest(ctx context.Context, senderID, receiverID string) (*models.FriendRequest, error) {
	if m.sendRequestFunc != nil {
		return m.sendRequestFunc(ctx, senderID, receiverID)
	}
	return &models.FriendRequest{ID: 1, SenderID: senderID, ReceiverID: receiverID, Status: models.FriendshipPending}, nil
}

func (m *mockFriendshipService) ReceivedRequests(ctx context.Context, userID string) ([]models.FriendRequest, error) {
	if m.receivedRequestsFunc != nil {
		return m.receivedRequestsFunc(ctx, userID)
	}
	return []models.FriendRequest{}, nil
}

func (m *mockFriendshipService) SentRequests(ctx context.Context, userID string) ([]models.FriendRequest, error) {
	if m.sentRequestsFunc != nil {
		return m.sentRequestsFunc(ctx, userID)
	}
	return []models.FriendRequest{}, nil
}

func (m *mockFriendshipService) AcceptRequest(ctx context.Context, requestID int, receiverID string) (*models.FriendRequest, error) {
	if m.acceptRequestFunc != nil {
		return m.acceptRequestFunc(ctx, requestID, receiverID)
	}
	return &models.FriendRequest{ID: requestID, SenderID: "sender", ReceiverID: receiverID, Status: models.FriendshipAccepted}, nil
}

func (m *mockFriendshipService) RejectRequest(ctx context.Context, requestID int, receiverID string) (*models.FriendRequest, error) {
	if m.rejectRequestFunc != nil {
		return m.rejectRequestFunc(ctx, requestID, receiverID)
	}
	return &models.FriendRequest{ID: requestID, SenderID: "sender", ReceiverID: receiverID, Status: models.FriendshipRejected}, nil
}

func (m *mockFriendshipService) CancelRequest(ctx context.Context, requestID int, senderID string) (*models.FriendRequest, error) {
	if m.cancelRequestFunc != nil {
		return m.cancelRequestFunc(ctx, requestID, senderID)
	}
	return &models.FriendRequest{ID: requestID, SenderID: senderID, ReceiverID: "receiver", Status: models.FriendshipCanceled}, nil
}

func (m *mockFriendshipService) ListFriends(ctx context.Context, userID string) ([]models.Friend, error) {
	if m.listFriendsFunc != nil {
		return m.listFriendsFunc(ctx, userID)
	}
	return []models.Friend{
		{IDAmistad: 1, IDUsuario: "friend-1", Username: "user1", Estado: models.FriendshipAccepted},
	}, nil
}

func (m *mockFriendshipService) RemoveFriend(ctx context.Context, userID, targetID string) error {
	if m.removeFriendFunc != nil {
		return m.removeFriendFunc(ctx, userID, targetID)
	}
	return nil
}

func (m *mockFriendshipService) GetFriendProfile(ctx context.Context, userID, friendID string) (*models.FriendProfile, error) {
	if m.getFriendProfileFunc != nil {
		return m.getFriendProfileFunc(ctx, userID, friendID)
	}
	nivel := 3
	return &models.FriendProfile{
		IDUsuario:   friendID,
		Username:    "friend_user",
		Correo:      "friend@example.com",
		IDNivel:     &nivel,
		Experiencia: 1500,
		IDAmistad:   1,
	}, nil
}

func (m *mockFriendshipService) BlockUser(ctx context.Context, blockerID, blockedID string) error {
	if m.blockUserFunc != nil {
		return m.blockUserFunc(ctx, blockerID, blockedID)
	}
	return nil
}

func (m *mockFriendshipService) ValidateInteraction(ctx context.Context, user1ID, user2ID string) error {
	if m.validateInteractionFunc != nil {
		return m.validateInteractionFunc(ctx, user1ID, user2ID)
	}
	return nil
}

func setupTestRouter(svc services.FriendshipService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	ctrl := controllers.NewFriendshipController(svc)
	routes.SetupRoutes(router, ctrl)
	return router
}

// =========================================================================
// Tests HTTP para RF-12
// =========================================================================

func TestListFriends_HTTP_Exitoso(t *testing.T) {
	router := setupTestRouter(&mockFriendshipService{})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/friends", nil)
	req.Header.Set("X-User-ID", "usr-123")

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("se esperaba código 200 OK, se obtuvo %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("error deserializando json: %v", err)
	}
	if resp["success"] != true {
		t.Errorf("se esperaba success = true")
	}
	data, ok := resp["data"].([]interface{})
	if !ok || len(data) != 1 {
		t.Fatalf("se esperaba 1 amigo en la lista")
	}
}

func TestListFriends_HTTP_NoAutenticado(t *testing.T) {
	router := setupTestRouter(&mockFriendshipService{})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/friends", nil)

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("se esperaba 401 Unauthorized, se obtuvo %d", w.Code)
	}
}

func TestRemoveFriend_HTTP_Exitoso(t *testing.T) {
	router := setupTestRouter(&mockFriendshipService{})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodDelete, "/friends/1", nil)
	req.Header.Set("X-User-ID", "usr-123")

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("se esperaba 200 OK, se obtuvo %d", w.Code)
	}
}

func TestRemoveFriend_HTTP_NoEncontrado(t *testing.T) {
	mockSvc := &mockFriendshipService{
		removeFriendFunc: func(ctx context.Context, userID, targetID string) error {
			return services.ErrFriendshipNotFound
		},
	}
	router := setupTestRouter(mockSvc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodDelete, "/friends/999", nil)
	req.Header.Set("X-User-ID", "usr-123")

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("se esperaba 404 Not Found, se obtuvo %d", w.Code)
	}
}

func TestRemoveFriend_HTTP_NoAutorizado(t *testing.T) {
	mockSvc := &mockFriendshipService{
		removeFriendFunc: func(ctx context.Context, userID, targetID string) error {
			return services.ErrNotAuthorized
		},
	}
	router := setupTestRouter(mockSvc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodDelete, "/friends/12", nil)
	req.Header.Set("X-User-ID", "usr-intruder")

	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("se esperaba 403 Forbidden, se obtuvo %d", w.Code)
	}
}

func TestGetFriendProfile_HTTP_Exitoso(t *testing.T) {
	router := setupTestRouter(&mockFriendshipService{})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/friends/friend-1/profile", nil)
	req.Header.Set("X-User-ID", "usr-123")

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("se esperaba 200 OK, se obtuvo %d", w.Code)
	}

	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	if data["username"] != "friend_user" {
		t.Errorf("esperado username 'friend_user', obtenido %v", data["username"])
	}
}

func TestGetFriendProfile_HTTP_Bloqueado(t *testing.T) {
	mockSvc := &mockFriendshipService{
		getFriendProfileFunc: func(ctx context.Context, userID, friendID string) (*models.FriendProfile, error) {
			return nil, services.ErrBlocked
		},
	}
	router := setupTestRouter(mockSvc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/friends/friend-blocked/profile", nil)
	req.Header.Set("X-User-ID", "usr-123")

	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("se esperaba 403 Forbidden, se obtuvo %d", w.Code)
	}
}

func TestGetFriendProfile_HTTP_NoAmigos(t *testing.T) {
	mockSvc := &mockFriendshipService{
		getFriendProfileFunc: func(ctx context.Context, userID, friendID string) (*models.FriendProfile, error) {
			return nil, services.ErrNotFriends
		},
	}
	router := setupTestRouter(mockSvc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/friends/stranger/profile", nil)
	req.Header.Set("X-User-ID", "usr-123")

	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("se esperaba 403 Forbidden, se obtuvo %d", w.Code)
	}
}

// =========================================================================
// Tests HTTP para RF-13
// =========================================================================

func TestBlockUser_HTTP_Exitoso(t *testing.T) {
	router := setupTestRouter(&mockFriendshipService{})

	body := `{"id_usuario": "usr-bad"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/friends/blocks", bytes.NewBufferString(body))
	req.Header.Set("X-User-ID", "usr-123")
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("se esperaba 200 OK, se obtuvo %d", w.Code)
	}
}

func TestBlockUser_HTTP_Autobloqueo(t *testing.T) {
	mockSvc := &mockFriendshipService{
		blockUserFunc: func(ctx context.Context, blockerID, blockedID string) error {
			return services.ErrSelfBlock
		},
	}
	router := setupTestRouter(mockSvc)

	body := `{"id_usuario": "usr-123"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/friends/blocks", bytes.NewBufferString(body))
	req.Header.Set("X-User-ID", "usr-123")
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("se esperaba 400 Bad Request, se obtuvo %d", w.Code)
	}
}

func TestBlockUser_HTTP_YaBloqueado(t *testing.T) {
	mockSvc := &mockFriendshipService{
		blockUserFunc: func(ctx context.Context, blockerID, blockedID string) error {
			return services.ErrAlreadyBlocked
		},
	}
	router := setupTestRouter(mockSvc)

	body := `{"id_usuario": "usr-bad"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/friends/blocks", bytes.NewBufferString(body))
	req.Header.Set("X-User-ID", "usr-123")
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("se esperaba 409 Conflict, se obtuvo %d", w.Code)
	}
}

func TestValidateInteraction_HTTP_Permitido(t *testing.T) {
	router := setupTestRouter(&mockFriendshipService{})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/friends/validate-interaction?id_usuario=usr-ok", nil)
	req.Header.Set("X-User-ID", "usr-123")

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("se esperaba 200 OK, se obtuvo %d", w.Code)
	}

	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	if data["permitido"] != true {
		t.Errorf("se esperaba permitido = true, se obtuvo %v", data["permitido"])
	}
}

func TestValidateInteraction_HTTP_Bloqueado(t *testing.T) {
	mockSvc := &mockFriendshipService{
		validateInteractionFunc: func(ctx context.Context, user1ID, user2ID string) error {
			return services.ErrBlocked
		},
	}
	router := setupTestRouter(mockSvc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/friends/validate-interaction?id_usuario=usr-blocked", nil)
	req.Header.Set("X-User-ID", "usr-123")

	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("se esperaba 403 Forbidden, se obtuvo %d", w.Code)
	}
}

// =========================================================================
// Tests HTTP para RF-14: Sugerencias de Amigos
// =========================================================================

func TestGetFriendSuggestions_HTTP_Exitoso(t *testing.T) {
	router := setupTestRouter(&mockFriendshipService{})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/friends/suggestions", nil)
	req.Header.Set("X-User-ID", "usr-123")

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("se esperaba 200 OK, se obtuvo %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("error deserializando json: %v", err)
	}
	if resp["success"] != true {
		t.Errorf("se esperaba success = true")
	}
	data, ok := resp["data"].([]interface{})
	if !ok || len(data) != 1 {
		t.Fatalf("se esperaba 1 sugerencia, se obtuvo: %v", resp["data"])
	}
	sugg := data[0].(map[string]interface{})
	if sugg["id_usuario"] != "user-sugg-1" {
		t.Errorf("id_usuario inesperado: %v", sugg["id_usuario"])
	}
	if sugg["username"] != "suggested_user" {
		t.Errorf("username inesperado: %v", sugg["username"])
	}
	if sugg["porcentaje_compatibilidad"] != float64(75) {
		t.Errorf("porcentaje_compatibilidad inesperado: %v", sugg["porcentaje_compatibilidad"])
	}
	if sugg["motivo"] != "Ambos escuchan a The Strokes" {
		t.Errorf("motivo inesperado: %v", sugg["motivo"])
	}
}

func TestGetFriendSuggestions_HTTP_RutasAlternativas_Exitoso(t *testing.T) {
	router := setupTestRouter(&mockFriendshipService{})

	// Probar /friends/recommendations
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest(http.MethodGet, "/friends/recommendations", nil)
	req1.Header.Set("X-User-ID", "usr-123")
	router.ServeHTTP(w1, req1)
	if w1.Code != http.StatusOK {
		t.Fatalf("se esperaba 200 OK en /friends/recommendations, se obtuvo %d", w1.Code)
	}

	// Probar /suggestions
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest(http.MethodGet, "/suggestions", nil)
	req2.Header.Set("X-User-ID", "usr-123")
	router.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("se esperaba 200 OK en /suggestions, se obtuvo %d", w2.Code)
	}
}

func TestGetFriendSuggestions_HTTP_NoAutenticado(t *testing.T) {
	router := setupTestRouter(&mockFriendshipService{})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/friends/suggestions", nil)

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("se esperaba 401 Unauthorized, se obtuvo %d", w.Code)
	}

	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["success"] != false {
		t.Errorf("se esperaba success = false")
	}
	errBody := resp["error"].(map[string]interface{})
	if errBody["code"] != "UNAUTHORIZED" {
		t.Errorf("se esperaba código UNAUTHORIZED, se obtuvo %v", errBody["code"])
	}
}

func TestGetFriendSuggestions_HTTP_UsuarioInvalido(t *testing.T) {
	mockSvc := &mockFriendshipService{
		getFriendSuggestionsFunc: func(ctx context.Context, userID string) ([]models.FriendSuggestion, error) {
			return nil, services.ErrInvalidUser
		},
	}
	router := setupTestRouter(mockSvc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/friends/suggestions", nil)
	req.Header.Set("X-User-ID", "usr-123")

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("se esperaba 401 Unauthorized para ErrInvalidUser, se obtuvo %d", w.Code)
	}
}

func TestGetFriendSuggestions_HTTP_SinPreferenciasMusicales(t *testing.T) {
	mockSvc := &mockFriendshipService{
		getFriendSuggestionsFunc: func(ctx context.Context, userID string) ([]models.FriendSuggestion, error) {
			return nil, services.ErrNoMusicalPreferences
		},
	}
	router := setupTestRouter(mockSvc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/friends/suggestions", nil)
	req.Header.Set("X-User-ID", "usr-123")

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("se esperaba 400 Bad Request para ErrNoMusicalPreferences, se obtuvo %d", w.Code)
	}

	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["success"] != false {
		t.Errorf("se esperaba success = false")
	}
	errBody := resp["error"].(map[string]interface{})
	if errBody["code"] != "VALIDATION_ERROR" {
		t.Errorf("se esperaba código VALIDATION_ERROR, se obtuvo %v", errBody["code"])
	}
}

func TestGetFriendSuggestions_HTTP_SinCandidatosDisponibles(t *testing.T) {
	mockSvc := &mockFriendshipService{
		getFriendSuggestionsFunc: func(ctx context.Context, userID string) ([]models.FriendSuggestion, error) {
			return nil, services.ErrInsufficientCandidates
		},
	}
	router := setupTestRouter(mockSvc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/friends/suggestions", nil)
	req.Header.Set("X-User-ID", "usr-123")

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("se esperaba 404 Not Found para ErrInsufficientCandidates, se obtuvo %d", w.Code)
	}
}

func TestGetFriendSuggestions_HTTP_ListaVaciaRetorna200(t *testing.T) {
	mockSvc := &mockFriendshipService{
		getFriendSuggestionsFunc: func(ctx context.Context, userID string) ([]models.FriendSuggestion, error) {
			return []models.FriendSuggestion{}, nil
		},
	}
	router := setupTestRouter(mockSvc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/friends/suggestions", nil)
	req.Header.Set("X-User-ID", "usr-123")

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("se esperaba 200 OK con lista vacía, se obtuvo %d", w.Code)
	}

	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["success"] != true {
		t.Errorf("se esperaba success = true")
	}
	data, ok := resp["data"].([]interface{})
	if !ok || len(data) != 0 {
		t.Fatalf("se esperaba lista vacía, se obtuvo %v", resp["data"])
	}
}

func TestGetFriendSuggestions_HTTP_ErrorInterno_NoFugaDetalles(t *testing.T) {
	mockSvc := &mockFriendshipService{
		getFriendSuggestionsFunc: func(ctx context.Context, userID string) ([]models.FriendSuggestion, error) {
			return nil, errors.New("FATAL: postgres connection pool dropped: secret credentials leaked")
		},
	}
	router := setupTestRouter(mockSvc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/friends/suggestions", nil)
	req.Header.Set("X-User-ID", "usr-123")

	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("se esperaba 500 Internal Server Error, se obtuvo %d", w.Code)
	}

	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["success"] != false {
		t.Errorf("se esperaba success = false")
	}
	errBody := resp["error"].(map[string]interface{})
	if errBody["code"] != "INTERNAL_ERROR" {
		t.Errorf("se esperaba código INTERNAL_ERROR, se obtuvo %v", errBody["code"])
	}
	if errBody["message"] != "Ha ocurrido un error interno" {
		t.Errorf("se esperaba mensaje genérico, se obtuvo %v", errBody["message"])
	}
}

func TestSuggestionController_Directo(t *testing.T) {
	svc := &mockFriendshipService{}
	ctrl := controllers.NewSuggestionController(svc)
	router := gin.New()
	routes.SetupSuggestionRoutes(router, ctrl)

	// Prueba exitosa
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/friends/suggestions", nil)
	req.Header.Set("X-User-ID", "usr-123")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("se esperaba 200 OK en SuggestionController directo, se obtuvo %d", w.Code)
	}

	// Prueba no autenticado
	wNoAuth := httptest.NewRecorder()
	reqNoAuth, _ := http.NewRequest(http.MethodGet, "/friends/suggestions", nil)
	router.ServeHTTP(wNoAuth, reqNoAuth)

	if wNoAuth.Code != http.StatusUnauthorized {
		t.Fatalf("se esperaba 401 Unauthorized en SuggestionController directo, se obtuvo %d", wNoAuth.Code)
	}
}
