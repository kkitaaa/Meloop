package services

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

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/auth/validate", nil)
	if err != nil {
		return "", errors.New("could not create auth validation request")
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", errors.New("auth service request failed")
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return "", ErrInvalidSession
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("auth service returned status %d", resp.StatusCode)
	}

	var result struct {
		Success bool `json:"success"`
		Data    struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&result); err != nil {
		return "", errors.New("auth service returned an invalid response")
	}
	if !result.Success || strings.TrimSpace(result.Data.ID) == "" {
		return "", errors.New("auth service returned no authenticated user")
	}

	return strings.TrimSpace(result.Data.ID), nil
}
