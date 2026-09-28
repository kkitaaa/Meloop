package models

const (
	WeightGenres       = 0.30
	WeightArtists      = 0.30
	WeightTracks       = 0.30
	WeightInteractions = 0.10
)

const (
	PreferenceTypeGenre  = "GENERO"
	PreferenceTypeArtist = "ARTISTA"
	PreferenceTypeTrack  = "CANCION"
)

// UserMusicalData representa el perfil y preferencias musicales recopiladas de un usuario.
type UserMusicalData struct {
	UserID           string   `json:"id_usuario"`
	Genres           []string `json:"generos"`
	Artists          []string `json:"artistas"`
	Tracks           []string `json:"canciones"`
	InteractedTracks []string `json:"canciones_interactuadas"`
}

// CompatibilityScore representa el resultado detallado del cálculo de compatibilidad musical entre dos usuarios.
type CompatibilityScore struct {
	UserAID                string   `json:"id_usuario_a"`
	UserBID                string   `json:"id_usuario_b"`
	MatchScore             int      `json:"porcentaje_compatibilidad"`
	GenreSimilarity        float64  `json:"similitud_generos"`
	ArtistSimilarity       float64  `json:"similitud_artistas"`
	TrackSimilarity        float64  `json:"similitud_canciones"`
	InteractionSimilarity  float64  `json:"similitud_interacciones"`
	CommonGenres           []string `json:"generos_comunes"`
	CommonArtists          []string `json:"artistas_comunes"`
	CommonTracks           []string `json:"canciones_comunes"`
	CommonInteractedTracks []string `json:"canciones_interactuadas_comunes"`
}
