package main

import (
	"duif/internal/repository"
	"duif/internal/server"
	"duif/internal/usecase"
	pb "duif/proto"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "50051"
	}

	// Create TCP listener
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", port, err)
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
	appName := os.Getenv("APP_NAME")
	if appName == "" {
		appName = "GoMail API"
	}
	log.Printf("🚀 %s starting on port %s (env: %s)", appName, port, os.Getenv("APP_ENV"))
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
