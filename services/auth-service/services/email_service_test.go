package services

import (
	"bytes"
	"errors"
	"log/slog"
	"strings"
	"testing"

	"github.com/meloop/auth-service/config"
)

func TestPasswordRecoveryEmailSimulationDoesNotLogToken(t *testing.T) {
	var output bytes.Buffer
	service := &smtpEmailService{
		cfg:    &config.Config{},
		logger: slog.New(slog.NewJSONHandler(&output, nil)),
	}
	recoveryLink := "https://example.test/reset?token=recovery-test-token"

	if err := service.SendPasswordRecoveryEmail(t.Context(), "test@example.test", recoveryLink); err != nil {
		t.Fatalf("SendPasswordRecoveryEmail returned an unexpected error")
	}

	log := output.String()
	if strings.Contains(log, recoveryLink) || strings.Contains(log, "recovery-test-token") {
		t.Errorf("log output contains a password recovery token")
	}
	if !strings.Contains(log, "password_recovery_email_simulated") {
		t.Errorf("log output is missing the email simulation event")
	}
}

func TestLogSMTPFailureDoesNotExposeErrorDetails(t *testing.T) {
	var output bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&output, nil))
	err := errors.New("SMTP authentication failed: smtp-test-secret smtp-test-token")

	logSMTPFailure(logger, "test@example.test", err)

	log := output.String()
	if strings.Contains(log, "smtp-test-secret") || strings.Contains(log, "smtp-test-token") {
		t.Errorf("log output contains sensitive SMTP error details")
	}
	if !strings.Contains(log, "smtp_send_mail_failed") {
		t.Errorf("log output is missing the SMTP failure event")
	}
}
