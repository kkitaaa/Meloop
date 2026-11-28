package models

// MusicSearchItem representa un resultado musical normalizado para el ecosistema Meloop
type MusicSearchItem struct {
	Type        string   `json:"type"`
	SpotifyID   string   `json:"spotify_id"`
	Name        string   `json:"name"`
	Artists     []string `json:"artists"`
	Album       string   `json:"album"`
	ImageURL    string   `json:"image_url"`
	ExternalURL string   `json:"external_url"`
}

// MusicSearchResponse estructura la respuesta paginada de búsqueda musical
type MusicSearchResponse struct {
	Results []MusicSearchItem `json:"results"`
	Limit   int               `json:"limit"`
	Offset  int               `json:"offset"`
	Total   int               `json:"total"`
}

// ErrorResponse estandariza las respuestas de error HTTP del servicio
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}
