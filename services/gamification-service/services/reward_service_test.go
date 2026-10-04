package services

import (
	"context"
	"errors"
	"testing"

	"github.com/meloop/gamification-service/models"
)

type rewardRepositoryFake struct {
	created models.Reward
	updated models.Reward
	set     models.Reward
	err     error
}

func (repository *rewardRepositoryFake) List(context.Context, bool) ([]models.Reward, error) {
	return nil, repository.err
}

func (repository *rewardRepositoryFake) Get(context.Context, string) (*models.Reward, error) {
	return nil, repository.err
}

func (repository *rewardRepositoryFake) Create(_ context.Context, reward models.Reward) (*models.Reward, error) {
	repository.created = reward
	reward.ID = "reward-1"
	return &reward, repository.err
}

func (repository *rewardRepositoryFake) Update(_ context.Context, _ string, reward models.Reward) (*models.Reward, error) {
	repository.updated = reward
	return &reward, repository.err
}

func (repository *rewardRepositoryFake) SetAvailable(_ context.Context, id string, available bool) (*models.Reward, error) {
	repository.set = models.Reward{ID: id, Available: available}
	return &repository.set, repository.err
}

func TestCreateRewardDefaultsToAvailable(t *testing.T) {
	repository := &rewardRepositoryFake{}
	service := NewRewardService(repository)
	reward, err := service.Create(context.Background(), models.RewardRequest{Name: "Badge", Type: "BADGE", RequiredLevel: 2})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if !reward.Available || repository.created.RequiredLevel != 2 {
		t.Fatalf("unexpected created reward: %+v", reward)
	}
}

func TestCreateRewardRequiresNameTypeAndLevel(t *testing.T) {
	service := NewRewardService(&rewardRepositoryFake{})
	if _, err := service.Create(context.Background(), models.RewardRequest{Name: "Badge", Type: "BADGE"}); !errors.Is(err, ErrInvalidReward) {
		t.Fatalf("Create() error = %v, want ErrInvalidReward", err)
	}
}

func TestUpdateRewardRequiresAvailabilityField(t *testing.T) {
	service := NewRewardService(&rewardRepositoryFake{})
	if _, err := service.Update(context.Background(), "reward-1", models.RewardRequest{Name: "Badge", Type: "BADGE", RequiredLevel: 2}); !errors.Is(err, ErrInvalidReward) {
		t.Fatalf("Update() error = %v, want ErrInvalidReward", err)
	}
}
