package main

import (
	"duif/internal/config"
	"duif/internal/repository"
	"duif/internal/server"
	"duif/internal/usecase"
	pb "duif/proto"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	// Initialize global configuration (call once at startup)
	if err := config.Init(); err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Access config globally
	cfg := config.Get()

	// Create TCP listener
	lis, err := net.Listen("tcp", ":"+cfg.Server.Port)
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", cfg.Server.Port, err)
	}

	// Initialize layers
	emailRepo := repository.NewInMemoryEmailRepository()
	mailUseCase := usecase.NewMailUseCase(emailRepo)
	mailServer := server.NewGRPCServer(mailUseCase)

	// Create gRPC server
	grpcServer := grpc.NewServer()
	pb.RegisterMailServiceServer(grpcServer, mailServer)

	// Register reflection service for tools like grpcurl
	reflection.Register(grpcServer)

	// Handle graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan
		log.Println("Shutting down gRPC server...")
		grpcServer.GracefulStop()
	}()

	// Start server
	log.Printf("🚀 %s starting on port %s (env: %s)", cfg.App.Name, cfg.Server.Port, cfg.App.Env)
	log.Printf("📧 Mail configured: %s:%d (user: %s)", cfg.Mail.Host, cfg.Mail.Port, cfg.Mail.Username)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
