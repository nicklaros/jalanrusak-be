<!--
================================================================================
SYNC IMPACT REPORT - Constitution Update
================================================================================

Version Change: 1.0.0 → 1.1.0

Amendment Summary:
- Modified Principle V: Test-Driven Quality → changed from mandatory to deferred
- Testing is now optional for initial implementation, can be added later
- Focus shifted to working implementation first, test coverage as enhancement

Modified Principles:
- Principle V: "Test-Driven Quality" → "Quality Through Testing (Deferred)"
  * Changed from mandatory requirement to optional/deferred approach
  * Tests no longer block feature implementation
  * Coverage targets removed (was >80% core coverage)
  * Testing framework (testify) retained for future use

Sections Modified:
✅ Core Principles - Principle V rewritten to reflect deferred testing policy
✅ Code Review Checklist - Removed mandatory test requirements
✅ Non-Negotiable Constraints - Removed test-related MUST requirements

Template Consistency Updates Required:
⚠ plan-template.md - Constitution Check section needs update (Principle V changed)
⚠ spec-template.md - User story testing sections can be optional now
⚠ tasks-template.md - Test task categorization should reflect deferred status
✅ .github/copilot-instructions.md - Already updated with Testing Policy

Deferred Items: None

Version Bump Rationale:
- MINOR: 1.0.0 → 1.1.0
- Material change to existing principle (testing methodology)
- Not backward incompatible (existing tests still valid, just not required)
- Expands flexibility without removing capabilities

Suggested Commit Message:
docs: amend constitution to v1.1.0 (defer testing requirements for initial implementation)

Date: October 12, 2025

================================================================================
-->

# JalanRusak Backend Constitution

**Project**: JalanRusak Backend Service  
**Mission**: Enable Indonesian citizens to report damaged roads through a reliable, scalable, and maintainable backend service that empowers citizens and helps authorities track infrastructure repairs.

## Core Principles

### I. Hexagonal Architecture (NON-NEGOTIABLE)

**Ports & Adapters pattern** is the foundation of this codebase. Business logic MUST remain independent of external frameworks and libraries.

- Core domain (`core/`) has **zero dependencies** on adapters or frameworks
- All external interactions abstracted through **port interfaces**
- Adapters (`adapters/in/`, `adapters/out/`) implement ports, never vice versa
- Clear separation: Core defines **what**, Adapters define **how**
- Dependency direction: Adapters → Core (never Core → Adapters)

**Rationale**: Enables independent testing, framework swapping, and long-term maintainability. Protects business logic from external change.

### II. Clean Code & Maintainability

Code MUST be **idiomatic Go** and self-documenting. Favor clarity over cleverness.

- Follow Go best practices and community conventions
- Functions are focused and single-purpose (max 50 lines as guideline)
- Clear naming that reveals intent (no `data`, `info`, `manager` suffixes)
- Prefer composition over inheritance
- Package names are lowercase, singular nouns (`user`, `report`, `storage`)
- Public functions and types have godoc comments
- Complex logic has inline comments explaining the **why**, not the **what**

**Rationale**: Reduces cognitive load, accelerates onboarding, and minimizes defects through clarity.

### III. Security & Privacy First

User data and system access MUST be protected at every layer.

- **Authentication**: JWT tokens with appropriate expiration (short-lived access tokens + refresh tokens)
- **Authorization**: Role-based access control (regular users vs verificators)
- **Input Validation**: All user inputs validated before processing
- **SQL Injection Prevention**: Parameterized queries only, no string concatenation
- **File Upload Security**: Validate file types, enforce size limits, sanitize filenames
- **Secrets Management**: Environment variables only, never commit credentials
- **HTTPS**: Required in production, no exceptions
- **Password Storage**: Bcrypt or Argon2 hashing, never plain text

**Rationale**: Protects citizen privacy and maintains trust in the reporting system.

### IV. API Design Excellence

RESTful API conventions MUST be followed consistently.

- Resource-based URLs (`/api/v1/damaged-roads`, `/api/v1/users`)
- Appropriate HTTP verbs (GET, POST, PUT, PATCH, DELETE)
- Standard status codes: 200 (success), 201 (created), 400 (validation), 401 (auth), 403 (forbidden), 404 (not found), 500 (server error)
- JSON request/response with consistent structure
- Pagination for list endpoints
- Comprehensive error messages with actionable details
- API versioning when breaking changes required (`/api/v1/`, `/api/v2/`)

**Rationale**: Predictable API behavior improves developer experience and reduces integration errors.

### V. Quality Through Testing (Deferred)

Testing provides **confidence and documentation** but is deferred for initial implementation. Tests can be added later as the project matures.

- **Initial Focus**: Working implementation first, tests as enhancement
- **Unit Tests**: Optional for core business logic (can target coverage later)
- **Integration Tests**: Optional for adapters with real dependencies
- **API Tests**: Optional for HTTP endpoints end-to-end
- **Testing Infrastructure**: Testify framework available when tests are added
- **Future Strategy**: Tests recommended before production deployment
- Use interfaces (ports) for easy mocking when tests are written

**Rationale**: Enables rapid feature development without test coverage blocking progress. Tests provide value for stability and regression prevention but can be added incrementally as needed.

**Note**: While testing is deferred, code MUST still be designed to be testable (use dependency injection, favor composition, avoid global state).

## Architecture Standards

### Hexagonal Layers

**Core Domain** (`core/`):
- `domain/entities/`: Pure business objects with no external dependencies
- `domain/errors/`: Business-specific error types
- `ports/usecases/`: Interfaces defining how external world drives the application
- `ports/external/`: Interfaces defining how core calls external services
- `services/`: Business logic implementing use case ports

**Adapters** (`adapters/`):
- `in/http/handlers/`: Convert HTTP requests to use case calls
- `in/http/middleware/`: Authentication, logging, CORS
- `in/http/routes/`: API endpoint definitions
- `out/repository/postgres/`: Database access implementations
- `out/storage/filesystem/`: File storage implementations
- `out/services/`: Third-party integrations

### Technology Choices

These technologies are **mandated** for consistency:

1. **Go 1.21+**: Type safety, performance, built-in concurrency
2. **Gin Framework**: HTTP routing and middleware
3. **PostgreSQL**: ACID compliance, geospatial capabilities, reliability
4. **JWT**: Stateless authentication (golang-jwt/jwt)
5. **Testify**: Testing assertions and mocking
6. **golang-migrate**: Database migrations

### Dependency Rules

- Core MUST NOT import adapter packages
- Core MAY import standard library and domain-focused libraries (e.g., validation)
- Adapters MUST import core port interfaces
- Configuration injected via `main.go`, not global variables
- Database drivers used only in repository adapters

## Development Workflow

### Feature Development Process

1. **Design First**:
   - Define or update domain entities
   - Design use case port interfaces
   - Design external port interfaces (repositories, services)

2. **Implement Core**:
   - Create/update domain entities in `core/domain/entities/`
   - Implement business logic in `core/services/`
   - Unit tests optional (can be added later)

3. **Implement Adapters**:
   - Create HTTP handlers in `adapters/in/http/handlers/`
   - Implement repositories in `adapters/out/repository/`
   - Add routes and middleware

4. **Integration**:
   - Wire dependencies in `cmd/server/main.go`
   - Create database migrations in `migrations/`
   - Integration and API tests optional (can be added later)

5. **Documentation**:
   - Update API documentation
   - Add godoc comments for public interfaces
   - Update README if needed

**Testing Note**: While tests are deferred, code should still follow testability principles (dependency injection, interface-based design, avoid global state).

### Code Review Checklist

Every pull request MUST verify:

- [ ] Hexagonal architecture principles maintained
- [ ] Core domain has no framework dependencies
- [ ] Proper error handling with wrapped context
- [ ] Input validation implemented at boundaries
- [ ] Code designed to be testable (dependency injection, interfaces)
- [ ] No hardcoded secrets or configuration
- [ ] Code follows Go conventions (gofmt, golint)
- [ ] API endpoints follow RESTful conventions
- [ ] Database migrations included for schema changes
- [ ] Godoc comments on public functions/types
- [ ] No SQL injection vulnerabilities (parameterized queries only)

**Note**: Tests are optional but recommended. If tests are written, they must pass before merge.

### Non-Negotiable Constraints

**MUST Have**:
- ✅ Hexagonal architecture maintained
- ✅ JWT authentication for protected endpoints
- ✅ Role-based authorization structure
- ✅ PostgreSQL as primary database
- ✅ Input validation at all boundaries
- ✅ Parameterized database queries
- ✅ Environment-based configuration
- ✅ RESTful API design

**MUST NOT Have**:
- ❌ Business logic in HTTP handlers
- ❌ Direct framework dependencies in core domain
- ❌ Hardcoded credentials or secrets
- ❌ Unvalidated user inputs reaching business logic
- ❌ SQL queries constructed with string concatenation
- ❌ Global mutable state
- ❌ Panic in business logic (return errors)

## Governance

This constitution **supersedes all other development practices**. All contributors must understand and follow these principles.

### Amendment Process

1. Proposed changes documented with rationale
2. Impact assessment on existing codebase
3. Team review and approval required
4. Version increment per semantic versioning
5. Migration plan for breaking changes

### Compliance Requirements

- All pull requests reviewed against this constitution
- Constitution violations must be explicitly justified
- Regular architecture reviews to ensure compliance
- Use `.specify/` workflow for feature development
- Complexity increases require strong justification

### Version Information

**Version**: 1.1.0 | **Ratified**: 2025-10-12 | **Last Amended**: 2025-10-12

---

**When in doubt**:
1. Keep business logic in the core domain
2. Depend on abstractions (ports), not concretions
3. Design for testability even if tests are deferred
4. Favor simplicity over cleverness
5. Document decisions and rationale, not obvious code