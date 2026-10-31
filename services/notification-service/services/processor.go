package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/meloop/notification-service/models"
)

type Repository interface {
	ResolvePostOwner(context.Context, string) (string, error)
	ResolveInteractionOwner(context.Context, string) (string, error)
	Create(context.Context, models.Notification) error
}

type Processor struct {
	repository Repository
}

const (
	eventPostLiked        = "post.liked"
	eventCommentCreated   = "comment.created"
	eventCommentCommented = "comment.commented"
	eventCommentReplied   = "comment.replied"
	eventFriendRequested  = "friend.requested"
	eventFriendAccepted   = "friend.accepted"
	eventMessageSent      = "message.sent"
	eventReportCreated    = "report.created"
	eventModerationAction = "moderation.action"
	eventUserLevelUp      = "user.level_up"
	eventRewardUnlocked   = "reward.unlocked"
)

type notificationFields struct {
	actor     string
	recipient string
	targetID  string
}

type permanentError struct{ err error }

func (e permanentError) Error() string { return e.err.Error() }
func (e permanentError) Unwrap() error { return e.err }

func IsPermanent(err error) bool {
	var target permanentError
	return errors.As(err, &target)
}

func NewProcessor(repository Repository) *Processor {
	return &Processor{repository: repository}
}

func (p *Processor) Process(ctx context.Context, routingKey, messageID string, body []byte) error {
	var event models.Event
	if err := json.Unmarshal(body, &event); err != nil {
		return permanentError{fmt.Errorf("decode event: %w", err)}
	}

	typeName := normalizeEventName(routingKey)
	if typeName == "" {
		typeName = normalizeEventName(event.Name)
	}
	if !supportedEvents[typeName] {
		return permanentError{fmt.Errorf("unsupported notification event %q", typeName)}
	}

	fields, err := p.resolveFields(ctx, typeName, event)
	if err != nil {
		return err
	}
	if fields.actor == fields.recipient {
		return nil
	}
	eventID := firstNonEmpty(messageID, event.ID)
	if eventID == "" {
		digest := sha256.Sum256(append([]byte(typeName+":"), body...))
		eventID = hex.EncodeToString(digest[:])
	}

	notification := models.Notification{
		EventID:   eventID,
		Recipient: fields.recipient,
		Type:      typeName,
		Data:      body,
	}
	if fields.actor != "" {
		notification.Actor = &fields.actor
	}
	if fields.targetID != "" {
		notification.TargetID = &fields.targetID
	}
	return p.repository.Create(ctx, notification)
}

func (p *Processor) resolveFields(ctx context.Context, typeName string, event models.Event) (notificationFields, error) {
	fields := notificationFields{
		actor:     firstNonEmpty(event.ActorID, event.ModeratorID),
		recipient: firstNonEmpty(event.RecipientID, event.TargetUserID, event.AffectedUserID),
	}
	switch typeName {
	case eventPostLiked, eventCommentCreated, eventCommentCommented:
		fields.actor = firstNonEmpty(event.ActorID, event.UserID)
		fields.targetID = event.PostID
		return p.resolveOwner(ctx, fields, event.PostID, p.repository.ResolvePostOwner, "post")
	case eventCommentReplied:
		fields.actor = firstNonEmpty(event.ActorID, event.UserID)
		fields.targetID = event.CommentID
		return p.resolveOwner(ctx, fields, event.CommentID, p.repository.ResolveInteractionOwner, "comment")
	case eventFriendRequested:
		fields.actor = firstNonEmpty(event.SenderID, fields.actor)
		fields.recipient = firstNonEmpty(fields.recipient, event.ReceiverID)
		fields.targetID = event.SenderID
	case eventFriendAccepted:
		fields.actor = firstNonEmpty(event.ReceiverID, fields.actor)
		fields.recipient = firstNonEmpty(fields.recipient, event.SenderID)
		fields.targetID = event.SenderID
	case eventMessageSent:
		fields.actor = firstNonEmpty(event.SenderID, fields.actor)
		fields.recipient = firstNonEmpty(fields.recipient, event.ReceiverID)
		fields.targetID = event.MessageID
	case eventUserLevelUp, eventRewardUnlocked:
		fields.recipient = firstNonEmpty(fields.recipient, event.UserID)
	case eventReportCreated, eventModerationAction:
		fields.actor = firstNonEmpty(event.ModeratorID, fields.actor)
	}
	if fields.recipient == "" {
		return notificationFields{}, permanentError{fmt.Errorf("event %q is missing recipient user ID", typeName)}
	}
	return fields, nil
}

func (p *Processor) resolveOwner(
	ctx context.Context,
	fields notificationFields,
	objectID string,
	resolve func(context.Context, string) (string, error),
	objectType string,
) (notificationFields, error) {
	if fields.actor == "" {
		return notificationFields{}, permanentError{errors.New("event is missing actor user ID")}
	}
	if fields.recipient == "" {
		if objectID == "" {
			return notificationFields{}, permanentError{fmt.Errorf("event is missing %s ID or recipient ID", objectType)}
		}
		recipient, err := resolve(ctx, objectID)
		if err != nil {
			return notificationFields{}, fmt.Errorf("resolve %s owner: %w", objectType, err)
		}
		fields.recipient = recipient
	}
	if fields.recipient == "" {
		return notificationFields{}, permanentError{fmt.Errorf("event is missing %s owner", objectType)}
	}
	return fields, nil
}

var supportedEvents = map[string]bool{
	eventPostLiked:        true,
	eventCommentCreated:   true,
	eventCommentCommented: true,
	eventCommentReplied:   true,
	eventFriendRequested:  true,
	eventFriendAccepted:   true,
	eventMessageSent:      true,
	eventReportCreated:    true,
	eventModerationAction: true,
	eventUserLevelUp:      true,
	eventRewardUnlocked:   true,
}

func normalizeEventName(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.NewReplacer("_", ".", "-", ".", " ", ".").Replace(value)
	switch value {
	case "postliked":
		return "post.liked"
	case "commentcreated":
		return "comment.created"
	case "friendrequested":
		return "friend.requested"
	case "friendaccepted":
		return "friend.accepted"
	case "user.level.up":
		return eventUserLevelUp
	case "userlevelup":
		return eventUserLevelUp
	case "rewardunlocked":
		return eventRewardUnlocked
	default:
		return value
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			return value
		}
	}
	return ""
}
