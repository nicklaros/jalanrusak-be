# Research: User Authentication for API Access

**Date**: October 12, 2025  
**Feature**: [spec.md](./spec.md)

## Overview

This document consolidates research findings for implementing authentication in the JalanRusak backend following hexagonal architecture principles with Go, Gin, PostgreSQL, and JWT.

---

## 1. JWT Token Strategy

### Decision: Dual Token System (Access + Refresh)

**Chosen Approach**: Short-lived access tokens (JWT, 24 hours) + long-lived refresh tokens (database-stored, 30 days)

**Rationale**:
- **Access tokens**: Stateless JWT enables horizontal scaling, no database lookup per request
- **Refresh tokens**: Database storage enables explicit revocation on logout/security events
- **Security balance**: Short access token lifetime limits exposure, refresh tokens provide seamless UX
- **Revocation support**: Database-stored refresh tokens can be invalidated immediately

**Alternatives Considered**:
1. **Stateless JWT only** (no refresh) - Rejected: Cannot revoke tokens before expiration, users must re-login frequently
2. **Database-stored access tokens** - Rejected: Every API request requires database lookup, scaling issues
3. **Redis-stored refresh tokens** - Deferred: Adds infrastructure complexity, can migrate later if needed

**Implementation**:
- Access token: JWT with `user_id`, `email`, `role`, `exp`, `iat` claims
- Refresh token: UUID v4 stored in `refresh_tokens` table with `user_id`, `expires_at`, `revoked` flag
- golang-jwt/jwt v5 for JWT generation/validation
- crypto/rand for secure refresh token generation

---

## 2. Password Hashing

### Decision: bcrypt with cost factor 12

**Chosen Approach**: bcrypt algorithm from golang.org/x/crypto/bcrypt

**Rationale**:
- **Industry standard**: Proven, widely audited, secure against rainbow tables
- **Adaptive cost**: Can increase cost factor as hardware improves
- **Built-in salt**: Automatic salt generation, no manual handling
- **Go native**: First-class support in Go ecosystem

**Alternatives Considered**:
1. **Argon2** - Deferred: More modern but less Go ecosystem adoption, can migrate later
2. **scrypt** - Rejected: Less mainstream in Go community
3. **PBKDF2** - Rejected: Requires more manual configuration

**Implementation**:
- bcrypt.GenerateFromPassword() with cost 12
- bcrypt.CompareHashAndPassword() for verification
- Cost 12 balances security (~250ms hash time) with user experience

---

## 3. Database Schema Design

### Decision: Four tables for authentication

**Tables**:
1. **users**: Core user data
2. **refresh_tokens**: Active refresh tokens
3. **password_reset_tokens**: Password reset tokens
4. **auth_event_logs**: Security audit trail

**Rationale**:
- **Separation of concerns**: Each entity has dedicated table
- **Normalization**: No redundant data, clear relationships
- **Performance**: Proper indexing on lookup columns
- **Auditability**: Event logs enable security monitoring

**Schema Details**:

```sql
-- users table
id UUID PRIMARY KEY
name VARCHAR(255) NOT NULL
email VARCHAR(255) UNIQUE NOT NULL (indexed)
password_hash VARCHAR(255) NOT NULL
role VARCHAR(50) NOT NULL DEFAULT 'user'
created_at TIMESTAMP NOT NULL DEFAULT NOW()
updated_at TIMESTAMP NOT NULL DEFAULT NOW()
last_login_at TIMESTAMP

-- refresh_tokens table
id UUID PRIMARY KEY
user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE (indexed)
token_hash VARCHAR(255) UNIQUE NOT NULL (indexed)
expires_at TIMESTAMP NOT NULL
revoked BOOLEAN NOT NULL DEFAULT false
created_at TIMESTAMP NOT NULL DEFAULT NOW()
last_used_at TIMESTAMP

-- password_reset_tokens table
id UUID PRIMARY KEY
user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE (indexed)
token_hash VARCHAR(255) UNIQUE NOT NULL (indexed)
expires_at TIMESTAMP NOT NULL
used BOOLEAN NOT NULL DEFAULT false
created_at TIMESTAMP NOT NULL DEFAULT NOW()

-- auth_event_logs table
id UUID PRIMARY KEY
user_id UUID REFERENCES users(id) ON DELETE SET NULL (indexed)
event_type VARCHAR(100) NOT NULL (indexed)
ip_address VARCHAR(45)
user_agent TEXT
success BOOLEAN NOT NULL
error_message TEXT
created_at TIMESTAMP NOT NULL DEFAULT NOW() (indexed)
```

**Indexing Strategy**:
- Primary keys (UUID) automatically indexed
- Email column: Unique index for fast lookup and uniqueness enforcement
- user_id foreign keys: Indexed for join performance
- Token hashes: Unique indexes for fast validation
- auth_event_logs: Composite index on (user_id, created_at) for audit queries

---

## 4. Email Service Strategy

### Decision: Interface-based with console/SMTP implementations

**Chosen Approach**: Port interface with two adapters

**Rationale**:
- **Development simplicity**: Console logging shows emails without infrastructure
- **Production ready**: SMTP adapter for real email delivery
- **Testability**: Easy to mock email service in tests
- **Hexagonal compliance**: External service abstracted through port

**Implementation**:
```go
// Port interface (core/ports/external/email_service.go)
type EmailService interface {
    SendPasswordReset(email, resetToken string) error
}

// Console adapter (development)
type ConsoleEmailService struct{}
func (s *ConsoleEmailService) SendPasswordReset(email, resetToken string) error {
    log.Printf("PASSWORD RESET EMAIL\nTo: %s\nReset Token: %s\n", email, resetToken)
    return nil
}

// SMTP adapter (production)
type SMTPEmailService struct {
    host string
    port int
    username string
    password string
}
// Uses net/smtp for delivery
```

**Configuration**:
- Environment variable `EMAIL_SERVICE_TYPE=console|smtp`
- SMTP credentials from environment (SMTP_HOST, SMTP_PORT, SMTP_USER, SMTP_PASS)

**Alternatives Considered**:
1. **Third-party service (SendGrid/AWS SES)** - Deferred: Adds external dependency and cost, can integrate later
2. **Queue-based async email** - Deferred: Adds complexity, can optimize later if needed

---

## 5. Hexagonal Architecture Mapping

### Decision: Strict port/adapter separation

**Core Domain** (`core/`):
- **Entities**: User, RefreshToken, PasswordResetToken (pure Go structs, no DB tags)
- **Services**: AuthService, UserService, PasswordService (business logic)
- **Ports Use Cases**: Interface definitions for use cases (what app does)
- **Ports External**: Interface definitions for external dependencies (repos, tokens, email)
- **Errors**: Domain-specific errors (InvalidCredentialsError, TokenExpiredError, etc.)

**Adapters** (`adapters/`):
- **HTTP In**: Gin handlers, middleware, DTOs (HTTP→domain conversion)
- **Postgres Out**: Repository implementations (domain→SQL conversion)
- **Security Out**: JWT generator, bcrypt hasher
- **Messaging Out**: Email service implementations

**Dependency Flow**:
```
HTTP Request → Handler (adapter) → Service (core) → Repository Port (core) → Postgres Repo (adapter)
                    ↓                                        ↓
                   DTO                                  Domain Entity
```

**Benefits**:
- Core domain testable without HTTP/database
- Adapters swappable (e.g., replace Postgres with MySQL)
- Business logic independent of frameworks

---

## 6. Token Expiration Handling

### Decision: Access token 24h, refresh token 30d, reset token 1h

**Rationale**:
- **Access token (24h)**: Balance between security and UX, long enough for daily usage
- **Refresh token (30d)**: Monthly re-authentication acceptable, prevents indefinite sessions
- **Reset token (1h)**: Tight window for password reset, reduces attack surface

**Cleanup Strategy**:
- Expired tokens remain in database for audit trail
- Background job (deferred) or manual cleanup script for old records
- Queries filter by `expires_at > NOW()` and `revoked = false`

---

## 7. Error Handling Strategy

### Decision: Domain errors with HTTP translation

**Approach**:
```go
// Domain errors (core/domain/errors/)
var (
    ErrInvalidCredentials = errors.New("invalid email or password")
    ErrTokenExpired = errors.New("token has expired")
    ErrTokenRevoked = errors.New("token has been revoked")
    ErrEmailAlreadyExists = errors.New("email already registered")
    ErrWeakPassword = errors.New("password does not meet requirements")
)

// HTTP handler translation
func translateError(err error) (int, gin.H) {
    switch {
    case errors.Is(err, domain.ErrInvalidCredentials):
        return 401, gin.H{"error": "Invalid credentials"}
    case errors.Is(err, domain.ErrTokenExpired):
        return 401, gin.H{"error": "Token expired", "code": "TOKEN_EXPIRED"}
    case errors.Is(err, domain.ErrEmailAlreadyExists):
        return 400, gin.H{"error": "Email already registered"}
    default:
        return 500, gin.H{"error": "Internal server error"}
    }
}
```

**Benefits**:
- Domain errors stay in core (no HTTP dependencies)
- HTTP adapter translates to status codes
- Testable error handling
- Client-friendly error messages

---

## 8. Testing Strategy

### Decision: Three-tier testing approach

**Unit Tests** (core/services/):
- Mock all ports (repositories, token generators, email service)
- Test business logic in isolation
- Examples: password validation, token expiration logic, email uniqueness checks

**Integration Tests** (adapters/repository/):
- Use testcontainers or docker-compose for PostgreSQL
- Test real database operations
- Examples: user CRUD, token storage/retrieval, transaction handling

**API Tests** (tests/api/):
- Full HTTP request/response cycle
- Test authentication flows end-to-end
- Examples: registration→login→refresh→logout, password reset flow

**Coverage Goal**: >80% in core domain services

---

## 9. Configuration Management

### Decision: Environment variables with viper

**Chosen Approach**:
```go
type Config struct {
    ServerPort        int    `mapstructure:"SERVER_PORT"`
    DatabaseURL       string `mapstructure:"DATABASE_URL"`
    JWTSecret         string `mapstructure:"JWT_SECRET"`
    AccessTokenTTL    int    `mapstructure:"ACCESS_TOKEN_TTL_HOURS"`
    RefreshTokenTTL   int    `mapstructure:"REFRESH_TOKEN_TTL_DAYS"`
    EmailServiceType  string `mapstructure:"EMAIL_SERVICE_TYPE"`
    SMTPHost          string `mapstructure:"SMTP_HOST"`
    SMTPPort          int    `mapstructure:"SMTP_PORT"`
    SMTPUser          string `mapstructure:"SMTP_USER"`
    SMTPPass          string `mapstructure:"SMTP_PASS"`
}
```

**Defaults** (development):
- SERVER_PORT: 8080
- DATABASE_URL: postgres://localhost:5432/jalanrusak_dev
- ACCESS_TOKEN_TTL_HOURS: 24
- REFRESH_TOKEN_TTL_DAYS: 30
- EMAIL_SERVICE_TYPE: console

**Rationale**:
- viper supports .env files and environment variables
- Type-safe configuration struct
- Easy to override per environment
- Secrets never committed to code

---

## 10. Middleware Design

### Decision: JWT validation middleware

**Implementation**:
```go
// adapters/in/http/middleware/auth_middleware.go
func AuthMiddleware(tokenService ports.TokenValidator) gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.AbortWithStatusJSON(401, gin.H{"error": "Authorization header required"})
            return
        }
        
        token := strings.TrimPrefix(authHeader, "Bearer ")
        claims, err := tokenService.ValidateAccessToken(token)
        if err != nil {
            c.AbortWithStatusJSON(401, gin.H{"error": "Invalid or expired token"})
            return
        }
        
        c.Set("user_id", claims.UserID)
        c.Set("user_role", claims.Role)
        c.Next()
    }
}
```

**Usage**:
```go
// Protected routes
authorized := r.Group("/api/v1")
authorized.Use(middleware.AuthMiddleware(tokenService))
{
    authorized.GET("/profile", handlers.GetProfile)
    authorized.POST("/logout", handlers.Logout)
}
```

---

## Summary

All technical decisions are documented with clear rationale. No "NEEDS CLARIFICATION" items remain. The architecture follows hexagonal principles, leverages Go best practices, and aligns with the JalanRusak constitution. Ready to proceed to Phase 1 (Design & Contracts).
