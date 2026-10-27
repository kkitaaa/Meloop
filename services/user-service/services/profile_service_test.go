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
	getProfileFunc         func(ctx context.Context, userID string) (*models.ProfileResponse, error)
	updateProfileFunc      func(ctx context.Context, userID string, profile *models.Profile) (*models.ProfileResponse, error)
	getFullProfileDataFunc func(ctx context.Context, userID string) (*models.UserProfileData, error)
	isBlockedFunc          func(ctx context.Context, user1ID, user2ID string) (bool, error)
	isFriendFunc           func(ctx context.Context, user1ID, user2ID string) (bool, error)
	updatedRecord          *models.Profile
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

func (m *mockProfileRepository) GetFullProfileData(ctx context.Context, userID string) (*models.UserProfileData, error) {
	if m.getFullProfileDataFunc != nil {
		return m.getFullProfileDataFunc(ctx, userID)
	}
	lvl := 3
	return &models.UserProfileData{
		IDUsuario:   userID,
		Username:    "usuario_test",
		Experiencia: 500,
		IDNivel:     &lvl,
		NivelNumero: lvl,
		Biografia:   "Bio de prueba",
		FotoPerfil:  "profiles/u1/foto.png",
		Banner:      "banners/u1/banner.png",
		Tema:        "default",
		Colores:     map[string]interface{}{"primario": "#1ABC9C"},
		Insignias: []models.InsigniaDTO{
			{IDRecompensa: "rew-ins-1", Tipo: "INSIGNIA", Equipada: true},
		},
		ElementosEquipados: []models.EquippedItemDTO{
			{IDInventario: "inv-1", IDRecompensa: "rew-marco-gold", Tipo: "MARCO"},
		},
		InformacionMusical: &models.MusicalPreferences{
			Generos:   []string{"rock", "indie"},
			Artistas:  []string{"radiohead"},
			Canciones: []string{"creep"},
		},
		Estadisticas: models.ProfileStatsDTO{
			Publicaciones: 8,
			Amigos:        4,
		},
	}, nil
}

func (m *mockProfileRepository) IsBlocked(ctx context.Context, user1ID, user2ID string) (bool, error) {
	if m.isBlockedFunc != nil {
		return m.isBlockedFunc(ctx, user1ID, user2ID)
	}
	return false, nil
}

func (m *mockProfileRepository) IsFriend(ctx context.Context, user1ID, user2ID string) (bool, error) {
	if m.isFriendFunc != nil {
		return m.isFriendFunc(ctx, user1ID, user2ID)
	}
	return false, nil
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

// =========================================================================
// Tests exhaustivos para Consulta de Perfiles (RF-06 / RN-09 / RF-55 / RNF-11)
// =========================================================================

func TestProfileService_GetOwnProfile_Success(t *testing.T) {
	profileRepo := &mockProfileRepository{}
	userRepo := &mockUserRepository{}
	mediaVerifier := &mockMediaVerifier{}
	privacyRepo := &mockPrivacyRepository{}

	srv := services.NewProfileService(profileRepo, userRepo, mediaVerifier, privacyRepo)

	profile, err := srv.GetOwnProfile(context.Background(), "usr-own")
	if err != nil {
		t.Fatalf("error inesperado al consultar perfil propio: %v", err)
	}

	if profile == nil {
		t.Fatal("se esperaba perfil no nulo")
	}
	if profile.IDUsuario != "usr-own" {
		t.Errorf("esperado IDUsuario 'usr-own', obtenido '%s'", profile.IDUsuario)
	}
	if profile.Username != "usuario_test" {
		t.Errorf("esperado username 'usuario_test', obtenido '%s'", profile.Username)
	}
	if profile.Nivel == nil || *profile.Nivel != 3 {
		t.Errorf("esperado nivel 3, obtenido %v", profile.Nivel)
	}
	if profile.Experiencia != 500 {
		t.Errorf("esperada experiencia 500, obtenido %d", profile.Experiencia)
	}
	if len(profile.Insignias) != 1 || profile.Insignias[0].IDRecompensa != "rew-ins-1" {
		t.Errorf("insignias incorrectas: %+v", profile.Insignias)
	}
	if len(profile.ElementosEquipados) != 1 || profile.ElementosEquipados[0].IDRecompensa != "rew-marco-gold" {
		t.Errorf("elementos equipados incorrectos: %+v", profile.ElementosEquipados)
	}
	if profile.InformacionMusical == nil || len(profile.InformacionMusical.Generos) != 2 {
		t.Errorf("información musical incorrecta: %+v", profile.InformacionMusical)
	}
	if profile.Estadisticas == nil || profile.Estadisticas.Publicaciones != 8 || profile.Estadisticas.Amigos != 4 {
		t.Errorf("estadísticas incorrectas: %+v", profile.Estadisticas)
	}
}

func TestProfileService_GetOwnProfile_UserNotFound(t *testing.T) {
	profileRepo := &mockProfileRepository{
		getFullProfileDataFunc: func(ctx context.Context, userID string) (*models.UserProfileData, error) {
			return nil, nil // Usuario no encontrado en BD
		},
	}
	userRepo := &mockUserRepository{}
	mediaVerifier := &mockMediaVerifier{}

	srv := services.NewProfileService(profileRepo, userRepo, mediaVerifier)

	_, err := srv.GetOwnProfile(context.Background(), "usr-inexistente")
	if !errors.Is(err, services.ErrUserNotFound) {
		t.Fatalf("se esperaba ErrUserNotFound, obtenido: %v", err)
	}
}

func TestProfileService_GetThirdPartyProfile_Public_Success(t *testing.T) {
	profileRepo := &mockProfileRepository{}
	userRepo := &mockUserRepository{}
	mediaVerifier := &mockMediaVerifier{}
	privacyRepo := &mockPrivacyRepository{
		configs: map[string]*models.ConfiguracionPrivacidad{
			"usr-target": {
				IDUsuario:                "usr-target",
				VisibilidadPerfil:        models.VisibilidadPublico,
				VisibilidadPublicaciones: models.VisibilidadPublico,
			},
		},
	}

	srv := services.NewProfileService(profileRepo, userRepo, mediaVerifier, privacyRepo)

	profile, err := srv.GetThirdPartyProfile(context.Background(), "usr-requester", "usr-target")
	if err != nil {
		t.Fatalf("error inesperado en perfil público: %v", err)
	}

	if profile.PerfilPrivado {
		t.Error("un perfil público no debe marcarse como privado")
	}
	if profile.Biografia != "Bio de prueba" {
		t.Errorf("biografía esperada 'Bio de prueba', obtenida '%s'", profile.Biografia)
	}
	if profile.InformacionMusical == nil || len(profile.InformacionMusical.Artistas) != 1 {
		t.Errorf("información musical pública ausente o incorrecta")
	}
	if profile.Estadisticas == nil || profile.Estadisticas.Publicaciones != 8 {
		t.Errorf("estadísticas públicas esperadas con publicaciones=8, obtenido %+v", profile.Estadisticas)
	}
}

func TestProfileService_GetThirdPartyProfile_FriendsOnly_Friend_Success(t *testing.T) {
	profileRepo := &mockProfileRepository{
		isFriendFunc: func(ctx context.Context, user1ID, user2ID string) (bool, error) {
			return true, nil // Son amigos aceptados
		},
	}
	userRepo := &mockUserRepository{}
	mediaVerifier := &mockMediaVerifier{}
	privacyRepo := &mockPrivacyRepository{
		configs: map[string]*models.ConfiguracionPrivacidad{
			"usr-friend": {
				IDUsuario:                "usr-friend",
				VisibilidadPerfil:        models.VisibilidadAmigos,
				VisibilidadPublicaciones: models.VisibilidadAmigos,
			},
		},
	}

	srv := services.NewProfileService(profileRepo, userRepo, mediaVerifier, privacyRepo)

	profile, err := srv.GetThirdPartyProfile(context.Background(), "usr-requester", "usr-friend")
	if err != nil {
		t.Fatalf("error inesperado consultando perfil de amigo: %v", err)
	}

	if profile.PerfilPrivado {
		t.Error("el perfil para un amigo no debe marcarse como privado")
	}
	if !profile.EsAmigo {
		t.Error("se esperaba EsAmigo=true")
	}
	if profile.Biografia == "" {
		t.Error("el amigo aceptado debe poder ver la biografía")
	}
	if profile.InformacionMusical == nil {
		t.Error("el amigo aceptado debe poder ver la información musical")
	}
	if profile.Estadisticas == nil || profile.Estadisticas.Publicaciones != 8 {
		t.Error("el amigo aceptado debe poder ver las estadísticas")
	}
}

func TestProfileService_GetThirdPartyProfile_FriendsOnly_NotFriend_Restricted(t *testing.T) {
	profileRepo := &mockProfileRepository{
		isFriendFunc: func(ctx context.Context, user1ID, user2ID string) (bool, error) {
			return false, nil // NO son amigos
		},
	}
	userRepo := &mockUserRepository{}
	mediaVerifier := &mockMediaVerifier{}
	privacyRepo := &mockPrivacyRepository{
		configs: map[string]*models.ConfiguracionPrivacidad{
			"usr-not-friend": {
				IDUsuario:                "usr-not-friend",
				VisibilidadPerfil:        models.VisibilidadAmigos,
				VisibilidadPublicaciones: models.VisibilidadAmigos,
			},
		},
	}

	srv := services.NewProfileService(profileRepo, userRepo, mediaVerifier, privacyRepo)

	profile, err := srv.GetThirdPartyProfile(context.Background(), "usr-requester", "usr-not-friend")
	if err != nil {
		t.Fatalf("error inesperado consultando perfil solo amigos: %v", err)
	}

	if !profile.PerfilPrivado {
		t.Error("debe marcarse como perfil privado para no amigos")
	}
	if profile.EsAmigo {
		t.Error("se esperaba EsAmigo=false")
	}
	// Restricción de campos por privacidad (RF-55)
	if profile.Biografia != "" {
		t.Errorf("la biografía debió restringirse para no amigos, se obtuvo: %s", profile.Biografia)
	}
	if profile.InformacionMusical != nil {
		t.Error("la información musical debió restringirse para no amigos")
	}
	if profile.Estadisticas != nil {
		t.Error("las estadísticas debieron restringirse para no amigos")
	}
	if len(profile.Insignias) > 0 {
		t.Error("las insignias debieron restringirse para no amigos")
	}
	if len(profile.ElementosEquipados) > 0 {
		t.Error("los elementos equipados debieron restringirse para no amigos")
	}
	// Datos de identificación permitidos
	if profile.Username != "usuario_test" || profile.FotoPerfil == "" {
		t.Error("los datos públicos básicos de identificación deben estar presentes")
	}
}

func TestProfileService_GetThirdPartyProfile_Private_Restricted(t *testing.T) {
	profileRepo := &mockProfileRepository{}
	userRepo := &mockUserRepository{}
	mediaVerifier := &mockMediaVerifier{}
	privacyRepo := &mockPrivacyRepository{
		configs: map[string]*models.ConfiguracionPrivacidad{
			"usr-private": {
				IDUsuario:                "usr-private",
				VisibilidadPerfil:        models.VisibilidadPrivado,
				VisibilidadPublicaciones: models.VisibilidadPrivado,
			},
		},
	}

	srv := services.NewProfileService(profileRepo, userRepo, mediaVerifier, privacyRepo)

	profile, err := srv.GetThirdPartyProfile(context.Background(), "usr-requester", "usr-private")
	if err != nil {
		t.Fatalf("error inesperado consultando perfil privado: %v", err)
	}

	if !profile.PerfilPrivado {
		t.Error("debe marcarse como perfil privado")
	}
	if profile.Biografia != "" || profile.InformacionMusical != nil || profile.Estadisticas != nil {
		t.Error("los campos sensibles de perfil privado deben estar omitidos para terceros")
	}
}

func TestProfileService_GetThirdPartyProfile_PublicationsPrivate_StatsHidden(t *testing.T) {
	profileRepo := &mockProfileRepository{}
	userRepo := &mockUserRepository{}
	mediaVerifier := &mockMediaVerifier{}
	privacyRepo := &mockPrivacyRepository{
		configs: map[string]*models.ConfiguracionPrivacidad{
			"usr-target": {
				IDUsuario:                "usr-target",
				VisibilidadPerfil:        models.VisibilidadPublico,
				VisibilidadPublicaciones: models.VisibilidadPrivado, // Publicaciones privadas
			},
		},
	}

	srv := services.NewProfileService(profileRepo, userRepo, mediaVerifier, privacyRepo)

	profile, err := srv.GetThirdPartyProfile(context.Background(), "usr-requester", "usr-target")
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}

	// El perfil es público pero las publicaciones son privadas
	if profile.PerfilPrivado {
		t.Error("el perfil es público, no debe ser privado")
	}
	if profile.Estadisticas == nil {
		t.Fatal("las estadísticas de perfil deben estar presentes")
	}
	if profile.Estadisticas.Publicaciones != 0 {
		t.Errorf("las publicaciones privadas deben tener contador 0 para terceros, obtenido: %d", profile.Estadisticas.Publicaciones)
	}
	if profile.Estadisticas.Amigos != 4 {
		t.Errorf("los amigos públicos deben mantenerse, obtenido: %d", profile.Estadisticas.Amigos)
	}
}

func TestProfileService_GetThirdPartyProfile_Blocked_RequesterBlockedOwner_RN09(t *testing.T) {
	profileRepo := &mockProfileRepository{
		isBlockedFunc: func(ctx context.Context, user1ID, user2ID string) (bool, error) {
			// El solicitante bloqueó al titular
			if user1ID == "usr-blocker" && user2ID == "usr-blocked" {
				return true, nil
			}
			return false, nil
		},
	}
	userRepo := &mockUserRepository{}
	mediaVerifier := &mockMediaVerifier{}

	srv := services.NewProfileService(profileRepo, userRepo, mediaVerifier)

	_, err := srv.GetThirdPartyProfile(context.Background(), "usr-blocker", "usr-blocked")
	if !errors.Is(err, services.ErrBlocked) {
		t.Fatalf("se esperaba ErrBlocked cuando el solicitante bloqueó al titular, se obtuvo: %v", err)
	}
}

func TestProfileService_GetThirdPartyProfile_Blocked_OwnerBlockedRequester_RN09(t *testing.T) {
	profileRepo := &mockProfileRepository{
		isBlockedFunc: func(ctx context.Context, user1ID, user2ID string) (bool, error) {
			// El titular bloqueó al solicitante
			if user1ID == "usr-requester" && user2ID == "usr-owner" {
				return true, nil
			}
			return false, nil
		},
	}
	userRepo := &mockUserRepository{}
	mediaVerifier := &mockMediaVerifier{}

	srv := services.NewProfileService(profileRepo, userRepo, mediaVerifier)

	_, err := srv.GetThirdPartyProfile(context.Background(), "usr-requester", "usr-owner")
	if !errors.Is(err, services.ErrBlocked) {
		t.Fatalf("se esperaba ErrBlocked cuando el titular bloqueó al solicitante, se obtuvo: %v", err)
	}
}

func TestProfileService_GetThirdPartyProfile_SelfID_DelegatesToOwn(t *testing.T) {
	profileRepo := &mockProfileRepository{}
	userRepo := &mockUserRepository{}
	mediaVerifier := &mockMediaVerifier{}

	srv := services.NewProfileService(profileRepo, userRepo, mediaVerifier)

	profile, err := srv.GetThirdPartyProfile(context.Background(), "usr-own", "usr-own")
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}

	if profile.IDUsuario != "usr-own" {
		t.Errorf("esperado IDUsuario usr-own, obtenido %s", profile.IDUsuario)
	}
	if profile.Biografia != "Bio de prueba" {
		t.Errorf("el titular debe ver su biografía propia completa")
	}
}

func TestProfileService_OptionalFieldsEmpty_NoError(t *testing.T) {
	profileRepo := &mockProfileRepository{
		getFullProfileDataFunc: func(ctx context.Context, userID string) (*models.UserProfileData, error) {
			// Simular usuario sin perfil previo, sin música, sin insignias, sin items equipados
			return &models.UserProfileData{
				IDUsuario:          userID,
				Username:           "nuevo_usuario",
				Experiencia:        0,
				IDNivel:            nil,
				NivelNumero:        1,
				Biografia:          "",
				FotoPerfil:         "",
				Banner:             "",
				Tema:               "default",
				Colores:            map[string]interface{}{},
				Insignias:          nil,
				ElementosEquipados: nil,
				InformacionMusical: nil,
				Estadisticas:       models.ProfileStatsDTO{Publicaciones: 0, Amigos: 0},
			}, nil
		},
	}
	userRepo := &mockUserRepository{}
	mediaVerifier := &mockMediaVerifier{}

	srv := services.NewProfileService(profileRepo, userRepo, mediaVerifier)

	// 1. Probar propio perfil
	own, err := srv.GetOwnProfile(context.Background(), "usr-empty")
	if err != nil {
		t.Fatalf("error inesperado con campos opcionales vacíos en propio perfil: %v", err)
	}
	if own.Insignias == nil || own.ElementosEquipados == nil || own.InformacionMusical == nil {
		t.Error("las listas y mapas opcionales no deben ser nulos")
	}

	// 2. Probar perfil de tercero
	third, err := srv.GetThirdPartyProfile(context.Background(), "usr-requester", "usr-empty")
	if err != nil {
		t.Fatalf("error inesperado con campos opcionales vacíos en perfil tercero: %v", err)
	}
	if third.Insignias == nil || third.ElementosEquipados == nil || third.InformacionMusical == nil {
		t.Error("las listas y mapas opcionales no deben ser nulos")
	}
}
