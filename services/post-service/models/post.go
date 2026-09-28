package models

import "time"

type Post struct {
	ID        string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	AuthorID  string    `json:"author_id" gorm:"not null"`
	Content   string    `json:"content" gorm:"type:text"`
	MediaURL  string    `json:"media_url"` // Guardará la URL de MinIO
	MusicID   string    `json:"music_id"`  // Referencia a la música validada
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}

type CreatePostRequest struct {
	AuthorID string `form:"author_id" binding:"required"`
	Content  string `form:"content"`
	MusicID  string `form:"music_id"`
	// El archivo vendrá en el multipart form
}