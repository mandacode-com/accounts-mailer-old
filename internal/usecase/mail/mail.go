package mail

import (
	"bytes"
	"html/template"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"gopkg.in/gomail.v2"
)

type MailUsecase struct {
	dialer              *gomail.Dialer
	verifyEmailTemplate *template.Template
	logger              *zap.Logger
	senderName          string
	senderEmail         string
}

// SendEmailVerificationMail sends an email verification mail to the user.
func (m *MailUsecase) SendEmailVerificationMail(email string, link string) error {
	data := struct {
		Link string
	}{
		Link: link,
	}

	var body bytes.Buffer
	if err := m.verifyEmailTemplate.Execute(&body, data); err != nil {
		m.logger.Error("failed to execute email template", zap.Error(err), zap.String("to", email))
		return err
	}

	msg := gomail.NewMessage()
	msg.SetAddressHeader("From", m.senderEmail, m.senderName)
	msg.SetHeader("To", email)
	msg.SetHeader("Subject", "[Mandacode] Email Verification")
	msg.SetBody("text/html", body.String())

	if err := m.dialer.DialAndSend(msg); err != nil {
		m.logger.Error("failed to send email", zap.Error(err), zap.String("to", email))
		return err
	}

	m.logger.Info("email sent successfully", zap.String("to", email))
	return nil
}

// NewMailUsecase creates a new instance of MailApp with the provided SMTP configuration.
func NewMailUsecase(host string, port int, senderName string, senderEmail string, dialer *gomail.Dialer, logger *zap.Logger) (*MailUsecase, error) {
	cwd, err := os.Getwd()
	tmplPath := filepath.Join(cwd, "template", "verify_email.html")
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		logger.Error("failed to parse email template", zap.Error(err))
		return nil, err
	}

	return &MailUsecase{
		dialer:              dialer,
		verifyEmailTemplate: tmpl,
		logger:              logger,
		senderName:          senderName,
		senderEmail:         senderEmail,
	}, nil
}
