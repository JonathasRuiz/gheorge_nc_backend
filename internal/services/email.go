package services

import (
	"fmt"
	"net/smtp"

	"gheorghe_nc_website/internal/config"
)

type EmailService struct {
	cfg *config.Config
}

func NewEmailService(cfg *config.Config) *EmailService {
	return &EmailService{cfg: cfg}
}

func (s *EmailService) SendContactEmail(name, email, phone, message string) error {
	if s.cfg.SMTPUser == "" || s.cfg.ContactForwardTo == "" {
		return fmt.Errorf("SMTP not configured")
	}

	subject := "New Contact Form Submission"
	body := fmt.Sprintf(
		"Name: %s\nEmail: %s\nPhone: %s\n\nMessage:\n%s",
		name, email, phone, message,
	)

	return s.send(s.cfg.ContactForwardTo, subject, body)
}

func (s *EmailService) send(to, subject, body string) error {
	addr := s.cfg.SMTPHost + ":" + s.cfg.SMTPPort
	auth := smtp.PlainAuth("", s.cfg.SMTPUser, s.cfg.SMTPPassword, s.cfg.SMTPHost)

	msg := []byte(
		"From: " + s.cfg.SMTPUser + "\r\n" +
			"To: " + to + "\r\n" +
			"Subject: " + subject + "\r\n" +
			"MIME-Version: 1.0\r\n" +
			"Content-Type: text/plain; charset=\"utf-8\"\r\n" +
			"\r\n" +
			body + "\r\n",
	)

	return smtp.SendMail(addr, auth, s.cfg.SMTPUser, []string{to}, msg)
}
