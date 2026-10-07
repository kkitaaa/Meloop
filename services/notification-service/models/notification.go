package models

type Event struct {
	Name           string
	ID             string
	UserID         string
	ActorID        string
	RecipientID    string
	ReceiverID     string
	SenderID       string
	TargetUserID   string
	AffectedUserID string
	ModeratorID    string
	PostID         string
	CommentID      string
	MessageID      string
}

type Notification struct {
	ID        string
	EventID   string
	Recipient string
	Actor     *string
	Type      string
	TargetID  *string
	Data      []byte
}
