package services

import (
	"context"
	"errors"
	"testing"

	"github.com/meloop/user-service/models"
)

type mockMusicPreferencesRepository struct {
	addRequest *models.AddMusicPreferenceRequest
	addLimit   int
	addCalls   int
}

func (m *mockMusicPreferencesRepository) GetMusicPreferences(
	context.Context,
	string,
) (*models.MusicPreferencesResponse, error) {
	return &models.MusicPreferencesResponse{}, nil
}

func (m *mockMusicPreferencesRepository) AddMusicPreference(
	_ context.Context,
	_ string,
	request *models.AddMusicPreferenceRequest,
	limit int,
) (*models.MusicPreferencesResponse, error) {
	m.addCalls++
	m.addRequest = request
	m.addLimit = limit
	return &models.MusicPreferencesResponse{}, nil
}

func (m *mockMusicPreferencesRepository) RemoveMusicPreference(
	context.Context,
	string,
	string,
	string,
) (*models.MusicPreferencesResponse, error) {
	return &models.MusicPreferencesResponse{}, nil
}

func TestMusicPreferencesServiceAddNormalizesAndEnforcesContract(t *testing.T) {
	repository := &mockMusicPreferencesRepository{}
	service := NewMusicPreferencesService(repository)
	artworkURL := "https://i.scdn.co/image/example"
	request := &models.AddMusicPreferenceRequest{
		Type:       " Track ",
		ItemID:     " 1234567890123456789012 ",
		Name:       " Example song ",
		ArtworkURL: &artworkURL,
	}

	if _, err := service.Add(context.Background(), "user-1", request); err != nil {
		t.Fatalf("Add returned error: %v", err)
	}
	if repository.addCalls != 1 || repository.addLimit != MusicPreferenceLimit {
		t.Fatalf("repository calls/limit = %d/%d", repository.addCalls, repository.addLimit)
	}
	if repository.addRequest.Type != "track" ||
		repository.addRequest.ItemID != "1234567890123456789012" ||
		repository.addRequest.Name != "Example song" {
		t.Fatalf("request was not normalized: %+v", repository.addRequest)
	}
	if request.Type != " Track " || request.ItemID != " 1234567890123456789012 " {
		t.Fatalf("service mutated caller request: %+v", request)
	}
}

func TestMusicPreferencesServiceAddRejectsInvalidArtworkWithoutPersistence(t *testing.T) {
	repository := &mockMusicPreferencesRepository{}
	service := NewMusicPreferencesService(repository)
	artworkURL := "http://example.test/image"

	_, err := service.Add(context.Background(), "user-1", &models.AddMusicPreferenceRequest{
		Type:       "track",
		ItemID:     "1234567890123456789012",
		Name:       "Example song",
		ArtworkURL: &artworkURL,
	})
	if !errors.Is(err, ErrInvalidMusicPreference) {
		t.Fatalf("Add error = %v, want ErrInvalidMusicPreference", err)
	}
	if repository.addCalls != 0 {
		t.Fatalf("invalid preference reached repository %d times", repository.addCalls)
	}
}
