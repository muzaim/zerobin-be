package email

import (
	"bytes"
	"fmt"
	"html/template"
	"net/smtp"
)

type Service interface {
	SendHTML(to string, subject string, body string) error
	SendPlain(to string, subject string, body string) error
	SendTemplate(to string, subject string, templatePath string, data interface{}) error
}

type emailService struct {
	host     string
	port     int
	user     string
	password string
	sender   string
}

func NewEmailService(host string, port int, user, password, sender string) Service {
	return &emailService{
		host:     host,
		port:     port,
		user:     user,
		password: password,
		sender:   sender,
	}
}

func (s *emailService) SendHTML(to string, subject string, body string) error {
	addr := fmt.Sprintf("%s:%d", s.host, s.port)
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	msg := []byte(fmt.Sprintf("To: %s\r\nSubject: %s\r\n%s%s", to, subject, mime, body))

	var auth smtp.Auth
	if s.user != "" && s.password != "" {
		auth = smtp.PlainAuth("", s.user, s.password, s.host)
	}

	return smtp.SendMail(addr, auth, s.sender, []string{to}, msg)
}

func (s *emailService) SendPlain(to string, subject string, body string) error {
	addr := fmt.Sprintf("%s:%d", s.host, s.port)
	msg := []byte(fmt.Sprintf("To: %s\r\nSubject: %s\r\n\r\n%s", to, subject, body))

	var auth smtp.Auth
	if s.user != "" && s.password != "" {
		auth = smtp.PlainAuth("", s.user, s.password, s.host)
	}

	return smtp.SendMail(addr, auth, s.sender, []string{to}, msg)
}

func (s *emailService) SendTemplate(to string, subject string, templatePath string, data interface{}) error {
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	var body bytes.Buffer
	if err := tmpl.Execute(&body, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return s.SendHTML(to, subject, body.String())
}
