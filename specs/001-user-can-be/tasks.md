# Tasks: User Authentication for API Access

**Input**: Design documents from `/specs/001-user-can-be/`  
**Prerequisites**: plan.md ✅, spec.md ✅, research.md ✅, data-model.md ✅, contracts/auth-api.yaml ✅

**Feature Branch**: `001-user-can-be`  
**Generated**: October 12, 2025

---

## Implementation Strategy

This feature implements user authentication following **hexagonal architecture** principles. Tasks are organized by user story to enable **independent implementation and testing** of each story.

**MVP Scope**: User Story 1 (Registration) + User Story 2 (Login) deliver a working authentication system.

**Testing Approach**: Tests are deferred and optional. Implementation can proceed without test coverage initially. Tests can be added later as needed.

---

## Task Format

`[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2, etc.)
- File paths are absolute from repository root

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and foundational structure needed by all user stories

- [X] **T001** Create Go module and project structure (`go mod init`, directories: `core/`, `adapters/`, `cmd/`, `config/`, `migrations/`, `tests/`)
- [X] **T002** [P] Install core dependencies (gin-gonic/gin, golang-jwt/jwt/v5, lib/pq, golang.org/x/crypto/bcrypt, google/uuid, spf13/viper)
- [X] **T003** [P] Install testing dependencies (stretchr/testify, DATA-DOG/go-sqlmock)
- [X] **T004** [P] Create `.env` template file with configuration variables (SERVER_PORT, DATABASE_URL, JWT_SECRET, ACCESS_TOKEN_TTL_HOURS, REFRESH_TOKEN_TTL_DAYS, EMAIL_SERVICE_TYPE)
- [X] **T005** [P] Add `.env` to `.gitignore` (security: never commit secrets)
- [X] **T006** Create database migrations for users table in `migrations/001_create_users_table.up.sql` and `migrations/001_create_users_table.down.sql`
- [X] **T007** Create database migrations for refresh_tokens table in `migrations/002_create_refresh_tokens_table.up.sql` and `migrations/002_create_refresh_tokens_table.down.sql`
- [X] **T008** Create database migrations for password_reset_tokens table in `migrations/003_create_password_reset_tokens_table.up.sql` and `migrations/003_create_password_reset_tokens_table.down.sql`
- [X] **T009** Create database migrations for auth_event_logs table in `migrations/004_create_auth_event_logs_table.up.sql` and `migrations/004_create_auth_event_logs_table.down.sql`
- [X] **T010** Create configuration loader in `config/config.go` (using viper to load environment variables)

**Checkpoint**: ✅ Database schema defined, dependencies installed, configuration ready

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core domain entities and ports that ALL user stories depend on. These MUST complete before implementing any user story.

### Domain Entities (Core Domain - No External Dependencies)

- [X] **T011** [P] Create User entity in `core/domain/entities/user.go` (ID, Name, Email, PasswordHash, Role, CreatedAt, UpdatedAt, LastLoginAt with validation methods)
- [X] **T012** [P] Create RefreshToken entity in `core/domain/entities/refresh_token.go` (ID, UserID, TokenHash, ExpiresAt, Revoked, CreatedAt, LastUsedAt with validation)
- [X] **T013** [P] Create PasswordResetToken entity in `core/domain/entities/password_reset_token.go` (ID, UserID, TokenHash, ExpiresAt, Used, CreatedAt with validation)
- [X] **T014** [P] Create AuthEventLog entity in `core/domain/entities/auth_event_log.go` (ID, UserID, EventType, IPAddress, UserAgent, Success, CreatedAt)

### Domain Errors

- [X] **T015** Create authentication domain errors in `core/domain/errors/auth_errors.go` (ErrInvalidCredentials, ErrUserAlreadyExists, ErrInvalidToken, ErrTokenExpired, ErrWeakPassword, ErrInvalidEmail, ErrUserNotFound)

### Port Interfaces (Contracts between Core and Adapters)

- [X] **T016** [P] Define AuthService port interface in `core/ports/usecases/auth_service.go` (Login, Logout, RefreshToken methods)
- [X] **T017** [P] Define UserService port interface in `core/ports/usecases/user_service.go` (Register, GetProfile methods)
- [X] **T018** [P] Define PasswordService port interface in `core/ports/usecases/password_service.go` (RequestReset, ConfirmReset methods)
- [X] **T019** [P] Define UserRepository port interface in `core/ports/external/user_repository.go` (Create, FindByEmail, FindByID, Update methods)
- [X] **T020** [P] Define TokenRepository port interface in `core/ports/external/token_repository.go` (SaveRefreshToken, FindRefreshToken, RevokeRefreshToken, RevokeAllUserTokens, SavePasswordResetToken, FindPasswordResetToken, MarkResetTokenUsed methods)
- [X] **T021** [P] Define TokenGenerator port interface in `core/ports/external/token_generator.go` (GenerateAccessToken, ValidateAccessToken, GenerateRefreshToken methods)
- [X] **T022** [P] Define PasswordHasher port interface in `core/ports/external/password_hasher.go` (Hash, Compare methods)
- [X] **T023** [P] Define EmailService port interface in `core/ports/external/email_service.go` (SendPasswordResetEmail method)

**Checkpoint**: ✅ Core domain defined, all port contracts established, ready for user story implementation

---

## Phase 3: User Story 1 - New User Registration (P1)

**Goal**: Enable new users to register with name, email, and password

**Independent Test**: Submit registration details → verify account created → verify can access protected endpoints

### Business Logic (Core Services)

- [X] **T024** [US1] Implement UserService in `core/services/user_service_impl.go` (Register method: validate email format, check uniqueness via UserRepository, validate password strength, hash password via PasswordHasher, create User entity with role="user", save via UserRepository, return user data)

### Output Adapters (Database & Security)

- [X] **T026** [US1] Implement PostgreSQL UserRepository in `adapters/out/repository/postgres/user_repository.go` (Create, FindByEmail, FindByID, Update methods with parameterized queries, handle unique constraint violations)
- [X] **T027** [US1] Implement bcrypt PasswordHasher in `adapters/out/security/bcrypt_password_hasher.go` (Hash with cost 12, Compare methods)
- [X] **T028** [US1] Implement AuthEventLogRepository in `adapters/out/repository/postgres/auth_event_log_repository.go` (Create, FindByUserID, FindFailedLoginAttempts methods)

### Input Adapters (HTTP Layer)

- [X] **T029** [US1] Create RegisterRequest and RegisterResponse DTOs in `adapters/in/http/dto/registration_dto.go`
- [X] **T030** [US1] Implement registration handler in `adapters/in/http/handlers/registration_handler.go` (POST /auth/register: parse RegisterRequest, call UserService.Register, return 201 with user data or 400 with validation errors)
- [X] **T031** [US1] Add registration route in `adapters/in/http/routes/routes.go` (POST /api/v1/auth/register → registrationHandler.Register)
- [X] **T032** [US1] Create main server entry point in `cmd/server/main.go` (wire all dependencies, start Gin server)

**Story Checkpoint**: ✅ US1 Complete - Users can register and accounts are created in database

---

## Phase 4: User Story 2 - User Login (P1)

**Goal**: Enable existing users to log in and receive authentication tokens

**Independent Test**: Provide valid credentials → verify authentication tokens received → verify tokens work for API access

### Business Logic (Core Services)

- [X] **T034** [US2] Implement AuthService in `core/services/auth_service_impl.go` (Login, RefreshToken, Logout, VerifyAccessToken methods with full business logic)

### Output Adapters (Tokens & Database)

- [X] **T036** [US2] Implement JWT TokenGenerator in `adapters/out/security/jwt_token_generator.go` (GenerateAccessToken with JWT claims, ValidateAccessToken, GenerateRefreshToken, HashToken methods)
- [X] **T037** [US2] Implement PostgreSQL RefreshTokenRepository in `adapters/out/repository/postgres/refresh_token_repository.go` (Create, FindByTokenHash, FindByUserID, Update, RevokeByUserID, RevokeByTokenHash, DeleteExpired methods)

### Input Adapters (HTTP Layer)

- [X] **T039** [US2] Create LoginRequest and LoginResponse DTOs in `adapters/in/http/dto/login_dto.go` (LoginRequest, LoginResponse, RefreshTokenRequest, RefreshTokenResponse, UserInfo)
- [X] **T040** [US2] Implement auth handlers in `adapters/in/http/handlers/auth_handler.go` (POST /auth/login, POST /auth/refresh, POST /auth/logout)
- [X] **T041** [US2] Create auth middleware in `adapters/in/http/middleware/auth_middleware.go` and update routes (public: login, refresh; protected: logout)

**Story Checkpoint**: ✅ US2 Complete - Users can log in and receive JWT tokens

---

## Phase 5: User Story 3 - Session Management (P2)

**Goal**: Token expiration validation and clear error messages

**Independent Test**: Login → wait for token expiration → verify expired tokens are rejected

**Note**: Session management validation is already implemented in Phase 4 (AuthService.VerifyAccessToken and auth middleware). No additional tasks required.

**Story Checkpoint**: ✅ US3 Complete - Token expiration is enforced and validated

---

## Phase 6: User Story 3a - Token Refresh (P2)

**Goal**: Users can refresh expired access tokens using refresh tokens

**Independent Test**: Login → wait for access token expiration → use refresh token → verify new access token works

**Note**: Token refresh is already implemented in Phase 4 (AuthService.RefreshToken and /auth/refresh endpoint). No additional tasks required.

**Story Checkpoint**: ✅ US3a Complete - Users can refresh tokens without re-login

---

## Phase 7: User Story 3b - User Logout (P2)

**Goal**: Users can explicitly logout and invalidate their refresh tokens

**Independent Test**: Login → logout → verify refresh token is revoked → verify cannot refresh

**Note**: Logout is already implemented in Phase 4 (AuthService.Logout and /auth/logout endpoint). No additional tasks required.

**Story Checkpoint**: ✅ US3b Complete - Users can securely logout

---

## Phase 8: User Story 5 - Password Security (P3)

**Goal**: Password reset via email with time-limited tokens

**Independent Test**: Request password reset → receive email → use reset token → verify password changed

### Business Logic (Core Services)

- [X] **T060** [US5] Implement PasswordService in `core/services/password_service_impl.go` (RequestPasswordReset, ResetPassword, ChangePassword methods with full business logic)

### Output Adapters (Database & Email)

- [X] **T062** [US5] Implement password reset token methods in `adapters/out/repository/postgres/password_reset_token_repository.go` (Create, FindByTokenHash, Update, DeleteByUserID, DeleteExpired methods)
- [X] **T063** [US5] Implement console EmailService in `adapters/out/messaging/console_email_service.go` (SendPasswordResetEmail, SendWelcomeEmail, SendPasswordChangedEmail to console)
- [ ] **T064** [P] [US5] Implement SMTP EmailService in `adapters/out/messaging/smtp_email_service.go` (SendPasswordResetEmail sends via SMTP for production)

### Input Adapters (HTTP Layer)

- [X] **T066** [US5] Create password DTOs in `adapters/in/http/dto/password_dto.go` (PasswordResetRequestRequest, PasswordResetConfirmRequest, PasswordChangeRequest with responses)
- [X] **T067** [US5] Implement password handlers in `adapters/in/http/handlers/password_handler.go` (POST /auth/password/reset-request, POST /auth/password/reset-confirm, POST /auth/password/change)
- [X] **T068** [US5] Add password routes in `adapters/in/http/routes/routes.go` (public: reset-request, reset-confirm; protected: change)

**Story Checkpoint**: ✅ US5 Complete - Password reset flow fully functional

---

## Phase 9: Application Wiring & Polish

**Purpose**: Complete application setup and cross-cutting concerns

### Dependency Injection & Main Entry Point

- [ ] **T070** Create main application in `cmd/server/main.go` (load config, initialize database connection, create repository instances, create service instances with dependency injection, create handler instances, setup Gin router with routes and middleware, start HTTP server)
- [ ] **T071** Add database connection pooling configuration in `cmd/server/main.go` (configure max open connections, max idle connections, connection lifetime)

### Authentication Event Logging (Cross-Cutting)

- [ ] **T072** [P] Create auth event logging in all AuthService methods in `core/services/auth_service_impl.go` (log login success/failure, token refresh, logout events with timestamp, user ID, IP address, user agent)
- [ ] **T073** [P] Implement AuthEventLog repository methods in `adapters/out/repository/postgres/auth_event_log_repository.go` (SaveEvent method with parameterized queries)

### Documentation & Deployment Preparation

- [ ] **T074** [P] Create API documentation from `contracts/auth-api.yaml` (ensure OpenAPI spec is up to date with implemented endpoints)
- [ ] **T075** [P] Update README.md with setup instructions (prerequisites, database setup, environment variables, running the application, API endpoints)
- [ ] **T076** [P] Create `.env.example` file with all required environment variables (sanitized, no real secrets)

**Final Checkpoint**: ✅ All user stories implemented, application fully wired, ready for deployment

---

## Task Dependencies

### Dependency Graph (Story Completion Order)

```
Phase 1 (Setup) → Phase 2 (Foundational)
                      ↓
    ┌─────────────────┼─────────────────┐
    ↓                 ↓                 ↓
 US1 (P1)         US2 (P1)         US5 (P3)
    ↓                 ↓                 
    └────→ US3 (P2) ←─┘
             ↓
    ┌────────┴────────┐
    ↓                 ↓
 US3a (P2)        US3b (P2)
    ↓                 ↓
    └────────┬────────┘
             ↓
      Phase 9 (Polish)
```

### Critical Path

Setup (T001-T010) → Foundational (T011-T023) → **US1 (T024-T033)** → **US2 (T034-T042)** → US3 (T043-T048) → US3a (T049-T054) → US3b (T055-T059) → US5 (T060-T069) → Polish (T070-T076)

**MVP = US1 + US2** (T001-T042): Registration and login provide core authentication functionality

---

## Parallel Execution Opportunities

### Setup Phase (T001-T010)
- T002 (install core deps) || T003 (install test deps) || T004 (create .env) || T005 (update .gitignore)
- T006, T007, T008, T009 (migrations) can all be created in parallel

### Foundational Phase (T011-T023)
- All domain entities (T011-T014) can be created in parallel
- All port interfaces (T016-T023) can be created in parallel

### User Story 1 (T024-T033)
- T026 (UserRepository) || T027 (PasswordHasher) can be implemented in parallel after T024-T025
- T029 (DTOs) || T031 (error middleware) can be created in parallel

### User Story 2 (T034-T042)
- T036 (JWT generator) || T037 (TokenRepository) can be implemented in parallel after T034-T035

### User Story 5 (T060-T069)
- T063 (console email) || T064 (SMTP email) can be implemented in parallel

### Polish Phase (T070-T076)
- T072 (event logging) || T074 (API docs) || T075 (README) || T076 (.env.example) can be completed in parallel

---

## Testing Strategy Summary

**Tests are deferred**: Unit tests, integration tests, and API tests are not required for initial implementation. Test coverage can be added later as needed. Focus on working implementation first.

---

## Summary

**Total Tasks**: 48 (after removing 28 test tasks)  
**By Phase**:
- Setup: 10 tasks
- Foundational: 13 tasks
- US1 (Registration): 7 tasks (removed 3 test tasks)
- US2 (Login): 6 tasks (removed 3 test tasks)
- US3 (Session Management): Already complete in Phase 4
- US3a (Token Refresh): Already complete in Phase 4
- US3b (Logout): Already complete in Phase 4
- US5 (Password Security): 6 tasks (removed 3 test tasks)
- Polish: 6 tasks (removed 1 test task)

**By User Story**:
- US1 (P1): 7 tasks - Registration functionality (tests removed)
- US2 (P1): 6 tasks - Login functionality (tests removed)
- US3 (P2): Complete - Session management (implemented in US2)
- US3a (P2): Complete - Token refresh (implemented in US2)
- US3b (P2): Complete - Logout (implemented in US2)
- US4: DEFERRED - Role-based access control (not implemented in this phase)
- US5 (P3): 6 tasks - Password reset (tests removed)

**MVP Scope** (Recommended First Iteration):
- Phase 1 (Setup): T001-T010
- Phase 2 (Foundational): T011-T023
- Phase 3 (US1): T024, T026-T032
- Phase 4 (US2): T034, T036-T037, T039-T041
- **Total MVP**: 36 tasks delivering registration + login (tests removed)

**Parallel Opportunities**: 15+ tasks can be executed in parallel (marked with [P])

**Independent Verification**: Each user story can be manually verified through API testing using tools like request.http, Postman, or curl.

---

## Next Steps

1. **Start with MVP**: Complete T001-T041 (Setup + Foundational + US1 + US2)
2. **Verify MVP**: Register user, login, access protected endpoint using request.http
3. **Complete**: Add US5 (T060-T068) for password reset
4. **Polish**: Complete Phase 9 (T070-T076) for production readiness
5. **Tests (Optional)**: Add test coverage later as needed

Each task is specific enough for immediate implementation. Follow hexagonal architecture principles throughout: **core domain first, then output adapters, then input adapters, finally wire in main**.
