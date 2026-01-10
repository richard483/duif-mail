package domain

import "time"

// Email represents the email entity
type Email struct {
	ID        string     `json:"id"`
	To        string     `json:"to"`
	From      string     `json:"from"`
	Subject   string     `json:"subject"`
	Body      string     `json:"body"`
	Status    string     `json:"status"`
	CreatedAt time.Time  `json:"created_at"`
	SentAt    *time.Time `json:"sent_at,omitempty"`
}

// EmailRepository defines the interface for email storage
type EmailRepository interface {
	Save(email *Email) error
	GetByID(id string) (*Email, error)
	GetAll() ([]*Email, error)
}
