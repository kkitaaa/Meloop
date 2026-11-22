package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

var (
	ErrInvalidSession         = errors.New("invalid or expired session")
	ErrAuthServiceUnavailable = errors.New("auth-service validation unavailable")
)

const authServiceTimeout = 5 * time.Second

type AuthClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewAuthClient(baseURL string, httpClients ...*http.Client) *AuthClient {
	httpClient := &http.Client{Timeout: authServiceTimeout}
	if len(httpClients) > 0 && httpClients[0] != nil {
		httpClient = httpClients[0]
	}
	return &AuthClient{
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: httpClient,
	}
}

func (c *AuthClient) ValidateToken(ctx context.Context, token string) (string, error) {
	if c == nil || c.httpClient == nil || c.baseURL == "" {
		return "", fmt.Errorf("%w: auth-service is not configured", ErrAuthServiceUnavailable)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/auth/validate", nil)
	if err != nil {
		return "", fmt.Errorf("%w: could not create validation request", ErrAuthServiceUnavailable)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("%w: validation request failed: %v", ErrAuthServiceUnavailable, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return "", ErrInvalidSession
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%w: auth-service returned status %d", ErrAuthServiceUnavailable, resp.StatusCode)
	}

	var result struct {
		Success bool `json:"success"`
		Data    struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&result); err != nil {
		return "", fmt.Errorf("%w: auth-service returned an invalid response", ErrAuthServiceUnavailable)
	}
	userID := strings.TrimSpace(result.Data.ID)
	if !result.Success || userID == "" {
		return "", fmt.Errorf("%w: auth-service returned no authenticated user", ErrAuthServiceUnavailable)
	}
	return userID, nil
}
