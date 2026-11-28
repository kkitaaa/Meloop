package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
)

type musicSearchItem struct {
	Type        string   `json:"type"`
	SpotifyID   string   `json:"spotify_id"`
	Name        string   `json:"name"`
	Artists     []string `json:"artists"`
	Album       string   `json:"album"`
	ImageURL    string   `json:"image_url"`
	ExternalURL string   `json:"external_url"`
}

type spotifySearchItem struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Artists []struct {
		Name string `json:"name"`
	} `json:"artists"`
	Images []struct {
		URL string `json:"url"`
	} `json:"images"`
	ExternalURLs struct {
		Spotify string `json:"spotify"`
	} `json:"external_urls"`
	Album *struct {
		Name   string `json:"name"`
		Images []struct {
			URL string `json:"url"`
		} `json:"images"`
	} `json:"album"`
}

type spotifySearchPage struct {
	Items []spotifySearchItem `json:"items"`
	Total int                 `json:"total"`
}

func (s *SpotifyCatalog) searchSpotify(ctx context.Context, query, itemType string, offset, limit int) (map[string]any, error) {
	results := make([]musicSearchItem, 0, limit)
	total := 0
	for len(results) < limit {
		nextOffset := offset + len(results)
		requestOffset := min(nextOffset, 1000)
		skip := nextOffset - requestOffset
		requestLimit := min(10, limit-len(results)+skip)
		params := url.Values{
			"q": {query}, "type": {itemType}, "market": {s.client.market},
			"offset": {strconv.Itoa(requestOffset)}, "limit": {strconv.Itoa(requestLimit)},
		}
		payload, err := s.client.get(ctx, "/search", params)
		if err != nil {
			return nil, err
		}
		key := map[string]string{"track": "tracks", "artist": "artists", "album": "albums"}[itemType]
		rawPage, exists := payload[key]
		if !exists || rawPage == nil {
			return nil, fmt.Errorf("Spotify search response missing %s", key)
		}
		encoded, err := json.Marshal(rawPage)
		if err != nil {
			return nil, fmt.Errorf("encode Spotify search page: %w", err)
		}
		var page spotifySearchPage
		if err := json.Unmarshal(encoded, &page); err != nil {
			return nil, fmt.Errorf("decode Spotify search page: %w", err)
		}
		total = page.Total
		if skip >= len(page.Items) {
			break
		}
		for _, item := range page.Items[skip:] {
			if item.ID == "" || item.Name == "" {
				return nil, fmt.Errorf("Spotify search item missing id or name")
			}
			results = append(results, normalizeSearchItem(item, itemType))
		}
		if len(page.Items) < requestLimit || offset+len(results) >= total {
			break
		}
	}
	return map[string]any{"results": results, "limit": limit, "offset": offset, "total": total}, nil
}

func normalizeSearchItem(item spotifySearchItem, itemType string) musicSearchItem {
	result := musicSearchItem{
		Type: itemType, SpotifyID: item.ID, Name: item.Name,
		Artists: make([]string, 0, len(item.Artists)), ExternalURL: item.ExternalURLs.Spotify,
	}
	for _, artist := range item.Artists {
		if artist.Name != "" {
			result.Artists = append(result.Artists, artist.Name)
		}
	}
	if item.Album != nil {
		result.Album = item.Album.Name
		if len(item.Album.Images) > 0 {
			result.ImageURL = item.Album.Images[0].URL
		}
	} else if len(item.Images) > 0 {
		result.ImageURL = item.Images[0].URL
	}
	return result
}
