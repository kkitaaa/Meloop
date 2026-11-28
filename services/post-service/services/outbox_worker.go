package services

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/meloop/post-service/repositories"
	"github.com/meloop/services/common/logging"
)

type OutboxWorker struct {
	repo         repositories.OutboxRepository
	publishFn    func(routingKey string, payload []byte) error
	wakeCh       chan struct{}
	pollInterval time.Duration
	batchSize    int
	leaseTime    time.Duration
}

func NewOutboxWorker(repo repositories.OutboxRepository, publishFn func(routingKey string, payload []byte) error) *OutboxWorker {
	return &OutboxWorker{
		repo:         repo,
		publishFn:    publishFn,
		wakeCh:       make(chan struct{}, 1),
		pollInterval: 3 * time.Second,
		batchSize:    25,
		leaseTime:    30 * time.Second,
	}
}

// Notify despierta de inmediato al worker tras confirmar una transacción en BD
func (w *OutboxWorker) Notify() {
	select {
	case w.wakeCh <- struct{}{}:
	default:
	}
}

// Start ejecuta el ciclo de despacho en segundo plano y recupera eventos pendientes al iniciar
func (w *OutboxWorker) Start(ctx context.Context) {
	logger := logging.New("post-service")
	logger.Info("outbox_worker_started", "poll_interval", w.pollInterval.String())

	// Procesar pendientes acumulados antes de un reinicio
	_, _ = w.ProcessPendingOnce(ctx)

	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Info("outbox_worker_stopped")
			return
		case <-w.wakeCh:
			_, _ = w.ProcessPendingOnce(ctx)
		case <-ticker.C:
			_, _ = w.ProcessPendingOnce(ctx)
		}
	}
}

func ensureOutboxEventID(outboxID, rawPayload string) []byte {
	id := strings.TrimSpace(outboxID)
	if id == "" {
		return []byte(rawPayload)
	}

	var obj map[string]any
	if err := json.Unmarshal([]byte(rawPayload), &obj); err != nil || obj == nil {
		return []byte(rawPayload)
	}

	if existing, ok := obj["event_id"].(string); ok && strings.TrimSpace(existing) != "" {
		return []byte(rawPayload)
	}

	obj["event_id"] = id
	enriched, err := json.Marshal(obj)
	if err != nil {
		return []byte(rawPayload)
	}
	return enriched
}

// ProcessPendingOnce reclama y publica un lote de eventos pendientes
func (w *OutboxWorker) ProcessPendingOnce(ctx context.Context) (int, error) {
	logger := logging.New("post-service")

	events, err := w.repo.ClaimPendingEvents(ctx, w.batchSize, w.leaseTime)
	if err != nil {
		logger.Error("outbox_claim_failed", "error", err)
		return 0, err
	}

	publishedCount := 0
	for _, ev := range events {
		payloadBytes := ensureOutboxEventID(ev.ID, ev.Payload)
		if err := w.publishFn(ev.RoutingKey, payloadBytes); err != nil {
			nextAttempt := ev.Attempts + 1
			backoff := computeOutboxBackoff(nextAttempt)
			nextRetry := time.Now().UTC().Add(backoff)
			_ = w.repo.MarkRetry(ctx, ev.ID, nextAttempt, nextRetry, err.Error())
			logger.Error("outbox_publish_retry_scheduled",
				"outbox_id", ev.ID,
				"routing_key", ev.RoutingKey,
				"attempt", nextAttempt,
				"next_retry_at", nextRetry.Format(time.RFC3339),
				"error", err,
			)
			continue
		}

		if err := w.repo.MarkPublished(ctx, ev.ID); err != nil {
			logger.Error("outbox_mark_published_failed", "outbox_id", ev.ID, "error", err)
			continue
		}
		publishedCount++
		logger.Info("outbox_event_published", "outbox_id", ev.ID, "routing_key", ev.RoutingKey)
	}

	return publishedCount, nil
}

func computeOutboxBackoff(attempt int) time.Duration {
	if attempt < 1 {
		attempt = 1
	}
	shift := attempt - 1
	if shift > 7 {
		shift = 7
	}
	backoff := 2 * time.Second * time.Duration(1<<shift)
	if backoff > 5*time.Minute {
		return 5 * time.Minute
	}
	return backoff
}
