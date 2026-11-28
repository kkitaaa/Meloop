package services

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHTTPFriendshipValidatorForwardsSessionAndFriendID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost || request.URL.Path != "/friends/validate-friendship" {
			t.Errorf("request = %s %s", request.Method, request.URL.Path)
		}
		if request.Header.Get("Authorization") != "Bearer session-token" {
			t.Errorf("Authorization = %q", request.Header.Get("Authorization"))
		}
		body, err := io.ReadAll(request.Body)
		if err != nil {
			t.Errorf("read request body: %v", err)
		}
		if string(body) != `{"friend_id":"friend-1"}` {
			t.Errorf("request body = %s", body)
		}
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"success":true,"data":{"is_friend":true}}`))
	}))
	defer server.Close()

	validator := NewHTTPFriendshipValidator(server.URL, 0)
	if err := validator.ValidateFriendship(context.Background(), "sender-1", "friend-1", "Bearer session-token"); err != nil {
		t.Fatalf("ValidateFriendship() error = %v", err)
	}
}

func TestHTTPFriendshipValidatorRejectsBlockedOrNonFriend(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusForbidden)
	}))
	defer server.Close()

	err := NewHTTPFriendshipValidator(server.URL, 0).
		ValidateFriendship(context.Background(), "sender-1", "blocked-1", "Bearer token")
	if !errors.Is(err, ErrNotFriend) {
		t.Fatalf("error = %v, want ErrNotFriend", err)
	}
}

func TestHTTPPostShareSenderUsesInternalChatContract(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost || request.URL.Path != "/internal/messages/share" {
			t.Errorf("request = %s %s", request.Method, request.URL.Path)
		}
		if request.Header.Get("Authorization") != "Bearer chat-secret" {
			t.Errorf("Authorization = %q", request.Header.Get("Authorization"))
		}
		body, err := io.ReadAll(request.Body)
		if err != nil {
			t.Errorf("read request body: %v", err)
		}
		if string(body) != `{"sender_id":"sender-1","recipient_id":"friend-1","post_id":"post-1"}` {
			t.Errorf("request body = %s", body)
		}
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusCreated)
		_, _ = writer.Write([]byte(`{"success":true,"data":{"message_id":"message-1"}}`))
	}))
	defer server.Close()

	sender := NewHTTPPostShareSender(server.URL, "chat-secret", 0)
	messageID, err := sender.SendPostShare(context.Background(), "sender-1", "friend-1", "post-1")
	if err != nil {
		t.Fatalf("SendPostShare() error = %v", err)
	}
	if messageID != "message-1" {
		t.Fatalf("message ID = %q, want message-1", messageID)
	}
}

func TestHTTPPostShareSenderReportsUnavailableChat(t *testing.T) {
	sender := NewHTTPPostShareSender("http://chat.test", "chat-secret", 1)
	_, err := sender.SendPostShare(context.Background(), "sender-1", "friend-1", "post-1")
	if !errors.Is(err, ErrSharingUnavailable) {
		t.Fatalf("error = %v, want ErrSharingUnavailable", err)
	}
	if !strings.Contains(err.Error(), "call chat-service") {
		t.Fatalf("error should identify the failed dependency: %v", err)
	}
}
