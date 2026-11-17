package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

var ErrUnauthenticated = errors.New("unauthenticated session")

type HTTPIdentityResolver struct {
	authServiceURL string
	client         *http.Client
}

func NewHTTPIdentityResolver(authServiceURL string, client *http.Client) *HTTPIdentityResolver {
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}
	return &HTTPIdentityResolver{
		authServiceURL: strings.TrimRight(authServiceURL, "/"),
		client:         client,
	}
}

func (r *HTTPIdentityResolver) Resolve(ctx context.Context, authorization string) (string, error) {
	if strings.TrimSpace(authorization) == "" {
		return "", ErrUnauthenticated
	}
	request, err := http.NewRequestWithContext(
		ctx, http.MethodGet, r.authServiceURL+"/auth/validate", nil,
	)
	if err != nil {
		return "", fmt.Errorf("create auth-service request: %w", err)
	}
	request.Header.Set("Authorization", authorization)

	response, err := r.client.Do(request)
	if err != nil {
		return "", fmt.Errorf("validate report session: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode == http.StatusUnauthorized || response.StatusCode == http.StatusForbidden {
		return "", ErrUnauthenticated
	}
	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("auth-service returned HTTP %d", response.StatusCode)
	}

	var result struct {
		Success bool `json:"success"`
		Data    struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode auth-service response: %w", err)
	}
	if !result.Success || strings.TrimSpace(result.Data.ID) == "" {
		return "", ErrUnauthenticated
	}
	return strings.TrimSpace(result.Data.ID), nil
}
