package models

type FeedSnapshot struct {
	ID      string   `json:"id"`
	PostIDs []string `json:"post_ids"`
}
