package services

import (
	"context"
	"errors"
	"testing"

	"github.com/meloop/moderation-service/models"
	"github.com/meloop/moderation-service/repositories"
)

type fakeReportRepository struct {
	report models.Report
	err    error
}

func (r *fakeReportRepository) Create(_ context.Context, report models.Report) (models.CreatedReport, error) {
	r.report = report
	if r.err != nil {
		return models.CreatedReport{}, r.err
	}
	return models.CreatedReport{ID: "report-1", Status: models.ReportPending, DueAt: "2026-11-18T00:00:00Z"}, nil
}

type fakeContentResolver struct {
	target models.TargetContent
	err    error
}

func (r fakeContentResolver) Resolve(context.Context, string, models.TargetType, string) (models.TargetContent, error) {
	return r.target, r.err
}

type fakeRateLimiter struct {
	allowed bool
	err     error
	calls   int
}

func (l *fakeRateLimiter) Allow(context.Context, string) (bool, error) {
	l.calls++
	return l.allowed, l.err
}

func TestReportServiceCreatesSnapshotForAllTargetTypes(t *testing.T) {
	for _, targetType := range []models.TargetType{models.TargetPost, models.TargetComment, models.TargetUser} {
		t.Run(string(targetType), func(t *testing.T) {
			repository := &fakeReportRepository{}
			limiter := &fakeRateLimiter{allowed: true}
			service := NewReportService(repository, fakeContentResolver{
				target: models.TargetContent{
					AuthorID: "author-2",
					Snapshot: models.ContentSnapshot{
						Text: "snapshot text", Author: "author-2", AuthorName: "Artist",
						URLs: []string{"https://example.test/image"},
					},
				},
			}, limiter)

			created, err := service.Create(context.Background(), "reporter-1", models.CreateReportRequest{
				TargetType: targetType, TargetID: "target-3", Reason: "spam",
			})
			if err != nil {
				t.Fatalf("Create() error = %v", err)
			}
			if created.Status != models.ReportPending || created.ID != "report-1" {
				t.Fatalf("unexpected report response: %+v", created)
			}
			if repository.report.ContentSnapshot.Text != "snapshot text" {
				t.Fatalf("snapshot was not stored: %+v", repository.report.ContentSnapshot)
			}
			if limiter.calls != 1 {
				t.Fatalf("rate limiter called %d times, want 1", limiter.calls)
			}
		})
	}
}

func TestReportServiceRejectsSelfReportsBeforeRateLimit(t *testing.T) {
	limiter := &fakeRateLimiter{allowed: true}
	service := NewReportService(
		&fakeReportRepository{},
		fakeContentResolver{target: models.TargetContent{AuthorID: "reporter-1"}},
		limiter,
	)

	_, err := service.Create(context.Background(), "reporter-1", models.CreateReportRequest{
		TargetType: models.TargetPost, TargetID: "post-1", Reason: "spam",
	})
	if !errors.Is(err, ErrCannotReportSelf) {
		t.Fatalf("Create() error = %v, want ErrCannotReportSelf", err)
	}
	if limiter.calls != 0 {
		t.Fatalf("rate limiter called %d times, want 0", limiter.calls)
	}
}

func TestReportServiceRejectsSameUserAndNotVisibleTargets(t *testing.T) {
	t.Run("same user", func(t *testing.T) {
		service := NewReportService(&fakeReportRepository{}, fakeContentResolver{}, &fakeRateLimiter{allowed: true})
		_, err := service.Create(context.Background(), "reporter-1", models.CreateReportRequest{
			TargetType: models.TargetUser, TargetID: "reporter-1", Reason: "spam",
		})
		if !errors.Is(err, ErrCannotReportSelf) {
			t.Fatalf("Create() error = %v, want ErrCannotReportSelf", err)
		}
	})

	t.Run("not visible", func(t *testing.T) {
		service := NewReportService(&fakeReportRepository{}, fakeContentResolver{err: ErrTargetNotFound}, &fakeRateLimiter{allowed: true})
		_, err := service.Create(context.Background(), "reporter-1", models.CreateReportRequest{
			TargetType: models.TargetComment, TargetID: "comment-1", Reason: "spam",
		})
		if !errors.Is(err, ErrTargetNotFound) {
			t.Fatalf("Create() error = %v, want ErrTargetNotFound", err)
		}
	})
}

func TestReportServiceMapsDuplicateReportError(t *testing.T) {
	repository := &fakeReportRepository{err: repositories.ErrAlreadyReported}
	service := NewReportService(repository, fakeContentResolver{
		target: models.TargetContent{AuthorID: "author-2"},
	}, &fakeRateLimiter{allowed: true})

	_, err := service.Create(context.Background(), "reporter-1", models.CreateReportRequest{
		TargetType: models.TargetPost, TargetID: "post-1", Reason: "spam",
	})
	if !errors.Is(err, ErrAlreadyReported) {
		t.Fatalf("Create() error = %v, want ErrAlreadyReported", err)
	}
}

func TestReportServiceEnforcesRateLimitAndDetailsValidation(t *testing.T) {
	t.Run("rate limit", func(t *testing.T) {
		service := NewReportService(&fakeReportRepository{}, fakeContentResolver{
			target: models.TargetContent{AuthorID: "author-2"},
		}, &fakeRateLimiter{allowed: false})
		_, err := service.Create(context.Background(), "reporter-1", models.CreateReportRequest{
			TargetType: models.TargetPost, TargetID: "post-1", Reason: "spam",
		})
		if !errors.Is(err, ErrRateLimit) {
			t.Fatalf("Create() error = %v, want ErrRateLimit", err)
		}
	})

	t.Run("invalid details", func(t *testing.T) {
		limiter := &fakeRateLimiter{allowed: true}
		service := NewReportService(&fakeReportRepository{}, fakeContentResolver{}, limiter)
		_, err := service.Create(context.Background(), "reporter-1", models.CreateReportRequest{
			TargetType: models.TargetPost, TargetID: "post-1", Reason: "otro",
		})
		if !errors.Is(err, ErrInvalidReport) {
			t.Fatalf("Create() error = %v, want ErrInvalidReport", err)
		}
		if limiter.calls != 0 {
			t.Fatalf("rate limiter called %d times, want 0", limiter.calls)
		}
	})
}
