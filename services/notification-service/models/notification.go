package models

import (
	"encoding/json"
	"time"
)

type Event struct {
	Name            string
	ID              string
	EventID         string
	UserID          string
	ActorID         string
	RecipientID     string
	ReceiverID      string
	SenderID        string
	TargetUserID    string
	AffectedUserID  string
	ModeratorID     string
	PostID          string
	CommentID       string
	MessageID       string
	CommentAuthorID string
	TargetAuthorID  string
	TargetType      string
	TargetID        string
}

type Notification struct {
	ID             string
	EventID        string
	Recipient      string
	Actor          *string
	Type           string
	PreferenceType string
	TargetID       *string
	Data           []byte
}

type NotificationResponse struct {
	ID        string          `json:"id_notificacion"`
	Recipient string          `json:"id_usuario"`
	Actor     *string         `json:"id_emisor"`
	Type      string          `json:"tipo"`
	TargetID  *string         `json:"id_objetivo"`
	Read      bool            `json:"leida"`
	Count     int             `json:"contador"`
	EventID   *string         `json:"id_evento"`
	Data      json.RawMessage `json:"datos"`
	CreatedAt time.Time       `json:"fecha_creacion"`
}
