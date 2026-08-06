package email

import (
	"fmt"
	"log"
	"net/smtp"
	"stationery-management/internal/config"
)

type EmailService struct {
	config *config.Config
}

func NewEmailService(cfg *config.Config) *EmailService {
	return &EmailService{config: cfg}
}

func (e *EmailService) SendEmail(to string, subject string, body string) error {
	if e.config.SMTPHost == "" || e.config.SMTPUser == "" {
		log.Printf("[MOCK EMAIL] To: %s | Subject: %s | Body: %s\n", to, subject, body)
		return nil
	}

	auth := smtp.PlainAuth("", e.config.SMTPUser, e.config.SMTPPassword, e.config.SMTPHost)
	msg := []byte(fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s", e.config.SMTPFrom, to, subject, body))

	err := smtp.SendMail(fmt.Sprintf("%s:%s", e.config.SMTPHost, e.config.SMTPPort), auth, e.config.SMTPUser, []string{to}, msg)
	if err != nil {
		log.Printf("[EMAIL SENDER ERROR] %v - Fallback logging email instead", err)
		log.Printf("[FALLBACK EMAIL] To: %s | Subject: %s | Body: %s\n", to, subject, body)
		return nil
	}

	return nil
}
