package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/meloop/media-service/models"
	"github.com/meloop/media-service/services"
	"github.com/meloop/services/common/httpresponse"
)

type mockMediaService struct {
	createPresignedUploadFunc func(ctx context.Context, userID string, req *models.PresignedUploadRequest) (*models.PresignedUploadResponse, error)
	verifyMediaFunc           func(ctx context.Context, objectKey string) (*models.VerifyMediaResponse, error)
}

func (m *mockMediaService) CreatePresignedUpload(ctx context.Context, userID string, req *models.PresignedUploadRequest) (*models.PresignedUploadResponse, error) {
	if m.createPresignedUploadFunc != nil {
		return m.createPresignedUploadFunc(ctx, userID, req)
	}
	return &models.PresignedUploadResponse{
		UploadURL: "http://localhost:9000/meloop-media/profiles/u1/med1.png",
		ObjectKey: "profiles/u1/med1.png",
		MediaID:   "med1",
		ExpiresIn: 900,
	}, nil
}

func (m *mockMediaService) VerifyMedia(ctx context.Context, objectKey string) (*models.VerifyMediaResponse, error) {
	if m.verifyMediaFunc != nil {
		return m.verifyMediaFunc(ctx, objectKey)
	}
	return &models.VerifyMediaResponse{
		Exists:      true,
		ObjectKey:   objectKey,
		ContentType: "image/png",
		SizeBytes:   2048,
	}, nil
}

func setupTestRouter(srv services.MediaService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	ctrl := NewMediaController(srv)

	r.POST("/media/upload", ctrl.GeneratePresignedUpload)
	r.POST("/v1/media/upload", ctrl.GeneratePresignedUpload)
	r.POST("/media/verify", ctrl.VerifyMedia)
	r.GET("/media/verify", ctrl.VerifyMedia)

	return r
}

func TestMediaController_GeneratePresignedUpload_200(t *testing.T) {
	srv := &mockMediaService{}
	router := setupTestRouter(srv)

	body, _ := json.Marshal(models.PresignedUploadRequest{
		ResourceType: "profile_photo",
		ContentType:  "image/png",
		FileName:     "avatar.png",
		FileSize:     1024,
	})

	req, _ := http.NewRequest(http.MethodPost, "/v1/media/upload", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", "usr-123")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("se esperaba código 200, se obtuvo: %d, body: %s", w.Code, w.Body.String())
	}

	var resp httpresponse.Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("error al deserializar respuesta: %v", err)
	}

	if !resp.Success {
		t.Error("se esperaba success=true")
	}
}

func TestMediaController_GeneratePresignedUpload_NoAutenticado_401(t *testing.T) {
	srv := &mockMediaService{}
	router := setupTestRouter(srv)

	body, _ := json.Marshal(models.PresignedUploadRequest{
		ResourceType: "profile_photo",
		ContentType:  "image/png",
		FileName:     "avatar.png",
		FileSize:     1024,
	})

	req, _ := http.NewRequest(http.MethodPost, "/v1/media/upload", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	// Sin cabecera X-User-ID

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("se esperaba código 401 para petición no autenticada, se obtuvo: %d", w.Code)
	}
}

func TestMediaController_GeneratePresignedUpload_Validacion_400(t *testing.T) {
	srv := &mockMediaService{
		createPresignedUploadFunc: func(ctx context.Context, userID string, req *models.PresignedUploadRequest) (*models.PresignedUploadResponse, error) {
			return nil, &services.ValidationError{
				Field:   "content_type",
				Issue:   "invalid_format",
				Message: "Formato no permitido",
			}
		},
	}
	router := setupTestRouter(srv)

	body, _ := json.Marshal(models.PresignedUploadRequest{
		ResourceType: "profile_photo",
		ContentType:  "application/pdf",
		FileSize:     1024,
	})

	req, _ := http.NewRequest(http.MethodPost, "/v1/media/upload", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", "usr-123")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("se esperaba código 400, se obtuvo: %d", w.Code)
	}

	var resp httpresponse.Response
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Error == nil || resp.Error.Code != httpresponse.ErrValidation {
		t.Errorf("error esperado VALIDATION_ERROR, se obtuvo: %+v", resp.Error)
	}
}

func TestMediaController_VerifyMedia_404(t *testing.T) {
	srv := &mockMediaService{
		verifyMediaFunc: func(ctx context.Context, objectKey string) (*models.VerifyMediaResponse, error) {
			return nil, services.ErrObjectNotFound
		},
	}
	router := setupTestRouter(srv)

	req, _ := http.NewRequest(http.MethodGet, "/media/verify?key=profiles/u1/inexistente.png", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("se esperaba código 404, se obtuvo: %d", w.Code)
	}
}

func TestMediaController_VerifyMedia_FalloAlmacenamiento_502(t *testing.T) {
	srv := &mockMediaService{
		verifyMediaFunc: func(ctx context.Context, objectKey string) (*models.VerifyMediaResponse, error) {
			return nil, fmtError()
		},
	}
	router := setupTestRouter(srv)

	req, _ := http.NewRequest(http.MethodGet, "/media/verify?key=profiles/u1/test.png", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadGateway {
		t.Fatalf("se esperaba código 502, se obtuvo: %d", w.Code)
	}
}

func fmtError() error {
	return services.ErrStorageUnreachable
}
