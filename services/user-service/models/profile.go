package models

// MusicalPreferences representa las preferencias musicales del usuario asociadas a su perfil
type MusicalPreferences struct {
	Generos   []string `json:"generos,omitempty"`
	Artistas  []string `json:"artistas,omitempty"`
	Canciones []string `json:"canciones,omitempty"`
}

// Profile representa la entidad de perfil del usuario en la base de datos
type Profile struct {
	IDUsuario          string                 `json:"id_usuario"`
	Username           string                 `json:"username"`
	Biografia          string                 `json:"biografia"`
	FotoPerfil         string                 `json:"foto_perfil"`
	Banner             string                 `json:"banner"`
	Tema               string                 `json:"tema"`
	Colores            map[string]interface{} `json:"colores"`
	InformacionMusical *MusicalPreferences    `json:"informacion_musical,omitempty"`
}

// UpdateProfileRequest contiene los campos editables del perfil según RF-07
type UpdateProfileRequest struct {
	Biografia             *string                `json:"biografia"`
	Bio                   *string                `json:"bio"` // Alias
	FotoPerfil            *string                `json:"foto_perfil"`
	Avatar                *string                `json:"avatar"`     // Alias
	AvatarURL             *string                `json:"avatar_url"` // Alias
	Banner                *string                `json:"banner"`
	BannerURL             *string                `json:"banner_url"` // Alias
	Tema                  *string                `json:"tema"`
	Theme                 *string                `json:"theme"` // Alias
	Colores               map[string]interface{} `json:"colores"`
	Colors                map[string]interface{} `json:"colors"` // Alias
	InformacionMusical    *MusicalPreferences    `json:"informacion_musical"`
	PreferenciasMusicales *MusicalPreferences    `json:"preferencias_musicales"` // Alias
}

// GetBiografia devuelve la biografía considerando alias
func (r *UpdateProfileRequest) GetBiografia() *string {
	if r.Biografia != nil {
		return r.Biografia
	}
	return r.Bio
}

// GetFotoPerfil devuelve la referencia a foto de perfil considerando alias
func (r *UpdateProfileRequest) GetFotoPerfil() *string {
	if r.FotoPerfil != nil {
		return r.FotoPerfil
	}
	if r.Avatar != nil {
		return r.Avatar
	}
	return r.AvatarURL
}

// GetBanner devuelve la referencia al banner considerando alias
func (r *UpdateProfileRequest) GetBanner() *string {
	if r.Banner != nil {
		return r.Banner
	}
	return r.BannerURL
}

// GetTema devuelve el tema configurado considerando alias
func (r *UpdateProfileRequest) GetTema() *string {
	if r.Tema != nil {
		return r.Tema
	}
	return r.Theme
}

// GetColores devuelve la paleta de colores considerando alias
func (r *UpdateProfileRequest) GetColores() map[string]interface{} {
	if r.Colores != nil {
		return r.Colores
	}
	return r.Colors
}

// GetInformacionMusical devuelve la información musical considerando alias
func (r *UpdateProfileRequest) GetInformacionMusical() *MusicalPreferences {
	if r.InformacionMusical != nil {
		return r.InformacionMusical
	}
	return r.PreferenciasMusicales
}

// ProfileResponse representa la respuesta JSON estructurada al consultar o actualizar el perfil
type ProfileResponse struct {
	IDUsuario          string                 `json:"id_usuario"`
	Username           string                 `json:"username"`
	Biografia          string                 `json:"biografia"`
	FotoPerfil         string                 `json:"foto_perfil"`
	Banner             string                 `json:"banner"`
	Tema               string                 `json:"tema"`
	Colores            map[string]interface{} `json:"colores"`
	InformacionMusical *MusicalPreferences    `json:"informacion_musical"`
}
