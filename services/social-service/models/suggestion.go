package models

// CandidateUser representa un usuario candidato elegible para sugerencia de amistad.
type CandidateUser struct {
	IDUsuario   string `json:"id_usuario"`
	Username    string `json:"username"`
	Correo      string `json:"correo,omitempty"`
	IDNivel     *int   `json:"id_nivel,omitempty"`
	Experiencia int    `json:"experiencia"`
	IsActive    bool   `json:"activo"`
}

// FriendSuggestion representa una sugerencia de amistad completa para RF-14,
// con identificación, nombre del candidato, porcentaje de compatibilidad, amigos en común
// y motivo de la recomendación estructurado para su consumo por el frontend.
type FriendSuggestion struct {
	IDUsuario              string   `json:"id_usuario"`
	Username               string   `json:"username"`
	Correo                 string   `json:"correo,omitempty"`
	IDNivel                *int     `json:"id_nivel,omitempty"`
	Experiencia            int      `json:"experiencia"`
	MatchScore             int      `json:"porcentaje_compatibilidad"`
	MutualFriends          int      `json:"amigos_en_comun"`
	CommonGenres           []string `json:"generos_comunes,omitempty"`
	CommonArtists          []string `json:"artistas_comunes,omitempty"`
	CommonTracks           []string `json:"canciones_comunes,omitempty"`
	CommonInteractedTracks []string `json:"canciones_interactuadas_comunes,omitempty"`
	Motivo                 string   `json:"motivo"`
}

// FriendSuggestionResponse representa el formato de respuesta estructurado para la API REST (RF-14).
type FriendSuggestionResponse = FriendSuggestion
