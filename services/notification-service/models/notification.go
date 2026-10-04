package models

import "encoding/json"

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

type NotificationRecord struct {
	ID        string          `json:"id"`
	SenderID  string          `json:"sender_id,omitempty"`
	Type      string          `json:"type"`
	TargetID  string          `json:"target_id,omitempty"`
	Read      bool            `json:"read"`
	Count     int             `json:"count"`
	CreatedAt string          `json:"created_at"`
	Data      json.RawMessage `json:"data"`
}

type NotificationPreference struct {
	Type    string `json:"type"`
	Enabled bool   `json:"enabled"`
}

const (
	NotificationTypePostLiked        = "post.liked"
	NotificationTypeCommentCreated   = "comment.created"
	NotificationTypeCommentCommented = "comment.commented"
	NotificationTypeCommentReplied   = "comment.replied"
	NotificationTypeFriendRequested  = "friend.requested"
	NotificationTypeFriendAccepted   = "friend.accepted"
	NotificationTypeMessageSent      = "message.sent"
	NotificationTypeReportCreated    = "report.created"
	NotificationTypeModerationAction = "moderation.action"
	NotificationTypeUserLevelUp      = "user.level_up"
	NotificationTypeRewardUnlocked   = "reward.unlocked"
)

var notificationTypes = []string{
	NotificationTypePostLiked,
	NotificationTypeCommentCreated,
	NotificationTypeCommentCommented,
	NotificationTypeCommentReplied,
	NotificationTypeFriendRequested,
	NotificationTypeFriendAccepted,
	NotificationTypeMessageSent,
	NotificationTypeReportCreated,
	NotificationTypeModerationAction,
	NotificationTypeUserLevelUp,
	NotificationTypeRewardUnlocked,
}

func SupportedNotificationTypes() []string {
	return append([]string(nil), notificationTypes...)
}

func IsSupportedNotificationType(value string) bool {
	for _, notificationType := range notificationTypes {
		if value == notificationType {
			return true
		}
	}
	return false
}
