INSERT INTO email_templates (id, name, subject, body, is_html)
VALUES
(
    '9f1b5f8b-9bb2-4a8e-9b8f-2dc92f9c0f11',
    'Welcome',
    'Welcome to GoMail',
    '<!DOCTYPE html><html><body><h1>Welcome {{.Name}}!</h1><p>{{.Message}}</p></body></html>',
    TRUE
),
(
    '0b74f6d5-0f9f-4d26-9f90-90a43d4d4f22',
    'Graduation Announcement',
    'Graduation Announcement',
    '<!DOCTYPE html><html><body><h1>Graduation Announcement</h1><p>Congratulations {{index . "name"}}!</p><p>{{index . "detail"}}</p></body></html>',
    TRUE
)
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    subject = EXCLUDED.subject,
    body = EXCLUDED.body,
    is_html = EXCLUDED.is_html,
    updated_at = NOW();

INSERT INTO emails (id, recipient_to, sender_from, subject, body, status, created_at, sent_at)
VALUES
(
    '11111111-1111-1111-1111-111111111111',
    'alice@example.com',
    'bob@example.com',
    'Welcome to GoMail',
    'This is a seeded sample email.',
    'sent',
    NOW() - INTERVAL '2 hours',
    NOW() - INTERVAL '2 hours'
),
(
    '22222222-2222-2222-2222-222222222222',
    'charlie@example.com',
    'admin@example.com',
    'System Notification',
    'Your account has been activated successfully.',
    'sent',
    NOW() - INTERVAL '1 hour',
    NOW() - INTERVAL '1 hour'
)
ON CONFLICT (id) DO NOTHING;

INSERT INTO email_send_logs (email_id, recipient_to, sender_from, subject, provider, result_status, error_message, created_at)
VALUES
(
    '11111111-1111-1111-1111-111111111111',
    'alice@example.com',
    'bob@example.com',
    'Welcome to GoMail',
    'smtp',
    'sent',
    NULL,
    NOW() - INTERVAL '2 hours'
),
(
    NULL,
    'failed@example.com',
    'system@example.com',
    'Delivery Retry',
    'smtp',
    'failed',
    'dial tcp: connection refused',
    NOW() - INTERVAL '10 minutes'
);
