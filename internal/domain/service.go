package domain

// MailService defines the business logic interface for mail operations
type MailService interface {
	SendEmail(req *SendEmailRequest) (*SendEmailResponse, error)
	GetEmail(id string) (*Email, error)
	GetAllEmails() ([]*Email, error)
	GetTemplates() ([]*EmailTemplate, error)
}
