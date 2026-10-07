package services_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/meloop/user-service/models"
	"github.com/meloop/user-service/services"
)

type mockProfileRepository struct {
	getProfileFunc    func(ctx context.Context, userID string) (*models.ProfileResponse, error)
	updateProfileFunc func(ctx context.Context, userID string, profile *models.Profile) (*models.ProfileResponse, error)
	updatedRecord     *models.Profile
}

func (m *mockProfileRepository) GetProfile(ctx context.Context, userID string) (*models.ProfileResponse, error) {
	if m.getProfileFunc != nil {
		return m.getProfileFunc(ctx, userID)
	}
	return &models.ProfileResponse{
		IDUsuario:  userID,
		Username:   "usuario_test",
		Biografia:  "Bio inicial",
		FotoPerfil: "profiles/usr-1/foto_antigua.png",
		Banner:     "banners/usr-1/banner_antiguo.png",
		Tema:       "default",
		Colores:    map[string]interface{}{"primario": "#1ABC9C"},
		InformacionMusical: &models.MusicalPreferences{
			Generos:  []string{"rock"},
			Artistas: []string{"queen"},
		},
	}, nil
}

func (m *mockProfileRepository) UpdateProfile(ctx context.Context, userID string, profile *models.Profile) (*models.ProfileResponse, error) {
	m.updatedRecord = profile
	if m.updateProfileFunc != nil {
		return m.updateProfileFunc(ctx, userID, profile)
	}
	return &models.ProfileResponse{
		IDUsuario:          userID,
		Username:           profile.Username,
		Biografia:          profile.Biografia,
		FotoPerfil:         profile.FotoPerfil,
		Banner:             profile.Banner,
		Tema:               profile.Tema,
		Colores:            profile.Colores,
		InformacionMusical: profile.InformacionMusical,
	}, nil
}

type mockMediaVerifier struct {
	verifyMediaFunc func(ctx context.Context, objectKey string) (bool, error)
	verifiedKeys    []string
}

func (m *mockMediaVerifier) VerifyMedia(ctx context.Context, objectKey string) (bool, error) {
	m.verifiedKeys = append(m.verifiedKeys, objectKey)
	if m.verifyMediaFunc != nil {
		return m.verifyMediaFunc(ctx, objectKey)
	}
	return true, nil
}

func TestProfileService_UpdateProfile_Exitoso(t *testing.T) {
	userRepo := &mockUserRepository{
		users: map[string]*models.Usuario{
			"usr-123": {IDUsuario: "usr-123", Username: "carlos_music", Correo: "carlos@test.com"},
		},
	}
	profileRepo := &mockProfileRepository{}
	mediaVerifier := &mockMediaVerifier{}

	srv := services.NewProfileService(profileRepo, userRepo, mediaVerifier)

	newBio := "Amante del rock progresivo y jazz"
	newTema := "oscuro"
	newFoto := "profiles/usr-123/nueva_foto.png"
	newBanner := "banners/usr-123/nuevo_banner.jpg"

	req := &models.UpdateProfileRequest{
		Biografia:  &newBio,
		Tema:       &newTema,
		FotoPerfil: &newFoto,
		Banner:     &newBanner,
		Colores: map[string]interface{}{
			"primario":   "#1ABC9C",
			"secundario": "#0D5C5E",
		},
		InformacionMusical: &models.MusicalPreferences{
			Generos:   []string{"rock", "jazz", "ambient"},
			Artistas:  []string{"pink floyd", "miles davis"},
			Canciones: []string{"time", "so what"},
		},
	}

	resp, err := srv.UpdateProfile(context.Background(), "usr-123", req)
	if err != nil {
		t.Fatalf("se esperaba éxito, se obtuvo error: %v", err)
	}

	if resp.Biografia != newBio {
		t.Errorf("biografía esperada '%s', se obtuvo '%s'", newBio, resp.Biografia)
	}
	if resp.Tema != newTema {
		t.Errorf("tema esperado '%s', se obtuvo '%s'", newTema, resp.Tema)
	}
	if resp.FotoPerfil != newFoto {
		t.Errorf("foto_perfil esperada '%s', se obtuvo '%s'", newFoto, resp.FotoPerfil)
	}
	if resp.Banner != newBanner {
		t.Errorf("banner esperado '%s', se obtuvo '%s'", newBanner, resp.Banner)
	}

	// Verificar que mediaVerifier fue invocado para validar foto y banner en media-service
	if len(mediaVerifier.verifiedKeys) != 2 {
		t.Errorf("se esperaba verificación de 2 recursos multimedia, se verificaron: %v", mediaVerifier.verifiedKeys)
	}
}

func TestProfileService_UpdateProfile_Validaciones(t *testing.T) {
	userRepo := &mockUserRepository{
		users: map[string]*models.Usuario{
			"usr-1": {IDUsuario: "usr-1", Username: "user1", Correo: "u1@test.com"},
		},
	}
	profileRepo := &mockProfileRepository{}
	mediaVerifier := &mockMediaVerifier{}
	srv := services.NewProfileService(profileRepo, userRepo, mediaVerifier)

	t.Run("Biografía mayor a 500 caracteres rechazada", func(t *testing.T) {
		bioLarga := strings.Repeat("a", 501)
		req := &models.UpdateProfileRequest{Biografia: &bioLarga}

		_, err := srv.UpdateProfile(context.Background(), "usr-1", req)
		if err == nil {
			t.Fatal("se esperaba error de validación por longitud de biografía")
		}

		var valErr *services.ValidationError
		if !errors.As(err, &valErr) || valErr.Field != "biografia" || valErr.Issue != "too_long" {
			t.Fatalf("se esperaba error too_long en biografia, se obtuvo: %v", err)
		}
	})

	t.Run("Tema con caracteres inválidos rechazado", func(t *testing.T) {
		temaInvalido := "tema<script>alert(1)</script>"
		req := &models.UpdateProfileRequest{Tema: &temaInvalido}

		_, err := srv.UpdateProfile(context.Background(), "usr-1", req)
		if err == nil {
			t.Fatal("se esperaba error de validación para tema con caracteres no permitidos")
		}

		var valErr *services.ValidationError
		if !errors.As(err, &valErr) || valErr.Field != "tema" {
			t.Fatalf("se esperaba error en tema, se obtuvo: %v", err)
		}
	})

	t.Run("Colores con formato no hexadecimal/CSS rechazados", func(t *testing.T) {
		req := &models.UpdateProfileRequest{
			Colores: map[string]interface{}{
				"primario": "color-no-valido-12345",
			},
		}

		_, err := srv.UpdateProfile(context.Background(), "usr-1", req)
		if err == nil {
			t.Fatal("se esperaba error de validación de color")
		}

		var valErr *services.ValidationError
		if !errors.As(err, &valErr) || valErr.Issue != "invalid_color_format" {
			t.Fatalf("se esperaba issue invalid_color_format, se obtuvo: %v", err)
		}
	})

	t.Run("Información musical con más de 50 elementos rechazada", func(t *testing.T) {
		manyGenres := make([]string, 51)
		for i := 0; i < 51; i++ {
			manyGenres[i] = "genre"
		}
		req := &models.UpdateProfileRequest{
			InformacionMusical: &models.MusicalPreferences{
				Generos: manyGenres,
			},
		}

		_, err := srv.UpdateProfile(context.Background(), "usr-1", req)
		if err == nil {
			t.Fatal("se esperaba error por superar límite de elementos musicales")
		}

		var valErr *services.ValidationError
		if !errors.As(err, &valErr) || valErr.Issue != "too_many_items" {
			t.Fatalf("se esperaba issue too_many_items, se obtuvo: %v", err)
		}
	})

	t.Run("Ruta multimedia con path traversal rechazada", func(t *testing.T) {
		fotoInsegura := "../../../etc/passwd"
		req := &models.UpdateProfileRequest{
			FotoPerfil: &fotoInsegura,
		}

		_, err := srv.UpdateProfile(context.Background(), "usr-1", req)
		if err == nil {
			t.Fatal("se esperaba rechazo por path traversal")
		}

		var valErr *services.ValidationError
		if !errors.As(err, &valErr) || valErr.Issue != "invalid_path" {
			t.Fatalf("se esperaba issue invalid_path, se obtuvo: %v", err)
		}
	})
}

// TestProfileService_ConsistenciaAnteFallos_CU03 prueba CU-03 E1/E2
func TestProfileService_ConsistenciaAnteFallos_CU03(t *testing.T) {
	userRepo := &mockUserRepository{
		users: map[string]*models.Usuario{
			"usr-1": {IDUsuario: "usr-1", Username: "user1", Correo: "u1@test.com"},
		},
	}

	t.Run("Si falla MinIO/S3, no se modifica la base de datos y se cancela la operación", func(t *testing.T) {
		profileRepo := &mockProfileRepository{}
		mediaVerifier := &mockMediaVerifier{
			verifyMediaFunc: func(ctx context.Context, objectKey string) (bool, error) {
				return false, services.ErrMediaStorageFailed
			},
		}

		srv := services.NewProfileService(profileRepo, userRepo, mediaVerifier)

		newBio := "Nueva biografia"
		newFoto := "profiles/usr-1/nueva_foto.png"

		req := &models.UpdateProfileRequest{
			Biografia:  &newBio,
			FotoPerfil: &newFoto,
		}

		_, err := srv.UpdateProfile(context.Background(), "usr-1", req)
		if err == nil {
			t.Fatal("se esperaba error por fallo en almacenamiento multimedia")
		}

		if !errors.Is(err, services.ErrMediaStorageFailed) {
			t.Fatalf("se esperaba ErrMediaStorageFailed, se obtuvo: %v", err)
		}

		// Confirmar que la base de datos NO fue modificada (consistencia ante fallos)
		if profileRepo.updatedRecord != nil {
			t.Fatalf("FALLO DE CONSISTENCIA: la base de datos fue modificada a pesar del fallo en MinIO/S3: %+v", profileRepo.updatedRecord)
		}
	})

	t.Run("Si el archivo no existe en MinIO/S3, se rechaza la actualización", func(t *testing.T) {
		profileRepo := &mockProfileRepository{}
		mediaVerifier := &mockMediaVerifier{
			verifyMediaFunc: func(ctx context.Context, objectKey string) (bool, error) {
				return false, nil // No existe
			},
		}

		srv := services.NewProfileService(profileRepo, userRepo, mediaVerifier)

		newFoto := "profiles/usr-1/archivo_inexistente.png"
		req := &models.UpdateProfileRequest{
			FotoPerfil: &newFoto,
		}

		_, err := srv.UpdateProfile(context.Background(), "usr-1", req)
		if err == nil {
			t.Fatal("se esperaba error por archivo no encontrado en almacenamiento")
		}

		var valErr *services.ValidationError
		if !errors.As(err, &valErr) || valErr.Issue != "file_not_found" {
			t.Fatalf("se esperaba issue file_not_found, se obtuvo: %v", err)
		}

		if profileRepo.updatedRecord != nil {
			t.Fatalf("FALLO DE CONSISTENCIA: se modificó la BD para un archivo inexistente")
		}
	})
}
