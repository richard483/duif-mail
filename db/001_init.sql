CREATE TABLE IF NOT EXISTS email_templates (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    subject TEXT NOT NULL,
    body TEXT NOT NULL,
    is_html BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS emails (
    id TEXT PRIMARY KEY,
    recipient_to TEXT NOT NULL,
    sender_from TEXT NOT NULL,
    subject TEXT NOT NULL,
    body TEXT NOT NULL,
    status TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    sent_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_emails_created_at ON emails (created_at DESC);
CREATE INDEX IF NOT EXISTS idx_emails_recipient_to ON emails (recipient_to);

CREATE TABLE IF NOT EXISTS email_send_logs (
    id BIGSERIAL PRIMARY KEY,
    email_id TEXT REFERENCES emails(id) ON DELETE SET NULL,
    recipient_to TEXT NOT NULL,
    sender_from TEXT NOT NULL,
    subject TEXT NOT NULL,
    provider TEXT NOT NULL DEFAULT 'smtp',
    result_status TEXT NOT NULL,
    error_message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_email_send_logs_email_id ON email_send_logs (email_id);
