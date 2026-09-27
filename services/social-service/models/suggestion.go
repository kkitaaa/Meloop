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

// FriendSuggestion representa una sugerencia de amistad completa, con compatibilidad RF-15,
// amigos en común, señales y motivo principal.
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
