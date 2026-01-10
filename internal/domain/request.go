package domain

// SendEmailRequest represents the request to send an email
type SendEmailRequest struct {
	To           string                 `json:"to" validate:"required,email"`
	From         string                 `json:"from" validate:"required,email"`
	Subject      string                 `json:"subject" validate:"required"`
	Body         string                 `json:"body" validate:"required"`
	IsHTML       bool                   `json:"is_html,omitempty"`       // Whether the body is HTML
	ReplyTo      string                 `json:"reply_to,omitempty"`      // Optional reply-to address
	Cc           string                 `json:"cc,omitempty"`            // Optional CC addresses
	TemplateId   string                 `json:"template_id,omitempty"`   // Optional template ID
	TemplateData map[string]interface{} `json:"template_data,omitempty"` // Data for template
}

// SendEmailResponse represents the response after sending an email
type SendEmailResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	EmailID string `json:"email_id,omitempty"`
}
