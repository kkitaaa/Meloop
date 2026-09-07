package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/meloop/recommendation-service/models"
)

type MLClient struct {
	baseURL string
	client  *http.Client
}

func NewMLClient(baseURL string) *MLClient {
	return &MLClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		client:  &http.Client{Timeout: 5 * time.Second},
	}
}

func (client *MLClient) Predict(ctx context.Context, request models.RecommendationRequest) (models.RecommendationResponse, error) {
	body, err := json.Marshal(request)
	if err != nil {
		return models.RecommendationResponse{}, fmt.Errorf("encode ML request: %w", err)
	}
	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodPost, client.baseURL+"/predict", bytes.NewReader(body))
	if err != nil {
		return models.RecommendationResponse{}, fmt.Errorf("create ML request: %w", err)
	}
	httpRequest.Header.Set("Content-Type", "application/json")

	response, err := client.client.Do(httpRequest)
	if err != nil {
		return models.RecommendationResponse{}, fmt.Errorf("call ML service: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return models.RecommendationResponse{}, fmt.Errorf("ML service returned status %d", response.StatusCode)
	}

	var recommendation models.RecommendationResponse
	if err := json.NewDecoder(response.Body).Decode(&recommendation); err != nil {
		return models.RecommendationResponse{}, fmt.Errorf("decode ML response: %w", err)
	}
	return recommendation, nil
}
