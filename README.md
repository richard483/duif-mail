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

## Configuration

The application uses environment variables for configuration. You can set them via:
1. `.env` file (recommended for development)
2. System environment variables (recommended for production)

### Setup .env file

```bash
# Copy the example file
cp .env.example .env

# Edit .env with your values
```

### Available Configuration

| Variable        | Description                          | Default          |
| --------------- | ------------------------------------ | ---------------- |
| `PORT`          | gRPC server port                     | `50051`          |
| `APP_NAME`      | Application name                     | `GoMail API`     |
| `APP_ENV`       | Environment (development/production) | `development`    |
| `STORAGE_DRIVER`| Storage backend (`memory`/`postgres`) | `memory`        |
| `DATABASE_URL`  | PostgreSQL DSN (required for postgres) | -              |
| `MAIL_HOST`     | SMTP server host                     | `smtp.gmail.com` |
| `MAIL_PORT`     | SMTP server port                     | `587`            |
| `MAIL_USERNAME` | SMTP username (required)             | -                |
| `MAIL_PASSWORD` | SMTP password (required)             | -                |
| `LOG_LEVEL`     | Logging level                        | `info`           |

### Accessing Configuration in Code

Configuration is loaded in `main.go` and injected into layers that need it:

```go
import "duif/internal/config"

// Load configuration
cfg, err := config.Load()
if err != nil {
    log.Fatal(err)
}

// Access configuration values
fmt.Println(cfg.Server.Port)      // "50051"
fmt.Println(cfg.Mail.Host)        // "smtp.gmail.com"
fmt.Println(cfg.App.Name)         // "GoMail API"
fmt.Println(cfg.IsDevelopment())  // true/false
```

See [internal/config/config.go](internal/config/config.go) for the complete configuration structure.

## Regenerating Protobuf Files

Whenever you modify any `.proto` files in the `proto/` directory, regenerate the Go code:

```bash
protoc --proto_path=. --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/*.proto

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

### Storage Driver

By default, the app uses in-memory storage (`STORAGE_DRIVER=memory`).

To use PostgreSQL:
1. Set `STORAGE_DRIVER=postgres`
2. Set `DATABASE_URL` (example: `postgres://postgres:postgres@localhost:5432/gomail?sslmode=disable`)
3. Run SQL files in order:
   - `db/001_init.sql`
   - `db/002_seed.sql`

### PostgreSQL Setup (psql Example)

```bash
# create database
psql -U postgres -c "CREATE DATABASE gomail;"

# apply schema and seed
psql "postgres://postgres:postgres@localhost:5432/gomail?sslmode=disable" -f db/001_init.sql
psql "postgres://postgres:postgres@localhost:5432/gomail?sslmode=disable" -f db/002_seed.sql
```

### Template IDs (UUID)

Current seeded templates:
- `9f1b5f8b-9bb2-4a8e-9b8f-2dc92f9c0f11` (`Welcome`)
- `0b74f6d5-0f9f-4d26-9f90-90a43d4d4f22` (`Graduation Announcement`)

Use template by sending `template_id` and `template_data` in `SendEmail`.

```bash
grpcurl -plaintext -d '{
  "to": "recipient@example.com",
  "from": "sender@example.com",
  "subject": "",
  "template_id": "9f1b5f8b-9bb2-4a8e-9b8f-2dc92f9c0f11",
  "template_data": {
    "Name": "Nico",
    "Message": "Welcome aboard"
  }
}' localhost:50051 mail.MailService/SendEmail
```

### Template Lifecycle

- Template source of truth is database table `email_templates`.
- Update template safely with SQL:

```sql
UPDATE email_templates
SET subject = 'New Subject', body = '<h1>Hello {{.Name}}</h1>', updated_at = NOW()
WHERE id = '9f1b5f8b-9bb2-4a8e-9b8f-2dc92f9c0f11';
```
- Application uses template body at send time and records send outcomes in `email_send_logs`.

## Running with Docker

```bash
# Build image
docker build -t gomail:local .

# Run container (memory storage)
docker run --rm -p 50051:50051 --env-file .env gomail:local
```

For PostgreSQL, include `STORAGE_DRIVER=postgres` and `DATABASE_URL` in your `.env`.

## Health Check

gRPC health service is enabled. You can check readiness:

```bash
grpcurl -plaintext localhost:50051 grpc.health.v1.Health/Check
```

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

# List templates (UUID IDs)
grpcurl -plaintext -d '{}' localhost:50051 mail.MailService/ListTemplates
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
Retrieves all emails, including `sent`, `failed`, and `pending` statuses.

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

### ListTemplates
Retrieves available templates (UUID-based IDs).

**Request:**
```protobuf
{}
```

**Response:**
```protobuf
{
  "templates": [
    {
      "id": "9f1b5f8b-9bb2-4a8e-9b8f-2dc92f9c0f11",
      "name": "Welcome",
      "subject": "Welcome!"
    }
  ],
  "count": 2
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
