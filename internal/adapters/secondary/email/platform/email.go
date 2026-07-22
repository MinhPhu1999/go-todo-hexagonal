package platform

import (
	"fmt"
	"log/slog"
	"net/smtp"

	"go-crud-db-p2/config"
	ports "go-crud-db-p2/internal/core/ports/platform"
)

type ConsoleEmailSender struct {
	logger *slog.Logger
}

func NewConsoleEmailSender() *ConsoleEmailSender {
	return &ConsoleEmailSender{logger: slog.Default()}
}

func (s *ConsoleEmailSender) SendOTP(to string, otp string) error {
	s.logger.Info("password reset otp",
		"to", to,
		"otp", otp,
	)
	return nil
}

type SMTPSender struct {
	host     string
	port     int
	user     string
	password string
	fromAddr string
	fromName string
	logger   *slog.Logger
}

func NewSMTPSender(cfg config.EmailConfig) *SMTPSender {
	return &SMTPSender{
		host:     cfg.SMTPHost,
		port:     cfg.SMTPPort,
		user:     cfg.SMTPUser,
		password: cfg.SMTPPassword,
		fromAddr: cfg.FromAddress,
		fromName: cfg.FromName,
		logger:   slog.Default(),
	}
}

func (s *SMTPSender) SendOTP(to string, otp string) error {
	subject := "Password Reset OTP"
	body := fmt.Sprintf("Your OTP for password reset is: %s\n\nThis OTP will expire shortly.", otp)

	msg := []byte(fmt.Sprintf("From: %s <%s>\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=\"UTF-8\"\r\n\r\n%s",
		s.fromName, s.fromAddr, to, subject, body))

	addr := fmt.Sprintf("%s:%d", s.host, s.port)

	var auth smtp.Auth
	if s.user != "" {
		auth = smtp.PlainAuth("", s.user, s.password, s.host)
	}

	if err := smtp.SendMail(addr, auth, s.fromAddr, []string{to}, msg); err != nil {
		s.logger.Error("failed to send email",
			"to", to,
			"error", err,
		)
		return fmt.Errorf("send email: %w", err)
	}

	s.logger.Info("email sent", "to", to)
	return nil
}

func NewEmailSender(cfg config.Config) ports.IEmailSender {
	if cfg.Email.SMTPHost != "" {
		return NewSMTPSender(cfg.Email)
	}
	return NewConsoleEmailSender()
}
