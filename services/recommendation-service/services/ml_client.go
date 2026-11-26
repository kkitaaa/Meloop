package services

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/meloop/recommendation-service/models"
)

const (
	mlRequestTimeout = 5 * time.Second
	mlRetryCount     = 2
	mlRetryBackoff   = 100 * time.Millisecond
)

var ErrMLCircuitOpen = errors.New("ML service circuit is open")

type MLClient struct {
	baseURL string
	client  *http.Client
	breaker *mlCircuitBreaker
}

type mlCircuitBreaker struct {
	mu            sync.Mutex
	failures      int
	threshold     int
	resetTimeout  time.Duration
	openedAt      time.Time
	probeInFlight bool
	logger        *slog.Logger
}

func NewMLClient(baseURL string) *MLClient {
	return &MLClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		client:  &http.Client{Timeout: mlRequestTimeout},
		breaker: &mlCircuitBreaker{
			threshold:    3,
			resetTimeout: 30 * time.Second,
			logger:       slog.Default(),
		},
	}
}

func (client *MLClient) Ready(ctx context.Context) error {
	return client.call(ctx, func(requestCtx context.Context) error {
		_, err := client.doWithRetry(requestCtx, http.MethodGet, client.baseURL+"/ready", nil)
		if err != nil {
			return fmt.Errorf("check ML service readiness: %w", err)
		}
		return nil
	})
}

func (client *MLClient) Predict(ctx context.Context, request models.RecommendationRequest) (models.RecommendationResponse, error) {
	if request.Preferences == nil {
		request.Preferences = []string{}
	}
	if request.Profile.Genres == nil {
		request.Profile.Genres = []string{}
	}
	if request.Profile.Artists == nil {
		request.Profile.Artists = []string{}
	}
	if request.Profile.Songs == nil {
		request.Profile.Songs = []string{}
	}
	if request.Interactions == nil {
		request.Interactions = []models.Interaction{}
	}

	body, err := json.Marshal(request)
	if err != nil {
		return models.RecommendationResponse{}, fmt.Errorf("encode ML request: %w", err)
	}

	var recommendation models.RecommendationResponse
	err = client.call(ctx, func(requestCtx context.Context) error {
		responseBody, err := client.doWithRetry(requestCtx, http.MethodPost, client.baseURL+"/predict", body)
		if err != nil {
			return fmt.Errorf("call ML service: %w", err)
		}
		if err := json.Unmarshal(responseBody, &recommendation); err != nil {
			return fmt.Errorf("decode ML response: %w", err)
		}
		return nil
	})
	if err != nil {
		return models.RecommendationResponse{}, err
	}
	return recommendation, nil
}

func (client *MLClient) call(parentCtx context.Context, operation func(context.Context) error) error {
	if !client.breaker.allow() {
		return ErrMLCircuitOpen
	}

	ctx, cancel := context.WithTimeout(parentCtx, mlRequestTimeout)
	defer cancel()
	err := operation(ctx)
	if err == nil {
		client.breaker.recordSuccess()
		return nil
	}
	if parentCtx.Err() != nil {
		client.breaker.abandonProbe()
		return err
	}
	client.breaker.recordFailure()
	return err
}

func (client *MLClient) doWithRetry(ctx context.Context, method, url string, body []byte) ([]byte, error) {
	var lastErr error
	for attempt := 0; attempt <= mlRetryCount; attempt++ {
		var requestBody io.Reader
		if body != nil {
			requestBody = bytes.NewReader(body)
		}
		request, err := http.NewRequestWithContext(ctx, method, url, requestBody)
		if err != nil {
			return nil, fmt.Errorf("create ML request: %w", err)
		}
		if body != nil {
			request.Header.Set("Content-Type", "application/json")
		}

		response, err := client.client.Do(request)
		if err != nil {
			lastErr = fmt.Errorf("perform ML request: %w", err)
			if ctx.Err() != nil {
				return nil, lastErr
			}
		} else {
			responseBody, readErr := io.ReadAll(response.Body)
			_ = response.Body.Close()
			if readErr != nil {
				lastErr = fmt.Errorf("read ML response: %w", readErr)
			} else if response.StatusCode != http.StatusOK {
				lastErr = mlStatusError(response.StatusCode, responseBody)
				if response.StatusCode < http.StatusInternalServerError && response.StatusCode != http.StatusTooManyRequests {
					return nil, lastErr
				}
			} else {
				return responseBody, nil
			}
		}

		if attempt < mlRetryCount {
			delay := mlRetryBackoff * time.Duration(1<<attempt)
			timer := time.NewTimer(delay)
			select {
			case <-ctx.Done():
				if !timer.Stop() {
					<-timer.C
				}
				return nil, ctx.Err()
			case <-timer.C:
			}
		}
	}
	return nil, fmt.Errorf("ML request failed after %d attempts: %w", mlRetryCount+1, lastErr)
}

func mlStatusError(status int, responseBody []byte) error {
	message := fmt.Sprintf("ML service returned status %d", status)
	if status != http.StatusUnprocessableEntity {
		return errors.New(message)
	}

	var response struct {
		Detail []struct {
			Type string `json:"type"`
			Loc  []any  `json:"loc"`
		} `json:"detail"`
	}
	if err := json.Unmarshal(responseBody, &response); err != nil || len(response.Detail) == 0 {
		return errors.New(message)
	}

	issues := make([]string, 0, len(response.Detail))
	for _, detail := range response.Detail {
		location := make([]string, len(detail.Loc))
		for index, part := range detail.Loc {
			location[index] = fmt.Sprint(part)
		}
		issues = append(issues, fmt.Sprintf("%s (%s)", strings.Join(location, "."), detail.Type))
	}
	return fmt.Errorf("%s: validation failed at %s", message, strings.Join(issues, ", "))
}

func (breaker *mlCircuitBreaker) allow() bool {
	breaker.mu.Lock()
	defer breaker.mu.Unlock()

	if breaker.openedAt.IsZero() {
		return true
	}
	if time.Since(breaker.openedAt) < breaker.resetTimeout || breaker.probeInFlight {
		return false
	}
	breaker.probeInFlight = true
	return true
}

func (breaker *mlCircuitBreaker) recordSuccess() {
	breaker.mu.Lock()
	defer breaker.mu.Unlock()

	wasOpen := !breaker.openedAt.IsZero()
	breaker.failures = 0
	breaker.openedAt = time.Time{}
	breaker.probeInFlight = false
	if wasOpen {
		breaker.logger.Info("ml_circuit_closed", "event", "ml_circuit_closed")
	}
}

func (breaker *mlCircuitBreaker) recordFailure() {
	breaker.mu.Lock()
	defer breaker.mu.Unlock()

	breaker.probeInFlight = false
	breaker.failures++
	if breaker.failures >= breaker.threshold {
		breaker.openedAt = time.Now()
		breaker.logger.Warn("ml_circuit_opened", "event", "ml_circuit_opened", "failures", breaker.failures, "retry_after", breaker.resetTimeout)
	}
}

func (breaker *mlCircuitBreaker) abandonProbe() {
	breaker.mu.Lock()
	breaker.probeInFlight = false
	breaker.mu.Unlock()
}
