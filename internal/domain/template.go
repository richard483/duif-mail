package domain

import "time"

// EmailTemplate represents a reusable email template.
type EmailTemplate struct {
	ID        string
	Name      string
	Subject   string
	Body      string
	IsHTML    bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

// TemplateRepository defines template lookup behavior.
type TemplateRepository interface {
	GetTemplateByID(id string) (*EmailTemplate, error)
	GetAllTemplates() ([]*EmailTemplate, error)
}
