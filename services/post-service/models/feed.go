package models

import "time"

// FeedPost representa la estructura exacta que el frontend necesita para renderizar el feed
type FeedPost struct {
	ID             string    `json:"id"`
	Content        string    `json:"content"`
	MediaURL       string    `json:"media_url"` // Enlace a MinIO
	CreatedAt      time.Time `json:"created_at"`
	AuthorID       string    `json:"author_id"`
	AuthorUsername string    `json:"author_username"`
	AuthorAvatar   string    `json:"author_avatar"`
	LikesCount     int       `json:"likes_count"`
	CommentsCount  int       `json:"comments_count"`
}