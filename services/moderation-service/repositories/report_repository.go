package repositories

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/meloop/moderation-service/models"
)

var ErrAlreadyReported = errors.New("report already exists for this target")

const openReportConstraint = "reports_open_reporter_target_unique"

type ReportRepository struct {
	db *pgxpool.Pool
}

func NewReportRepository(db *pgxpool.Pool) *ReportRepository {
	return &ReportRepository{db: db}
}

func (r *ReportRepository) Create(ctx context.Context, report models.Report) (models.CreatedReport, error) {
	snapshot, err := json.Marshal(report.ContentSnapshot)
	if err != nil {
		return models.CreatedReport{}, fmt.Errorf("encode report snapshot: %w", err)
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return models.CreatedReport{}, fmt.Errorf("begin report transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var id, status string
	var createdAt, dueAt time.Time
	err = tx.QueryRow(ctx, `
		INSERT INTO reports (
			reporter_id, target_type, target_id, reason, details, status,
			created_at, due_at, content_snapshot, version
		)
		VALUES (
			$1, $2, $3, $4, $5, 'pendiente',
			transaction_timestamp(), transaction_timestamp() + INTERVAL '48 hours',
			$6::jsonb, 1
		)
		RETURNING id::text, status, created_at, due_at
	`, report.ReporterID, report.TargetType, report.TargetID, report.Reason, report.Details, snapshot).
		Scan(&id, &status, &createdAt, &dueAt)
	if err != nil {
		var pgError *pgconn.PgError
		if errors.As(err, &pgError) && pgError.Code == "23505" && pgError.ConstraintName == openReportConstraint {
			return models.CreatedReport{}, ErrAlreadyReported
		}
		return models.CreatedReport{}, fmt.Errorf("insert report: %w", err)
	}

	eventPayload, err := json.Marshal(struct {
		Event      string    `json:"event"`
		ReportID   string    `json:"report_id"`
		ReporterID *string   `json:"reporter_id"`
		TargetType string    `json:"target_type"`
		TargetID   string    `json:"target_id"`
		Status     string    `json:"status"`
		CreatedAt  time.Time `json:"created_at"`
		DueAt      time.Time `json:"due_at"`
	}{
		Event: "report.created", ReportID: id, ReporterID: report.ReporterID,
		TargetType: string(report.TargetType), TargetID: report.TargetID, Status: status,
		CreatedAt: createdAt, DueAt: dueAt,
	})
	if err != nil {
		return models.CreatedReport{}, fmt.Errorf("encode report event: %w", err)
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO report_event_outbox (routing_key, payload)
		VALUES ('report.created', $1::jsonb)
	`, eventPayload); err != nil {
		return models.CreatedReport{}, fmt.Errorf("queue report event: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return models.CreatedReport{}, fmt.Errorf("commit report transaction: %w", err)
	}

	return models.CreatedReport{
		ID:     id,
		Status: status,
		DueAt:  dueAt.UTC().Format(time.RFC3339Nano),
	}, nil
}

type EventPublisher interface {
	Publish(context.Context, string, []byte) error
}

func RunOutboxDispatcher(
	ctx context.Context,
	db *pgxpool.Pool,
	publisher EventPublisher,
	logger interface {
		Error(string, ...any)
	},
) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		if err := dispatchOne(ctx, db, publisher); err != nil && ctx.Err() == nil {
			logger.Error("report_event_dispatch_failed", "error", err)
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func dispatchOne(ctx context.Context, db *pgxpool.Pool, publisher EventPublisher) error {
	tx, err := db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var id int64
	var routingKey string
	var payload []byte
	err = tx.QueryRow(ctx, `
		SELECT id, routing_key, payload
		FROM report_event_outbox
		ORDER BY id
		LIMIT 1
		FOR UPDATE SKIP LOCKED
	`).Scan(&id, &routingKey, &payload)
	if errors.Is(err, pgx.ErrNoRows) {
		return tx.Commit(ctx)
	}
	if err != nil {
		return err
	}
	publishCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := publisher.Publish(publishCtx, routingKey, payload); err != nil {
		_, updateErr := tx.Exec(ctx, `
			UPDATE report_event_outbox
			SET attempts = attempts + 1, last_error = $2
			WHERE id = $1
		`, id, err.Error())
		if updateErr != nil {
			return fmt.Errorf("publish report event: %v; record failure: %w", err, updateErr)
		}
		if commitErr := tx.Commit(ctx); commitErr != nil {
			return fmt.Errorf("publish report event: %v; commit retry state: %w", err, commitErr)
		}
		return err
	}
	if _, err := tx.Exec(ctx, "DELETE FROM report_event_outbox WHERE id = $1", id); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
