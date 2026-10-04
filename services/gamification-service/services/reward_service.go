package services

import (
	"context"
	"errors"
	"strings"

	"github.com/meloop/gamification-service/models"
	"github.com/meloop/gamification-service/repositories"
)

var (
	ErrRewardNotFound        = errors.New("reward not found")
	ErrInvalidReward         = errors.New("invalid reward")
	ErrRequiredLevelNotFound = repositories.ErrRequiredLevelNotFound
)

type RewardRepository interface {
	List(context.Context, bool) ([]models.Reward, error)
	Get(context.Context, string) (*models.Reward, error)
	Create(context.Context, models.Reward) (*models.Reward, error)
	Update(context.Context, string, models.Reward) (*models.Reward, error)
	SetAvailable(context.Context, string, bool) (*models.Reward, error)
}

type RewardService struct {
	repository RewardRepository
}

func NewRewardService(repository RewardRepository) *RewardService {
	return &RewardService{repository: repository}
}

func (s *RewardService) List(ctx context.Context, includeUnavailable bool) ([]models.Reward, error) {
	return s.repository.List(ctx, includeUnavailable)
}

func (s *RewardService) Get(ctx context.Context, rewardID string) (*models.Reward, error) {
	reward, err := s.repository.Get(ctx, rewardID)
	if err != nil {
		return nil, err
	}
	if reward == nil {
		return nil, ErrRewardNotFound
	}
	return reward, nil
}

func (s *RewardService) Create(ctx context.Context, request models.RewardRequest) (*models.Reward, error) {
	reward, err := rewardFromRequest(request)
	if err != nil {
		return nil, err
	}
	reward.Available = true
	if request.Available != nil {
		reward.Available = *request.Available
	}
	return s.repository.Create(ctx, reward)
}

func (s *RewardService) Update(ctx context.Context, rewardID string, request models.RewardRequest) (*models.Reward, error) {
	reward, err := rewardFromRequest(request)
	if err != nil {
		return nil, err
	}
	if request.Available == nil {
		return nil, ErrInvalidReward
	}
	reward.Available = *request.Available
	updated, err := s.repository.Update(ctx, rewardID, reward)
	if err != nil {
		return nil, err
	}
	if updated == nil {
		return nil, ErrRewardNotFound
	}
	return updated, nil
}

func (s *RewardService) SetAvailable(ctx context.Context, rewardID string, available bool) (*models.Reward, error) {
	reward, err := s.repository.SetAvailable(ctx, rewardID, available)
	if err != nil {
		return nil, err
	}
	if reward == nil {
		return nil, ErrRewardNotFound
	}
	return reward, nil
}

func rewardFromRequest(request models.RewardRequest) (models.Reward, error) {
	request.Name = strings.TrimSpace(request.Name)
	request.Type = strings.TrimSpace(request.Type)
	if request.Name == "" || request.Type == "" || request.RequiredLevel <= 0 {
		return models.Reward{}, ErrInvalidReward
	}
	return models.Reward{
		Name:          request.Name,
		Description:   strings.TrimSpace(request.Description),
		Type:          request.Type,
		RequiredLevel: request.RequiredLevel,
	}, nil
}
