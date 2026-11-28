package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/meloop/music-service/config"
)

const (
	spotifyAPIURL   = "https://api.spotify.com/v1"
	spotifyTokenURL = "https://accounts.spotify.com/api/token"
	maxResponseSize = 4 << 20
)

type SpotifyAPIError struct {
	StatusCode int
	RetryAfter string
}

func (e *SpotifyAPIError) Error() string {
	return fmt.Sprintf("Spotify API returned status %d", e.StatusCode)
}

type spotifyTokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

type spotifyAccessToken struct {
	value     string
	expiresAt time.Time
}

type spotifyClient struct {
	clientID     string
	clientSecret string
	market       string
	httpClient   *http.Client
	apiURL       string
	tokenURL     string

	tokenMu sync.Mutex
	token   spotifyAccessToken
}

type SpotifyCatalog struct {
	client *spotifyClient
}

func NewSpotifyCatalog(cfg config.Config) (*SpotifyCatalog, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &SpotifyCatalog{
		client: &spotifyClient{
			clientID:     cfg.SpotifyClientID,
			clientSecret: cfg.SpotifyClientSecret,
			market:       cfg.SpotifyMarket,
			httpClient:   &http.Client{Timeout: 12 * time.Second},
			apiURL:       spotifyAPIURL,
			tokenURL:     spotifyTokenURL,
		},
	}, nil
}

func (s *SpotifyCatalog) Search(
	ctx context.Context,
	query string,
	itemType string,
	offset int,
	limit int,
) (map[string]any, error) {
	if itemType == "genre" {
		return searchGenres(query, offset, limit), nil
	}

	params := url.Values{
		"q":      {query},
		"type":   {itemType},
		"offset": {strconv.Itoa(offset)},
		"limit":  {strconv.Itoa(limit)},
		"market": {s.client.market},
	}
	return s.client.get(ctx, "/search", params)
}

func (s *SpotifyCatalog) Detail(
	ctx context.Context,
	itemType string,
	spotifyID string,
) (map[string]any, error) {
	path := "/" + itemType + "/" + url.PathEscape(spotifyID)
	params := make(url.Values)
	if itemType != "artists" {
		params.Set("market", s.client.market)
	}
	return s.client.get(ctx, path, params)
}

func (c *spotifyClient) get(
	ctx context.Context,
	path string,
	params url.Values,
) (map[string]any, error) {
	for attempt := 0; attempt < 2; attempt++ {
		token, err := c.accessToken(ctx)
		if err != nil {
			return nil, err
		}

		endpoint, err := url.JoinPath(c.apiURL, path)
		if err != nil {
			return nil, err
		}
		requestURL, err := url.Parse(endpoint)
		if err != nil {
			return nil, err
		}
		requestURL.RawQuery = params.Encode()

		request, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL.String(), nil)
		if err != nil {
			return nil, err
		}
		request.Header.Set("Authorization", "Bearer "+token)

		response, err := c.httpClient.Do(request)
		if err != nil {
			return nil, err
		}

		if response.StatusCode == http.StatusUnauthorized && attempt == 0 {
			_ = response.Body.Close()
			c.invalidateToken(token)
			continue
		}
		if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
			apiErr := &SpotifyAPIError{
				StatusCode: response.StatusCode,
				RetryAfter: response.Header.Get("Retry-After"),
			}
			_ = response.Body.Close()
			return nil, apiErr
		}

		var result map[string]any
		decodeErr := json.NewDecoder(io.LimitReader(response.Body, maxResponseSize)).Decode(&result)
		closeErr := response.Body.Close()
		if decodeErr != nil {
			return nil, fmt.Errorf("decode Spotify response: %w", decodeErr)
		}
		if closeErr != nil {
			return nil, fmt.Errorf("close Spotify response: %w", closeErr)
		}
		return result, nil
	}

	return nil, errors.New("Spotify authorization failed after token refresh")
}

func (c *spotifyClient) accessToken(ctx context.Context) (string, error) {
	c.tokenMu.Lock()
	defer c.tokenMu.Unlock()

	if c.token.value != "" && time.Now().Before(c.token.expiresAt) {
		return c.token.value, nil
	}

	form := url.Values{"grant_type": {"client_credentials"}}
	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.tokenURL,
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		return "", err
	}
	request.SetBasicAuth(c.clientID, c.clientSecret)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	response, err := c.httpClient.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return "", &SpotifyAPIError{StatusCode: response.StatusCode}
	}

	var tokenResponse spotifyTokenResponse
	if err := json.NewDecoder(io.LimitReader(response.Body, 1<<20)).Decode(&tokenResponse); err != nil {
		return "", fmt.Errorf("decode Spotify token response: %w", err)
	}
	if tokenResponse.AccessToken == "" || tokenResponse.ExpiresIn <= 0 {
		return "", errors.New("Spotify returned an invalid access token response")
	}

	cacheDuration := time.Duration(tokenResponse.ExpiresIn)*time.Second - 30*time.Second
	if cacheDuration <= 0 {
		cacheDuration = time.Duration(tokenResponse.ExpiresIn) * time.Second / 2
	}
	c.token = spotifyAccessToken{
		value:     tokenResponse.AccessToken,
		expiresAt: time.Now().Add(cacheDuration),
	}
	return c.token.value, nil
}

func (c *spotifyClient) invalidateToken(token string) {
	c.tokenMu.Lock()
	defer c.tokenMu.Unlock()
	if c.token.value == token {
		c.token = spotifyAccessToken{}
	}
}

type genre struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

var availableGenres = []genre{
	{ID: "alternative", Name: "Alternativo"},
	{ID: "ambient", Name: "Ambiental"},
	{ID: "blues", Name: "Blues"},
	{ID: "classical", Name: "Clásica"},
	{ID: "country", Name: "Country"},
	{ID: "dance", Name: "Dance"},
	{ID: "electronic", Name: "Electrónica"},
	{ID: "folk", Name: "Folk"},
	{ID: "funk", Name: "Funk"},
	{ID: "hip-hop", Name: "Hip Hop"},
	{ID: "indie", Name: "Indie"},
	{ID: "jazz", Name: "Jazz"},
	{ID: "k-pop", Name: "K-pop"},
	{ID: "latin", Name: "Latino"},
	{ID: "metal", Name: "Metal"},
	{ID: "pop", Name: "Pop"},
	{ID: "punk", Name: "Punk"},
	{ID: "r-and-b", Name: "R&B"},
	{ID: "reggae", Name: "Reggae"},
	{ID: "reggaeton", Name: "Reggaetón"},
	{ID: "rock", Name: "Rock"},
	{ID: "salsa", Name: "Salsa"},
	{ID: "soul", Name: "Soul"},
	{ID: "soundtrack", Name: "Bandas sonoras"},
	{ID: "trap", Name: "Trap"},
	{ID: "urbano-chileno", Name: "Urbano chileno"},
}

func searchGenres(query string, offset int, limit int) map[string]any {
	query = strings.ToLower(strings.TrimSpace(query))
	filtered := make([]genre, 0, len(availableGenres))
	for _, item := range availableGenres {
		if strings.Contains(strings.ToLower(item.Name), query) ||
			strings.Contains(strings.ToLower(item.ID), query) {
			filtered = append(filtered, item)
		}
	}

	start := min(offset, len(filtered))
	end := min(start+limit, len(filtered))
	items := filtered[start:end]
	nextOffset := any(nil)
	if end < len(filtered) {
		nextOffset = end
	}
	return map[string]any{
		"genres": map[string]any{
			"items":       items,
			"offset":      offset,
			"limit":       limit,
			"total":       len(filtered),
			"has_more":    end < len(filtered),
			"next_offset": nextOffset,
		},
	}
}
