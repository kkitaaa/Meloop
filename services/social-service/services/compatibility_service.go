package services

import (
	"context"
	"errors"
	"math"
	"sort"
	"strings"

	"github.com/meloop/social-service/models"
	"github.com/meloop/social-service/repositories"
)

var (
	ErrSelfCompatibility = errors.New("cannot calculate compatibility with the same user")
)

type CompatibilityService interface {
	CalculateCompatibility(ctx context.Context, userAID, userBID string) (*models.CompatibilityScore, error)
}

type compatibilityService struct {
	repo repositories.CompatibilityRepository
}

func NewCompatibilityService(repo repositories.CompatibilityRepository) CompatibilityService {
	return &compatibilityService{repo: repo}
}

func (s *compatibilityService) CalculateCompatibility(ctx context.Context, userAID, userBID string) (*models.CompatibilityScore, error) {
	userAID = strings.TrimSpace(userAID)
	userBID = strings.TrimSpace(userBID)

	if userAID == "" || userBID == "" {
		return nil, ErrInvalidUser
	}
	if userAID == userBID {
		return nil, ErrSelfCompatibility
	}

	dataA, dataB, err := s.repo.GetMusicalProfiles(ctx, userAID, userBID)
	if err != nil {
		return nil, err
	}

	return CalculateCompatibilityScore(dataA, dataB), nil
}

// CalculateCompatibilityScore realiza el cálculo determinista de compatibilidad musical entre dos perfiles musicales.
func CalculateCompatibilityScore(dataA, dataB *models.UserMusicalData) *models.CompatibilityScore {
	if dataA == nil {
		dataA = &models.UserMusicalData{}
	}
	if dataB == nil {
		dataB = &models.UserMusicalData{}
	}

	simGenres, commonGenres := calculateJaccard(dataA.Genres, dataB.Genres)
	simArtists, commonArtists := calculateJaccard(dataA.Artists, dataB.Artists)
	simTracks, commonTracks := calculateJaccard(dataA.Tracks, dataB.Tracks)
	simInteractions, commonInteractions := calculateJaccard(dataA.InteractedTracks, dataB.InteractedTracks)

	score := (simGenres * models.WeightGenres) +
		(simArtists * models.WeightArtists) +
		(simTracks * models.WeightTracks) +
		(simInteractions * models.WeightInteractions)

	matchPercentage := int(math.Round(score * 100.0))
	if matchPercentage < 0 {
		matchPercentage = 0
	} else if matchPercentage > 100 {
		matchPercentage = 100
	}

	return &models.CompatibilityScore{
		UserAID:                dataA.UserID,
		UserBID:                dataB.UserID,
		MatchScore:             matchPercentage,
		GenreSimilarity:        simGenres,
		ArtistSimilarity:       simArtists,
		TrackSimilarity:        simTracks,
		InteractionSimilarity:  simInteractions,
		CommonGenres:           commonGenres,
		CommonArtists:          commonArtists,
		CommonTracks:           commonTracks,
		CommonInteractedTracks: commonInteractions,
	}
}

// calculateJaccard calcula la similitud de Jaccard (|A ∩ B| / |A ∪ B|) y retorna los elementos compartidos ordenados.
func calculateJaccard(sliceA, sliceB []string) (float64, []string) {
	setA := make(map[string]struct{}, len(sliceA))
	for _, item := range sliceA {
		trimmed := strings.TrimSpace(item)
		if trimmed != "" {
			setA[trimmed] = struct{}{}
		}
	}

	setB := make(map[string]struct{}, len(sliceB))
	for _, item := range sliceB {
		trimmed := strings.TrimSpace(item)
		if trimmed != "" {
			setB[trimmed] = struct{}{}
		}
	}

	if len(setA) == 0 && len(setB) == 0 {
		return 0.0, []string{}
	}

	commonMap := make(map[string]struct{})
	unionMap := make(map[string]struct{}, len(setA)+len(setB))

	for k := range setA {
		unionMap[k] = struct{}{}
	}

	for k := range setB {
		unionMap[k] = struct{}{}
		if _, exists := setA[k]; exists {
			commonMap[k] = struct{}{}
		}
	}

	unionSize := len(unionMap)
	if unionSize == 0 {
		return 0.0, []string{}
	}

	intersectionSize := len(commonMap)
	similarity := float64(intersectionSize) / float64(unionSize)

	commonList := make([]string, 0, intersectionSize)
	for k := range commonMap {
		commonList = append(commonList, k)
	}
	sort.Strings(commonList)

	return similarity, commonList
}
