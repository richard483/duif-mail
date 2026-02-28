package repository

import (
	"context"
	"database/sql"
	"duif/internal/domain"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(dsn string) (*PostgresRepository, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return &PostgresRepository{db: db}, nil
}

func (r *PostgresRepository) Close() error {
	return r.db.Close()
}

func (r *PostgresRepository) Save(email *domain.Email) error {
	const query = `
INSERT INTO emails (id, recipient_to, sender_from, subject, body, status, created_at, sent_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
`
	_, err := r.db.Exec(
		query,
		email.ID,
		email.To,
		email.From,
		email.Subject,
		email.Body,
		email.Status,
		email.CreatedAt.UTC(),
		email.SentAt,
	)
	if err != nil {
		return fmt.Errorf("insert email: %w", err)
	}
	return nil
}

func (r *PostgresRepository) GetByID(id string) (*domain.Email, error) {
	const query = `
SELECT id, recipient_to, sender_from, subject, body, status, created_at, sent_at
FROM emails
WHERE id = $1
`
	var email domain.Email
	var sentAt sql.NullTime
	err := r.db.QueryRow(query, id).Scan(
		&email.ID,
		&email.To,
		&email.From,
		&email.Subject,
		&email.Body,
		&email.Status,
		&email.CreatedAt,
		&sentAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("%w: email not found with ID: %s", domain.ErrNotFound, id)
		}
		return nil, fmt.Errorf("select email by id: %w", err)
	}
	if sentAt.Valid {
		t := sentAt.Time
		email.SentAt = &t
	}
	return &email, nil
}

func (r *PostgresRepository) GetAll() ([]*domain.Email, error) {
	const query = `
SELECT id, recipient_to, sender_from, subject, body, status, created_at, sent_at
FROM emails
ORDER BY created_at DESC
`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("select all emails: %w", err)
	}
	defer rows.Close()

	emails := make([]*domain.Email, 0)
	for rows.Next() {
		var email domain.Email
		var sentAt sql.NullTime
		if err := rows.Scan(
			&email.ID,
			&email.To,
			&email.From,
			&email.Subject,
			&email.Body,
			&email.Status,
			&email.CreatedAt,
			&sentAt,
		); err != nil {
			return nil, fmt.Errorf("scan email row: %w", err)
		}
		if sentAt.Valid {
			t := sentAt.Time
			email.SentAt = &t
		}
		emails = append(emails, &email)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate email rows: %w", err)
	}
	return emails, nil
}

func (r *PostgresRepository) GetTemplateByID(id string) (*domain.EmailTemplate, error) {
	const query = `
SELECT id, name, subject, body, is_html, created_at, updated_at
FROM email_templates
WHERE id = $1
`
	var tmpl domain.EmailTemplate
	err := r.db.QueryRow(query, id).Scan(
		&tmpl.ID,
		&tmpl.Name,
		&tmpl.Subject,
		&tmpl.Body,
		&tmpl.IsHTML,
		&tmpl.CreatedAt,
		&tmpl.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("%w: template not found with ID: %s", domain.ErrNotFound, id)
		}
		return nil, fmt.Errorf("select template by id: %w", err)
	}

	return &tmpl, nil
}

func (r *PostgresRepository) SaveSendLog(logEntry *domain.EmailSendLog) error {
	const query = `
INSERT INTO email_send_logs (email_id, recipient_to, sender_from, subject, provider, result_status, error_message, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
`

	_, err := r.db.Exec(
		query,
		logEntry.EmailID,
		logEntry.RecipientTo,
		logEntry.SenderFrom,
		logEntry.Subject,
		logEntry.Provider,
		logEntry.ResultStatus,
		logEntry.ErrorMessage,
		logEntry.CreatedAt.UTC(),
	)
	if err != nil {
		return fmt.Errorf("insert email send log: %w", err)
	}

	return nil
}

func (r *PostgresRepository) GetAllTemplates() ([]*domain.EmailTemplate, error) {
	const query = `
SELECT id, name, subject, body, is_html, created_at, updated_at
FROM email_templates
ORDER BY created_at ASC
`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("select all templates: %w", err)
	}
	defer rows.Close()

	templates := make([]*domain.EmailTemplate, 0)
	for rows.Next() {
		var tmpl domain.EmailTemplate
		if err := rows.Scan(
			&tmpl.ID,
			&tmpl.Name,
			&tmpl.Subject,
			&tmpl.Body,
			&tmpl.IsHTML,
			&tmpl.CreatedAt,
			&tmpl.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan template row: %w", err)
		}
		templates = append(templates, &tmpl)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate template rows: %w", err)
	}

	return templates, nil
}
