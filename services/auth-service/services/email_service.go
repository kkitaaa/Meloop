package services

import (
	"context"
	"fmt"
	"log/slog"
	"net/smtp"
	"strings"

	"github.com/meloop/auth-service/config"
	"github.com/meloop/services/common/logging"
)

// EmailService define la interfaz para el envío de correos electrónicos
type EmailService interface {
	SendPasswordRecoveryEmail(ctx context.Context, toEmail string, recoveryLink string) error
}

type smtpEmailService struct {
	cfg    *config.Config
	logger *slog.Logger
}

// NewEmailService crea una instancia de EmailService configurada
func NewEmailService(cfg *config.Config) EmailService {
	return &smtpEmailService{
		cfg:    cfg,
		logger: logging.New("email-service"),
	}
}

func (s *smtpEmailService) SendPasswordRecoveryEmail(ctx context.Context, toEmail string, recoveryLink string) error {
	subject := "Recuperación de contraseña - Meloop"
	body := fmt.Sprintf(
		"Hola,\r\n\r\n"+
			"Has solicitado restablecer tu contraseña en Meloop.\r\n\r\n"+
			"Para continuar, haz clic en el siguiente enlace:\r\n"+
			"%s\r\n\r\n"+
			"Este enlace expirará en 30 minutos y es de un solo uso.\r\n"+
			"Si no solicitaste este cambio, puedes ignorar este mensaje.\r\n\r\n"+
			"Equipo de Meloop",
		recoveryLink,
	)

	// Si no hay configuración SMTP (modo desarrollo/testing), registrar en logs y continuar
	if s.cfg.SMTPHost == "" {
		s.logger.Info("password_recovery_email_simulated", "to", toEmail, "link", recoveryLink)
		return nil
	}

	msg := []byte(strings.Join([]string{
		fmt.Sprintf("From: %s", s.cfg.SMTPFrom),
		fmt.Sprintf("To: %s", toEmail),
		fmt.Sprintf("Subject: %s", subject),
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"",
		body,
	}, "\r\n"))

	var auth smtp.Auth
	if s.cfg.SMTPUser != "" && s.cfg.SMTPPassword != "" {
		auth = smtp.PlainAuth("", s.cfg.SMTPUser, s.cfg.SMTPPassword, s.cfg.SMTPHost)
	}

	addr := fmt.Sprintf("%s:%d", s.cfg.SMTPHost, s.cfg.SMTPPort)
	err := smtp.SendMail(addr, auth, s.cfg.SMTPFrom, []string{toEmail}, msg)
	if err != nil {
		s.logger.Error("smtp_send_mail_failed", "error", err, "to", toEmail)
		return fmt.Errorf("error sending email via SMTP: %w", err)
	}

	s.logger.Info("password_recovery_email_sent", "to", toEmail)
	return nil
}
