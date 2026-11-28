package services

import (
	"context"
	"net/url"
	"regexp"
	"strings"

	"github.com/meloop/user-service/models"
	"github.com/meloop/user-service/repositories"
)

const MusicPreferenceLimit = models.MusicPreferenceLimit

var (
	ErrPreferenceAlreadyExists = repositories.ErrMusicPreferenceExists
	ErrPreferenceLimitReached  = repositories.ErrMusicPreferenceLimit
	ErrInvalidMusicPreference  = repositories.ErrInvalidMusicPreference
	musicPreferenceIDPattern   = regexp.MustCompile(`^[A-Za-z0-9-]{1,50}$`)
)

type MusicPreferencesRepository interface {
	GetMusicPreferences(ctx context.Context, userID string) (*models.MusicPreferencesResponse, error)
	AddMusicPreference(ctx context.Context, userID string, request *models.AddMusicPreferenceRequest, limit int) (*models.MusicPreferencesResponse, error)
	RemoveMusicPreference(ctx context.Context, userID, itemType, itemID string) (*models.MusicPreferencesResponse, error)
}

type MusicPreferencesService interface {
	Get(ctx context.Context, userID string) (*models.MusicPreferencesResponse, error)
	Add(ctx context.Context, userID string, request *models.AddMusicPreferenceRequest) (*models.MusicPreferencesResponse, error)
	Remove(ctx context.Context, userID, itemType, itemID string) (*models.MusicPreferencesResponse, error)
}

type musicPreferencesService struct {
	repository MusicPreferencesRepository
}

func NewMusicPreferencesService(repository MusicPreferencesRepository) MusicPreferencesService {
	return &musicPreferencesService{repository: repository}
}

func (s *musicPreferencesService) Get(
	ctx context.Context,
	userID string,
) (*models.MusicPreferencesResponse, error) {
	return s.repository.GetMusicPreferences(ctx, userID)
}

func (s *musicPreferencesService) Add(
	ctx context.Context,
	userID string,
	request *models.AddMusicPreferenceRequest,
) (*models.MusicPreferencesResponse, error) {
	if request == nil {
		return nil, ErrInvalidMusicPreference
	}

	normalizedRequest := *request
	normalizedRequest.Type = strings.ToLower(strings.TrimSpace(normalizedRequest.Type))
	normalizedRequest.ItemID = strings.TrimSpace(normalizedRequest.ItemID)
	normalizedRequest.Name = strings.TrimSpace(normalizedRequest.Name)
	if request.ArtworkURL != nil {
		artworkURL := strings.TrimSpace(*request.ArtworkURL)
		normalizedRequest.ArtworkURL = &artworkURL
	}
	if !validMusicPreferenceType(normalizedRequest.Type) ||
		!musicPreferenceIDPattern.MatchString(normalizedRequest.ItemID) ||
		normalizedRequest.Name == "" ||
		len(normalizedRequest.Name) > 200 ||
		!validArtworkURL(normalizedRequest.ArtworkURL) {
		return nil, ErrInvalidMusicPreference
	}

	return s.repository.AddMusicPreference(ctx, userID, &normalizedRequest, MusicPreferenceLimit)
}

func (s *musicPreferencesService) Remove(
	ctx context.Context,
	userID string,
	itemType string,
	itemID string,
) (*models.MusicPreferencesResponse, error) {
	itemType = strings.ToLower(strings.TrimSpace(itemType))
	itemID = strings.TrimSpace(itemID)
	if !validMusicPreferenceType(itemType) || !musicPreferenceIDPattern.MatchString(itemID) {
		return nil, ErrInvalidMusicPreference
	}
	return s.repository.RemoveMusicPreference(ctx, userID, itemType, itemID)
}

func validMusicPreferenceType(itemType string) bool {
	switch itemType {
	case "track", "artist", "genre":
		return true
	default:
		return false
	}
}

func validArtworkURL(rawURL *string) bool {
	if rawURL == nil || *rawURL == "" {
		return true
	}
	parsed, err := url.ParseRequestURI(*rawURL)
	return err == nil && parsed.Scheme == "https" && parsed.Host != ""
}
