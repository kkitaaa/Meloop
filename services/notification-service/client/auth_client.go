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

var ErrInvalidSession = errors.New("invalid or expired session")

type AuthClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewAuthClient(baseURL string) *AuthClient {
	return &AuthClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (c *AuthClient) ValidateToken(ctx context.Context, token string) (string, error) {
	if c == nil || c.httpClient == nil || c.baseURL == "" {
		return "", errors.New("auth service is not configured")
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/auth/validate", nil)
	if err != nil {
		return "", fmt.Errorf("create auth validation request: %w", err)
	}
	request.Header.Set("Authorization", "Bearer "+token)

	response, err := c.httpClient.Do(request)
	if err != nil {
		return "", fmt.Errorf("call auth service: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusUnauthorized {
		return "", ErrInvalidSession
	}
	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("auth service returned status %d", response.StatusCode)
	}

	var result struct {
		Success bool `json:"success"`
		Data    struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(io.LimitReader(response.Body, 1<<20)).Decode(&result); err != nil {
		return "", fmt.Errorf("decode auth service response: %w", err)
	}
	if !result.Success || strings.TrimSpace(result.Data.ID) == "" {
		return "", errors.New("auth service returned no authenticated user")
	}
	return strings.TrimSpace(result.Data.ID), nil
}
