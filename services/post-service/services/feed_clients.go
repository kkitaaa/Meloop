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

func NewSocialFeedClient(baseURL string, timeouts ...time.Duration) *SocialFeedClient {
	timeout := 100 * time.Millisecond
	if len(timeouts) > 0 && timeouts[0] > 0 {
		timeout = timeouts[0]
	}
	return &SocialFeedClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		client:  &http.Client{Timeout: timeout},
	}
}

func (client *SocialFeedClient) ExcludeBlocked(ctx context.Context, userID string, authorIDs []string) ([]string, error) {
	allowed, _, err := client.getExcludedAndAllowed(ctx, userID, authorIDs)
	return allowed, err
}

func (client *SocialFeedClient) GetBlockedCandidateIDs(ctx context.Context, userID string, candidateIDs []string) ([]string, error) {
	_, excluded, err := client.getExcludedAndAllowed(ctx, userID, candidateIDs)
	return excluded, err
}

func (client *SocialFeedClient) getExcludedAndAllowed(ctx context.Context, userID string, authorIDs []string) ([]string, []string, error) {
	if len(authorIDs) == 0 {
		return []string{}, []string{}, nil
	}

	payload, err := json.Marshal(struct {
		UserID       string   `json:"user_id"`
		CandidateIDs []string `json:"candidate_ids"`
	}{UserID: userID, CandidateIDs: authorIDs})
	if err != nil {
		return nil, nil, fmt.Errorf("encode block-exclusion request: %w", err)
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, client.baseURL+"/internal/blocks/exclude", bytes.NewReader(payload))
	if err != nil {
		return nil, nil, fmt.Errorf("create block-exclusion request: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := client.client.Do(request)
	if err != nil {
		return nil, nil, fmt.Errorf("call social-service block exclusion: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, nil, fmt.Errorf("social-service block exclusion returned status %d", response.StatusCode)
	}

	var result struct {
		Success bool `json:"success"`
		Data    struct {
			CandidateIDs []string `json:"candidate_ids"`
			ExcludedIDs  []string `json:"excluded_ids"`
		} `json:"data"`
	}
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		return nil, nil, fmt.Errorf("decode social-service block-exclusion response: %w", err)
	}
	if !result.Success || result.Data.CandidateIDs == nil {
		return nil, nil, errors.New("social-service returned an invalid block-exclusion response")
	}
	if result.Data.ExcludedIDs == nil {
		allowed := make(map[string]struct{}, len(result.Data.CandidateIDs))
		for _, userID := range result.Data.CandidateIDs {
			allowed[userID] = struct{}{}
		}
		result.Data.ExcludedIDs = make([]string, 0)
		for _, userID := range authorIDs {
			if _, isAllowed := allowed[userID]; !isAllowed {
				result.Data.ExcludedIDs = append(result.Data.ExcludedIDs, userID)
			}
		}
	}
	return result.Data.CandidateIDs, result.Data.ExcludedIDs, nil
}

type HTTPFeedRecommendationProvider struct {
	baseURL string
	client  *http.Client
}

func NewHTTPFeedRecommendationProvider(baseURL string, timeouts ...time.Duration) *HTTPFeedRecommendationProvider {
	timeout := 50 * time.Millisecond
	if len(timeouts) > 0 && timeouts[0] > 0 {
		timeout = timeouts[0]
	}
	return &HTTPFeedRecommendationProvider{
		baseURL: strings.TrimRight(baseURL, "/"),
		client:  &http.Client{Timeout: timeout},
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

type HTTPFeedIdentityVerifier struct {
	baseURL string
	client  *http.Client
}

func NewHTTPFeedIdentityVerifier(baseURL string, timeout time.Duration) *HTTPFeedIdentityVerifier {
	if timeout <= 0 {
		timeout = 60 * time.Millisecond
	}
	return &HTTPFeedIdentityVerifier{
		baseURL: strings.TrimRight(baseURL, "/"),
		client:  &http.Client{Timeout: timeout},
	}
}

func (verifier *HTTPFeedIdentityVerifier) AuthenticatedUserID(ctx context.Context, authorization string) (string, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, verifier.baseURL+"/auth/validate", nil)
	if err != nil {
		return "", fmt.Errorf("create feed session validation request: %w", err)
	}
	request.Header.Set("Authorization", authorization)
	response, err := verifier.client.Do(request)
	if err != nil {
		return "", fmt.Errorf("validate feed session: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode == http.StatusUnauthorized {
		return "", ErrInvalidFeedIdentity
	}
	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("auth-service returned status %d during feed session validation", response.StatusCode)
	}

	var result struct {
		Success bool `json:"success"`
		Data    struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode feed session validation response: %w", err)
	}
	if !result.Success || !ValidateFeedUserID(result.Data.ID) {
		return "", ErrInvalidFeedIdentity
	}
	return result.Data.ID, nil
}
