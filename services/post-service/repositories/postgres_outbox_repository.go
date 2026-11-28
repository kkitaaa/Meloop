package repositories

import (
	"context"
	"time"

	"github.com/meloop/post-service/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type OutboxRepository interface {
	ClaimPendingEvents(ctx context.Context, limit int, leaseDuration time.Duration) ([]models.OutboxEvent, error)
	MarkPublished(ctx context.Context, id string) error
	MarkRetry(ctx context.Context, id string, attempts int, nextRetryAt time.Time, lastError string) error
}

type PostgresOutboxRepository struct {
	db *gorm.DB
}

func NewPostgresOutboxRepository(db *gorm.DB) *PostgresOutboxRepository {
	return &PostgresOutboxRepository{db: db}
}

// ClaimPendingEvents reserva eventos pendientes o cuyo lease expiró usando FOR UPDATE SKIP LOCKED
func (r *PostgresOutboxRepository) ClaimPendingEvents(ctx context.Context, limit int, leaseDuration time.Duration) ([]models.OutboxEvent, error) {
	var claimed []models.OutboxEvent
	now := time.Now().UTC()
	leaseUntil := now.Add(leaseDuration)

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var rows []models.OutboxEvent
		if err := tx.
			Where("status IN ? AND next_retry_at <= ?", []string{models.OutboxStatusPending, models.OutboxStatusProcessing}, now).
			Order("created_at ASC").
			Limit(limit).
			Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
			Find(&rows).Error; err != nil {
			return err
		}

		if len(rows) == 0 {
			return nil
		}

		ids := make([]string, len(rows))
		for i, row := range rows {
			ids[i] = row.ID
		}

		if err := tx.Model(&models.OutboxEvent{}).
			Where("id IN ?", ids).
			Updates(map[string]interface{}{
				"status":        models.OutboxStatusProcessing,
				"next_retry_at": leaseUntil,
			}).Error; err != nil {
			return err
		}

		claimed = rows
		return nil
	})

	return claimed, err
}

func (r *PostgresOutboxRepository) MarkPublished(ctx context.Context, id string) error {
	now := time.Now().UTC()
	return r.db.WithContext(ctx).
		Model(&models.OutboxEvent{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":       models.OutboxStatusPublished,
			"published_at": now,
			"last_error":   "",
		}).Error
}

func (r *PostgresOutboxRepository) MarkRetry(ctx context.Context, id string, attempts int, nextRetryAt time.Time, lastError string) error {
	return r.db.WithContext(ctx).
		Model(&models.OutboxEvent{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":        models.OutboxStatusPending,
			"attempts":      attempts,
			"next_retry_at": nextRetryAt.UTC(),
			"last_error":    lastError,
		}).Error
}
