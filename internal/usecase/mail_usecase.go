package usecase

import (
	"bytes"
	"duif/internal/config"
	"duif/internal/domain"
	"fmt"
	"html/template"
	"log"
	"net/mail"
	"net/smtp"
	"strings"
	"time"

	"github.com/google/uuid"
)

// MailUseCase handles business logic for email operations
type MailUseCase struct {
	repo         domain.EmailRepository
	templateRepo domain.TemplateRepository
	sendLogRepo  domain.SendLogRepository
	sendMail     func(to, from, subject, body string, isHTML bool, replyTo string, cc []string) error
}

// Ensure MailUseCase implements MailService interface
var _ domain.MailService = (*MailUseCase)(nil)

// NewMailUseCase creates a new mail use case
func NewMailUseCase(repo domain.EmailRepository) *MailUseCase {
	var templateRepo domain.TemplateRepository
	if typedRepo, ok := repo.(domain.TemplateRepository); ok {
		templateRepo = typedRepo
	}
	var sendLogRepo domain.SendLogRepository
	if typedRepo, ok := repo.(domain.SendLogRepository); ok {
		sendLogRepo = typedRepo
	}

	return &MailUseCase{
		repo:         repo,
		templateRepo: templateRepo,
		sendLogRepo:  sendLogRepo,
		sendMail:     sendMail,
	}
}

// NewMailUseCaseWithSender allows tests to inject a sender implementation.
func NewMailUseCaseWithSender(repo domain.EmailRepository, sender func(to, from, subject, body string, isHTML bool, replyTo string, cc []string) error) *MailUseCase {
	if sender == nil {
		sender = sendMail
	}
	var templateRepo domain.TemplateRepository
	if typedRepo, ok := repo.(domain.TemplateRepository); ok {
		templateRepo = typedRepo
	}
	var sendLogRepo domain.SendLogRepository
	if typedRepo, ok := repo.(domain.SendLogRepository); ok {
		sendLogRepo = typedRepo
	}

	return &MailUseCase{
		repo:         repo,
		templateRepo: templateRepo,
		sendLogRepo:  sendLogRepo,
		sendMail:     sender,
	}
}

// SendEmail processes the email sending request
func (u *MailUseCase) SendEmail(req *domain.SendEmailRequest) (*domain.SendEmailResponse, error) {
	ccRecipients, err := validateSendEmailRequest(req)
	if err != nil {
		return &domain.SendEmailResponse{
			Success: false,
			Message: err.Error(),
		}, err
	}

	log.Printf("Sending email to %s from %s with subject '%s'", req.To, req.From, req.Subject)

	// Process template if provided
	body := req.Body
	if req.TemplateId != "" {
		if u.templateRepo == nil {
			return &domain.SendEmailResponse{
				Success: false,
				Message: "template repository is not configured",
			}, fmt.Errorf("template repository is not configured")
		}

		templateEntity, err := u.templateRepo.GetTemplateByID(req.TemplateId)
		if err != nil {
			return &domain.SendEmailResponse{
				Success: false,
				Message: fmt.Sprintf("Failed to load template: %v", err),
			}, err
		}

		renderedBody, err := renderTemplate(templateEntity.Body, req.TemplateData)
		if err != nil {
			return &domain.SendEmailResponse{
				Success: false,
				Message: fmt.Sprintf("Failed to render template: %v", err),
			}, err
		}
		body = renderedBody
		if templateEntity.IsHTML {
			req.IsHTML = true
		}
		if strings.TrimSpace(req.Subject) == "" {
			req.Subject = templateEntity.Subject
		}
	}

	// Create email entity with the final body that is actually sent.
	now := time.Now()
	email := &domain.Email{
		ID:        uuid.New().String(),
		To:        req.To,
		From:      req.From,
		Subject:   req.Subject,
		Body:      body,
		Status:    "pending",
		CreatedAt: now,
		SentAt:    nil,
	}

	err = u.sendMail(req.To, req.From, req.Subject, body, req.IsHTML, req.ReplyTo, ccRecipients)

	if err != nil {
		email.Status = "failed"
		if saveErr := u.repo.Save(email); saveErr != nil {
			email.ID = ""
		}
		u.saveSendLog(email, "failed", err)
		return &domain.SendEmailResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to send email: %v", err),
		}, err
	}

	sentAt := time.Now()
	email.Status = "sent"
	email.SentAt = &sentAt

	// Save to repository
	if err := u.repo.Save(email); err != nil {
		email.ID = ""
		u.saveSendLog(email, "failed", err)
		return &domain.SendEmailResponse{
			Success: false,
			Message: "Failed to save email",
		}, err
	}
	u.saveSendLog(email, "sent", nil)

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

// GetTemplates retrieves all templates.
func (u *MailUseCase) GetTemplates() ([]*domain.EmailTemplate, error) {
	if u.templateRepo == nil {
		return nil, fmt.Errorf("template repository is not configured")
	}
	return u.templateRepo.GetAllTemplates()
}

func sendMail(to, from, subject, body string, isHTML bool, replyTo string, cc []string) error {
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
	if len(cc) > 0 {
		headers["Cc"] = strings.Join(cc, ", ")
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
	if len(cc) > 0 {
		recipients = append(recipients, cc...)
	}

	err := smtp.SendMail(smtpAddr, auth, from, recipients, msg.Bytes())
	return err
}

func validateSendEmailRequest(req *domain.SendEmailRequest) ([]string, error) {
	if req == nil {
		return nil, fmt.Errorf("%w: request is required", domain.ErrInvalidInput)
	}
	if strings.TrimSpace(req.To) == "" || strings.TrimSpace(req.From) == "" {
		return nil, fmt.Errorf("%w: to and from are required", domain.ErrInvalidInput)
	}
	if strings.TrimSpace(req.Subject) == "" && strings.TrimSpace(req.TemplateId) == "" {
		return nil, fmt.Errorf("%w: subject is required when template_id is empty", domain.ErrInvalidInput)
	}
	if strings.TrimSpace(req.Body) == "" && strings.TrimSpace(req.TemplateId) == "" {
		return nil, fmt.Errorf("%w: body or template_id is required", domain.ErrInvalidInput)
	}
	if _, err := mail.ParseAddress(req.To); err != nil {
		return nil, fmt.Errorf("%w: invalid to address", domain.ErrInvalidInput)
	}
	if _, err := mail.ParseAddress(req.From); err != nil {
		return nil, fmt.Errorf("%w: invalid from address", domain.ErrInvalidInput)
	}
	if strings.TrimSpace(req.ReplyTo) != "" {
		if _, err := mail.ParseAddress(req.ReplyTo); err != nil {
			return nil, fmt.Errorf("%w: invalid reply_to address", domain.ErrInvalidInput)
		}
	}

	ccRecipients, err := parseRecipients(req.Cc)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid cc address", domain.ErrInvalidInput)
	}

	return ccRecipients, nil
}

func parseRecipients(raw string) ([]string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}

	parts := strings.Split(raw, ",")
	recipients := make([]string, 0, len(parts))
	for _, part := range parts {
		candidate := strings.TrimSpace(part)
		if candidate == "" {
			continue
		}
		if _, err := mail.ParseAddress(candidate); err != nil {
			return nil, err
		}
		recipients = append(recipients, candidate)
	}
	return recipients, nil
}

// renderTemplate renders an HTML template with the provided data
func renderTemplate(templateBody string, data map[string]interface{}) (string, error) {
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

	tmpl, err := template.New("email_template").Parse(templateBody)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, normalizedData); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

func (u *MailUseCase) saveSendLog(email *domain.Email, resultStatus string, sendErr error) {
	if u.sendLogRepo == nil {
		return
	}

	var errorMessage *string
	if sendErr != nil {
		msg := sendErr.Error()
		errorMessage = &msg
	}

	var emailID *string
	if strings.TrimSpace(email.ID) != "" {
		id := email.ID
		emailID = &id
	}
	logEntry := &domain.EmailSendLog{
		EmailID:      emailID,
		RecipientTo:  email.To,
		SenderFrom:   email.From,
		Subject:      email.Subject,
		Provider:     "smtp",
		ResultStatus: resultStatus,
		ErrorMessage: errorMessage,
		CreatedAt:    time.Now(),
	}

	if err := u.sendLogRepo.SaveSendLog(logEntry); err != nil {
		log.Printf("Failed to save send log for email %s: %v", email.ID, err)
	}
}
