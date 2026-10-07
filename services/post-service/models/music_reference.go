package models

// ReferenceType define un tipo estricto para evitar strings mágicos
type ReferenceType string

const (
	TrackReference  ReferenceType = "track"
	ArtistReference ReferenceType = "artist"
	AlbumReference  ReferenceType = "album"
)

// MusicReference representa la entidad musical (canción, artista, álbum) enlazada al post
type MusicReference struct {
	Provider      string        `json:"provider"`                 // ej: "spotify", "youtube"
	ReferenceID   string        `json:"reference_id"`             // ej: el spotify_id
	ReferenceType ReferenceType `json:"reference_type"`           // track, artist, album
	Title         string        `json:"title,omitempty"`          // Metadato opcional para caché/UI rápida
	CoverURL      string        `json:"cover_url,omitempty"`      // Metadato opcional
}

// CreatePostRequest es el payload esperado en el endpoint HTTP/REST
type CreatePostRequest struct {
	UserID          string           `json:"user_id" binding:"required"`
	Content         string           `json:"content"`
	MusicReferences []MusicReference `json:"music_references"` // Arreglo para soportar múltiples referencias
}