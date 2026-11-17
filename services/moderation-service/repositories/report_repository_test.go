package repositories

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/meloop/moderation-service/models"
)

func TestReportRepositoryUsesServerDeadlineAndRejectsOpenDuplicates(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL no está configurada; se omite la prueba de integración")
	}

	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		t.Fatalf("parse TEST_DATABASE_URL: %v", err)
	}
	if !strings.Contains(strings.ToLower(config.ConnConfig.Database), "test") {
		t.Fatal("TEST_DATABASE_URL debe apuntar a una base de datos cuyo nombre incluya 'test'")
	}

	ctx := context.Background()
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	t.Cleanup(pool.Close)
	if err := pool.Ping(ctx); err != nil {
		t.Fatalf("connect to test database: %v", err)
	}

	migration, err := os.ReadFile(filepath.Join("..", "..", "..", "supabase", "migrations", "20261116230000_moderation_reports.sql"))
	if err != nil {
		t.Fatalf("read moderation migration: %v", err)
	}
	if _, err := pool.Exec(ctx, string(migration)); err != nil {
		t.Fatalf("apply moderation migration to test database: %v", err)
	}

	reporterID := fmt.Sprintf("moderation-test-%d", time.Now().UnixNano())
	repository := NewReportRepository(pool)
	created, err := repository.Create(ctx, models.Report{
		ReporterID: &reporterID,
		TargetType: models.TargetPost,
		TargetID:   "post-1",
		Reason:     "spam",
		ContentSnapshot: models.ContentSnapshot{
			Text: "snapshot before edits", Author: "author-1", AuthorName: "Artist",
			URLs: []string{"https://example.test/post"},
		},
	})
	if err != nil {
		t.Fatalf("create report: %v", err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(ctx, "DELETE FROM report_event_outbox WHERE payload->>'report_id' = $1", created.ID)
		_, _ = pool.Exec(ctx, "DELETE FROM reports WHERE id = $1", created.ID)
	})

	var createdAt, dueAt time.Time
	if err := pool.QueryRow(ctx, "SELECT created_at, due_at FROM reports WHERE id = $1", created.ID).
		Scan(&createdAt, &dueAt); err != nil {
		t.Fatalf("read server timestamps: %v", err)
	}
	if dueAt.Sub(createdAt) != 48*time.Hour {
		t.Fatalf("deadline interval = %s, want exactly 48h", dueAt.Sub(createdAt))
	}
	responseDeadline, err := time.Parse(time.RFC3339Nano, created.DueAt)
	if err != nil || !responseDeadline.Equal(dueAt) {
		t.Fatalf("response deadline %q does not match stored deadline %s", created.DueAt, dueAt)
	}

	_, err = repository.Create(ctx, models.Report{
		ReporterID: &reporterID,
		TargetType: models.TargetPost,
		TargetID:   "post-1",
		Reason:     "spam",
		ContentSnapshot: models.ContentSnapshot{
			Text: "repeated report", Author: "author-1", URLs: []string{},
		},
	})
	if !errors.Is(err, ErrAlreadyReported) {
		t.Fatalf("duplicate create error = %v, want ErrAlreadyReported", err)
	}

}
