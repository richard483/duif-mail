package server

import (
	"context"
	"duif/internal/domain"
	pb "duif/proto"
	"errors"
	"fmt"
	"net"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GRPCServer implements the gRPC MailService
type GRPCServer struct {
	pb.UnimplementedMailServiceServer
	mailService domain.MailService
}

// NewGRPCServer creates a new gRPC server
func NewGRPCServer(mailService domain.MailService) *GRPCServer {
	return &GRPCServer{
		mailService: mailService,
	}
}

// SendEmail handles the SendEmail RPC
func (s *GRPCServer) SendEmail(ctx context.Context, req *pb.EmailRequest) (*pb.EmailResponse, error) {
	templateData := make(map[string]interface{}, len(req.TemplateData))
	for key, value := range req.TemplateData {
		if value == nil {
			templateData[key] = nil
			continue
		}
		templateData[key] = value.AsInterface()
	}

	// Convert protobuf request to domain request
	domainReq := &domain.SendEmailRequest{
		To:           req.To,
		From:         req.From,
		Subject:      req.Subject,
		Body:         req.Body,
		IsHTML:       req.IsHtml,
		ReplyTo:      req.ReplyTo,
		Cc:           req.Cc,
		TemplateId:   req.TemplateId,
		TemplateData: templateData,
	}

	// Call use case
	resp, err := s.mailService.SendEmail(domainReq)
	if err != nil {
		return nil, mapSendEmailError(err)
	}

	// Convert domain response to protobuf response
	return &pb.EmailResponse{
		Success: resp.Success,
		Message: resp.Message,
		EmailId: resp.EmailID,
	}, nil
}

func mapSendEmailError(err error) error {
	if errors.Is(err, domain.ErrInvalidInput) {
		return status.Error(codes.InvalidArgument, err.Error())
	}
	if errors.Is(err, domain.ErrNotFound) {
		return status.Error(codes.NotFound, err.Error())
	}

	var netErr net.Error
	if errors.As(err, &netErr) {
		return status.Error(codes.Unavailable, fmt.Sprintf("smtp/network error: %v", err))
	}

	return status.Error(codes.Internal, fmt.Sprintf("failed to send email: %v", err))
}

// GetEmail handles the GetEmail RPC
func (s *GRPCServer) GetEmail(ctx context.Context, req *pb.GetEmailRequest) (*pb.Email, error) {
	email, err := s.mailService.GetEmail(req.Id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "email not found")
		}
		if errors.Is(err, domain.ErrInvalidInput) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Error(codes.Internal, "failed to retrieve email")
	}

	// Convert domain email to protobuf email
	pbEmail := &pb.Email{
		Id:        email.ID,
		To:        email.To,
		From:      email.From,
		Subject:   email.Subject,
		Body:      email.Body,
		Status:    email.Status,
		CreatedAt: email.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	if email.SentAt != nil {
		pbEmail.SentAt = email.SentAt.Format("2006-01-02T15:04:05Z07:00")
	}

	return pbEmail, nil
}

// GetAllEmails handles the GetAllEmails RPC
func (s *GRPCServer) GetAllEmails(ctx context.Context, req *pb.GetAllEmailsRequest) (*pb.GetAllEmailsResponse, error) {
	// Return all statuses (sent/failed/pending) for internal operational visibility.
	emails, err := s.mailService.GetAllEmails()
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to retrieve emails")
	}

	// Convert domain emails to protobuf emails
	pbEmails := make([]*pb.Email, 0, len(emails))
	for _, email := range emails {
		pbEmail := &pb.Email{
			Id:        email.ID,
			To:        email.To,
			From:      email.From,
			Subject:   email.Subject,
			Body:      email.Body,
			Status:    email.Status,
			CreatedAt: email.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}

		if email.SentAt != nil {
			pbEmail.SentAt = email.SentAt.Format("2006-01-02T15:04:05Z07:00")
		}

		pbEmails = append(pbEmails, pbEmail)
	}

	return &pb.GetAllEmailsResponse{
		Emails: pbEmails,
		Count:  int32(len(pbEmails)),
	}, nil
}

// ListTemplates handles the ListTemplates RPC.
func (s *GRPCServer) ListTemplates(ctx context.Context, req *pb.ListTemplatesRequest) (*pb.ListTemplatesResponse, error) {
	templates, err := s.mailService.GetTemplates()
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to retrieve templates")
	}

	pbTemplates := make([]*pb.Template, 0, len(templates))
	for _, tmpl := range templates {
		pbTemplates = append(pbTemplates, &pb.Template{
			Id:        tmpl.ID,
			Name:      tmpl.Name,
			Subject:   tmpl.Subject,
			IsHtml:    tmpl.IsHTML,
			CreatedAt: tmpl.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt: tmpl.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	return &pb.ListTemplatesResponse{
		Templates: pbTemplates,
		Count:     int32(len(pbTemplates)),
	}, nil
}
