package services

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/meloop/media-service/models"
)

type mockStorageRepository struct {
	presignedUploadURLFunc func(ctx context.Context, objectKey, contentType string, expires time.Duration) (string, error)
	objectExistsFunc       func(ctx context.Context, objectKey string) (bool, int64, string, error)
	deleteObjectFunc       func(ctx context.Context, objectKey string) error
}

func (m *mockStorageRepository) GeneratePresignedUploadURL(ctx context.Context, objectKey, contentType string, expires time.Duration) (string, error) {
	if m.presignedUploadURLFunc != nil {
		return m.presignedUploadURLFunc(ctx, objectKey, contentType, expires)
	}
	return "http://localhost:9000/meloop-media/" + objectKey + "?signature=test", nil
}

func (m *mockStorageRepository) ObjectExists(ctx context.Context, objectKey string) (bool, int64, string, error) {
	if m.objectExistsFunc != nil {
		return m.objectExistsFunc(ctx, objectKey)
	}
	return true, 1024, "image/png", nil
}

func (m *mockStorageRepository) DeleteObject(ctx context.Context, objectKey string) error {
	if m.deleteObjectFunc != nil {
		return m.deleteObjectFunc(ctx, objectKey)
	}
	return nil
}

func TestCreatePresignedUpload_Exitoso_FotoPerfil(t *testing.T) {
	repo := &mockStorageRepository{}
	srv := NewMediaService(repo)

	req := &models.PresignedUploadRequest{
		ResourceType: "profile_photo",
		ContentType:  "image/png",
		FileName:     "avatar.png",
		FileSize:     1024 * 500, // 500 KB
	}

	resp, err := srv.CreatePresignedUpload(context.Background(), "usr-123", req)
	if err != nil {
		t.Fatalf("se esperaba éxito pero ocurrió error: %v", err)
	}

	if resp == nil {
		t.Fatal("se esperaba respuesta no nula")
	}

	if !strings.HasPrefix(resp.ObjectKey, "profiles/usr-123/") {
		t.Errorf("clave de objeto incorrecta: %s, esperaba prefijo 'profiles/usr-123/'", resp.ObjectKey)
	}

	if !strings.HasSuffix(resp.ObjectKey, ".png") {
		t.Errorf("extensión de objeto incorrecta: %s, esperaba sufijo '.png'", resp.ObjectKey)
	}

	if resp.UploadURL == "" {
		t.Error("se esperaba una URL prefirmada de subida no vacía")
	}

	if resp.ExpiresIn != 900 {
		t.Errorf("se esperaba tiempo de expiración de 900s, se obtuvo: %d", resp.ExpiresIn)
	}
}

func TestCreatePresignedUpload_Exitoso_Banner(t *testing.T) {
	repo := &mockStorageRepository{}
	srv := NewMediaService(repo)

	req := &models.PresignedUploadRequest{
		TipoRecurso: "banner",
		TipoMIME:    "image/jpeg",
		Nombre:      "banner.jpg",
		TamanoBytes: 1024 * 1024 * 2, // 2 MB
	}

	resp, err := srv.CreatePresignedUpload(context.Background(), "usr-456", req)
	if err != nil {
		t.Fatalf("se esperaba éxito pero ocurrió error: %v", err)
	}

	if !strings.HasPrefix(resp.ObjectKey, "banners/usr-456/") {
		t.Errorf("clave de objeto incorrecta: %s, esperaba prefijo 'banners/usr-456/'", resp.ObjectKey)
	}

	if !strings.HasSuffix(resp.ObjectKey, ".jpg") {
		t.Errorf("extensión de objeto incorrecta: %s, esperaba sufijo '.jpg'", resp.ObjectKey)
	}
}

func TestCreatePresignedUpload_RechazoFormatoInvalido(t *testing.T) {
	repo := &mockStorageRepository{}
	srv := NewMediaService(repo)

	casos := []struct {
		nombre       string
		resourceType string
		contentType  string
	}{
		{"GIF no permitido", "profile_photo", "image/gif"},
		{"PDF no permitido", "profile_photo", "application/pdf"},
		{"SVG no permitido", "banner", "image/svg+xml"},
		{"Video en lugar de imagen", "profile_photo", "video/mp4"},
		{"Texto plano", "banner", "text/plain"},
		{"Audio no permitido en perfil", "profile_photo", "audio/mpeg"},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			req := &models.PresignedUploadRequest{
				ResourceType: c.resourceType,
				ContentType:  c.contentType,
				FileSize:     1024 * 100,
			}

			_, err := srv.CreatePresignedUpload(context.Background(), "usr-1", req)
			if err == nil {
				t.Fatalf("se esperaba error de validación para formato %s", c.contentType)
			}

			var valErr *ValidationError
			if !errors.As(err, &valErr) {
				t.Fatalf("se esperaba ValidationError, se obtuvo: %T", err)
			}

			if valErr.Issue != "invalid_format" {
				t.Errorf("se esperaba issue 'invalid_format', se obtuvo: '%s'", valErr.Issue)
			}
		})
	}
}

func TestCreatePresignedUpload_RechazoTamanoExcedido(t *testing.T) {
	repo := &mockStorageRepository{}
	srv := NewMediaService(repo)

	t.Run("Imagen superior a 5MB rechazada", func(t *testing.T) {
		req := &models.PresignedUploadRequest{
			ResourceType: "profile_photo",
			ContentType:  "image/png",
			FileSize:     (5 * 1024 * 1024) + 1, // 5MB + 1 byte
		}

		_, err := srv.CreatePresignedUpload(context.Background(), "usr-1", req)
		if err == nil {
			t.Fatal("se esperaba error por superar 5MB")
		}

		var valErr *ValidationError
		if !errors.As(err, &valErr) || valErr.Issue != "size_exceeded" {
			t.Fatalf("se esperaba error size_exceeded, se obtuvo: %v", err)
		}
	})

	t.Run("Tamaño 0 bytes rechazado", func(t *testing.T) {
		req := &models.PresignedUploadRequest{
			ResourceType: "profile_photo",
			ContentType:  "image/jpeg",
			FileSize:     0,
		}

		_, err := srv.CreatePresignedUpload(context.Background(), "usr-1", req)
		if err == nil {
			t.Fatal("se esperaba error por tamaño 0")
		}
	})
}

func TestCreatePresignedUpload_RechazoSinAutenticacion(t *testing.T) {
	repo := &mockStorageRepository{}
	srv := NewMediaService(repo)

	req := &models.PresignedUploadRequest{
		ResourceType: "profile_photo",
		ContentType:  "image/png",
		FileSize:     1024,
	}

	_, err := srv.CreatePresignedUpload(context.Background(), "", req)
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("se esperaba ErrUnauthorized, se obtuvo: %v", err)
	}
}

func TestVerifyMedia(t *testing.T) {
	t.Run("Objeto existente verificado exitosamente", func(t *testing.T) {
		repo := &mockStorageRepository{
			objectExistsFunc: func(ctx context.Context, objectKey string) (bool, int64, string, error) {
				return true, 2048, "image/jpeg", nil
			},
		}
		srv := NewMediaService(repo)

		resp, err := srv.VerifyMedia(context.Background(), "profiles/usr-1/abc.jpg")
		if err != nil {
			t.Fatalf("error inesperado: %v", err)
		}
		if !resp.Exists || resp.SizeBytes != 2048 {
			t.Errorf("datos de verificación inesperados: %+v", resp)
		}
	})

	t.Run("Objeto no existente devuelve ErrObjectNotFound", func(t *testing.T) {
		repo := &mockStorageRepository{
			objectExistsFunc: func(ctx context.Context, objectKey string) (bool, int64, string, error) {
				return false, 0, "", nil
			},
		}
		srv := NewMediaService(repo)

		_, err := srv.VerifyMedia(context.Background(), "profiles/usr-1/inexistente.jpg")
		if !errors.Is(err, ErrObjectNotFound) {
			t.Fatalf("se esperaba ErrObjectNotFound, se obtuvo: %v", err)
		}
	})

	t.Run("Fallo de almacenamiento devuelve ErrStorageUnreachable", func(t *testing.T) {
		repo := &mockStorageRepository{
			objectExistsFunc: func(ctx context.Context, objectKey string) (bool, int64, string, error) {
				return false, 0, "", errors.New("connection refused to minio:9000")
			},
		}
		srv := NewMediaService(repo)

		_, err := srv.VerifyMedia(context.Background(), "profiles/usr-1/abc.jpg")
		if !errors.Is(err, ErrStorageUnreachable) {
			t.Fatalf("se esperaba ErrStorageUnreachable, se obtuvo: %v", err)
		}
	})
}
