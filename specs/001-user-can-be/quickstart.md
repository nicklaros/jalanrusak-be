# Quickstart: User Authentication Implementation

**Date**: October 12, 2025  
**Feature**: [spec.md](./spec.md)  
**Contracts**: [auth-api.yaml](./contracts/auth-api.yaml)

## Overview

This guide provides step-by-step instructions for implementing the user authentication feature following hexagonal architecture principles.

---

## Prerequisites

- Go 1.21+ installed
- PostgreSQL 14+ running locally or accessible
- golang-migrate CLI installed (`brew install golang-migrate` or download from GitHub)
- Git configured on feature branch `001-user-can-be`

---

## Step 1: Database Setup

### Create Database

```bash
# Create development database
createdb jalanrusak_dev

# Or using psql
psql -U postgres
CREATE DATABASE jalanrusak_dev;
```

### Run Migrations

```bash
# Navigate to project root
cd /home/nicklaros/dev/nicklaros/jalanrusak-be

# Run migrations
migrate -path migrations -database "postgres://localhost:5432/jalanrusak_dev?sslmode=disable" up

# Verify tables created
psql -U postgres -d jalanrusak_dev -c "\dt"
# Should see: users, refresh_tokens, password_reset_tokens, auth_event_logs
```

---

## Step 2: Install Dependencies

```bash
# Core dependencies
go get github.com/gin-gonic/gin
go get github.com/golang-jwt/jwt/v5
go get github.com/lib/pq
go get golang.org/x/crypto/bcrypt
go get github.com/google/uuid
go get github.com/spf13/viper

# Testing dependencies
go get github.com/stretchr/testify
go get github.com/DATA-DOG/go-sqlmock
```

---

## Step 3: Configuration

Create `.env` file in project root:

```bash
# Server
SERVER_PORT=8080

# Database
DATABASE_URL=postgres://localhost:5432/jalanrusak_dev?sslmode=disable

# JWT Configuration
JWT_SECRET=your-secret-key-change-in-production
ACCESS_TOKEN_TTL_HOURS=24
REFRESH_TOKEN_TTL_DAYS=30

# Email Service
EMAIL_SERVICE_TYPE=console
# For production SMTP:
# EMAIL_SERVICE_TYPE=smtp
# SMTP_HOST=smtp.gmail.com
# SMTP_PORT=587
# SMTP_USER=your-email@gmail.com
# SMTP_PASS=your-app-password
```

**⚠️ Important**: Add `.env` to `.gitignore` (never commit secrets!)

---

## Step 4: Implementation Order

Follow this sequence to maintain hexagonal architecture principles:

### Phase A: Core Domain (No External Dependencies)

1. **Domain Entities** (`core/domain/entities/`)
   - [ ] `user.go` - User entity with validation
   - [ ] `refresh_token.go` - RefreshToken entity
   - [ ] `password_reset_token.go` - PasswordResetToken entity

2. **Domain Errors** (`core/domain/errors/`)
   - [ ] `auth_errors.go` - All authentication errors

3. **Port Interfaces** (`core/ports/`)
   - [ ] `usecases/auth_service.go` - Authentication use cases
   - [ ] `usecases/user_service.go` - User management use cases
   - [ ] `usecases/password_service.go` - Password management use cases
   - [ ] `external/user_repository.go` - User repository port
   - [ ] `external/token_repository.go` - Token repository ports
   - [ ] `external/token_generator.go` - Token generation port
   - [ ] `external/password_hasher.go` - Password hashing port
   - [ ] `external/email_service.go` - Email service port

4. **Service Implementations** (`core/services/`)
   - [ ] `auth_service_impl.go` - Login, logout, refresh logic
   - [ ] `user_service_impl.go` - Registration, profile logic
   - [ ] `password_service_impl.go` - Password reset logic

5. **Unit Tests** (`tests/unit/core/services/`)
   - [ ] `auth_service_test.go` - Mock all ports, test business logic
   - [ ] `user_service_test.go`
   - [ ] `password_service_test.go`

### Phase B: Output Adapters

6. **PostgreSQL Repositories** (`adapters/out/repository/postgres/`)
   - [ ] `user_repository.go` - Implement UserRepository port
   - [ ] `refresh_token_repository.go` - Implement TokenRepository port
   - [ ] `password_reset_repository.go` - Implement ResetTokenRepository port

7. **Security Adapters** (`adapters/out/security/`)
   - [ ] `jwt_token_generator.go` - Implement TokenGenerator port (JWT)
   - [ ] `bcrypt_password_hasher.go` - Implement PasswordHasher port (bcrypt)

8. **Email Adapters** (`adapters/out/messaging/`)
   - [ ] `console_email_service.go` - Console logging implementation
   - [ ] `smtp_email_service.go` - SMTP implementation

9. **Integration Tests** (`tests/integration/adapters/repository/`)
   - [ ] `user_repository_test.go` - Test with real PostgreSQL
   - [ ] `refresh_token_repository_test.go`
   - [ ] `password_reset_repository_test.go`

### Phase C: Input Adapters

10. **DTOs** (`adapters/in/http/dto/`)
    - [ ] `auth_request.go` - Register, Login, Refresh request DTOs
    - [ ] `auth_response.go` - Auth, User, Error response DTOs

11. **HTTP Handlers** (`adapters/in/http/handlers/`)
    - [ ] `auth_handler.go` - Register, Login, Refresh, Logout handlers
    - [ ] `password_handler.go` - Password reset request/confirm handlers
    - [ ] `user_handler.go` - Get profile handler

12. **Middleware** (`adapters/in/http/middleware/`)
    - [ ] `auth_middleware.go` - JWT validation middleware
    - [ ] `error_middleware.go` - Global error handling

13. **Routes** (`adapters/in/http/routes/`)
    - [ ] `routes.go` - Wire all endpoints

14. **API Tests** (`tests/api/`)
    - [ ] `auth_api_test.go` - Full end-to-end HTTP tests

### Phase D: Application Entry Point

15. **Configuration** (`config/`)
    - [ ] `config.go` - Load from environment with viper

16. **Main Application** (`cmd/server/`)
    - [ ] `main.go` - Dependency injection, wire everything together

---

## Step 5: Running the Application

```bash
# From project root
go run cmd/server/main.go

# Should see:
# [GIN-debug] Listening and serving HTTP on :8080
```

---

## Step 6: Testing the API

### Register a new user

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john.doe@example.com",
    "password": "SecureP@ss123"
  }'

# Expected response (201 Created):
{
  "accessToken": "eyJhbGciOiJIUzI1NiIs...",
  "refreshToken": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "tokenType": "Bearer",
  "expiresIn": 86400,
  "user": {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "name": "John Doe",
    "email": "john.doe@example.com",
    "role": "user",
    "createdAt": "2025-10-12T10:30:00Z"
  }
}
```

### Login

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john.doe@example.com",
    "password": "SecureP@ss123"
  }'

# Expected response (200 OK): Same as register response
```

### Get Profile (Protected Endpoint)

```bash
# Replace <ACCESS_TOKEN> with actual token from login/register
curl -X GET http://localhost:8080/api/v1/users/profile \
  -H "Authorization: Bearer <ACCESS_TOKEN>"

# Expected response (200 OK):
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "name": "John Doe",
  "email": "john.doe@example.com",
  "role": "user",
  "createdAt": "2025-10-12T10:30:00Z",
  "lastLoginAt": "2025-10-12T14:45:00Z"
}
```

### Refresh Token

```bash
# Replace <REFRESH_TOKEN> with actual refresh token
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refreshToken": "<REFRESH_TOKEN>"
  }'

# Expected response (200 OK): New access and refresh tokens
```

### Logout

```bash
curl -X POST http://localhost:8080/api/v1/auth/logout \
  -H "Authorization: Bearer <ACCESS_TOKEN>"

# Expected response (200 OK):
{
  "message": "Logout successful"
}
```

### Request Password Reset

```bash
curl -X POST http://localhost:8080/api/v1/auth/password-reset/request \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john.doe@example.com"
  }'

# Expected response (200 OK):
{
  "message": "Password reset email sent"
}

# Check console logs for reset token (EMAIL_SERVICE_TYPE=console)
```

### Confirm Password Reset

```bash
# Replace <RESET_TOKEN> from console logs
curl -X POST http://localhost:8080/api/v1/auth/password-reset/confirm \
  -H "Content-Type: application/json" \
  -d '{
    "token": "<RESET_TOKEN>",
    "newPassword": "NewSecureP@ss456"
  }'

# Expected response (200 OK):
{
  "message": "Password reset successful"
}
```

---

## Step 7: Running Tests

```bash
# Unit tests (core services)
go test ./tests/unit/core/services/... -v

# Integration tests (repositories)
go test ./tests/integration/adapters/repository/... -v

# API tests (end-to-end)
go test ./tests/api/... -v

# All tests with coverage
go test ./... -cover

# Coverage report
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

---

## Architecture Validation

### Verify Hexagonal Architecture Compliance

**✅ Core domain has no external dependencies:**
```bash
# Should show NO imports from gin, pq, jwt, etc.
grep -r "github.com/gin-gonic" core/
grep -r "github.com/lib/pq" core/
# Both should return nothing
```

**✅ Adapters depend on core, not vice versa:**
```bash
# Core should not import adapters
grep -r "adapters/" core/
# Should return nothing
```

**✅ All tests pass:**
```bash
go test ./... -v
# All tests should pass with >80% coverage in core/
```

---

## Troubleshooting

### Database Connection Issues

```bash
# Test database connection
psql -U postgres -d jalanrusak_dev -c "SELECT 1;"

# Check if migrations ran
psql -U postgres -d jalanrusak_dev -c "\dt"
```

### JWT Token Issues

```bash
# Verify JWT_SECRET is set
echo $JWT_SECRET

# Test token generation manually
go run -m jwt # (create test script)
```

### Port Already in Use

```bash
# Find process using port 8080
lsof -i :8080

# Kill process
kill -9 <PID>

# Or change SERVER_PORT in .env
```

---

## Next Steps

After completing this implementation:

1. ✅ All authentication endpoints functional
2. ✅ Unit tests passing (>80% coverage in core)
3. ✅ Integration tests passing
4. ✅ API tests passing
5. ✅ Hexagonal architecture verified

**Future Enhancements** (Deferred):
- Role-based access control enforcement
- Rate limiting for brute force protection
- Email verification requirement
- Multi-factor authentication
- Session management dashboard

---

## Reference Documentation

- **Spec**: [spec.md](./spec.md) - Complete feature specification
- **Research**: [research.md](./research.md) - Technical decisions and rationale
- **Data Model**: [data-model.md](./data-model.md) - Entity definitions and relationships
- **API Contract**: [contracts/auth-api.yaml](./contracts/auth-api.yaml) - OpenAPI specification
- **Constitution**: [/.github/CONSTITUTION.md](/.github/CONSTITUTION.md) - Project principles

---

**Ready to implement!** Follow the phase order above to maintain clean architecture throughout development.
