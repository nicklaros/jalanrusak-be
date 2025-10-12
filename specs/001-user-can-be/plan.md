# Implementation Plan: User Authentication for API Access

**Branch**: `001-user-can-be` | **Date**: October 12, 2025 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/001-user-can-be/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

Implement a comprehensive authentication system for the JalanRusak backend API that enables users to register, log in, manage sessions, and reset passwords. The system uses JWT-based authentication with refresh tokens for seamless user experience, PostgreSQL for data persistence, and follows hexagonal architecture principles. This phase focuses on core authentication functionality while deferring role-based access control enforcement and rate limiting to future iterations.

## Technical Context

**Language/Version**: Go 1.21+  
**Primary Dependencies**: gin-gonic/gin (HTTP), golang-jwt/jwt (JWT), lib/pq or pgx (PostgreSQL driver), testify (testing)  
**Storage**: PostgreSQL (users, refresh tokens, password reset tokens, auth event logs)  
**Testing**: Go testing package with testify for assertions and mocking  
**Target Platform**: Linux server (containerizable)  
**Project Type**: Backend API (single project following hexagonal architecture)  
**Performance Goals**: <200ms response time for authentication endpoints, support 100+ concurrent authentication requests  
**Constraints**: Must follow hexagonal architecture (core domain independent of frameworks), stateless JWT access tokens, database-stored refresh tokens for revocation  
**Scale/Scope**: Initial MVP focused on authentication only (registration, login, token refresh, logout, password reset); role field stored but authorization deferred

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### ✅ Hexagonal Architecture Compliance

- **Core Domain Independence**: ✅ PASS - Authentication business logic will reside in `core/services/`, domain entities in `core/domain/entities/`, with no framework dependencies
- **Ports & Adapters**: ✅ PASS - Use case interfaces in `core/ports/usecases/`, repository interfaces in `core/ports/external/`, HTTP handlers in `adapters/in/http/handlers/`, PostgreSQL repositories in `adapters/out/repository/postgres/`
- **Business Logic Isolation**: ✅ PASS - All authentication rules, password hashing, token generation in core services, HTTP/database details in adapters

### ✅ Security Requirements

- **JWT Authentication**: ✅ PASS - Access tokens (stateless JWT) + refresh tokens (database-stored for revocation)
- **Password Security**: ✅ PASS - bcrypt hashing, strength validation (8+ chars, mixed case, numbers)
- **Input Validation**: ✅ PASS - Email format, password strength, SQL injection prevention
- **Secrets Management**: ✅ PASS - JWT secrets, database credentials via environment variables
- **HTTPS**: ✅ DEFERRED - Enforced at deployment/infrastructure level (not application code)

### ✅ API Design

- **RESTful Conventions**: ✅ PASS - `/api/v1/auth/register`, `/api/v1/auth/login`, `/api/v1/auth/refresh`, `/api/v1/auth/logout`, `/api/v1/auth/password-reset`
- **HTTP Status Codes**: ✅ PASS - 200 (success), 201 (created), 400 (validation), 401 (auth failed), 500 (server error)
- **JSON Format**: ✅ PASS - Consistent request/response structures
- **Error Messages**: ✅ PASS - Clear, actionable error responses

### ✅ Testing Strategy

- **Unit Tests**: ✅ PASS - Core authentication service logic tested in isolation with mocked ports
- **Integration Tests**: ✅ PASS - PostgreSQL repository tests with real database
- **API Tests**: ✅ PASS - End-to-end HTTP handler tests
- **Coverage Goal**: ✅ PASS - >80% coverage in core domain

### ✅ Database Guidelines

- **Migrations**: ✅ PASS - golang-migrate for schema versioning (users, refresh_tokens, password_reset_tokens, auth_event_logs tables)
- **Transactions**: ✅ PASS - Multi-step operations (e.g., token refresh) wrapped in transactions
- **Naming**: ✅ PASS - snake_case tables/columns
- **Indexing**: ✅ PASS - Index on email (unique), refresh token string, reset token string

### ⚠️ Deferred Features (Documented)

- **Role-Based Authorization**: ⏸️ DEFERRED - Role field stored but not enforced in this phase
- **Rate Limiting**: ⏸️ DEFERRED - Brute force protection deferred to future implementation

**Gate Status**: ✅ **PASSED** - All constitutional requirements met; deferred features explicitly documented in spec

## Project Structure

### Documentation (this feature)

```
specs/001-user-can-be/
├── spec.md              # Feature specification
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
│   ├── auth-api.yaml    # OpenAPI spec for authentication endpoints
│   └── schemas.yaml     # Shared schema definitions
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```
jalanrusak-be/
├── cmd/
│   └── server/
│       └── main.go                          # Application entry point, dependency injection
├── core/
│   ├── domain/
│   │   ├── entities/
│   │   │   ├── user.go                      # User entity
│   │   │   ├── refresh_token.go             # Refresh token entity
│   │   │   └── password_reset_token.go      # Password reset token entity
│   │   └── errors/
│   │       └── auth_errors.go               # Authentication domain errors
│   ├── ports/
│   │   ├── usecases/
│   │   │   ├── auth_service.go              # Authentication use case interface
│   │   │   ├── user_service.go              # User management use case interface
│   │   │   └── password_service.go          # Password management use case interface
│   │   └── external/
│   │       ├── user_repository.go           # User repository port
│   │       ├── token_repository.go          # Token repository port
│   │       ├── token_generator.go           # Token generation port
│   │       ├── password_hasher.go           # Password hashing port
│   │       └── email_service.go             # Email service port
│   └── services/
│       ├── auth_service_impl.go             # Authentication business logic
│       ├── user_service_impl.go             # User management business logic
│       └── password_service_impl.go         # Password management business logic
├── adapters/
│   ├── in/
│   │   └── http/
│   │       ├── handlers/
│   │       │   ├── auth_handler.go          # Auth HTTP handlers (register, login, refresh, logout)
│   │       │   ├── password_handler.go      # Password reset handlers
│   │       │   └── user_handler.go          # User profile handlers
│   │       ├── middleware/
│   │       │   ├── auth_middleware.go       # JWT validation middleware
│   │       │   └── error_middleware.go      # Error handling middleware
│   │       ├── routes/
│   │       │   └── routes.go                # Route definitions
│   │       └── dto/
│   │           ├── auth_request.go          # Request DTOs
│   │           └── auth_response.go         # Response DTOs
│   └── out/
│       ├── repository/
│       │   └── postgres/
│       │       ├── user_repository.go       # PostgreSQL user repository
│       │       ├── refresh_token_repository.go  # PostgreSQL refresh token repository
│       │       └── password_reset_repository.go # PostgreSQL password reset repository
│       ├── security/
│       │   ├── jwt_token_generator.go       # JWT token implementation
│       │   └── bcrypt_password_hasher.go    # bcrypt password hashing
│       └── messaging/
│           ├── console_email_service.go     # Console email (development)
│           └── smtp_email_service.go        # SMTP email (production)
├── config/
│   └── config.go                            # Configuration loading
├── migrations/
│   ├── 001_create_users_table.up.sql
│   ├── 001_create_users_table.down.sql
│   ├── 002_create_refresh_tokens_table.up.sql
│   ├── 002_create_refresh_tokens_table.down.sql
│   ├── 003_create_password_reset_tokens_table.up.sql
│   ├── 003_create_password_reset_tokens_table.down.sql
│   ├── 004_create_auth_event_logs_table.up.sql
│   └── 004_create_auth_event_logs_table.down.sql
└── tests/
    ├── unit/
    │   └── core/
    │       └── services/
    │           ├── auth_service_test.go
    │           ├── user_service_test.go
    │           └── password_service_test.go
    ├── integration/
    │   └── adapters/
    │       └── repository/
    │           ├── user_repository_test.go
    │           ├── refresh_token_repository_test.go
    │           └── password_reset_repository_test.go
    └── api/
        └── auth_api_test.go                 # End-to-end API tests
```

**Structure Decision**: Single backend project following hexagonal architecture as per JalanRusak constitution. Core domain (`core/`) contains business logic and port definitions. Adapters (`adapters/`) contain HTTP handlers and PostgreSQL implementations. Clear separation ensures testability and maintainability.

## Complexity Tracking

*Fill ONLY if Constitution Check has violations that must be justified*

No violations - all constitutional requirements are met. Deferred features (role-based authorization, rate limiting) are explicitly documented in the specification and do not violate the constitution.
