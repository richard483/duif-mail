package repository

import (
	"duif/internal/domain"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// InMemoryEmailRepository implements EmailRepository with in-memory storage
type InMemoryEmailRepository struct {
	emails    map[string]*domain.Email
	templates map[string]*domain.EmailTemplate
	sendLogs  []*domain.EmailSendLog
	mu        sync.RWMutex
}

// NewInMemoryEmailRepository creates a new in-memory repository with static demo data
func NewInMemoryEmailRepository() *InMemoryEmailRepository {
	repo := &InMemoryEmailRepository{
		emails:    make(map[string]*domain.Email),
		templates: make(map[string]*domain.EmailTemplate),
		sendLogs:  make([]*domain.EmailSendLog, 0),
	}

	// Add some static demo data
	repo.seedData()
	repo.seedTemplates()

	return repo
}

// seedData adds static demo emails
func (r *InMemoryEmailRepository) seedData() {
	now := time.Now()

	demoEmails := []*domain.Email{
		{
			ID:        uuid.New().String(),
			To:        "alice@example.com",
			From:      "bob@example.com",
			Subject:   "Welcome to GoMail",
			Body:      "This is a demo email showing the mail service in action.",
			Status:    "sent",
			CreatedAt: now.Add(-2 * time.Hour),
			SentAt:    func() *time.Time { t := now.Add(-2 * time.Hour); return &t }(),
		},
		{
			ID:        uuid.New().String(),
			To:        "charlie@example.com",
			From:      "admin@example.com",
			Subject:   "System Notification",
			Body:      "Your account has been activated successfully.",
			Status:    "sent",
			CreatedAt: now.Add(-1 * time.Hour),
			SentAt:    func() *time.Time { t := now.Add(-1 * time.Hour); return &t }(),
		},
		{
			ID:        uuid.New().String(),
			To:        "david@example.com",
			From:      "sales@example.com",
			Subject:   "Special Offer",
			Body:      "Check out our latest products with 20% discount!",
			Status:    "sent",
			CreatedAt: now.Add(-30 * time.Minute),
			SentAt:    func() *time.Time { t := now.Add(-30 * time.Minute); return &t }(),
		},
	}

	for _, email := range demoEmails {
		r.emails[email.ID] = email
	}
}

func (r *InMemoryEmailRepository) seedTemplates() {
	now := time.Now()
	r.templates["9f1b5f8b-9bb2-4a8e-9b8f-2dc92f9c0f11"] = &domain.EmailTemplate{
		ID:        "9f1b5f8b-9bb2-4a8e-9b8f-2dc92f9c0f11",
		Name:      "Welcome",
		Subject:   "Welcome!",
		IsHTML:    true,
		CreatedAt: now,
		UpdatedAt: now,
		Body: `
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
	}
	r.templates["0b74f6d5-0f9f-4d26-9f90-90a43d4d4f22"] = &domain.EmailTemplate{
		ID:        "0b74f6d5-0f9f-4d26-9f90-90a43d4d4f22",
		Name:      "Graduation Announcement",
		Subject:   "Graduation Announcement",
		IsHTML:    true,
		CreatedAt: now,
		UpdatedAt: now,
		Body: `
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
}

// Save saves an email to the repository
func (r *InMemoryEmailRepository) Save(email *domain.Email) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if email.ID == "" {
		return fmt.Errorf("email ID cannot be empty")
	}

	r.emails[email.ID] = email
	return nil
}

// GetByID retrieves an email by its ID
func (r *InMemoryEmailRepository) GetByID(id string) (*domain.Email, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	email, exists := r.emails[id]
	if !exists {
		return nil, fmt.Errorf("%w: email not found with ID: %s", domain.ErrNotFound, id)
	}

	return email, nil
}

// GetAll retrieves all emails
func (r *InMemoryEmailRepository) GetAll() ([]*domain.Email, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	emails := make([]*domain.Email, 0, len(r.emails))
	for _, email := range r.emails {
		emails = append(emails, email)
	}

	return emails, nil
}

// GetTemplateByID retrieves an email template by its ID.
func (r *InMemoryEmailRepository) GetTemplateByID(id string) (*domain.EmailTemplate, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tmpl, exists := r.templates[id]
	if !exists {
		return nil, fmt.Errorf("%w: template not found with ID: %s", domain.ErrNotFound, id)
	}

	return tmpl, nil
}

// GetAllTemplates returns all templates.
func (r *InMemoryEmailRepository) GetAllTemplates() ([]*domain.EmailTemplate, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	templates := make([]*domain.EmailTemplate, 0, len(r.templates))
	for _, tmpl := range r.templates {
		templates = append(templates, tmpl)
	}
	return templates, nil
}

// SaveSendLog stores send attempt metadata for in-memory runtime.
func (r *InMemoryEmailRepository) SaveSendLog(logEntry *domain.EmailSendLog) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.sendLogs = append(r.sendLogs, logEntry)
	return nil
}
