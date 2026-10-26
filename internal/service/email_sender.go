package service

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"html/template"
	"log/slog"
	"net"
	"net/smtp"
	"strings"

	"shortix-api/internal/config"
)

//go:embed templates/*.html
var emailTemplatesFS embed.FS

type EmailSender interface {
	SendTemplate(ctx context.Context, to, subject, templateName string, data any) error
}

type NoopEmailSender struct {
	logger *slog.Logger
}

func NewNoopEmailSender(logger *slog.Logger) *NoopEmailSender {
	if logger == nil {
		logger = slog.Default()
	}
	return &NoopEmailSender{logger: logger}
}

func (n *NoopEmailSender) SendTemplate(_ context.Context, to, subject, templateName string, _ any) error {
	n.logger.Info("email sending disabled", "to", to, "subject", subject, "template", templateName)
	return nil
}

type SMTPSender struct {
	host      string
	port      string
	username  string
	password  string
	fromEmail string
	fromName  string
	logger    *slog.Logger
	tpls      *template.Template
}

func NewSMTPSender(cfg *config.Config, logger *slog.Logger) EmailSender {
	if logger == nil {
		logger = slog.Default()
	}

	host := strings.TrimSpace(cfg.SMTPHost)
	from := strings.TrimSpace(cfg.SMTPFromEmail)
	if host == "" || from == "" {
		logger.Info("smtp not configured; falling back to noop email sender")
		return NewNoopEmailSender(logger)
	}

	tpls, err := template.ParseFS(emailTemplatesFS, "templates/*.html")
	if err != nil {
		logger.Error("failed parsing email templates; falling back to noop sender", "error", err)
		return NewNoopEmailSender(logger)
	}

	return &SMTPSender{
		host:      host,
		port:      cfg.SMTPPort,
		username:  cfg.SMTPUsername,
		password:  cfg.SMTPPassword,
		fromEmail: from,
		fromName:  cfg.SMTPFromName,
		logger:    logger,
		tpls:      tpls,
	}
}

func (s *SMTPSender) SendTemplate(_ context.Context, to, subject, templateName string, data any) error {
	var body bytes.Buffer
	if err := s.tpls.ExecuteTemplate(&body, templateName, data); err != nil {
		return fmt.Errorf("render template %q: %w", templateName, err)
	}

	fromHeader := s.fromEmail
	if strings.TrimSpace(s.fromName) != "" {
		fromHeader = fmt.Sprintf("%s <%s>", s.fromName, s.fromEmail)
	}

	var msg bytes.Buffer
	msg.WriteString(fmt.Sprintf("From: %s\r\n", fromHeader))
	msg.WriteString(fmt.Sprintf("To: %s\r\n", to))
	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	msg.WriteString("MIME-Version: 1.0\r\n")
	msg.WriteString("Content-Type: text/html; charset=\"UTF-8\"\r\n")
	msg.WriteString("\r\n")
	msg.Write(body.Bytes())

	addr := net.JoinHostPort(s.host, s.port)
	var auth smtp.Auth
	if strings.TrimSpace(s.username) != "" {
		auth = smtp.PlainAuth("", s.username, s.password, s.host)
	}

	if err := smtp.SendMail(addr, auth, s.fromEmail, []string{to}, msg.Bytes()); err != nil {
		return fmt.Errorf("smtp send failed: %w", err)
	}

	s.logger.Info("email sent", "to", to, "subject", subject, "template", templateName)
	return nil
}
