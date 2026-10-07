package models

import "time"

type Comment struct {
	ID              string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	PostID          string    `json:"post_id" gorm:"not null"`
	AuthorID        string    `json:"author_id" gorm:"not null"`
	Content         string    `json:"content" gorm:"type:text;not null"`
	ParentCommentID *string   `json:"parent_comment_id" gorm:"type:uuid;index"`
	CreatedAt       time.Time `json:"created_at" gorm:"autoCreateTime"`
}