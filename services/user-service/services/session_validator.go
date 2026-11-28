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

var (
	ErrInvalidSession         = errors.New("invalid session")
	ErrAuthServiceUnavailable = errors.New("auth service unavailable")
)

type SessionValidator interface {
	ValidateSession(ctx context.Context, authorization string) (string, error)
}

type httpSessionValidator struct {
	authServiceURL string
	httpClient     *http.Client
}

func NewHTTPSessionValidator(authServiceURL string) SessionValidator {
	return &httpSessionValidator{
		authServiceURL: strings.TrimRight(authServiceURL, "/"),
		httpClient:     &http.Client{Timeout: 3 * time.Second},
	}
}

func (v *httpSessionValidator) ValidateSession(
	ctx context.Context,
	authorization string,
) (string, error) {
	parts := strings.Fields(authorization)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", ErrInvalidSession
	}

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		v.authServiceURL+"/auth/validate",
		nil,
	)
	if err != nil {
		return "", fmt.Errorf("%w: create validation request", ErrAuthServiceUnavailable)
	}
	request.Header.Set("Authorization", "Bearer "+parts[1])

	response, err := v.httpClient.Do(request)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrAuthServiceUnavailable, err)
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusUnauthorized || response.StatusCode == http.StatusForbidden {
		return "", ErrInvalidSession
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return "", fmt.Errorf("%w: auth service returned status %d", ErrAuthServiceUnavailable, response.StatusCode)
	}

	var result struct {
		Success bool `json:"success"`
		Data    struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(io.LimitReader(response.Body, 1<<20)).Decode(&result); err != nil {
		return "", fmt.Errorf("%w: invalid validation response", ErrAuthServiceUnavailable)
	}
	if !result.Success || strings.TrimSpace(result.Data.ID) == "" {
		return "", ErrInvalidSession
	}
	return strings.TrimSpace(result.Data.ID), nil
}
