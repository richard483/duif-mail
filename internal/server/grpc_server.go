package server

import (
	"context"
	"duif/internal/domain"
	pb "duif/proto"
	"fmt"
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
		return &pb.EmailResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to send email: %v", err),
		}, nil
	}

	// Convert domain response to protobuf response
	return &pb.EmailResponse{
		Success: resp.Success,
		Message: resp.Message,
		EmailId: resp.EmailID,
	}, nil
}

// GetEmail handles the GetEmail RPC
func (s *GRPCServer) GetEmail(ctx context.Context, req *pb.GetEmailRequest) (*pb.Email, error) {
	email, err := s.mailService.GetEmail(req.Id)
	if err != nil {
		return nil, fmt.Errorf("email not found: %v", err)
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
	emails, err := s.mailService.GetAllEmails()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve emails: %v", err)
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
