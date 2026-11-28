package services

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type HTTPFriendshipValidator struct {
	baseURL string
	client  *http.Client
}

func NewHTTPFriendshipValidator(baseURL string, timeout time.Duration) *HTTPFriendshipValidator {
	if timeout <= 0 {
		timeout = 3 * time.Second
	}
	return &HTTPFriendshipValidator{
		baseURL: strings.TrimRight(baseURL, "/"),
		client:  &http.Client{Timeout: timeout},
	}
}

func (validator *HTTPFriendshipValidator) ValidateFriendship(ctx context.Context, _, friendID, authorization string) error {
	payload, err := json.Marshal(struct {
		FriendID string `json:"friend_id"`
	}{FriendID: friendID})
	if err != nil {
		return fmt.Errorf("%w: encode friendship validation request: %v", ErrSharingUnavailable, err)
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, validator.baseURL+"/friends/validate-friendship", bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("%w: create friendship validation request: %v", ErrSharingUnavailable, err)
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", authorization)
	response, err := validator.client.Do(request)
	if err != nil {
		return fmt.Errorf("%w: call social-service: %v", ErrSharingUnavailable, err)
	}
	defer response.Body.Close()
	if response.StatusCode == http.StatusForbidden || response.StatusCode == http.StatusNotFound {
		return ErrNotFriend
	}
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("%w: social-service returned status %d", ErrSharingUnavailable, response.StatusCode)
	}

	var result struct {
		Success bool `json:"success"`
	}
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		return fmt.Errorf("%w: decode friendship validation response: %v", ErrSharingUnavailable, err)
	}
	if !result.Success {
		return errors.Join(ErrSharingUnavailable, errors.New("social-service returned an unsuccessful response"))
	}
	return nil
}

type HTTPPostShareSender struct {
	baseURL string
	token   string
	client  *http.Client
}

func NewHTTPPostShareSender(baseURL, token string, timeout time.Duration) *HTTPPostShareSender {
	if timeout <= 0 {
		timeout = 3 * time.Second
	}
	return &HTTPPostShareSender{
		baseURL: strings.TrimRight(baseURL, "/"),
		token:   strings.TrimSpace(token),
		client:  &http.Client{Timeout: timeout},
	}
}

func (sender *HTTPPostShareSender) SendPostShare(ctx context.Context, senderID, friendID, postID string) (string, error) {
	if sender.baseURL == "" || sender.token == "" {
		return "", ErrSharingUnavailable
	}
	payload, err := json.Marshal(struct {
		SenderID    string `json:"sender_id"`
		RecipientID string `json:"recipient_id"`
		PostID      string `json:"post_id"`
	}{SenderID: senderID, RecipientID: friendID, PostID: postID})
	if err != nil {
		return "", fmt.Errorf("%w: encode chat message request: %v", ErrSharingUnavailable, err)
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, sender.baseURL+"/internal/messages/share", bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("%w: create chat message request: %v", ErrSharingUnavailable, err)
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+sender.token)
	response, err := sender.client.Do(request)
	if err != nil {
		return "", fmt.Errorf("%w: call chat-service: %v", ErrSharingUnavailable, err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("%w: chat-service returned status %d", ErrSharingUnavailable, response.StatusCode)
	}
	var result struct {
		Success bool `json:"success"`
		Data    struct {
			MessageID string `json:"message_id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("%w: decode chat-service response: %v", ErrSharingUnavailable, err)
	}
	if !result.Success || result.Data.MessageID == "" {
		return "", fmt.Errorf("%w: chat-service returned an invalid message response", ErrSharingUnavailable)
	}
	return result.Data.MessageID, nil
}
