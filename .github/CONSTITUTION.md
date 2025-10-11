# JalanRusak Backend - Project Constitution

## Project Overview

**JalanRusak** is a backend service enabling Indonesian citizens to report damaged roads. The service provides photo uploads, GPS tracking, location-based organization, and a verification workflow for tracking road repairs.

**Mission**: To create a reliable, scalable, and maintainable backend service that empowers citizens to report infrastructure issues and helps authorities track and verify repairs.

---

## Core Principles

### 1. **Architecture First**
- **Hexagonal Architecture (Ports & Adapters)** is the foundation
- Business logic remains independent of external frameworks and libraries
- Clear separation between core domain and adapters
- All external dependencies are abstracted through ports

### 2. **Clean Code & Maintainability**
- Code should be self-documenting with clear naming
- Follow Go best practices and idiomatic patterns
- Keep functions focused and single-purpose
- Prefer composition over inheritance
- Write code that's easy to test and refactor

### 3. **Security & Privacy**
- User authentication via JWT tokens
- Role-based authorization (regular users vs verificators)
- Secure handling of user data and location information
- Image uploads must be validated and sanitized
- Never expose sensitive configuration in code

### 4. **API Design Excellence**
- RESTful API conventions
- Consistent response formats
- Proper HTTP status codes
- Comprehensive error messages
- API versioning when needed

### 5. **Data Integrity**
- Location data must be accurate and validated
- GPS coordinates must be within valid ranges
- Report status transitions must be controlled
- Image associations must be properly maintained

---

## Architecture Decisions

### Hexagonal Architecture Layers

#### **Core Domain** (`core/`)
- **Domain Entities** (`domain/entities/`): Pure business objects with no dependencies
- **Domain Errors** (`domain/errors/`): Business-specific error types
- **Ports - Use Cases** (`ports/usecases/`): Interfaces defining how external world drives the application (use cases)
- **Ports - External** (`ports/external/`): Interfaces defining how core calls external services (repositories, storage)
- **Services** (`services/`): Business logic implementing use cases ports

**Rules:**
- Core must never depend on adapters
- All external dependencies accessed through ports
- Business logic isolated from frameworks

#### **Adapters** (`adapters/`)

**Input Adapters** (`adapters/in/`):
- **HTTP Handlers** (`http/handlers/`): Convert HTTP requests to use case calls
- **Middleware** (`http/middleware/`): Authentication, logging, CORS
- **Routes** (`http/routes/`): API endpoint definitions

**Output Adapters** (`adapters/out/`):
- **Repository** (`repository/postgres/`): Database access implementations
- **Storage** (`storage/filesystem/`): File storage implementations
- **Services**: Third-party integrations

**Rules:**
- Adapters depend on core ports, never vice versa
- Each adapter implements a port interface
- Adapters are swappable without affecting core

### Technology Choices

1. **Go 1.21+**: Type safety, performance, built-in concurrency
2. **Gin Framework**: Lightweight, fast HTTP routing with middleware support
3. **PostgreSQL**: ACID compliance, geospatial capabilities, reliability
4. **JWT Authentication**: Stateless, scalable authentication
5. **Local Filesystem Storage**: Default storage (swappable design for cloud storage)

---

## Development Guidelines

### Code Organization

```
Core Domain (core/)
├── Business rules and entities (no external dependencies)
├── Define what the application does
└── Independent, testable, reusable

Adapters (adapters/)
├── How the application interacts with the world
├── Implement core ports
└── Replaceable without affecting business logic
```

### Naming Conventions

- **Packages**: Lowercase, singular nouns (`user`, `report`, `storage`)
- **Interfaces** (Ports): Descriptive with context (`ReportRepository`, `ImageStorageService`)
- **Implementations**: Include adapter type (`PostgresReportRepository`, `FilesystemStorageService`)
- **Files**: Snake case (`report_service.go`, `user_handler.go`)
- **Constants**: PascalCase or ALL_CAPS for exported constants

### Error Handling

- Use domain-specific errors (`domain/errors/`)
- Wrap errors with context using `fmt.Errorf` with `%w`
- Return errors, don't panic (except for truly exceptional cases)
- HTTP handlers translate domain errors to appropriate status codes
- Log errors at the boundary (handlers), not in business logic

### Testing Strategy

1. **Unit Tests**: Test core business logic in isolation
2. **Integration Tests**: Test adapters with real dependencies
3. **API Tests**: Test HTTP endpoints end-to-end
4. **Test Coverage**: Aim for >80% coverage in core domain
5. **Mocking**: Use interfaces (ports) for easy mocking

### Database Guidelines

- **Migrations**: All schema changes via migration files
- **Transactions**: Use for multi-step operations
- **Indexing**: Index foreign keys and frequently queried fields
- **Naming**: Snake_case for tables and columns
- **Null Handling**: Prefer NOT NULL with defaults where sensible

### API Design Rules

1. **Endpoints**: RESTful resource-based URLs (`/api/v1/damaged-roads`, `/api/v1/users`)
2. **Methods**: Use appropriate HTTP verbs (GET, POST, PUT, PATCH, DELETE)
3. **Status Codes**:
   - 200: Success
   - 201: Created
   - 400: Bad Request (validation errors)
   - 401: Unauthorized
   - 403: Forbidden
   - 404: Not Found
   - 500: Internal Server Error
4. **Request/Response**: JSON format with consistent structure
5. **Pagination**: Use for list endpoints
6. **Filtering**: Support location-based queries

### Security Requirements

- **Authentication**: JWT tokens with expiration
- **Authorization**: Role-based access control
- **Input Validation**: Validate all user inputs
- **SQL Injection**: Use parameterized queries only
- **File Uploads**: Validate file types, size limits
- **Secrets Management**: Use environment variables, never commit secrets
- **HTTPS**: Enforce in production

### Configuration Management

- Use environment variables for configuration
- Provide sensible defaults for development
- Document all configuration options
- Separate config for different environments (dev, staging, prod)
- Never commit `.env` files

---

## Feature Development Workflow

### Adding New Features

1. **Design First**
   - Define domain entities if needed
   - Design input port (use case interface)
   - Design output ports (repository/service interfaces)

2. **Implement Core**
   - Create/update domain entities
   - Implement business logic in services
   - Write unit tests

3. **Implement Adapters**
   - Create HTTP handlers
   - Implement repositories
   - Add routes and middleware

4. **Integration**
   - Wire dependencies in `main.go`
   - Add database migrations if needed
   - Write integration tests

5. **Documentation**
   - Update API documentation
   - Add code comments for complex logic
   - Update README if needed

### Code Review Checklist

- [ ] Follows hexagonal architecture principles
- [ ] Core domain has no framework dependencies
- [ ] Proper error handling and propagation
- [ ] Input validation implemented
- [ ] Tests written and passing
- [ ] No hardcoded secrets or configuration
- [ ] Code is idiomatic Go
- [ ] API endpoints follow REST conventions
- [ ] Database migrations included if schema changed
- [ ] Documentation updated

---

## Constraints & Non-Negotiables

### Must Have
✅ Hexagonal architecture maintained
✅ JWT authentication for protected endpoints
✅ Role-based authorization (user/verificator)
✅ PostgreSQL as primary database
✅ Image upload with validation
✅ Location hierarchy (province → city → district → subdistrict)
✅ Report verification workflow
✅ RESTful API design

### Must Not Have
❌ Business logic in HTTP handlers
❌ Direct framework dependencies in core domain
❌ Hardcoded credentials or secrets
❌ Unvalidated user inputs
❌ SQL queries in business logic
❌ Global mutable state
❌ Panic in business logic

---

## Dependencies

### Core Dependencies
- **gin-gonic/gin**: HTTP web framework
- **lib/pq** or **pgx**: PostgreSQL driver
- **golang-jwt/jwt**: JWT token handling
- **testify**: Testing assertions and mocking

### Development Dependencies
- **golang-migrate**: Database migrations
- **air** (optional): Hot reload for development

---

## Quality Standards

### Performance
- API response time < 200ms for simple queries
- Support concurrent requests efficiently
- Optimize database queries (proper indexing)
- Image upload size limits enforced

### Reliability
- Graceful error handling
- Database connection pooling
- Transaction management for data consistency
- Proper logging for debugging

### Maintainability
- Clear separation of concerns
- Documented complex business logic
- Consistent code style (use `gofmt`, `golint`)
- Modular and testable code

---

## Documentation Requirements

### Code Documentation
- Public functions and types must have Go doc comments
- Complex algorithms need inline comments
- API handlers describe expected request/response

### API Documentation
- Endpoint descriptions
- Request/response examples
- Authentication requirements
- Error response formats

### README Maintenance
- Keep technology stack updated
- Document setup and run instructions
- List environment variables
- Provide example configuration

---

## Migration & Evolution

### Adding New Storage Backends
1. Define port interface in `core/ports/out/`
2. Implement adapter in `adapters/out/storage/`
3. Update dependency injection in `main.go`
4. Add configuration for storage selection

### Database Schema Changes
1. Create migration file in `migrations/`
2. Test migration up and down
3. Update repository implementations
4. Update domain entities if needed
5. Document breaking changes

### API Versioning
- Use URL versioning (`/api/v1/`, `/api/v2/`)
- Maintain backward compatibility when possible
- Document deprecations with sunset dates
- Provide migration guides for breaking changes

---

## Success Metrics

### Technical Health
- Test coverage > 80% in core domain
- Build time < 2 minutes
- All tests pass before merge
- No critical security vulnerabilities
- API response time within SLA

### Code Quality
- No circular dependencies
- Low coupling between modules
- High cohesion within modules
- Consistent code style
- Minimal code duplication

---

## Conclusion

This constitution ensures JalanRusak Backend remains **clean, maintainable, and scalable** while delivering value to citizens reporting damaged roads. All contributors must understand and follow these principles to maintain the integrity of the codebase.

**When in doubt:**
1. Keep business logic in the core
2. Depend on abstractions (ports), not concretions
3. Write tests first or immediately after
4. Favor simplicity over cleverness
5. Document decisions and rationale

---

*Last Updated: October 11, 2025*
*Version: 1.0*
