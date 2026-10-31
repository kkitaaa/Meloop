package services

import (
	"context"
	"errors"
	"testing"

	"github.com/meloop/gamification-service/models"
	"github.com/meloop/gamification-service/repositories"
)

type fakeGamificationRepository struct {
	result   *models.ExperienceResult
	progress *models.UserProgress
	err      error
	called   bool
}

func (f *fakeGamificationRepository) AddExperience(ctx context.Context, userID string, amount int) (*models.ExperienceResult, error) {
	f.called = true
	if f.err != nil {
		return nil, f.err
	}
	return f.result, nil
}

func (f *fakeGamificationRepository) GetUserProgress(ctx context.Context, userID string) (*models.UserProgress, error) {
	f.called = true
	if f.err != nil {
		return nil, f.err
	}
	return f.progress, nil
}

type fakeEventPublisher struct {
	levelUpEvents        []models.LevelUpEvent
	rewardUnlockedEvents []models.RewardUnlockedEvent
	levelUpErr           error
	rewardUnlockedErr    error
}

func (f *fakeEventPublisher) PublishLevelUp(ctx context.Context, event models.LevelUpEvent) error {
	if f.levelUpErr != nil {
		return f.levelUpErr
	}
	f.levelUpEvents = append(f.levelUpEvents, event)
	return nil
}

func (f *fakeEventPublisher) PublishRewardUnlocked(ctx context.Context, event models.RewardUnlockedEvent) error {
	if f.rewardUnlockedErr != nil {
		return f.rewardUnlockedErr
	}
	f.rewardUnlockedEvents = append(f.rewardUnlockedEvents, event)
	return nil
}

// Caso A — Sin subida de nivel
// XP procesada correctamente -> Nivel no cambia -> No publicar level.up -> No publicar reward.unlocked
func TestCasoA_SinSubidaDeNivel(t *testing.T) {
	repo := &fakeGamificationRepository{
		result: &models.ExperienceResult{
			UserID:          "user-100",
			PreviousXP:      50,
			CurrentXP:       60,
			PreviousLevel:   1,
			CurrentLevel:    1,
			LeveledUp:       false,
			UnlockedRewards: nil,
		},
	}
	publisher := &fakeEventPublisher{}
	service := NewGamificationService(repo, publisher)

	result, err := service.ProcessExperience(context.Background(), "user-100", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.LeveledUp {
		t.Errorf("expected LeveledUp to be false, got true")
	}
	if len(publisher.levelUpEvents) != 0 {
		t.Fatalf("expected 0 level.up events, got %d", len(publisher.levelUpEvents))
	}
	if len(publisher.rewardUnlockedEvents) != 0 {
		t.Fatalf("expected 0 reward.unlocked events, got %d", len(publisher.rewardUnlockedEvents))
	}
}

// Caso B — Subida de nivel sin recompensas
// XP procesada correctamente -> Nuevo nivel detectado -> Publicar 1 level.up -> No publicar reward.unlocked
func TestCasoB_SubidaDeNivelSinRecompensas(t *testing.T) {
	repo := &fakeGamificationRepository{
		result: &models.ExperienceResult{
			UserID:          "user-200",
			PreviousXP:      95,
			CurrentXP:       105,
			PreviousLevel:   1,
			CurrentLevel:    2,
			LeveledUp:       true,
			UnlockedRewards: []models.Reward{},
		},
	}
	publisher := &fakeEventPublisher{}
	service := NewGamificationService(repo, publisher)

	result, err := service.ProcessExperience(context.Background(), "user-200", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.LeveledUp {
		t.Errorf("expected LeveledUp to be true, got false")
	}
	if len(publisher.levelUpEvents) != 1 {
		t.Fatalf("expected 1 level.up event, got %d", len(publisher.levelUpEvents))
	}
	event := publisher.levelUpEvents[0]
	if event.UserID != "user-200" || event.Nivel != 2 || event.Event != "level.up" {
		t.Fatalf("unexpected level.up event payload: %+v", event)
	}
	if len(publisher.rewardUnlockedEvents) != 0 {
		t.Fatalf("expected 0 reward.unlocked events, got %d", len(publisher.rewardUnlockedEvents))
	}
}

// Caso C — Subida de nivel con una recompensa
// XP procesada correctamente -> Nuevo nivel detectado -> Recompensa incorporada al inventario -> Publicar 1 level.up -> Publicar 1 reward.unlocked
func TestCasoC_SubidaDeNivelConUnaRecompensa(t *testing.T) {
	repo := &fakeGamificationRepository{
		result: &models.ExperienceResult{
			UserID:        "user-300",
			PreviousXP:    190,
			CurrentXP:     210,
			PreviousLevel: 2,
			CurrentLevel:  3,
			LeveledUp:     true,
			UnlockedRewards: []models.Reward{
				{
					IDRecompensa: "rew-badge-gold",
					Tipo:         "INSIGNIA",
					IDNivel:      3,
				},
			},
		},
	}
	publisher := &fakeEventPublisher{}
	service := NewGamificationService(repo, publisher)

	result, err := service.ProcessExperience(context.Background(), "user-300", 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.LeveledUp {
		t.Errorf("expected LeveledUp to be true, got false")
	}
	if len(publisher.levelUpEvents) != 1 {
		t.Fatalf("expected 1 level.up event, got %d", len(publisher.levelUpEvents))
	}
	lvlEvent := publisher.levelUpEvents[0]
	if lvlEvent.UserID != "user-300" || lvlEvent.Nivel != 3 || lvlEvent.Event != "level.up" {
		t.Fatalf("unexpected level.up event: %+v", lvlEvent)
	}

	if len(publisher.rewardUnlockedEvents) != 1 {
		t.Fatalf("expected 1 reward.unlocked event, got %d", len(publisher.rewardUnlockedEvents))
	}
	rewEvent := publisher.rewardUnlockedEvents[0]
	if rewEvent.UserID != "user-300" || rewEvent.RewardID != "rew-badge-gold" || rewEvent.RewardType != "INSIGNIA" || rewEvent.Nivel != 3 || rewEvent.Event != "reward.unlocked" {
		t.Fatalf("unexpected reward.unlocked event: %+v", rewEvent)
	}
}

// Caso D — Subida de nivel con varias recompensas
// XP procesada correctamente -> Nuevo nivel detectado -> Recompensa A incorporada, Recompensa B incorporada -> Publicar 1 level.up -> Publicar 1 reward.unlocked para A -> Publicar 1 reward.unlocked para B
func TestCasoD_SubidaDeNivelConVariasRecompensas(t *testing.T) {
	repo := &fakeGamificationRepository{
		result: &models.ExperienceResult{
			UserID:        "user-400",
			PreviousXP:    450,
			CurrentXP:     550,
			PreviousLevel: 4,
			CurrentLevel:  5,
			LeveledUp:     true,
			UnlockedRewards: []models.Reward{
				{
					IDRecompensa: "rew-theme-neon",
					Tipo:         "TEMA_PERFIL",
					IDNivel:      5,
				},
				{
					IDRecompensa: "rew-avatar-vip",
					Tipo:         "MARCO_AVATAR",
					IDNivel:      5,
				},
			},
		},
	}
	publisher := &fakeEventPublisher{}
	service := NewGamificationService(repo, publisher)

	result, err := service.ProcessExperience(context.Background(), "user-400", 100)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.LeveledUp {
		t.Errorf("expected LeveledUp to be true, got false")
	}

	// Debe publicarse exactamente 1 level.up
	if len(publisher.levelUpEvents) != 1 {
		t.Fatalf("expected 1 level.up event, got %d", len(publisher.levelUpEvents))
	}
	if publisher.levelUpEvents[0].Nivel != 5 || publisher.levelUpEvents[0].UserID != "user-400" {
		t.Fatalf("unexpected level.up event: %+v", publisher.levelUpEvents[0])
	}

	// Debe publicarse UN evento por cada recompensa (NO agrupadas)
	if len(publisher.rewardUnlockedEvents) != 2 {
		t.Fatalf("expected 2 reward.unlocked events, got %d", len(publisher.rewardUnlockedEvents))
	}

	rewA := publisher.rewardUnlockedEvents[0]
	if rewA.UserID != "user-400" || rewA.RewardID != "rew-theme-neon" || rewA.RewardType != "TEMA_PERFIL" || rewA.Nivel != 5 || rewA.Event != "reward.unlocked" {
		t.Fatalf("unexpected reward A event: %+v", rewA)
	}

	rewB := publisher.rewardUnlockedEvents[1]
	if rewB.UserID != "user-400" || rewB.RewardID != "rew-avatar-vip" || rewB.RewardType != "MARCO_AVATAR" || rewB.Nivel != 5 || rewB.Event != "reward.unlocked" {
		t.Fatalf("unexpected reward B event: %+v", rewB)
	}
}

// Caso E — Fallo de la transacción
// Procesamiento de XP -> Error / rollback -> No publicar level.up -> No publicar reward.unlocked
func TestCasoE_FalloDeTransaccion(t *testing.T) {
	dbErr := errors.New("database connection failed during transaction")
	repo := &fakeGamificationRepository{
		err: dbErr,
	}
	publisher := &fakeEventPublisher{}
	service := NewGamificationService(repo, publisher)

	result, err := service.ProcessExperience(context.Background(), "user-500", 50)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, dbErr) {
		t.Fatalf("error = %v, want %v", err, dbErr)
	}
	if result != nil {
		t.Fatalf("expected nil result on failure, got %+v", result)
	}

	// ATOMICIDAD: Ningún evento debe haber sido publicado
	if len(publisher.levelUpEvents) != 0 {
		t.Fatalf("expected 0 level.up events on rollback, got %d", len(publisher.levelUpEvents))
	}
	if len(publisher.rewardUnlockedEvents) != 0 {
		t.Fatalf("expected 0 reward.unlocked events on rollback, got %d", len(publisher.rewardUnlockedEvents))
	}
}

func TestValidationErrors_DoNotPublishEvents(t *testing.T) {
	repo := &fakeGamificationRepository{}
	publisher := &fakeEventPublisher{}
	service := NewGamificationService(repo, publisher)

	t.Run("empty user ID", func(t *testing.T) {
		_, err := service.ProcessExperience(context.Background(), "   ", 10)
		if !errors.Is(err, ErrInvalidUserID) {
			t.Fatalf("expected ErrInvalidUserID, got %v", err)
		}
		if repo.called {
			t.Fatal("repository should not be called with invalid input")
		}
		if len(publisher.levelUpEvents) > 0 || len(publisher.rewardUnlockedEvents) > 0 {
			t.Fatal("events should not be published for invalid user")
		}
	})

	t.Run("zero or negative XP", func(t *testing.T) {
		_, err := service.ProcessExperience(context.Background(), "user-1", 0)
		if !errors.Is(err, ErrInvalidExperienceAmount) {
			t.Fatalf("expected ErrInvalidExperienceAmount, got %v", err)
		}
		_, errNeg := service.ProcessExperience(context.Background(), "user-1", -15)
		if !errors.Is(errNeg, ErrInvalidExperienceAmount) {
			t.Fatalf("expected ErrInvalidExperienceAmount for negative XP, got %v", errNeg)
		}
		if len(publisher.levelUpEvents) > 0 || len(publisher.rewardUnlockedEvents) > 0 {
			t.Fatal("events should not be published for invalid XP amount")
		}
	})

	t.Run("user not found in repository", func(t *testing.T) {
		repoNotFound := &fakeGamificationRepository{err: repositories.ErrUserNotFound}
		publisherNotFound := &fakeEventPublisher{}
		serviceNotFound := NewGamificationService(repoNotFound, publisherNotFound)

		_, err := serviceNotFound.ProcessExperience(context.Background(), "user-not-found", 20)
		if !errors.Is(err, repositories.ErrUserNotFound) {
			t.Fatalf("expected ErrUserNotFound, got %v", err)
		}
		if len(publisherNotFound.levelUpEvents) > 0 || len(publisherNotFound.rewardUnlockedEvents) > 0 {
			t.Fatal("events should not be published when user is not found")
		}
	})
}

func TestPublisherError_IsPropagated(t *testing.T) {
	pubErr := errors.New("rabbitmq connection broker dropped")
	repo := &fakeGamificationRepository{
		result: &models.ExperienceResult{
			UserID:        "user-600",
			PreviousXP:    100,
			CurrentXP:     200,
			PreviousLevel: 1,
			CurrentLevel:  2,
			LeveledUp:     true,
		},
	}
	publisher := &fakeEventPublisher{levelUpErr: pubErr}
	service := NewGamificationService(repo, publisher)

	_, err := service.ProcessExperience(context.Background(), "user-600", 100)
	if !errors.Is(err, pubErr) {
		t.Fatalf("expected publisher error %v, got %v", pubErr, err)
	}
}
