package messaging

import (
	"encoding/json"
	"regexp"
	"testing"

	"github.com/meloop/social-service/models"
)

func TestNewFriendRequestCancelledEventUsesContractAndUniqueUUID(t *testing.T) {
	request := &models.FriendRequest{
		ID: "request-123", SenderID: "sender-1", ReceiverID: "receiver-2",
		Status: models.FriendshipCanceled,
	}

	first, err := newFriendRequestCancelledEvent(request)
	if err != nil {
		t.Fatalf("newFriendRequestCancelledEvent() error = %v", err)
	}
	second, err := newFriendRequestCancelledEvent(request)
	if err != nil {
		t.Fatalf("newFriendRequestCancelledEvent() error = %v", err)
	}

	if !regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`).MatchString(first.EventID) {
		t.Fatalf("event_id %q is not a UUID v4", first.EventID)
	}
	if first.EventID == second.EventID {
		t.Fatal("event_id must be unique for each event")
	}
	if first.EventType != friendRequestCancelledRoutingKey ||
		first.RequestID != request.ID ||
		first.SenderID != request.SenderID ||
		first.ReceiverID != request.ReceiverID ||
		first.Reason != "user_blocked" {
		t.Fatalf("unexpected event: %+v", first)
	}

	body, err := json.Marshal(first)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(body, &fields); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	wantFields := []string{"event_id", "event_type", "request_id", "sender_id", "receiver_id", "reason"}
	if len(fields) != len(wantFields) {
		t.Fatalf("JSON fields = %v, want exactly %v", fields, wantFields)
	}
	for _, field := range wantFields {
		if _, ok := fields[field]; !ok {
			t.Errorf("JSON is missing %q", field)
		}
	}
}

func TestNewFriendRequestCancelledEventRejectsNilRequest(t *testing.T) {
	if _, err := newFriendRequestCancelledEvent(nil); err == nil {
		t.Fatal("expected an error for a nil request")
	}
}
