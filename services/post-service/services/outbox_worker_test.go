package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/meloop/post-service/models"
)

type memoryOutboxRepo struct {
	events       map[string]models.OutboxEvent
	publishedIDs []string
}

func (m *memoryOutboxRepo) ClaimPendingEvents(ctx context.Context, limit int, leaseDuration time.Duration) ([]models.OutboxEvent, error) {
	var out []models.OutboxEvent
	now := time.Now().UTC()
	for id, ev := range m.events {
		if (ev.Status == models.OutboxStatusPending || ev.Status == models.OutboxStatusProcessing) && !ev.NextRetryAt.After(now) {
			ev.Status = models.OutboxStatusProcessing
			ev.NextRetryAt = now.Add(leaseDuration)
			m.events[id] = ev
			out = append(out, ev)
			if len(out) >= limit {
				break
			}
		}
	}
	return out, nil
}

func (m *memoryOutboxRepo) MarkPublished(ctx context.Context, id string) error {
	ev := m.events[id]
	ev.Status = models.OutboxStatusPublished
	now := time.Now().UTC()
	ev.PublishedAt = &now
	m.events[id] = ev
	m.publishedIDs = append(m.publishedIDs, id)
	return nil
}

func (m *memoryOutboxRepo) MarkRetry(ctx context.Context, id string, attempts int, nextRetryAt time.Time, lastError string) error {
	ev := m.events[id]
	ev.Status = models.OutboxStatusPending
	ev.Attempts = attempts
	ev.NextRetryAt = nextRetryAt
	ev.LastError = lastError
	m.events[id] = ev
	return nil
}

func TestOutboxWorker_RetriesAfterBrokerOrUnroutableFailureAndPublishes(t *testing.T) {
	repo := &memoryOutboxRepo{
		events: map[string]models.OutboxEvent{
			"ev-1": {
				ID:          "ev-1",
				RoutingKey:  "comment.created",
				Payload:     `{"event":"comment.created","comment_id":"c-1"}`,
				Status:      models.OutboxStatusPending,
				NextRetryAt: time.Now().UTC().Add(-time.Second),
			},
		},
	}

	shouldFail := true
	worker := NewOutboxWorker(repo, func(routingKey string, payload []byte) error {
		if shouldFail {
			return errors.New("mensaje comment.created no enrutable (sin cola enlazada)")
		}
		return nil
	})

	// 1. Primer intento falla: debe mantenerse PENDING con attempts=1
	published, err := worker.ProcessPendingOnce(context.Background())
	if err != nil {
		t.Fatalf("error inesperado en ProcessPendingOnce: %v", err)
	}
	if published != 0 {
		t.Fatalf("esperados 0 publicados cuando RabbitMQ falla, obtenidos %d", published)
	}
	if repo.events["ev-1"].Status != models.OutboxStatusPending || repo.events["ev-1"].Attempts != 1 {
		t.Fatalf("el evento debía quedar PENDING para reintento: %+v", repo.events["ev-1"])
	}

	// 2. Simular vencimiento del backoff y recuperación de cola enlazada
	ev := repo.events["ev-1"]
	ev.NextRetryAt = time.Now().UTC().Add(-time.Second)
	repo.events["ev-1"] = ev
	shouldFail = false

	published, err = worker.ProcessPendingOnce(context.Background())
	if err != nil {
		t.Fatalf("error inesperado en segundo intento: %v", err)
	}
	if published != 1 || repo.events["ev-1"].Status != models.OutboxStatusPublished {
		t.Fatalf("el evento debía marcarse PUBLISHED tras recuperarse RabbitMQ: %+v", repo.events["ev-1"])
	}
}
