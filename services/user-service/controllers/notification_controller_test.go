package controllers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/meloop/user-service/controllers"
	"github.com/meloop/user-service/models"
	"github.com/meloop/user-service/services"
)

type mockNotificationConfigService struct {
	getSettingsFunc   func(ctx context.Context, userID string) ([]models.NotificationSettingResponse, error)
	updateSettingFunc func(ctx context.Context, userID string, tipo string, habilitada bool) (*models.NotificationSettingResponse, error)
	updateBatchFunc   func(ctx context.Context, userID string, items []models.NotificationSettingItem) ([]models.NotificationSettingResponse, error)
}

func (m *mockNotificationConfigService) GetSettings(ctx context.Context, userID string) ([]models.NotificationSettingResponse, error) {
	if m.getSettingsFunc != nil {
		return m.getSettingsFunc(ctx, userID)
	}
	var res []models.NotificationSettingResponse
	for _, t := range models.AllNotificationTypes {
		res = append(res, models.NotificationSettingResponse{
			TipoNotificacion: t,
			Habilitada:       true,
		})
	}
	return res, nil
}

func (m *mockNotificationConfigService) UpdateSetting(ctx context.Context, userID string, tipo string, habilitada bool) (*models.NotificationSettingResponse, error) {
	if m.updateSettingFunc != nil {
		return m.updateSettingFunc(ctx, userID, tipo, habilitada)
	}
	return &models.NotificationSettingResponse{
		TipoNotificacion: tipo,
		Habilitada:       habilitada,
	}, nil
}

func (m *mockNotificationConfigService) UpdateBatch(ctx context.Context, userID string, items []models.NotificationSettingItem) ([]models.NotificationSettingResponse, error) {
	if m.updateBatchFunc != nil {
		return m.updateBatchFunc(ctx, userID, items)
	}
	var res []models.NotificationSettingResponse
	for _, item := range items {
		h := true
		if item.Habilitada != nil {
			h = *item.Habilitada
		}
		res = append(res, models.NotificationSettingResponse{
			TipoNotificacion: item.TipoNotificacion,
			Habilitada:       h,
		})
	}
	return res, nil
}

func setupNotifTestRouter(srv services.NotificationConfigService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	ctrl := controllers.NewNotificationController(srv)

	protected := router.Group("/users")
	protected.Use(controllers.AuthRequired())
	{
		protected.GET("/me/notifications/settings", ctrl.GetNotificationSettings)
		protected.PATCH("/me/notifications/settings", ctrl.UpdateNotificationSettings)
		protected.PATCH("/me/notifications/settings/:tipo", ctrl.UpdateSingleSetting)
	}

	return router
}

func TestGetNotificationSettings_HTTP_Exitoso(t *testing.T) {
	router := setupNotifTestRouter(&mockNotificationConfigService{})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/users/me/notifications/settings", nil)
	req.Header.Set("X-User-ID", "usr-123")

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("se esperaba código 200 OK, se obtuvo %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("error al deserializar respuesta JSON: %v", err)
	}

	if resp["success"] != true {
		t.Errorf("se esperaba success = true, se obtuvo %v", resp["success"])
	}

	data, ok := resp["data"].([]interface{})
	if !ok {
		t.Fatal("se esperaba que 'data' sea un array")
	}

	if len(data) != len(models.AllNotificationTypes) {
		t.Errorf("se esperaban %d elementos, se obtuvieron %d", len(models.AllNotificationTypes), len(data))
	}
}

func TestGetNotificationSettings_HTTP_NoAutenticado(t *testing.T) {
	router := setupNotifTestRouter(&mockNotificationConfigService{})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/users/me/notifications/settings", nil)

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("se esperaba 401 Unauthorized, se obtuvo %d", w.Code)
	}
}

func TestUpdateNotificationSettings_HTTP_ArrayPayload(t *testing.T) {
	router := setupNotifTestRouter(&mockNotificationConfigService{})

	w := httptest.NewRecorder()
	body := `[{"tipo_notificacion": "LIKE", "habilitada": false}]`
	req, _ := http.NewRequest(http.MethodPatch, "/users/me/notifications/settings", bytes.NewBufferString(body))
	req.Header.Set("X-User-ID", "usr-123")
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("se esperaba 200 OK, se obtuvo %d", w.Code)
	}
}

func TestUpdateNotificationSettings_HTTP_SingleObjectPayload(t *testing.T) {
	router := setupNotifTestRouter(&mockNotificationConfigService{})

	w := httptest.NewRecorder()
	body := `{"tipo_notificacion": "LIKE", "habilitada": false}`
	req, _ := http.NewRequest(http.MethodPatch, "/users/me/notifications/settings", bytes.NewBufferString(body))
	req.Header.Set("X-User-ID", "usr-123")
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("se esperaba 200 OK, se obtuvo %d", w.Code)
	}

	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	if data["tipo_notificacion"] != "LIKE" || data["habilitada"] != false {
		t.Errorf("datos inesperados: %v", data)
	}
}

func TestUpdateNotificationSettings_HTTP_WrapperObjectPayload(t *testing.T) {
	router := setupNotifTestRouter(&mockNotificationConfigService{})

	w := httptest.NewRecorder()
	body := `{"configuraciones": [{"tipo_notificacion": "LIKE", "habilitada": false}]}`
	req, _ := http.NewRequest(http.MethodPatch, "/users/me/notifications/settings", bytes.NewBufferString(body))
	req.Header.Set("X-User-ID", "usr-123")
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("se esperaba 200 OK, se obtuvo %d", w.Code)
	}
}

func TestUpdateSingleSetting_HTTP_Param(t *testing.T) {
	router := setupNotifTestRouter(&mockNotificationConfigService{})

	w := httptest.NewRecorder()
	body := `{"habilitada": false}`
	req, _ := http.NewRequest(http.MethodPatch, "/users/me/notifications/settings/LIKE", bytes.NewBufferString(body))
	req.Header.Set("X-User-ID", "usr-123")
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("se esperaba 200 OK, se obtuvo %d", w.Code)
	}

	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	if data["tipo_notificacion"] != "LIKE" || data["habilitada"] != false {
		t.Errorf("datos inesperados: %v", data)
	}
}

func TestUpdateNotificationSettings_HTTP_Invalido(t *testing.T) {
	mockSvc := &mockNotificationConfigService{
		updateSettingFunc: func(ctx context.Context, userID string, tipo string, habilitada bool) (*models.NotificationSettingResponse, error) {
			return nil, &services.ValidationError{
				Field:   "tipo_notificacion",
				Issue:   "invalid_type",
				Message: "Tipo de notificación inválido",
			}
		},
	}
	router := setupNotifTestRouter(mockSvc)

	w := httptest.NewRecorder()
	body := `{"tipo_notificacion": "INVALID", "habilitada": true}`
	req, _ := http.NewRequest(http.MethodPatch, "/users/me/notifications/settings", bytes.NewBufferString(body))
	req.Header.Set("X-User-ID", "usr-123")
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("se esperaba 400 Bad Request, se obtuvo %d", w.Code)
	}
}
