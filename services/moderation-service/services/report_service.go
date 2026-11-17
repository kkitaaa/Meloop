package services

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/meloop/moderation-service/models"
	"github.com/meloop/moderation-service/repositories"
)

var (
	ErrCannotReportSelf = errors.New("CANNOT_REPORT_SELF")
	ErrTargetNotFound   = errors.New("TARGET_NOT_FOUND")
	ErrAlreadyReported  = errors.New("ALREADY_REPORTED")
	ErrRateLimit        = errors.New("RATE_LIMIT_EXCEEDED")
	ErrInvalidReport    = errors.New("INVALID_REPORT")
)

type ReportRepository interface {
	Create(context.Context, models.Report) (models.CreatedReport, error)
}

type ContentResolver interface {
	Resolve(context.Context, string, models.TargetType, string) (models.TargetContent, error)
}

type RateLimiter interface {
	Allow(context.Context, string) (bool, error)
}

type ReportService struct {
	reports ReportRepository
	targets ContentResolver
	limiter RateLimiter
}

func NewReportService(reports ReportRepository, targets ContentResolver, limiter RateLimiter) *ReportService {
	return &ReportService{reports: reports, targets: targets, limiter: limiter}
}

func (s *ReportService) Create(
	ctx context.Context,
	reporterID string,
	request models.CreateReportRequest,
) (models.CreatedReport, error) {
	reporterID = strings.TrimSpace(reporterID)
	if reporterID == "" {
		return models.CreatedReport{}, ErrInvalidReport
	}
	if err := request.Validate(); err != nil {
		return models.CreatedReport{}, fmt.Errorf("%w: %v", ErrInvalidReport, err)
	}
	request.TargetID = strings.TrimSpace(request.TargetID)
	if request.TargetType == models.TargetUser && request.TargetID == reporterID {
		return models.CreatedReport{}, ErrCannotReportSelf
	}

	target, err := s.targets.Resolve(ctx, reporterID, request.TargetType, request.TargetID)
	if err != nil {
		return models.CreatedReport{}, err
	}
	if target.AuthorID == reporterID {
		return models.CreatedReport{}, ErrCannotReportSelf
	}

	allowed, err := s.limiter.Allow(ctx, reporterID)
	if err != nil {
		return models.CreatedReport{}, fmt.Errorf("check report rate limit: %w", err)
	}
	if !allowed {
		return models.CreatedReport{}, ErrRateLimit
	}

	created, err := s.reports.Create(ctx, models.Report{
		ReporterID:      &reporterID,
		TargetType:      request.TargetType,
		TargetID:        request.TargetID,
		Reason:          request.Reason,
		Details:         strings.TrimSpace(request.Details),
		ContentSnapshot: target.Snapshot,
	})
	if err != nil {
		return models.CreatedReport{}, MapRepositoryError(err)
	}
	return created, nil
}

func MapRepositoryError(err error) error {
	if errors.Is(err, repositories.ErrAlreadyReported) {
		return ErrAlreadyReported
	}
	return err
}
