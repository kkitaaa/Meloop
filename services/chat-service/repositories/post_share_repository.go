package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

type PostShareRepository interface {
	CreatePostShare(ctx context.Context, senderID, recipientID, postID string) (string, error)
}

type SQLPostShareRepository struct {
	db *sql.DB
}

func NewSQLPostShareRepository(db *sql.DB) *SQLPostShareRepository {
	return &SQLPostShareRepository{db: db}
}

func (repository *SQLPostShareRepository) CreatePostShare(ctx context.Context, senderID, recipientID, postID string) (string, error) {
	content, err := json.Marshal(struct {
		Type   string `json:"type"`
		PostID string `json:"post_id"`
	}{Type: "post_share", PostID: postID})
	if err != nil {
		return "", fmt.Errorf("encode post-share message: %w", err)
	}

	messageID := uuid.NewString()
	_, err = repository.db.ExecContext(ctx, `
		INSERT INTO mensaje_chat (
			id_mensaje, id_emisor, id_receptor, contenido, fecha_envio, estado
		)
		VALUES ($1, $2, $3, $4, NOW(), 'ENVIADO')
	`, messageID, senderID, recipientID, string(content))
	if err != nil {
		return "", fmt.Errorf("persist post-share message: %w", err)
	}
	return messageID, nil
}
