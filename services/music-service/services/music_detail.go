package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
)

type spotifyDetailPage struct {
	Items  []spotifySearchItem `json:"items"`
	Offset int                 `json:"offset"`
	Next   *string             `json:"next"`
}

type spotifyDetail struct {
	spotifySearchItem
	Genres []string           `json:"genres"`
	Tracks *spotifyDetailPage `json:"tracks"`
}

func (s *SpotifyCatalog) getDetail(ctx context.Context, itemType, spotifyID string) (map[string]any, error) {
	market := s.client.market
	key := fmt.Sprintf("music:detail:v1:%s:%s:%s", itemType, market, spotifyID)
	if itemType == "tracks" && s.songs != nil {
		cached, found, err := s.songs.Get(ctx, spotifyID, market)
		if err != nil {
			searchLogger.Warn("song_detail_cache_read_failed", "error", err)
		} else if found && validDetail(cached) {
			return cached, nil
		} else if found {
			searchLogger.Warn("song_detail_cache_invalid")
		}
	} else if itemType != "tracks" && s.cache != nil {
		cached, found, err := s.cache.Get(ctx, key)
		if err != nil {
			searchLogger.Warn("music_detail_cache_read_failed", "error", err)
		} else if found && validDetail(cached) {
			return cached, nil
		} else if found {
			searchLogger.Warn("music_detail_cache_invalid")
		}
	}

	params := make(url.Values)
	if itemType != "artists" {
		params.Set("market", market)
	}
	payload, err := s.client.get(ctx, "/"+itemType+"/"+url.PathEscape(spotifyID), params)
	if err != nil {
		return nil, err
	}
	detail, err := s.normalizeDetail(ctx, itemType, spotifyID, payload)
	if err != nil {
		return nil, err
	}
	if itemType == "tracks" && s.songs != nil {
		if err := s.songs.Set(ctx, spotifyID, market, detail); err != nil {
			searchLogger.Warn("song_detail_cache_write_failed", "error", err)
		}
	} else if itemType != "tracks" && s.cache != nil {
		if err := s.cache.Set(ctx, key, detail, detailCacheTTL); err != nil {
			searchLogger.Warn("music_detail_cache_write_failed", "error", err)
		}
	}
	return detail, nil
}

func validDetail(detail map[string]any) bool {
	id, idOK := detail["spotify_id"].(string)
	name, nameOK := detail["name"].(string)
	return idOK && id != "" && nameOK && name != ""
}

func (s *SpotifyCatalog) normalizeDetail(ctx context.Context, itemType, spotifyID string, payload map[string]any) (map[string]any, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("encode Spotify detail: %w", err)
	}
	var item spotifyDetail
	if err := json.Unmarshal(raw, &item); err != nil {
		return nil, fmt.Errorf("decode Spotify detail: %w", err)
	}
	if item.ID == "" || item.Name == "" {
		return nil, errors.New("Spotify detail is missing id or name")
	}

	singular := map[string]string{"tracks": "track", "artists": "artist", "albums": "album"}[itemType]
	if singular == "" {
		return nil, errors.New("invalid music detail type")
	}
	base := normalizeSearchItem(item.spotifySearchItem, singular)
	result := map[string]any{
		"spotify_id":   base.SpotifyID,
		"type":         singular,
		"name":         base.Name,
		"artists":      base.Artists,
		"album":        optionalDetailString(base.Album),
		"image_url":    optionalDetailString(base.ImageURL),
		"external_url": optionalDetailString(base.ExternalURL),
		"genres":       []string{},
		"tracks":       []musicSearchItem{},
	}
	if item.Genres != nil {
		result["genres"] = item.Genres
	}
	if itemType != "albums" {
		return result, nil
	}
	if item.Tracks == nil {
		return nil, errors.New("Spotify album is missing its track page")
	}
	tracks := make([]musicSearchItem, 0)
	page := item.Tracks
	for {
		if len(tracks)+len(page.Items) > 1000 {
			return nil, errors.New("Spotify album track pagination exceeded its safe limit")
		}
		for _, track := range page.Items {
			if track.ID == "" || track.Name == "" {
				return nil, errors.New("Spotify album track is missing id or name")
			}
			tracks = append(tracks, normalizeSearchItem(track, "track"))
		}
		if page.Next == nil || *page.Next == "" {
			break
		}
		if len(page.Items) == 0 || len(tracks) >= 1000 {
			return nil, errors.New("Spotify album track pagination exceeded its safe limit")
		}
		nextOffset := page.Offset + len(page.Items)
		params := url.Values{
			"market": {s.client.market},
			"limit":  {"50"},
			"offset": {strconv.Itoa(nextOffset)},
		}
		next, err := s.client.get(ctx, "/albums/"+url.PathEscape(spotifyID)+"/tracks", params)
		if err != nil {
			return nil, err
		}
		raw, err := json.Marshal(next)
		if err != nil {
			return nil, fmt.Errorf("encode Spotify album tracks: %w", err)
		}
		page = &spotifyDetailPage{}
		if err := json.Unmarshal(raw, page); err != nil {
			return nil, fmt.Errorf("decode Spotify album tracks: %w", err)
		}
		if page.Offset != nextOffset {
			return nil, errors.New("Spotify album returned an unexpected track offset")
		}
	}
	result["tracks"] = tracks
	return result, nil
}

func optionalDetailString(value string) any {
	if value == "" {
		return nil
	}
	return value
}
