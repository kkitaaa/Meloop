package services

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/meloop/user-service/models"
	"github.com/meloop/user-service/repositories"
)

var (
	themeRegex    = regexp.MustCompile(`^[a-zA-Z0-9_\-\s]{1,50}$`)
	hexColorRegex = regexp.MustCompile(`^(#([0-9a-fA-F]{3}|[0-9a-fA-F]{6}|[0-9a-fA-F]{8})|0x[0-9a-fA-F]{6,8}|rgba?\([0-9\s,\.]+\))$`)

	// ErrBlocked indica que existe una restricción de bloqueo entre los usuarios involucrados (RN-09)
	ErrBlocked = errors.New("USER_BLOCKED")
)

// ProfileService define las operaciones de negocio sobre el perfil de usuario (RF-06, RF-07)
type ProfileService interface {
	GetOwnProfile(ctx context.Context, userID string) (*models.UserProfileResponse, error)
	GetThirdPartyProfile(ctx context.Context, requestingUserID, targetUserID string) (*models.ThirdPartyProfileResponse, error)
	GetProfile(ctx context.Context, userID string) (*models.ProfileResponse, error)
	UpdateProfile(ctx context.Context, userID string, req *models.UpdateProfileRequest) (*models.ProfileResponse, error)
}

type profileService struct {
	profileRepo   repositories.ProfileRepository
	userRepo      repositories.UserRepository
	privacyRepo   repositories.PrivacyRepository
	mediaVerifier MediaVerifier
}

// NewProfileService crea una nueva instancia de ProfileService
func NewProfileService(
	profileRepo repositories.ProfileRepository,
	userRepo repositories.UserRepository,
	mediaVerifier MediaVerifier,
	privacyRepo ...repositories.PrivacyRepository,
) ProfileService {
	var privRepo repositories.PrivacyRepository
	if len(privacyRepo) > 0 {
		privRepo = privacyRepo[0]
	}
	return &profileService{
		profileRepo:   profileRepo,
		userRepo:      userRepo,
		privacyRepo:   privRepo,
		mediaVerifier: mediaVerifier,
	}
}

// GetOwnProfile recupera la información completa del perfil propio del usuario autenticado (RF-06 / RNF-11)
func (s *profileService) GetOwnProfile(ctx context.Context, userID string) (*models.UserProfileResponse, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, ErrUserNotFound
	}

	data, err := s.profileRepo.GetFullProfileData(ctx, userID)
	if err != nil {
		return nil, err
	}
	if data == nil {
		return nil, ErrUserNotFound
	}

	insignias := data.Insignias
	if insignias == nil {
		insignias = make([]models.InsigniaDTO, 0)
	}

	equipados := data.ElementosEquipados
	if equipados == nil {
		equipados = make([]models.EquippedItemDTO, 0)
	}

	music := data.InformacionMusical
	if music == nil {
		music = &models.MusicalPreferences{
			Generos:   make([]string, 0),
			Artistas:  make([]string, 0),
			Canciones: make([]string, 0),
		}
	}

	resp := &models.UserProfileResponse{
		IDUsuario:          data.IDUsuario,
		Username:           data.Username,
		FotoPerfil:         data.FotoPerfil,
		Banner:             data.Banner,
		Biografia:          data.Biografia,
		Nivel:              &data.NivelNumero,
		Experiencia:        data.Experiencia,
		Tema:               data.Tema,
		Colores:            data.Colores,
		Insignias:          insignias,
		ElementosEquipados: equipados,
		InformacionMusical: music,
		Estadisticas:       &data.Estadisticas,
	}

	return resp, nil
}

// GetThirdPartyProfile recupera la información de perfil de otro usuario aplicando bloqueo y políticas de privacidad (RF-06 / RF-55 / RN-09 / RNF-11)
func (s *profileService) GetThirdPartyProfile(ctx context.Context, requestingUserID, targetUserID string) (*models.ThirdPartyProfileResponse, error) {
	requestingUserID = strings.TrimSpace(requestingUserID)
	targetUserID = strings.TrimSpace(targetUserID)

	if requestingUserID == "" || targetUserID == "" {
		return nil, ErrUserNotFound
	}

	// Si el solicitante consulta su propio ID a través de este método, delegar a perfil propio
	if requestingUserID == targetUserID {
		own, err := s.GetOwnProfile(ctx, targetUserID)
		if err != nil {
			return nil, err
		}
		return &models.ThirdPartyProfileResponse{
			IDUsuario:          own.IDUsuario,
			Username:           own.Username,
			FotoPerfil:         own.FotoPerfil,
			Banner:             own.Banner,
			Biografia:          own.Biografia,
			Nivel:              own.Nivel,
			Experiencia:        own.Experiencia,
			Tema:               own.Tema,
			Colores:            own.Colores,
			Insignias:          own.Insignias,
			ElementosEquipados: own.ElementosEquipados,
			InformacionMusical: own.InformacionMusical,
			Estadisticas:       own.Estadisticas,
			EsAmigo:            false,
			PerfilPrivado:      false,
		}, nil
	}

	// 1. Restricción por bloqueo (RN-09): Comprobar bloqueo bidireccional
	blocked, err := s.profileRepo.IsBlocked(ctx, requestingUserID, targetUserID)
	if err != nil {
		return nil, err
	}
	if blocked {
		return nil, ErrBlocked
	}

	// 2. Obtener datos consolidados del perfil destino
	data, err := s.profileRepo.GetFullProfileData(ctx, targetUserID)
	if err != nil {
		return nil, err
	}
	if data == nil {
		return nil, ErrUserNotFound
	}

	// 3. Consultar configuración de privacidad del titular (RF-55)
	var privacy *models.ConfiguracionPrivacidad
	if s.privacyRepo != nil {
		privacy, err = s.privacyRepo.GetByUserID(ctx, targetUserID)
		if err != nil {
			return nil, err
		}
	}
	if privacy == nil {
		privacy = &models.ConfiguracionPrivacidad{
			VisibilidadPerfil:        models.VisibilidadPublico,
			VisibilidadPublicaciones: models.VisibilidadPublico,
			VisibilidadInteracciones: models.VisibilidadPublico,
		}
	}

	// 4. Comprobar relación de amistad aceptada
	isFriend, err := s.profileRepo.IsFriend(ctx, requestingUserID, targetUserID)
	if err != nil {
		return nil, err
	}

	// 5. Aplicar reglas de privacidad del perfil (RF-55)
	visPerfil := strings.ToUpper(strings.TrimSpace(privacy.VisibilidadPerfil))
	if visPerfil == "" {
		visPerfil = models.VisibilidadPublico
	}

	// Caso: Perfil privado o visible solo para amigos cuando NO existe amistad aceptada
	if visPerfil == models.VisibilidadPrivado || (visPerfil == models.VisibilidadAmigos && !isFriend) {
		// Retornar exclusivamente la información básica de identificación permitida
		return &models.ThirdPartyProfileResponse{
			IDUsuario:     data.IDUsuario,
			Username:      data.Username,
			FotoPerfil:    data.FotoPerfil,
			Nivel:         &data.NivelNumero,
			Experiencia:   data.Experiencia,
			EsAmigo:       isFriend,
			PerfilPrivado: true,
		}, nil
	}

	// Caso: Perfil público o visible para amigos con amistad aceptada
	visPub := strings.ToUpper(strings.TrimSpace(privacy.VisibilidadPublicaciones))
	if visPub == "" {
		visPub = models.VisibilidadPublico
	}

	stats := &models.ProfileStatsDTO{
		Amigos: data.Estadisticas.Amigos,
	}
	if visPub == models.VisibilidadPublico || (visPub == models.VisibilidadAmigos && isFriend) {
		stats.Publicaciones = data.Estadisticas.Publicaciones
	}

	insignias := data.Insignias
	if insignias == nil {
		insignias = make([]models.InsigniaDTO, 0)
	}

	equipados := data.ElementosEquipados
	if equipados == nil {
		equipados = make([]models.EquippedItemDTO, 0)
	}

	music := data.InformacionMusical
	if music == nil {
		music = &models.MusicalPreferences{
			Generos:   make([]string, 0),
			Artistas:  make([]string, 0),
			Canciones: make([]string, 0),
		}
	}

	return &models.ThirdPartyProfileResponse{
		IDUsuario:          data.IDUsuario,
		Username:           data.Username,
		FotoPerfil:         data.FotoPerfil,
		Banner:             data.Banner,
		Biografia:          data.Biografia,
		Nivel:              &data.NivelNumero,
		Experiencia:        data.Experiencia,
		Tema:               data.Tema,
		Colores:            data.Colores,
		Insignias:          insignias,
		ElementosEquipados: equipados,
		InformacionMusical: music,
		Estadisticas:       stats,
		EsAmigo:            isFriend,
		PerfilPrivado:      false,
	}, nil
}

// GetProfile recupera el perfil del usuario autenticado
func (s *profileService) GetProfile(ctx context.Context, userID string) (*models.ProfileResponse, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, ErrUserNotFound
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	profile, err := s.profileRepo.GetProfile(ctx, userID)
	if err != nil {
		return nil, err
	}

	if profile == nil {
		return &models.ProfileResponse{
			IDUsuario:          user.IDUsuario,
			Username:           user.Username,
			Biografia:          "",
			FotoPerfil:         "",
			Banner:             "",
			Tema:               "default",
			Colores:            make(map[string]interface{}),
			InformacionMusical: &models.MusicalPreferences{},
		}, nil
	}

	return profile, nil
}

// UpdateProfile actualiza la información personalizable del perfil aplicando validaciones y consistencia multimedia (RF-07 / REI-03 / CU-03)
func (s *profileService) UpdateProfile(ctx context.Context, userID string, req *models.UpdateProfileRequest) (*models.ProfileResponse, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, ErrUserNotFound
	}

	if req == nil {
		return nil, &ValidationError{
			Field:   "profile",
			Issue:   "required",
			Message: "Los datos de actualización de perfil son obligatorios",
		}
	}

	// 1. Verificar que el usuario exista
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	// 2. Obtener estado actual del perfil
	currentProfile, err := s.profileRepo.GetProfile(ctx, userID)
	if err != nil {
		return nil, err
	}

	profileToUpdate := models.Profile{
		IDUsuario:  userID,
		Username:   user.Username,
		Biografia:  "",
		FotoPerfil: "",
		Banner:     "",
		Tema:       "default",
		Colores:    make(map[string]interface{}),
	}

	if currentProfile != nil {
		profileToUpdate.Biografia = currentProfile.Biografia
		profileToUpdate.FotoPerfil = currentProfile.FotoPerfil
		profileToUpdate.Banner = currentProfile.Banner
		profileToUpdate.Tema = currentProfile.Tema
		if currentProfile.Colores != nil {
			profileToUpdate.Colores = currentProfile.Colores
		}
		profileToUpdate.InformacionMusical = currentProfile.InformacionMusical
	}

	// 3. Validar y procesar Biografía
	if reqBio := req.GetBiografia(); reqBio != nil {
		bio := strings.TrimSpace(*reqBio)
		if len(bio) > 500 {
			return nil, &ValidationError{
				Field:   "biografia",
				Issue:   "too_long",
				Message: "La biografía no puede superar los 500 caracteres",
			}
		}
		profileToUpdate.Biografia = bio
	}

	// 4. Validar y procesar Tema
	if reqTema := req.GetTema(); reqTema != nil {
		tema := strings.TrimSpace(*reqTema)
		if len(tema) > 50 {
			return nil, &ValidationError{
				Field:   "tema",
				Issue:   "too_long",
				Message: "El nombre del tema no puede superar los 50 caracteres",
			}
		}
		if tema != "" && !themeRegex.MatchString(tema) {
			return nil, &ValidationError{
				Field:   "tema",
				Issue:   "invalid_format",
				Message: "El tema contiene caracteres inválidos",
			}
		}
		if tema == "" {
			tema = "default"
		}
		profileToUpdate.Tema = tema
	}

	// 5. Validar y procesar Colores
	if reqColores := req.GetColores(); reqColores != nil {
		for key, val := range reqColores {
			if strVal, ok := val.(string); ok && strVal != "" {
				trimmed := strings.TrimSpace(strVal)
				if !hexColorRegex.MatchString(trimmed) && !isValidNamedColor(trimmed) {
					return nil, &ValidationError{
						Field:   fmt.Sprintf("colores.%s", key),
						Issue:   "invalid_color_format",
						Message: fmt.Sprintf("El valor de color '%s' no tiene un formato hexadecimal o CSS válido", strVal),
					}
				}
			}
		}
		profileToUpdate.Colores = reqColores
	}

	// 6. Validar y procesar Información Musical
	if reqMusic := req.GetInformacionMusical(); reqMusic != nil {
		if len(reqMusic.Generos) > 50 {
			return nil, &ValidationError{
				Field:   "informacion_musical.generos",
				Issue:   "too_many_items",
				Message: "No puede especificar más de 50 géneros musicales",
			}
		}
		if len(reqMusic.Artistas) > 50 {
			return nil, &ValidationError{
				Field:   "informacion_musical.artistas",
				Issue:   "too_many_items",
				Message: "No puede especificar más de 50 artistas musicales",
			}
		}
		if len(reqMusic.Canciones) > 50 {
			return nil, &ValidationError{
				Field:   "informacion_musical.canciones",
				Issue:   "too_many_items",
				Message: "No puede especificar más de 50 canciones musicales",
			}
		}

		for _, item := range append(append(reqMusic.Generos, reqMusic.Artistas...), reqMusic.Canciones...) {
			if len(item) > 100 {
				return nil, &ValidationError{
					Field:   "informacion_musical",
					Issue:   "item_too_long",
					Message: "Los identificadores o nombres musicales no pueden superar los 100 caracteres",
				}
			}
		}
		profileToUpdate.InformacionMusical = reqMusic
	}

	// 7. Validar y verificar Foto de Perfil (CU-03: Consistencia ante fallos)
	if reqFoto := req.GetFotoPerfil(); reqFoto != nil {
		fotoKey := strings.TrimSpace(*reqFoto)
		if len(fotoKey) > 255 {
			return nil, &ValidationError{
				Field:   "foto_perfil",
				Issue:   "too_long",
				Message: "La referencia de la foto de perfil no puede superar los 255 caracteres",
			}
		}
		if strings.Contains(fotoKey, "..") {
			return nil, &ValidationError{
				Field:   "foto_perfil",
				Issue:   "invalid_path",
				Message: "La referencia de archivo contiene caracteres no permitidos",
			}
		}

		// Si cambió y no está vacía, verificar su existencia física en MinIO/S3 a través de media-service
		if fotoKey != "" && fotoKey != profileToUpdate.FotoPerfil && s.mediaVerifier != nil {
			exists, err := s.mediaVerifier.VerifyMedia(ctx, fotoKey)
			if err != nil {
				return nil, fmt.Errorf("%w: error al verificar foto de perfil en media-service: %v", ErrMediaStorageFailed, err)
			}
			if !exists {
				return nil, &ValidationError{
					Field:   "foto_perfil",
					Issue:   "file_not_found",
					Message: "La foto de perfil especificada no existe en el almacenamiento",
				}
			}
		}
		profileToUpdate.FotoPerfil = fotoKey
	}

	// 8. Validar y verificar Banner (CU-03: Consistencia ante fallos)
	if reqBanner := req.GetBanner(); reqBanner != nil {
		bannerKey := strings.TrimSpace(*reqBanner)
		if len(bannerKey) > 255 {
			return nil, &ValidationError{
				Field:   "banner",
				Issue:   "too_long",
				Message: "La referencia del banner no puede superar los 255 caracteres",
			}
		}
		if strings.Contains(bannerKey, "..") {
			return nil, &ValidationError{
				Field:   "banner",
				Issue:   "invalid_path",
				Message: "La referencia de archivo contiene caracteres no permitidos",
			}
		}

		// Si cambió y no está vacía, verificar su existencia física en MinIO/S3 a través de media-service
		if bannerKey != "" && bannerKey != profileToUpdate.Banner && s.mediaVerifier != nil {
			exists, err := s.mediaVerifier.VerifyMedia(ctx, bannerKey)
			if err != nil {
				return nil, fmt.Errorf("%w: error al verificar banner en media-service: %v", ErrMediaStorageFailed, err)
			}
			if !exists {
				return nil, &ValidationError{
					Field:   "banner",
					Issue:   "file_not_found",
					Message: "El banner especificado no existe en el almacenamiento",
				}
			}
		}
		profileToUpdate.Banner = bannerKey
	}

	// 9. Persistir atómicamente todos los cambios en PostgreSQL
	updated, err := s.profileRepo.UpdateProfile(ctx, userID, &profileToUpdate)
	if err != nil {
		return nil, err
	}

	return updated, nil
}

func isValidNamedColor(c string) bool {
	switch strings.ToLower(c) {
	case "red", "green", "blue", "yellow", "cyan", "magenta", "black", "white",
		"teal", "purple", "orange", "grey", "gray", "transparent":
		return true
	default:
		return false
	}
}
