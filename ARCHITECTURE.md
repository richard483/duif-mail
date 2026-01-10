# Clean Architecture Implementation

This project follows **Clean Architecture** principles (also known as Hexagonal Architecture or Ports & Adapters).

## Architecture Layers

```
┌─────────────────────────────────────────────────────────┐
│                    cmd/api (main.go)                    │
│                   Entry Point / Composition Root        │
└─────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────┐
│              internal/server (gRPC Server)              │
│              Delivery Layer / Adapters                  │
│         (Converts gRPC requests to domain calls)        │
└─────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────┐
│            internal/usecase (MailUseCase)               │
│              Business Logic / Use Cases                 │
│         (Implements domain.MailService interface)       │
└─────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────┐
│     internal/domain (Entities + Interfaces/Ports)       │
│                    Core Business Domain                 │
│        (No dependencies on outer layers/frameworks)     │
└─────────────────────────────────────────────────────────┘
                              ▲
                              │
┌─────────────────────────────────────────────────────────┐
│    internal/repository (InMemoryEmailRepository)        │
│         Data Access Layer / Infrastructure              │
│    (Implements domain.EmailRepository interface)        │
└─────────────────────────────────────────────────────────┘
```

## Dependency Rule

✅ **Dependencies point INWARD** - Outer layers depend on inner layers, never the reverse:
- `server` → `domain` (uses `MailService` interface)
- `usecase` → `domain` (implements `MailService`, uses `EmailRepository` interface)
- `repository` → `domain` (implements `EmailRepository` interface)
- `domain` → **NOTHING** (pure business logic, no external dependencies)

## Layer Responsibilities

### 1. Domain Layer (`internal/domain/`)
**The core of the application - contains business entities and business rules.**

- **Files:**
  - `email.go` - Core entity (`Email`) and repository interface (`EmailRepository`)
  - `request.go` - DTOs for business operations
  - `service.go` - Business service interface (`MailService`)

- **Characteristics:**
  - ✅ No framework dependencies
  - ✅ No imports from other internal layers
  - ✅ Defines interfaces that outer layers implement
  - ✅ Contains pure business logic

### 2. Use Case Layer (`internal/usecase/`)
**Application-specific business rules and orchestration.**

- **Files:**
  - `mail_usecase.go` - Implements `domain.MailService` interface

- **Responsibilities:**
  - ✅ Orchestrates domain entities
  - ✅ Implements business workflows
  - ✅ Depends only on domain interfaces
  - ✅ Framework-agnostic

- **Example:**
  ```go
  func (u *MailUseCase) SendEmail(req *domain.SendEmailRequest) (*domain.SendEmailResponse, error) {
      // 1. Validate
      // 2. Create domain entity
      // 3. Call repository (via interface)
      // 4. Return response
  }
  ```

### 3. Repository Layer (`internal/repository/`)
**Data access implementations.**

- **Files:**
  - `email_repository.go` - Implements `domain.EmailRepository` interface

- **Characteristics:**
  - ✅ Implements domain interfaces
  - ✅ Handles data persistence (in-memory, database, etc.)
  - ✅ Can be swapped without affecting business logic
  - ✅ Converts between storage format and domain entities

### 4. Server/Delivery Layer (`internal/server/`)
**External interface adapters (gRPC in this case).**

- **Files:**
  - `grpc_server.go` - Implements gRPC service interface

- **Responsibilities:**
  - ✅ Converts gRPC requests to domain calls
  - ✅ Depends on domain interfaces (not concrete implementations)
  - ✅ Handles protocol-specific concerns (protobuf conversion)
  - ✅ Framework/protocol-specific code isolated here

- **Example:**
  ```go
  func (s *GRPCServer) SendEmail(ctx context.Context, req *pb.EmailRequest) (*pb.EmailResponse, error) {
      // 1. Convert protobuf to domain request
      // 2. Call business service (via interface)
      // 3. Convert domain response to protobuf
  }
  ```

### 5. Main/Composition Root (`cmd/api/main.go`)
**Wires everything together.**

- **Responsibilities:**
  - ✅ Creates concrete implementations
  - ✅ Injects dependencies
  - ✅ Starts the server
  - ✅ Only place where all layers are visible

## Key Benefits

1. **Testability** - Each layer can be tested independently with mocks
2. **Flexibility** - Easy to swap implementations (e.g., replace in-memory with PostgreSQL)
3. **Independence** - Business logic doesn't depend on frameworks, UI, or databases
4. **Maintainability** - Clear separation of concerns
5. **Scalability** - Easy to add new delivery mechanisms (REST API, CLI, etc.)

## Design Patterns Used

1. **Dependency Inversion Principle (DIP)**
   - Outer layers depend on inner layer interfaces
   - Example: `GRPCServer` depends on `domain.MailService` interface, not `MailUseCase` struct

2. **Repository Pattern**
   - Abstracts data access behind `EmailRepository` interface
   - Example: Can swap `InMemoryEmailRepository` with `PostgresEmailRepository`

3. **Ports & Adapters**
   - Domain defines "ports" (interfaces)
   - Infrastructure provides "adapters" (implementations)
   - Example: `domain.EmailRepository` (port), `InMemoryEmailRepository` (adapter)

## Testing Strategy

```go
// Unit test use case with mock repository
func TestMailUseCase_SendEmail(t *testing.T) {
    mockRepo := &MockEmailRepository{}
    useCase := usecase.NewMailUseCase(mockRepo)
    
    resp, err := useCase.SendEmail(&domain.SendEmailRequest{...})
    // assertions
}

// Integration test gRPC server with mock service
func TestGRPCServer_SendEmail(t *testing.T) {
    mockService := &MockMailService{}
    server := server.NewGRPCServer(mockService)
    
    resp, err := server.SendEmail(context.Background(), &pb.EmailRequest{...})
    // assertions
}
```

## Adding a New Feature

Example: Add email templates

1. **Domain** - Add `Template` entity and `TemplateRepository` interface
2. **Repository** - Implement `InMemoryTemplateRepository`
3. **Use Case** - Update `SendEmail` to use templates
4. **Server** - Add gRPC method for template management
5. **Proto** - Define template messages and service methods
6. **Main** - Wire new dependencies

No changes needed in existing tests for unrelated features! 🎉
