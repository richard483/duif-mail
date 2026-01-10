package domain

// SendEmailRequest represents the request to send an email
type SendEmailRequest struct {
	To      string `json:"to" validate:"required,email"`
	From    string `json:"from" validate:"required,email"`
	Subject string `json:"subject" validate:"required"`
	Body    string `json:"body" validate:"required"`
}

// SendEmailResponse represents the response after sending an email
type SendEmailResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	EmailID string `json:"email_id,omitempty"`
}
