package models

import "time"

type Comment struct {
	ID              string    `json:"id"`
	PostID          string    `json:"post_id"`
	AuthorID        string    `json:"author_id"`
	Content         string    `json:"content"`
	ParentCommentID *string   `json:"parent_comment_id,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	Replies         []Comment `json:"replies,omitempty"`
}

type CreateCommentRequest struct {
	AuthorID string `json:"author_id"`
	Content  string `json:"content"`
}
