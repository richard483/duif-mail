# GoMail API

A clean architecture mail service built with Go and gRPC.

## Project Structure

```
gomail/
├── cmd/
│   └── api/
│       └── main.go          # Application entry point & dependency injection
├── internal/
│   ├── domain/              # Domain layer (entities & interfaces)
│   │   ├── email.go         # Core entity & repository interface
│   │   ├── request.go       # DTOs
│   │   └── service.go       # Business service interface
│   ├── usecase/             # Business logic layer
│   │   └── mail_usecase.go  # Implements domain.MailService
│   ├── repository/          # Data access layer
│   │   └── email_repository.go  # Implements domain.EmailRepository
│   └── server/              # gRPC server implementation
│       └── grpc_server.go   # Protocol adapter
├── proto/
│   ├── mail.proto           # Protobuf definitions
│   ├── mail.pb.go           # Generated protobuf code
│   └── mail_grpc.pb.go      # Generated gRPC code
├── ARCHITECTURE.md          # Clean architecture documentation
├── README.md
└── go.mod
```

**Architecture:** This project follows [Clean Architecture](ARCHITECTURE.md) principles with clear separation between domain, use cases, and infrastructure.

## Prerequisites

Install the following tools:
- Go 1.25+
- Protocol Buffer Compiler (protoc)
- Go protobuf plugins:

```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

Make sure `$GOPATH/bin` is in your PATH.

## Regenerating Protobuf Files

Whenever you modify any `.proto` files in the `proto/` directory, regenerate the Go code:

```bash
protoc --proto_path=. \
  --go_out=. --go_opt=paths=source_relative \
  --go-grpc_out=. --go-grpc_opt=paths=source_relative \
  proto/*.proto
```

This will generate/update:
- `proto/mail.pb.go` - Protocol buffer message definitions
- `proto/mail_grpc.pb.go` - gRPC service definitions

## Running the Application

```bash
# Install dependencies
go mod tidy

# Run the server
go run cmd/api/main.go

# Or build and run
go build -o gomail.exe cmd/api/main.go
./gomail.exe
```

The gRPC server will start on `localhost:50051`

## Testing the gRPC API

### Using grpcurl

Install grpcurl: https://github.com/fullstorydev/grpcurl

```bash
# List available services
grpcurl -plaintext localhost:50051 list

# List methods in MailService
grpcurl -plaintext localhost:50051 list mail.MailService

# Send an email
grpcurl -plaintext -d '{
  "to": "recipient@example.com",
  "from": "sender@example.com",
  "subject": "Test Email",
  "body": "This is a test email via gRPC"
}' localhost:50051 mail.MailService/SendEmail

# Get all emails (includes 3 demo emails)
grpcurl -plaintext -d '{}' localhost:50051 mail.MailService/GetAllEmails

# Get email by ID
grpcurl -plaintext -d '{"id": "YOUR-EMAIL-ID"}' localhost:50051 mail.MailService/GetEmail
```

### Using Go Client

```go
package main

import (
    "context"
    "log"
    pb "duif/proto"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
)

func main() {
    conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()

    client := pb.NewMailServiceClient(conn)
    
    resp, err := client.SendEmail(context.Background(), &pb.EmailRequest{
        To:      "test@example.com",
        From:    "admin@example.com",
        Subject: "Hello",
        Body:    "Testing gRPC API",
    })
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("Success: %v, Message: %s, EmailID: %s", 
        resp.Success, resp.Message, resp.EmailId)
}
```

## Available gRPC Methods

### SendEmail
Sends a new email

**Request:**
```protobuf
{
  "to": "recipient@example.com",
  "from": "sender@example.com",
  "subject": "Subject",
  "body": "Email body"
}
```

**Response:**
```protobuf
{
  "success": true,
  "message": "Email sent successfully",
  "email_id": "uuid"
}
```

### GetAllEmails
Retrieves all emails

**Request:**
```protobuf
{}
```

**Response:**
```protobuf
{
  "emails": [...],
  "count": 3
}
```

### GetEmail
Retrieves a specific email by ID

**Request:**
```protobuf
{
  "id": "email-uuid"
}
```

**Response:**
```protobuf
{
  "id": "uuid",
  "to": "recipient@example.com",
  "from": "sender@example.com",
  "subject": "Subject",
  "body": "Body",
  "status": "sent",
  "created_at": "2026-01-10T10:00:00Z",
  "sent_at": "2026-01-10T10:00:00Z"
}
```

## Demo Data

The application comes with 3 pre-seeded demo emails for testing purposes.

## Technologies

- **Go 1.25+**
- **gRPC** - RPC framework
- **Protocol Buffers** - Interface definition language
- **Clean Architecture** - Project structure
- **In-Memory Repository** - Static demo data
