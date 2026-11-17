package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var ErrSocialCounterUnavailable = errors.New("social counter unavailable")

type SocialFriendCounter interface {
	GetFriendsCount(ctx context.Context, userID string) (int, error)
}

type httpSocialFriendCounter struct {
	baseURL string
	token   string
	client  *http.Client
}

func NewHTTPSocialFriendCounter(baseURL, token string) SocialFriendCounter {
	return &httpSocialFriendCounter{
		baseURL: strings.TrimRight(strings.TrimSpace(baseURL), "/"),
		token:   strings.TrimSpace(token),
		client:  &http.Client{Timeout: 5 * time.Second},
	}
}

func (c *httpSocialFriendCounter) GetFriendsCount(ctx context.Context, userID string) (int, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" || c.baseURL == "" || c.token == "" {
		return 0, fmt.Errorf("%w: service URL, token, and user ID are required", ErrSocialCounterUnavailable)
	}

	endpoint := fmt.Sprintf("%s/internal/users/%s/friends/count", c.baseURL, url.PathEscape(userID))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return 0, fmt.Errorf("%w: create request: %v", ErrSocialCounterUnavailable, err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("%w: request failed: %v", ErrSocialCounterUnavailable, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("%w: social-service returned HTTP %d", ErrSocialCounterUnavailable, resp.StatusCode)
	}

	var result struct {
		Success bool `json:"success"`
		Data    struct {
			Count int `json:"count"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("%w: decode response: %v", ErrSocialCounterUnavailable, err)
	}
	if !result.Success || result.Data.Count < 0 {
		return 0, fmt.Errorf("%w: social-service returned an invalid count response", ErrSocialCounterUnavailable)
	}
	return result.Data.Count, nil
}
