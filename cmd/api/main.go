package main

import (
	"duif/internal/config"
	"duif/internal/domain"
	"duif/internal/repository"
	"duif/internal/server"
	"duif/internal/usecase"
	pb "duif/proto"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
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
	var emailRepo domain.EmailRepository
	var shutdown func()

	switch strings.ToLower(cfg.Storage.Driver) {
	case "postgres":
		pgRepo, err := repository.NewPostgresRepository(cfg.Storage.DSN)
		if err != nil {
			log.Fatalf("Failed to initialize postgres repository: %v", err)
		}
		emailRepo = pgRepo
		shutdown = func() {
			if err := pgRepo.Close(); err != nil {
				log.Printf("Failed closing postgres connection: %v", err)
			}
		}
	default:
		emailRepo = repository.NewInMemoryEmailRepository()
		shutdown = func() {}
	}
	defer shutdown()

	mailUseCase := usecase.NewMailUseCase(emailRepo)
	mailServer := server.NewGRPCServer(mailUseCase)

	// Create gRPC server
	grpcServer := grpc.NewServer()
	pb.RegisterMailServiceServer(grpcServer, mailServer)

	// Register health service for runtime readiness checks.
	healthServer := health.NewServer()
	healthServer.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	healthpb.RegisterHealthServer(grpcServer, healthServer)

	// Register reflection only in development mode.
	if cfg.IsDevelopment() {
		reflection.Register(grpcServer)
	}

	// Handle graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan
		log.Println("Shutting down gRPC server...")
		grpcServer.GracefulStop()
	}()

	// Start server
	log.Printf("%s starting on port %s (env: %s)", cfg.App.Name, cfg.Server.Port, cfg.App.Env)
	log.Printf("Mail configured: %s:%d", cfg.Mail.Host, cfg.Mail.Port)
	log.Printf("Storage driver: %s", cfg.Storage.Driver)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
