package usecase

import (
	"bytes"
	"duif/internal/config"
	"duif/internal/domain"
	"fmt"
	"html/template"
	"log"
	"net/smtp"
	"strings"
	"time"

	"github.com/google/uuid"
)

// MailUseCase handles business logic for email operations
type MailUseCase struct {
	repo domain.EmailRepository
}

// Ensure MailUseCase implements MailService interface
var _ domain.MailService = (*MailUseCase)(nil)

// NewMailUseCase creates a new mail use case
func NewMailUseCase(repo domain.EmailRepository) *MailUseCase {
	return &MailUseCase{
		repo: repo,
	}
}

// SendEmail processes the email sending request
func (u *MailUseCase) SendEmail(req *domain.SendEmailRequest) (*domain.SendEmailResponse, error) {
	// Validate request (you can add more validation here)
	if req.To == "" || req.From == "" || req.Subject == "" {
		return &domain.SendEmailResponse{
			Success: false,
			Message: "Missing required fields",
		}, fmt.Errorf("validation failed: missing required fields")
	}

	// Create email entity
	now := time.Now()
	email := &domain.Email{
		ID:        uuid.New().String(),
		To:        req.To,
		From:      req.From,
		Subject:   req.Subject,
		Body:      req.Body,
		Status:    "sent",
		CreatedAt: now,
		SentAt:    &now,
	}

	log.Printf("Sending email to %s from %s with subject '%s'", req.To, req.From, req.Subject)

	// Process template if provided
	body := req.Body
	if req.TemplateId != "" && req.TemplateData != nil {
		renderedBody, err := renderTemplate(req.TemplateId, req.TemplateData)
		if err != nil {
			return &domain.SendEmailResponse{
				Success: false,
				Message: fmt.Sprintf("Failed to render template: %v", err),
			}, err
		}
		body = renderedBody
		// Templates are HTML by default
		req.IsHTML = true
	}

	err := sendMail(req.To, req.From, req.Subject, body, req.IsHTML, req.ReplyTo, req.Cc)

	if err != nil {
		return &domain.SendEmailResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to send email: %v", err),
		}, err
	}

	// Save to repository
	if err := u.repo.Save(email); err != nil {
		return &domain.SendEmailResponse{
			Success: false,
			Message: "Failed to save email",
		}, err
	}

	return &domain.SendEmailResponse{
		Success: true,
		Message: "Email sent successfully",
		EmailID: email.ID,
	}, nil
}

// GetEmail retrieves an email by ID
func (u *MailUseCase) GetEmail(id string) (*domain.Email, error) {
	return u.repo.GetByID(id)
}

// GetAllEmails retrieves all emails
func (u *MailUseCase) GetAllEmails() ([]*domain.Email, error) {
	return u.repo.GetAll()
}

func sendMail(to, from, subject, body string, isHTML bool, replyTo, cc string) error {
	auth := smtp.PlainAuth("", config.Get().Mail.Username, config.Get().Mail.Password, config.Get().Mail.Host)
	smtpAddr := fmt.Sprintf("%s:%d", config.Get().Mail.Host, config.Get().Mail.Port)

	// Build email headers
	headers := make(map[string]string)
	headers["From"] = from
	headers["To"] = to
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"

	if replyTo != "" {
		headers["Reply-To"] = replyTo
	}
	if cc != "" {
		headers["Cc"] = cc
	}

	// Set content type based on whether it's HTML or plain text
	if isHTML {
		headers["Content-Type"] = "text/html; charset=UTF-8"
	} else {
		headers["Content-Type"] = "text/plain; charset=UTF-8"
	}

	// Build the message
	var msg bytes.Buffer
	for k, v := range headers {
		msg.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	msg.WriteString("\r\n")
	msg.WriteString(body)

	// Prepare recipient list
	recipients := []string{to}
	if cc != "" {
		recipients = append(recipients, cc)
	}

	err := smtp.SendMail(smtpAddr, auth, from, recipients, msg.Bytes())
	return err
}

// renderTemplate renders an HTML template with the provided data
func renderTemplate(templateName string, data map[string]interface{}) (string, error) {
	// Define your templates here or load from files
	templates := map[string]string{
		"welcome": `
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #4CAF50; color: white; padding: 20px; text-align: center; }
        .content { padding: 20px; background-color: #f9f9f9; }
        .footer { text-align: center; padding: 10px; font-size: 12px; color: #777; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Welcome {{.Name}}!</h1>
        </div>
        <div class="content">
            <p>{{.Message}}</p>
        </div>
        <div class="footer">
            <p>This is an automated email. Please do not reply.</p>
        </div>
    </div>
</body>
</html>`,
		"graduation_announcement": `
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; background-color: #f5f2ec; }
        .container { max-width: 640px; margin: 0 auto; padding: 24px; }
        .card { background-color: #ffffff; border-radius: 12px; padding: 28px; box-shadow: 0 6px 24px rgba(0, 0, 0, 0.08); }
        .banner { background-color: #1f3a5f; color: #fff; padding: 18px 24px; border-radius: 10px; text-align: center; }
        .badge { display: inline-block; background-color: #f4c430; color: #1f3a5f; font-weight: bold; padding: 6px 12px; border-radius: 999px; margin-top: 16px; }
        .detail { background-color: #f7f7f7; padding: 16px; border-radius: 8px; margin-top: 16px; }
        .footer { text-align: center; margin-top: 24px; font-size: 12px; color: #666; }
    </style>
</head>
<body>
    <div class="container">
        <div class="card">
            <div class="banner">
                <h1>Graduation Announcement</h1>
                <p>Congratulations {{index . "name"}}!</p>
            </div>
            <div class="badge">Status: {{index . "success"}}</div>
            <div class="detail">
                <p>{{index . "detail"}}</p>
            </div>
            <div class="footer">
                <p>We are proud to celebrate your achievement.</p>
            </div>
        </div>
    </div>
</body>
</html>`,
	}

	tmplStr, exists := templates[templateName]
	if !exists {
		return "", fmt.Errorf("template '%s' not found", templateName)
	}

	normalizedData := make(map[string]interface{}, len(data)*2)
	for key, value := range data {
		normalizedData[key] = value
		lowerKey := strings.ToLower(key)
		if lowerKey != key {
			normalizedData[lowerKey] = value
		}
		if key != "" {
			titleKey := strings.ToUpper(key[:1]) + key[1:]
			if titleKey != key {
				normalizedData[titleKey] = value
			}
		}
	}

	tmpl, err := template.New(templateName).Parse(tmplStr)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, normalizedData); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}
