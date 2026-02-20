package server

import (
	"context"
	"duif/internal/domain"
	pb "duif/proto"
	"fmt"
	"net"
	"testing"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
)

type stubMailService struct {
	sendFn      func(req *domain.SendEmailRequest) (*domain.SendEmailResponse, error)
	getFn       func(id string) (*domain.Email, error)
	getAllFn    func() ([]*domain.Email, error)
	templatesFn func() ([]*domain.EmailTemplate, error)
}

func (s *stubMailService) SendEmail(req *domain.SendEmailRequest) (*domain.SendEmailResponse, error) {
	return s.sendFn(req)
}

func (s *stubMailService) GetEmail(id string) (*domain.Email, error) {
	return s.getFn(id)
}

func (s *stubMailService) GetAllEmails() ([]*domain.Email, error) {
	return s.getAllFn()
}

func (s *stubMailService) GetTemplates() ([]*domain.EmailTemplate, error) {
	return s.templatesFn()
}

func TestSendEmail_ReturnsInvalidArgumentStatus(t *testing.T) {
	svc := &stubMailService{
		sendFn: func(req *domain.SendEmailRequest) (*domain.SendEmailResponse, error) {
			return nil, fmt.Errorf("%w: invalid to address", domain.ErrInvalidInput)
		},
		getFn:       func(id string) (*domain.Email, error) { return nil, nil },
		getAllFn:    func() ([]*domain.Email, error) { return nil, nil },
		templatesFn: func() ([]*domain.EmailTemplate, error) { return nil, nil },
	}
	server := NewGRPCServer(svc)

	_, err := server.SendEmail(context.Background(), &pb.EmailRequest{})
	if err == nil {
		t.Fatal("expected gRPC error")
	}
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("expected InvalidArgument, got %v", status.Code(err))
	}
}

func TestGetEmail_ReturnsNotFoundStatus(t *testing.T) {
	svc := &stubMailService{
		sendFn: func(req *domain.SendEmailRequest) (*domain.SendEmailResponse, error) {
			return nil, nil
		},
		getFn: func(id string) (*domain.Email, error) {
			return nil, domain.ErrNotFound
		},
		getAllFn:    func() ([]*domain.Email, error) { return nil, nil },
		templatesFn: func() ([]*domain.EmailTemplate, error) { return nil, nil },
	}
	server := NewGRPCServer(svc)

	_, err := server.GetEmail(context.Background(), &pb.GetEmailRequest{Id: "missing"})
	if err == nil {
		t.Fatal("expected gRPC error")
	}
	if status.Code(err) != codes.NotFound {
		t.Fatalf("expected NotFound, got %v", status.Code(err))
	}
}

func TestSendEmail_Success(t *testing.T) {
	svc := &stubMailService{
		sendFn: func(req *domain.SendEmailRequest) (*domain.SendEmailResponse, error) {
			if req.TemplateData["name"] != "nico" {
				t.Fatalf("expected template data conversion, got: %+v", req.TemplateData)
			}
			return &domain.SendEmailResponse{
				Success: true,
				Message: "Email sent successfully",
				EmailID: "email-123",
			}, nil
		},
		getFn:       func(id string) (*domain.Email, error) { return nil, nil },
		getAllFn:    func() ([]*domain.Email, error) { return nil, nil },
		templatesFn: func() ([]*domain.EmailTemplate, error) { return nil, nil },
	}
	server := NewGRPCServer(svc)

	resp, err := server.SendEmail(context.Background(), &pb.EmailRequest{
		To:           "to@example.com",
		From:         "from@example.com",
		Subject:      "Subject",
		Body:         "Body",
		TemplateData: map[string]*structpb.Value{"name": structpb.NewStringValue("nico")},
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !resp.Success || resp.EmailId != "email-123" {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestSendEmail_InternalError(t *testing.T) {
	svc := &stubMailService{
		sendFn: func(req *domain.SendEmailRequest) (*domain.SendEmailResponse, error) {
			return nil, fmt.Errorf("smtp down")
		},
		getFn:       func(id string) (*domain.Email, error) { return nil, nil },
		getAllFn:    func() ([]*domain.Email, error) { return nil, nil },
		templatesFn: func() ([]*domain.EmailTemplate, error) { return nil, nil },
	}
	server := NewGRPCServer(svc)

	_, err := server.SendEmail(context.Background(), &pb.EmailRequest{})
	if err == nil {
		t.Fatal("expected gRPC error")
	}
	if status.Code(err) != codes.Internal {
		t.Fatalf("expected Internal, got %v", status.Code(err))
	}
}

func TestSendEmail_ReturnsNotFoundStatus(t *testing.T) {
	svc := &stubMailService{
		sendFn: func(req *domain.SendEmailRequest) (*domain.SendEmailResponse, error) {
			return nil, fmt.Errorf("%w: template not found", domain.ErrNotFound)
		},
		getFn:       func(id string) (*domain.Email, error) { return nil, nil },
		getAllFn:    func() ([]*domain.Email, error) { return nil, nil },
		templatesFn: func() ([]*domain.EmailTemplate, error) { return nil, nil },
	}
	server := NewGRPCServer(svc)

	_, err := server.SendEmail(context.Background(), &pb.EmailRequest{})
	if err == nil {
		t.Fatal("expected gRPC error")
	}
	if status.Code(err) != codes.NotFound {
		t.Fatalf("expected NotFound, got %v", status.Code(err))
	}
}

func TestSendEmail_ReturnsUnavailableStatusForNetworkError(t *testing.T) {
	svc := &stubMailService{
		sendFn: func(req *domain.SendEmailRequest) (*domain.SendEmailResponse, error) {
			return nil, &net.OpError{Op: "dial", Net: "tcp", Err: fmt.Errorf("connection refused")}
		},
		getFn:       func(id string) (*domain.Email, error) { return nil, nil },
		getAllFn:    func() ([]*domain.Email, error) { return nil, nil },
		templatesFn: func() ([]*domain.EmailTemplate, error) { return nil, nil },
	}
	server := NewGRPCServer(svc)

	_, err := server.SendEmail(context.Background(), &pb.EmailRequest{})
	if err == nil {
		t.Fatal("expected gRPC error")
	}
	if status.Code(err) != codes.Unavailable {
		t.Fatalf("expected Unavailable, got %v", status.Code(err))
	}
}

func TestGetEmail_Success(t *testing.T) {
	now := time.Now().UTC()
	sent := now.Add(1 * time.Minute)
	svc := &stubMailService{
		sendFn: func(req *domain.SendEmailRequest) (*domain.SendEmailResponse, error) {
			return nil, nil
		},
		getFn: func(id string) (*domain.Email, error) {
			return &domain.Email{
				ID:        "1",
				To:        "to@example.com",
				From:      "from@example.com",
				Subject:   "subject",
				Body:      "body",
				Status:    "sent",
				CreatedAt: now,
				SentAt:    &sent,
			}, nil
		},
		getAllFn:    func() ([]*domain.Email, error) { return nil, nil },
		templatesFn: func() ([]*domain.EmailTemplate, error) { return nil, nil },
	}
	server := NewGRPCServer(svc)

	resp, err := server.GetEmail(context.Background(), &pb.GetEmailRequest{Id: "1"})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if resp.Id != "1" || resp.SentAt == "" {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestGetEmail_InvalidArgument(t *testing.T) {
	svc := &stubMailService{
		sendFn: func(req *domain.SendEmailRequest) (*domain.SendEmailResponse, error) {
			return nil, nil
		},
		getFn: func(id string) (*domain.Email, error) {
			return nil, fmt.Errorf("%w: bad id", domain.ErrInvalidInput)
		},
		getAllFn:    func() ([]*domain.Email, error) { return nil, nil },
		templatesFn: func() ([]*domain.EmailTemplate, error) { return nil, nil },
	}
	server := NewGRPCServer(svc)

	_, err := server.GetEmail(context.Background(), &pb.GetEmailRequest{Id: ""})
	if err == nil {
		t.Fatal("expected gRPC error")
	}
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("expected InvalidArgument, got %v", status.Code(err))
	}
}

func TestGetAllEmails_Success(t *testing.T) {
	now := time.Now()
	svc := &stubMailService{
		sendFn: func(req *domain.SendEmailRequest) (*domain.SendEmailResponse, error) {
			return nil, nil
		},
		getFn: func(id string) (*domain.Email, error) {
			return nil, nil
		},
		getAllFn: func() ([]*domain.Email, error) {
			return []*domain.Email{
				{
					ID:        "1",
					To:        "to@example.com",
					From:      "from@example.com",
					Subject:   "Subject",
					Body:      "Body",
					Status:    "sent",
					CreatedAt: now,
				},
			}, nil
		},
		templatesFn: func() ([]*domain.EmailTemplate, error) { return nil, nil },
	}
	server := NewGRPCServer(svc)

	resp, err := server.GetAllEmails(context.Background(), &pb.GetAllEmailsRequest{})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if resp.Count != 1 || len(resp.Emails) != 1 {
		t.Fatalf("expected one email, got count=%d len=%d", resp.Count, len(resp.Emails))
	}
}

func TestGetAllEmails_InternalError(t *testing.T) {
	svc := &stubMailService{
		sendFn: func(req *domain.SendEmailRequest) (*domain.SendEmailResponse, error) {
			return nil, nil
		},
		getFn: func(id string) (*domain.Email, error) {
			return nil, nil
		},
		getAllFn: func() ([]*domain.Email, error) {
			return nil, fmt.Errorf("db down")
		},
		templatesFn: func() ([]*domain.EmailTemplate, error) { return nil, nil },
	}
	server := NewGRPCServer(svc)

	_, err := server.GetAllEmails(context.Background(), &pb.GetAllEmailsRequest{})
	if err == nil {
		t.Fatal("expected gRPC error")
	}
	if status.Code(err) != codes.Internal {
		t.Fatalf("expected Internal, got %v", status.Code(err))
	}
}

func TestListTemplates_Success(t *testing.T) {
	now := time.Now().UTC()
	svc := &stubMailService{
		sendFn: func(req *domain.SendEmailRequest) (*domain.SendEmailResponse, error) { return nil, nil },
		getFn:  func(id string) (*domain.Email, error) { return nil, nil },
		getAllFn: func() ([]*domain.Email, error) {
			return nil, nil
		},
		templatesFn: func() ([]*domain.EmailTemplate, error) {
			return []*domain.EmailTemplate{
				{
					ID:        "9f1b5f8b-9bb2-4a8e-9b8f-2dc92f9c0f11",
					Name:      "Welcome",
					Subject:   "Welcome!",
					IsHTML:    true,
					CreatedAt: now,
					UpdatedAt: now,
				},
			}, nil
		},
	}
	server := NewGRPCServer(svc)

	resp, err := server.ListTemplates(context.Background(), &pb.ListTemplatesRequest{})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if resp.Count != 1 || len(resp.Templates) != 1 {
		t.Fatalf("expected one template, got count=%d len=%d", resp.Count, len(resp.Templates))
	}
	if resp.Templates[0].Id == "" {
		t.Fatal("expected template id")
	}
}
