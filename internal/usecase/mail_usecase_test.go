package usecase

import (
	"duif/internal/config"
	"duif/internal/domain"
	"fmt"
	"os"
	"strings"
	"testing"
)

type stubRepo struct {
	saved     map[string]*domain.Email
	templates map[string]*domain.EmailTemplate
	sendLogs  []*domain.EmailSendLog
}

func newStubRepo() *stubRepo {
	return &stubRepo{
		saved: make(map[string]*domain.Email),
		templates: map[string]*domain.EmailTemplate{
			"9f1b5f8b-9bb2-4a8e-9b8f-2dc92f9c0f11": {
				ID:      "9f1b5f8b-9bb2-4a8e-9b8f-2dc92f9c0f11",
				Subject: "Welcome!",
				IsHTML:  true,
				Body:    `<h1>Welcome {{.Name}}!</h1><p>{{.Message}}</p>`,
			},
			"0b74f6d5-0f9f-4d26-9f90-90a43d4d4f22": {
				ID:      "0b74f6d5-0f9f-4d26-9f90-90a43d4d4f22",
				Subject: "Graduation Announcement",
				IsHTML:  true,
				Body:    `<h1>Graduation Announcement</h1><p>Congratulations {{index . "name"}}!</p>`,
			},
		},
	}
}

func (r *stubRepo) Save(email *domain.Email) error {
	r.saved[email.ID] = email
	return nil
}

func (r *stubRepo) GetByID(id string) (*domain.Email, error) {
	email, ok := r.saved[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return email, nil
}

func (r *stubRepo) GetAll() ([]*domain.Email, error) {
	out := make([]*domain.Email, 0, len(r.saved))
	for _, e := range r.saved {
		out = append(out, e)
	}
	return out, nil
}

func (r *stubRepo) GetTemplateByID(id string) (*domain.EmailTemplate, error) {
	tmpl, ok := r.templates[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return tmpl, nil
}

func (r *stubRepo) GetAllTemplates() ([]*domain.EmailTemplate, error) {
	templates := make([]*domain.EmailTemplate, 0, len(r.templates))
	for _, tmpl := range r.templates {
		templates = append(templates, tmpl)
	}
	return templates, nil
}

func (r *stubRepo) SaveSendLog(logEntry *domain.EmailSendLog) error {
	r.sendLogs = append(r.sendLogs, logEntry)
	return nil
}

func TestSendEmail_ValidationError(t *testing.T) {
	repo := newStubRepo()
	uc := NewMailUseCaseWithSender(repo, func(to, from, subject, body string, isHTML bool, replyTo string, cc []string) error {
		t.Fatal("sender should not be called on invalid input")
		return nil
	})

	_, err := uc.SendEmail(&domain.SendEmailRequest{
		To:      "",
		From:    "sender@example.com",
		Subject: "Hello",
		Body:    "Body",
	})
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
	if !strings.Contains(err.Error(), domain.ErrInvalidInput.Error()) {
		t.Fatalf("expected invalid input error, got: %v", err)
	}
}

func TestSendEmail_InvalidCC(t *testing.T) {
	repo := newStubRepo()
	uc := NewMailUseCaseWithSender(repo, func(to, from, subject, body string, isHTML bool, replyTo string, cc []string) error {
		t.Fatal("sender should not be called on invalid cc")
		return nil
	})

	_, err := uc.SendEmail(&domain.SendEmailRequest{
		To:      "to@example.com",
		From:    "from@example.com",
		Subject: "Hello",
		Body:    "Body",
		Cc:      "invalid-email",
	})
	if err == nil {
		t.Fatal("expected cc validation error, got nil")
	}
}

func TestSendEmail_TemplateBodyIsPersistedAsRenderedBody(t *testing.T) {
	repo := newStubRepo()
	var sentBody string
	var sentCC []string
	var sentHTML bool

	uc := NewMailUseCaseWithSender(repo, func(to, from, subject, body string, isHTML bool, replyTo string, cc []string) error {
		sentBody = body
		sentCC = cc
		sentHTML = isHTML
		return nil
	})

	resp, err := uc.SendEmail(&domain.SendEmailRequest{
		To:         "to@example.com",
		From:       "from@example.com",
		Subject:    "Welcome",
		TemplateId: "9f1b5f8b-9bb2-4a8e-9b8f-2dc92f9c0f11",
		TemplateData: map[string]interface{}{
			"Name":    "Nico",
			"Message": "Welcome aboard",
		},
		Cc: "cc1@example.com, cc2@example.com",
	})
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if resp.EmailID == "" {
		t.Fatal("expected generated email id")
	}
	if !strings.Contains(sentBody, "Welcome Nico!") {
		t.Fatalf("expected rendered body to be sent, got: %s", sentBody)
	}
	if !sentHTML {
		t.Fatal("expected template mail to be sent as HTML")
	}
	if len(sentCC) != 2 {
		t.Fatalf("expected 2 CC recipients, got %d", len(sentCC))
	}

	stored, ok := repo.saved[resp.EmailID]
	if !ok {
		t.Fatal("expected email to be stored")
	}
	if stored.Body != sentBody {
		t.Fatal("expected stored body to match rendered sent body")
	}
	if len(repo.sendLogs) == 0 {
		t.Fatal("expected send log entry on success")
	}
	if repo.sendLogs[0].ResultStatus != "sent" {
		t.Fatalf("expected sent log status, got: %s", repo.sendLogs[0].ResultStatus)
	}
}

func TestSendEmail_RequiresBodyOrTemplate(t *testing.T) {
	repo := newStubRepo()
	uc := NewMailUseCaseWithSender(repo, func(to, from, subject, body string, isHTML bool, replyTo string, cc []string) error {
		t.Fatal("sender should not be called")
		return nil
	})

	_, err := uc.SendEmail(&domain.SendEmailRequest{
		To:      "to@example.com",
		From:    "from@example.com",
		Subject: "Subject",
	})
	if err == nil {
		t.Fatal("expected error when body and template are empty")
	}
}

func TestGetEmailAndGetAll(t *testing.T) {
	repo := newStubRepo()
	uc := NewMailUseCaseWithSender(repo, nil)

	emailSent := &domain.Email{ID: "id-1", To: "to@example.com", From: "from@example.com", Subject: "s", Body: "b", Status: "sent"}
	emailFailed := &domain.Email{ID: "id-2", To: "to2@example.com", From: "from@example.com", Subject: "s2", Body: "b2", Status: "failed"}
	if err := repo.Save(emailSent); err != nil {
		t.Fatalf("failed saving test email: %v", err)
	}
	if err := repo.Save(emailFailed); err != nil {
		t.Fatalf("failed saving test email: %v", err)
	}

	got, err := uc.GetEmail("id-1")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if got.ID != "id-1" {
		t.Fatalf("expected id-1, got %s", got.ID)
	}

	all, err := uc.GetAllEmails()
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(all) != 2 {
		t.Fatalf("expected 2 emails, got %d", len(all))
	}

	var foundFailed bool
	for _, e := range all {
		if e.Status == "failed" {
			foundFailed = true
			break
		}
	}
	if !foundFailed {
		t.Fatal("expected failed email to be included in GetAllEmails")
	}
}

func TestGetTemplates(t *testing.T) {
	repo := newStubRepo()
	uc := NewMailUseCaseWithSender(repo, nil)

	templates, err := uc.GetTemplates()
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(templates) == 0 {
		t.Fatal("expected at least one template")
	}
}

func TestParseRecipients(t *testing.T) {
	recipients, err := parseRecipients("a@example.com, b@example.com")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(recipients) != 2 {
		t.Fatalf("expected 2 recipients, got %d", len(recipients))
	}

	_, err = parseRecipients("not-an-email")
	if err == nil {
		t.Fatal("expected parse error")
	}
}

func TestValidateSendEmailRequest_ReplyToInvalid(t *testing.T) {
	_, err := validateSendEmailRequest(&domain.SendEmailRequest{
		To:      "to@example.com",
		From:    "from@example.com",
		Subject: "subject",
		Body:    "body",
		ReplyTo: "invalid",
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestRenderTemplate(t *testing.T) {
	rendered, err := renderTemplate(`<h1>Graduation Announcement</h1><p>Congratulations {{index . "name"}}!</p>`, map[string]interface{}{
		"name":    "Nico",
		"success": "Passed",
		"detail":  "You have completed all requirements.",
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !strings.Contains(rendered, "Congratulations Nico!") {
		t.Fatalf("expected rendered content to include name, got: %s", rendered)
	}

	_, err = renderTemplate(`{{if`, map[string]interface{}{})
	if err == nil {
		t.Fatal("expected parse error")
	}
}

func TestSendMail_ReturnsErrorForUnreachableSMTP(t *testing.T) {
	t.Setenv("MAIL_HOST", "127.0.0.1")
	t.Setenv("MAIL_PORT", "1")
	t.Setenv("MAIL_USERNAME", "user@example.com")
	t.Setenv("MAIL_PASSWORD", "secret")

	if err := config.Init(); err != nil {
		t.Fatalf("expected config init success, got %v", err)
	}

	err := sendMail(
		"to@example.com",
		"from@example.com",
		"subject",
		"body",
		false,
		"reply@example.com",
		[]string{"cc1@example.com", "cc2@example.com"},
	)
	if err == nil {
		t.Fatal("expected SMTP error due to unreachable server")
	}
}

func TestValidateSendEmailRequest_NilRequest(t *testing.T) {
	_, err := validateSendEmailRequest(nil)
	if err == nil {
		t.Fatal("expected nil request error")
	}
}

func TestSendEmail_SenderFailure(t *testing.T) {
	repo := newStubRepo()
	uc := NewMailUseCaseWithSender(repo, func(to, from, subject, body string, isHTML bool, replyTo string, cc []string) error {
		return fmt.Errorf("smtp failed")
	})

	resp, err := uc.SendEmail(&domain.SendEmailRequest{
		To:      "to@example.com",
		From:    "from@example.com",
		Subject: "subject",
		Body:    "body",
	})
	if err == nil {
		t.Fatal("expected sender error")
	}
	if resp.Success {
		t.Fatal("expected success=false when sender fails")
	}
	if len(repo.sendLogs) == 0 {
		t.Fatal("expected send log entry on failure")
	}
	if repo.sendLogs[0].ResultStatus != "failed" {
		t.Fatalf("expected failed log status, got: %s", repo.sendLogs[0].ResultStatus)
	}
}

func TestMain(m *testing.M) {
	// Ensure config.Init can run with valid defaults for tests that call sendMail.
	_ = os.Setenv("MAIL_HOST", "127.0.0.1")
	_ = os.Setenv("MAIL_PORT", "1")
	_ = os.Setenv("MAIL_USERNAME", "user@example.com")
	_ = os.Setenv("MAIL_PASSWORD", "secret")

	os.Exit(m.Run())
}
