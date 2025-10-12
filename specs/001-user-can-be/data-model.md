# Data Model: User Authentication

**Date**: October 12, 2025  
**Feature**: [spec.md](./spec.md)  
**Research**: [research.md](./research.md)

## Overview

This document defines the data entities for the authentication system following hexagonal architecture principles. Entities are defined as domain objects (core layer) separate from database schema (adapter layer).

---

## Domain Entities

### 1. User

**Purpose**: Represents a registered user account in the system

**Attributes**:
- `ID` (UUID): Unique identifier
- `Name` (string): Full name (1-255 characters)
- `Email` (string): Email address, unique, normalized to lowercase
- `PasswordHash` (string): bcrypt-hashed password (never store plain text)
- `Role` (string): User role - "user" or "verificator" (enum)
- `CreatedAt` (time.Time): Account creation timestamp
- `UpdatedAt` (time.Time): Last modification timestamp
- `LastLoginAt` (*time.Time): Last successful login (nullable)

**Validation Rules**:
- Email: Valid RFC 5322 format, unique, max 255 chars
- Name: Required, 1-255 characters
- Password (pre-hash): Min 8 chars, at least 1 uppercase, 1 lowercase, 1 number
- Role: Must be "user" or "verificator", defaults to "user"

**Relationships**:
- Has many `RefreshToken` (one user can have multiple active refresh tokens)
- Has many `PasswordResetToken` (one user can have multiple reset requests)
- Has many `AuthEventLog` (audit trail of authentication events)

**Business Rules**:
- Email must be unique across all users
- Password must be hashed with bcrypt before storage
- Role defaults to "user" for all new registrations
- Cannot delete user while active refresh tokens exist (cascade delete)

---

### 2. RefreshToken

**Purpose**: Represents a long-lived token for obtaining new access tokens without re-authentication

**Attributes**:
- `ID` (UUID): Unique identifier
- `UserID` (UUID): Foreign key to User
- `TokenHash` (string): SHA-256 hash of the actual token (for secure storage)
- `ExpiresAt` (time.Time): Expiration timestamp (30 days from creation)
- `Revoked` (bool): Whether token has been explicitly revoked
- `CreatedAt` (time.Time): Token creation timestamp
- `LastUsedAt` (*time.Time): Last time token was used for refresh (nullable)

**Validation Rules**:
- TokenHash: Required, unique, 64 characters (SHA-256 hex)
- UserID: Must reference existing user
- ExpiresAt: Must be future timestamp
- Revoked: Defaults to false

**Relationships**:
- Belongs to one `User`

**Business Rules**:
- Token is valid only if: `ExpiresAt > NOW()` AND `Revoked = false`
- Token is single-use: After refresh, old token marked as revoked
- Logout revokes all user's refresh tokens
- Cascade delete when user is deleted

**State Transitions**:
```
Created → Active (not revoked, not expired)
Active → Used (after successful refresh, marked revoked)
Active → Revoked (explicit logout)
Active → Expired (ExpiresAt passed)
```

---

### 3. PasswordResetToken

**Purpose**: Represents a time-limited token for password reset verification

**Attributes**:
- `ID` (UUID): Unique identifier
- `UserID` (UUID): Foreign key to User
- `TokenHash` (string): SHA-256 hash of the actual token
- `ExpiresAt` (time.Time): Expiration timestamp (1 hour from creation)
- `Used` (bool): Whether token has been used
- `CreatedAt` (time.Time): Token creation timestamp

**Validation Rules**:
- TokenHash: Required, unique, 64 characters (SHA-256 hex)
- UserID: Must reference existing user
- ExpiresAt: Must be future timestamp
- Used: Defaults to false

**Relationships**:
- Belongs to one `User`

**Business Rules**:
- Token is valid only if: `ExpiresAt > NOW()` AND `Used = false`
- Token is single-use: Marked as used after successful password reset
- Expired tokens cannot be used (even if not marked as used)
- User can have multiple reset tokens (but only latest valid one works)

**State Transitions**:
```
Created → Active (not used, not expired)
Active → Used (password successfully reset)
Active → Expired (1 hour passed)
```

---

### 4. AuthEventLog

**Purpose**: Audit trail for authentication-related events (security monitoring)

**Attributes**:
- `ID` (UUID): Unique identifier
- `UserID` (*UUID): Foreign key to User (nullable for failed login attempts)
- `EventType` (string): Type of authentication event (enum)
- `IPAddress` (string): IP address of the request (IPv4 or IPv6, max 45 chars)
- `UserAgent` (string): Browser/client user agent
- `Success` (bool): Whether the authentication action succeeded
- `ErrorMessage` (string): Error details if Success=false (nullable)
- `CreatedAt` (time.Time): Event timestamp

**Event Types** (enum):
- `LOGIN_SUCCESS`: Successful login
- `LOGIN_FAILURE`: Failed login attempt
- `REGISTER_SUCCESS`: Successful registration
- `REGISTER_FAILURE`: Failed registration
- `TOKEN_REFRESH_SUCCESS`: Successful token refresh
- `TOKEN_REFRESH_FAILURE`: Failed token refresh
- `LOGOUT`: User logout
- `PASSWORD_RESET_REQUEST`: Password reset requested
- `PASSWORD_RESET_SUCCESS`: Password successfully reset
- `PASSWORD_RESET_FAILURE`: Failed password reset

**Validation Rules**:
- EventType: Must be one of the defined types
- IPAddress: Valid IPv4 or IPv6 format
- CreatedAt: Auto-set on creation

**Relationships**:
- Belongs to one `User` (optional - failed attempts may not have valid user)

**Business Rules**:
- Immutable: Events cannot be updated or deleted (audit integrity)
- UserID nullable to track failed login attempts with unknown users
- Indexed on (UserID, CreatedAt) for efficient audit queries

---

## Entity Relationships Diagram

```
┌─────────────────┐
│      User       │
│─────────────────│
│ ID (PK)         │
│ Name            │
│ Email (unique)  │
│ PasswordHash    │
│ Role            │
│ CreatedAt       │
│ UpdatedAt       │
│ LastLoginAt     │
└────────┬────────┘
         │
         │ 1:N
         ├──────────────────┬──────────────────┬─────────────────┐
         │                  │                  │                 │
         ▼                  ▼                  ▼                 ▼
┌────────────────┐ ┌──────────────────┐ ┌─────────────────┐ ┌──────────────┐
│ RefreshToken   │ │PasswordResetToken│ │  AuthEventLog   │ │ (Future:     │
│────────────────│ │──────────────────│ │─────────────────│ │  Reports,    │
│ ID (PK)        │ │ ID (PK)          │ │ ID (PK)         │ │  etc.)       │
│ UserID (FK)    │ │ UserID (FK)      │ │ UserID (FK)     │ └──────────────┘
│ TokenHash      │ │ TokenHash        │ │ EventType       │
│ ExpiresAt      │ │ ExpiresAt        │ │ IPAddress       │
│ Revoked        │ │ Used             │ │ UserAgent       │
│ CreatedAt      │ │ CreatedAt        │ │ Success         │
│ LastUsedAt     │ └──────────────────┘ │ ErrorMessage    │
└────────────────┘                      │ CreatedAt       │
                                        └─────────────────┘
```

---

## Data Flow Examples

### Registration Flow
```
1. User submits: {name, email, password}
2. Validate: email format, password strength, email uniqueness
3. Hash password with bcrypt (cost 12)
4. Create User entity: {ID=UUID, Name, Email(lowercase), PasswordHash, Role="user"}
5. Save User to database
6. Log AuthEventLog: {EventType=REGISTER_SUCCESS, Success=true}
7. Return User (without PasswordHash)
```

### Login Flow
```
1. User submits: {email, password}
2. Find User by email
3. Verify password: bcrypt.Compare(password, user.PasswordHash)
4. If valid:
   a. Generate access token JWT (24h expiry)
   b. Generate refresh token (UUID)
   c. Hash refresh token with SHA-256
   d. Create RefreshToken entity: {UserID, TokenHash, ExpiresAt=+30days}
   e. Save RefreshToken to database
   f. Update User.LastLoginAt
   g. Log AuthEventLog: {EventType=LOGIN_SUCCESS, Success=true}
   h. Return {accessToken, refreshToken}
5. If invalid:
   a. Log AuthEventLog: {EventType=LOGIN_FAILURE, Success=false, ErrorMessage}
   b. Return error
```

### Token Refresh Flow
```
1. User submits refresh token
2. Hash submitted token with SHA-256
3. Find RefreshToken by TokenHash
4. Validate: ExpiresAt > NOW() AND Revoked = false
5. If valid:
   a. Generate new access token JWT (24h expiry)
   b. Generate new refresh token (UUID)
   c. Hash new refresh token with SHA-256
   d. Mark old RefreshToken as Revoked=true
   e. Create new RefreshToken entity: {UserID, TokenHash(new), ExpiresAt=+30days}
   f. Save changes
   g. Log AuthEventLog: {EventType=TOKEN_REFRESH_SUCCESS}
   h. Return {accessToken, refreshToken(new)}
6. If invalid:
   a. Log AuthEventLog: {EventType=TOKEN_REFRESH_FAILURE}
   b. Return error
```

### Password Reset Flow
```
1. User requests reset: {email}
2. Find User by email
3. If found:
   a. Generate reset token (UUID)
   b. Hash reset token with SHA-256
   c. Create PasswordResetToken: {UserID, TokenHash, ExpiresAt=+1hour}
   d. Save PasswordResetToken
   e. Send email with reset link containing token
   f. Log AuthEventLog: {EventType=PASSWORD_RESET_REQUEST}
4. User submits: {resetToken, newPassword}
5. Hash submitted token, find PasswordResetToken
6. Validate: ExpiresAt > NOW() AND Used = false
7. If valid:
   a. Validate new password strength
   b. Hash new password with bcrypt
   c. Update User.PasswordHash
   d. Mark PasswordResetToken as Used=true
   e. Revoke all user's RefreshTokens (force re-login)
   f. Log AuthEventLog: {EventType=PASSWORD_RESET_SUCCESS}
   g. Return success
```

---

## Security Considerations

### Password Storage
- **Never store plain text passwords**
- Use bcrypt with cost factor 12 (balance security/performance)
- Password hashing happens in domain service, not adapter

### Token Storage
- **Access tokens**: Stateless JWT, not stored in database
- **Refresh tokens**: Store SHA-256 hash, not plain token (prevents theft if DB compromised)
- **Reset tokens**: Store SHA-256 hash, single-use, short TTL (1 hour)

### Data Protection
- Email addresses normalized to lowercase for consistency
- Sensitive fields (PasswordHash, TokenHash) never returned in API responses
- Audit logs immutable for forensics

---

## Database Migration Strategy

**Migration Files** (in chronological order):

1. `001_create_users_table` - Create users table with indexes
2. `002_create_refresh_tokens_table` - Create refresh tokens with foreign key
3. `003_create_password_reset_tokens_table` - Create reset tokens
4. `004_create_auth_event_logs_table` - Create audit log table

**Rollback Support**: Each migration has `.up.sql` and `.down.sql` for reversibility

**Indexing Strategy**:
- Primary keys (UUID): Automatic index
- users.email: Unique index
- refresh_tokens.user_id: Foreign key index
- refresh_tokens.token_hash: Unique index
- password_reset_tokens.user_id: Foreign key index
- password_reset_tokens.token_hash: Unique index
- auth_event_logs.user_id: Index for user audit queries
- auth_event_logs.created_at: Index for time-range queries
- Composite index on auth_event_logs(user_id, created_at) for user activity queries

---

## Summary

Four core entities define the authentication domain:
1. **User**: Account data with secure password storage
2. **RefreshToken**: Session management with revocation support
3. **PasswordResetToken**: Secure password recovery
4. **AuthEventLog**: Security audit trail

All entities follow hexagonal architecture principles with clear separation between domain models (core) and database schema (adapters). Ready to generate API contracts.
