package domain

import "time"

// EmailSendLog records send attempt outcomes for operational tracking.
type EmailSendLog struct {
	EmailID      *string
	RecipientTo  string
	SenderFrom   string
	Subject      string
	Provider     string
	ResultStatus string
	ErrorMessage *string
	CreatedAt    time.Time
}

// SendLogRepository defines persistence for send attempt logs.
type SendLogRepository interface {
	SaveSendLog(log *EmailSendLog) error
}
