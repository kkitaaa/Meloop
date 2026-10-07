package models

import (
	"encoding/json"
	"strings"
)

func (e *Event) UnmarshalJSON(data []byte) error {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	values := make(map[string]string, len(fields))
	for key, raw := range fields {
		var value string
		if err := json.Unmarshal(raw, &value); err != nil {
			continue
		}
		values[normalizeField(key)] = value
	}
	get := func(keys ...string) string {
		for _, key := range keys {
			if value := values[normalizeField(key)]; value != "" {
				return value
			}
		}
		return ""
	}

	e.Name = get("event", "eventType", "type")
	e.ID = get("eventId", "id")
	e.UserID = get("userId")
	e.ActorID = get("actorId")
	e.RecipientID = get("recipientId")
	e.ReceiverID = get("receiverId")
	e.SenderID = get("senderId")
	e.TargetUserID = get("targetUserId")
	e.AffectedUserID = get("affectedUserId")
	e.ModeratorID = get("moderatorId")
	e.PostID = get("postId")
	e.CommentID = get("commentId", "interactionId", "replyToId", "parentCommentId", "targetCommentId")
	e.MessageID = get("messageId")
	return nil
}

func normalizeField(value string) string {
	return strings.NewReplacer("_", "", "-", "").Replace(strings.ToLower(value))
}
