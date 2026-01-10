package usecase

import (
	"duif/internal/domain"
	"fmt"
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
