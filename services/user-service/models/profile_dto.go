package models

import "time"

// InsigniaDTO representa una recompensa de tipo insignia en el perfil (RF-06)
type InsigniaDTO struct {
	IDRecompensa    string    `json:"id_recompensa"`
	Tipo            string    `json:"tipo"`
	Equipada        bool      `json:"equipada"`
	FechaDesbloqueo time.Time `json:"fecha_desbloqueo,omitempty"`
}

// EquippedItemDTO representa un elemento de personalización visual equipado en el perfil (RF-06)
type EquippedItemDTO struct {
	IDInventario    string    `json:"id_inventario"`
	IDRecompensa    string    `json:"id_recompensa"`
	Tipo            string    `json:"tipo"`
	FechaDesbloqueo time.Time `json:"fecha_desbloqueo,omitempty"`
}

// ProfileStatsDTO representa las estadísticas relevantes y autorizadas del perfil (RF-06)
type ProfileStatsDTO struct {
	Publicaciones int `json:"publicaciones"`
	Amigos        int `json:"amigos"`
}

// UserProfileResponse representa el DTO de respuesta para la consulta del propio perfil (RF-06 / RNF-11)
// Excluye estrictamente: correo, contrasena_hash, fecha_nacimiento e historial de reportes
type UserProfileResponse struct {
	IDUsuario          string                 `json:"id_usuario"`
	Username           string                 `json:"username"`
	FotoPerfil         string                 `json:"foto_perfil"`
	Banner             string                 `json:"banner"`
	Biografia          string                 `json:"biografia"`
	Nivel              *int                   `json:"nivel,omitempty"`
	Experiencia        int                    `json:"experiencia"`
	Tema               string                 `json:"tema,omitempty"`
	Colores            map[string]interface{} `json:"colores,omitempty"`
	Insignias          []InsigniaDTO          `json:"insignias"`
	ElementosEquipados []EquippedItemDTO      `json:"elementos_equipados"`
	InformacionMusical *MusicalPreferences    `json:"informacion_musical,omitempty"`
	Estadisticas       *ProfileStatsDTO       `json:"estadisticas,omitempty"`
}

// ThirdPartyProfileResponse representa el DTO de respuesta para la consulta de perfiles de otros usuarios (RF-06 / RF-55 / RNF-11)
// Excluye estrictamente: correo, contrasena_hash, fecha_nacimiento e historial de reportes.
// Se construye exclusivamente con los campos permitidos según las políticas de privacidad del titular.
type ThirdPartyProfileResponse struct {
	IDUsuario          string                 `json:"id_usuario"`
	Username           string                 `json:"username"`
	FotoPerfil         string                 `json:"foto_perfil"`
	Banner             string                 `json:"banner,omitempty"`
	Biografia          string                 `json:"biografia,omitempty"`
	Nivel              *int                   `json:"nivel,omitempty"`
	Experiencia        int                    `json:"experiencia"`
	Tema               string                 `json:"tema,omitempty"`
	Colores            map[string]interface{} `json:"colores,omitempty"`
	Insignias          []InsigniaDTO          `json:"insignias,omitempty"`
	ElementosEquipados []EquippedItemDTO      `json:"elementos_equipados,omitempty"`
	InformacionMusical *MusicalPreferences    `json:"informacion_musical,omitempty"`
	Estadisticas       *ProfileStatsDTO       `json:"estadisticas,omitempty"`
	EsAmigo            bool                   `json:"es_amigo,omitempty"`
	PerfilPrivado      bool                   `json:"perfil_privado,omitempty"`
}

// UserProfileData es la estructura interna de persistencia para consolidar los datos de perfil antes de construir los DTOs
type UserProfileData struct {
	IDUsuario          string
	Username           string
	Experiencia        int
	IDNivel            *int
	NivelNumero        int
	Biografia          string
	FotoPerfil         string
	Banner             string
	Tema               string
	Colores            map[string]interface{}
	Insignias          []InsigniaDTO
	ElementosEquipados []EquippedItemDTO
	InformacionMusical *MusicalPreferences
	Estadisticas       ProfileStatsDTO
}
