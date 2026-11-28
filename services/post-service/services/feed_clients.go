package services

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type SocialFeedClient struct {
	baseURL string
	client  *http.Client
}

func NewSocialFeedClient(baseURL string) *SocialFeedClient {
	return &SocialFeedClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		client:  &http.Client{Timeout: 150 * time.Millisecond},
	}
}

func (client *SocialFeedClient) ExcludeBlocked(ctx context.Context, userID string, authorIDs []string) ([]string, error) {
	if len(authorIDs) == 0 {
		return []string{}, nil
	}

	payload, err := json.Marshal(struct {
		UserID       string   `json:"user_id"`
		CandidateIDs []string `json:"candidate_ids"`
	}{UserID: userID, CandidateIDs: authorIDs})
	if err != nil {
		return nil, fmt.Errorf("encode block-exclusion request: %w", err)
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, client.baseURL+"/internal/blocks/exclude", bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("create block-exclusion request: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := client.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("call social-service block exclusion: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("social-service block exclusion returned status %d", response.StatusCode)
	}

	var result struct {
		Success bool `json:"success"`
		Data    struct {
			CandidateIDs []string `json:"candidate_ids"`
		} `json:"data"`
	}
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode social-service block-exclusion response: %w", err)
	}
	if !result.Success || result.Data.CandidateIDs == nil {
		return nil, errors.New("social-service returned an invalid block-exclusion response")
	}
	return result.Data.CandidateIDs, nil
}

type HTTPFeedRecommendationProvider struct {
	baseURL string
	client  *http.Client
}

func NewHTTPFeedRecommendationProvider(baseURL string) *HTTPFeedRecommendationProvider {
	return &HTTPFeedRecommendationProvider{
		baseURL: strings.TrimRight(baseURL, "/"),
		client:  &http.Client{Timeout: 100 * time.Millisecond},
	}
}

func (client *HTTPFeedRecommendationProvider) GetPostRecommendations(ctx context.Context, userID string, limit int) ([]FeedRecommendation, error) {
	numericUserID, valid := parseRecommendationUserID(userID)
	if !valid {
		return []FeedRecommendation{}, nil
	}

	endpoint := fmt.Sprintf("%s/recommendations/%s?limit=%s", client.baseURL, strconv.Itoa(numericUserID), strconv.Itoa(limit))
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("create recommendation request: %w", err)
	}
	response, err := client.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("call recommendation-service: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("recommendation-service returned status %d", response.StatusCode)
	}

	var result struct {
		Recommendations []struct {
			ItemID  int     `json:"item_id"`
			ItemKey string  `json:"item_key"`
			Score   float64 `json:"score"`
		} `json:"recommendations"`
	}
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode recommendation-service response: %w", err)
	}
	recommendations := make([]FeedRecommendation, 0, len(result.Recommendations))
	for _, item := range result.Recommendations {
		postID := strings.TrimSpace(item.ItemKey)
		if postID == "" && item.ItemID > 0 {
			postID = strconv.Itoa(item.ItemID)
		}
		if postID != "" {
			recommendations = append(recommendations, FeedRecommendation{PostID: postID, Score: item.Score})
		}
	}
	return recommendations, nil
}
