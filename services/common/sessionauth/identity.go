package sessionauth

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
	ErrUnauthorized = errors.New("session is invalid or missing")
	ErrUnavailable  = errors.New("authentication service is unavailable")
)

type sessionResponse struct {
	Success bool `json:"success"`
	Data    struct {
		ID string `json:"id"`
	} `json:"data"`
}

var client = &http.Client{Timeout: 3 * time.Second}

func UserID(ctx context.Context, authServiceURL, authorization string) (string, error) {
	if strings.TrimSpace(authorization) == "" {
		return "", ErrUnauthorized
	}
	validationURL := strings.TrimRight(authServiceURL, "/") + "/auth/validate"
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, validationURL, nil)
	if err != nil {
		return "", fmt.Errorf("build auth validation request: %w", err)
	}
	request.Header.Set("Authorization", authorization)
	response, err := client.Do(request)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	defer response.Body.Close()
	if response.StatusCode >= http.StatusInternalServerError {
		return "", ErrUnavailable
	}
	if response.StatusCode != http.StatusOK {
		return "", ErrUnauthorized
	}
	var result sessionResponse
	if err := json.NewDecoder(io.LimitReader(response.Body, 1<<20)).Decode(&result); err != nil {
		return "", fmt.Errorf("decode auth validation response: %w", err)
	}
	if !result.Success || strings.TrimSpace(result.Data.ID) == "" {
		return "", ErrUnauthorized
	}
	return result.Data.ID, nil
}
