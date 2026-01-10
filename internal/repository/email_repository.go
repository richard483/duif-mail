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
	emails map[string]*domain.Email
	mu     sync.RWMutex
}

// NewInMemoryEmailRepository creates a new in-memory repository with static demo data
func NewInMemoryEmailRepository() *InMemoryEmailRepository {
	repo := &InMemoryEmailRepository{
		emails: make(map[string]*domain.Email),
	}

	// Add some static demo data
	repo.seedData()

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
		return nil, fmt.Errorf("email not found with ID: %s", id)
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
