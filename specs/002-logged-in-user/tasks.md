# Implementation Tasks: Logged-In Users Report Damaged Roads

**Feature**: 002-logged-in-user
**Branch**: 002-logged-in-user
**Total Tasks**: 56
**Date**: 2025-10-20

## Implementation Strategy

**MVP Approach**: User Story 1 (P1) provides complete end-to-end functionality for basic report submission and serves as the Minimum Viable Product. User Stories 2 and 3 enhance the core feature with more sophisticated capabilities.

**Parallel Execution**: Tasks marked with `[P]` can be executed in parallel. Tasks without `[P]` must be executed sequentially.

**Testing Strategy**: Tests are optional per constitution guidance and not included in task list.

## User Story Summary

| Story | Priority | Tasks | Independent Test | Focus |
|-------|----------|-------|------------------|-------|
| US1 | P1 | 14 tasks | Submit report with required fields | Core report submission functionality |
| US2 | P2 | 7 tasks | Coordinate path accuracy validation | Geospatial validation and path handling |
| US3 | P3 | 7 tasks | Photo attachment and description | Evidence collection and validation |

---

## Phase 1: Project Setup (Shared Infrastructure)

**Goal**: Establish the foundation for all user stories

| ID | Task | File | Description |
|----|------|------|-------------|
| T001 | Initialize Go module | `go.mod` | Create go.mod with Go 1.21+ and required dependencies (gin, golang-jwt, testify, golang-migrate, orb, sqlx) |
| T002 | Setup project structure | Directory structure | Create hexagonal architecture directories (cmd/, core/, adapters/, migrations/, docs/) |
| T003 | Configure environment | `config/config.go` | Create configuration structure for database, JWT, server settings |
| T004 | Setup database connection | `adapters/out/repository/postgres/connection.go` | Create PostgreSQL connection pool with PostGIS support |
| T005 | Create base error types | `core/domain/errors/errors.go` | Define domain-specific error types for validation and business logic |
| T006 | Setup logging infrastructure | `pkg/logger/logger.go` | Create structured logging with context support |

---

## Phase 2: Foundational (Blocking Prerequisites)

**Goal**: Implement core services that all user stories depend on

| ID | Task | File | Description |
|----|------|------|-------------|
| T007 | Create JWT middleware | `adapters/in/http/middleware/auth.go` [P] | Implement JWT authentication middleware for protected endpoints |
| T008 | Create input validation middleware | `adapters/in/http/middleware/validation.go` [P] | Implement request validation middleware |
| T009 | Setup database migrations | `migrations/005_create_damaged_roads_table.up.sql` | Create damaged_roads and damaged_road_photos tables with PostGIS extension |
| T010 | Create repository interfaces | `core/ports/external/repository.go` | Define repository interfaces for data access |
| T011 | Create base HTTP server setup | `cmd/server/main.go` | Setup Gin server with basic middleware and route structure |
| T012 | Create shared value objects | `core/domain/entities/value_objects.go` | Implement Title, SubDistrictCode, Point, and Geometry value objects |

---

## Phase 3: User Story 1 - Basic Report Submission (P1)

**Story Goal**: Enable logged-in users to submit damaged road reports with required fields
**Independent Test**: Authenticated user can submit report with title, subdistrict code, at least one coordinate, and at least one photo URL → receive confirmation

### Story 1 Core Implementation
| ID | Task | File | Description | Story |
|----|------|------|-------------|-------|
| T014 | Create DamagedRoad entity | `core/domain/entities/damaged_road.go` | Implement core DamagedRoad entity with basic fields and validation | US1 |
| T015 | Create User entity | `core/domain/entities/user.go` [P] | Implement User entity for authentication context | US1 |
| T016 | Create ReportService interface | `core/ports/usecases/report_service.go` | Define use case interface for report operations | US1 |
| T017 | Implement ReportService | `core/services/report_service_impl.go` | Implement business logic for creating reports with basic validation | US1 |
| T018 | Create DamagedRoad repository | `adapters/out/repository/postgres/damaged_road_repo.go` | Implement PostgreSQL repository for report CRUD operations | US1 |
| T019 | Create ReportHandler | `adapters/in/http/handlers/report_handler.go` | Implement HTTP handler for report creation endpoint | US1 |
| T020 | Setup report routes | `adapters/in/http/routes/routes.go` | Configure /damaged-roads POST route with auth middleware | US1 |
| T021 | Add Swagger annotations | `adapters/in/http/handlers/report_handler.go` | Add comprehensive Swagger documentation for create report endpoint | US1 |
| T022 | Create basic validation service | `core/services/validation_service.go` [P] | Implement field validation (title length, required fields) | US1 |
| T023 | Add error handling | `adapters/in/http/handlers/report_handler.go` | Implement comprehensive error responses for validation failures | US1 |
| T024 | Performance optimization | `adapters/out/repository/postgres/damaged_road_repo.go` | Add database indexes for report queries | US1 |
| T025 | Documentation update | `docs/swagger.yaml` | Generate updated Swagger documentation | US1 |
| T026 | Manual testing guide | `docs/testing-guide.md` | Create manual testing procedures for US1 | US1 |
| T027 | Code review checklist | `docs/code-review-checklist.md` | Create code review checklist for story completion | US1 |

**✓ Phase 3 Checkpoint**: User Story 1 complete - MVP ready for demonstration

---

## Phase 4: User Story 2 - Coordinate Path Accuracy (P2)

**Story Goal**: Enable users to trace accurate road segments with coordinate validation
**Independent Test**: User can provide coordinate pairs forming damaged path → stored with ordered path mapping

### Story 2 Core Implementation
| ID | Task | File | Description | Story |
|----|------|------|-------------|-------|
| T029 | Create Geometry service | `core/services/geometry_service.go` | Implement coordinate validation and PostGIS LineString conversion | US2 |
| T030 | Add Indonesian boundaries data | `migrations/002_load_indonesian_boundaries.up.sql` | Load Indonesian national boundary data for validation | US2 |
| T031 | Create LocationValidator | `core/services/location_validator.go` | Implement coordinate boundary validation service | US2 |
| T032 | Update DamagedRoad entity | `core/domain/entities/damaged_road.go` | Enhance entity to support multiple coordinate points as path | US2 |
| T033 | Update ReportService | `core/services/report_service_impl.go` | Add path validation logic to report creation | US2 |
| T034 | Update repository for paths | `adapters/out/repository/postgres/damaged_road_repo.go` | Add PostGIS path storage and spatial queries | US2 |
| T035 | Add validation endpoints | `adapters/in/http/handlers/validation_handler.go` [P] | Create /validation/coordinates and /validation/subdistrict-codes endpoints | US2 |

**✓ Phase 4 Checkpoint**: User Story 2 complete - enhanced location accuracy

---

## Phase 5: User Story 3 - Photo Evidence & Description (P3)

**Story Goal**: Enable users to attach photos and descriptions for report credibility
**Independent Test**: User can add photo URLs and optional description → stored and retrievable with report metadata

### Story 3 Core Implementation
| ID | Task | File | Description | Story |
|----|------|------|-------------|-------|
| T036 | Create PhotoValidator service | `core/services/photo_validator.go` | Implement photo URL validation with SSRF protection (HTTP/HTTPS only, 5s timeout, no localhost/private IPs) | US3 |
| T037 | ~~Create photo attachments table~~ | ~~migrations/003_create_photo_attachments.up.sql~~ | **REMOVED**: Photo table already created in migration 005 | ~~US3~~ |
| T038-T039 | Add PhotoURLs and Description fields | `core/domain/entities/damaged_road.go` | Add PhotoURLs field with validation (1-10 photos) and optional Description field (max 500 chars) - **MERGED TASK** | US3 |
| T040 | Update ReportService | `core/services/report_service_impl.go` | Add photo validation and description handling | US3 |
| T041 | Update repository for photos | `adapters/out/repository/postgres/damaged_road_repo.go` | Add photo attachment storage and retrieval using damaged_road_photos table | US3 |
| T042 | Add photo validation endpoint | `adapters/in/http/handlers/validation_handler.go` | Create /validation/photo-urls endpoint with SSRF protection | US3 |

**✓ Phase 5 Checkpoint**: User Story 3 complete - full evidence collection

---

## Phase 6: Polish & Cross-Cutting Concerns

**Goal**: Finalize implementation with additional features and optimizations

| ID | Task | File | Description |
|----|------|------|-------------|
| T043 | Implement report listing | `adapters/in/http/handlers/report_handler.go` | Add GET /damaged-roads endpoint for user's reports |
| T044 | Add pagination support | `adapters/out/repository/postgres/damaged_road_repo.go` | Implement pagination for list endpoints |
| T045 | Create report detail endpoint | `adapters/in/http/handlers/report_handler.go` | Add GET /damaged-roads/{id} endpoint |
| T046 | Add status update endpoint | `adapters/in/http/handlers/report_handler.go` | Add PATCH /damaged-roads/{id}/status for admins |
| T047 | Implement rate limiting | `adapters/in/http/middleware/rate_limit.go` | Add rate limiting for validation endpoints |
| T048 | Add caching layer | `pkg/cache/cache.go` | Implement caching for validation results |
| T049 | Create health check endpoint | `adapters/in/http/handlers/health_handler.go` | Add /health endpoint for monitoring |
| T050 | Performance monitoring | `pkg/metrics/metrics.go` | Add performance metrics collection |
| T051 | Security hardening | `adapters/in/http/middleware/security.go` | Add security headers and CORS configuration |
| T052 | Update all Swagger docs | `docs/swagger.yaml` | Ensure all endpoints have complete Swagger documentation |
| T053 | Deployment configuration | `docker/Dockerfile` | Create Docker configuration for deployment |
| T054 | Environment documentation | `docs/deployment.md` | Document deployment process and environment setup |
| T055 | User documentation | `docs/api-usage.md` | Create API usage documentation for frontend developers |
| T056 | Final code review | - | Comprehensive code review against constitution requirements |

---

## Dependencies Graph

```
Phase 1 (Setup) → Phase 2 (Foundational) → Phase 3 (US1 MVP)
                                        ↓
                                   Phase 4 (US2) → Phase 5 (US3) → Phase 6 (Polish)
```

**Critical Path**: T001-T012 → T014-T028 → T030-T037 → T039-T046 → T047-T063

**Parallel Execution Groups**:
- **Group 1**: T001-T006 (Setup infrastructure)
- **Group 2**: T007, T008, T012 (Middleware and value objects)
- **Group 3**: T022 (Shared services)
- **Group 4**: T015, T036 (Parallel components)
- **Group 5**: T043-T056 (Final features)

---

## Parallel Execution Examples

### User Story 1 (Phase 3) - 4 Parallel Streams

**Stream A (Core Logic)**: T014 → T016 → T017 → T024
**Stream B (Data Layer)**: T015 → T018 → T025
**Stream C (API Layer)**: T019 → T020 → T021 → T026
**Stream D (Quality)**: T025 → T026 → T027

### User Story 2 (Phase 4) - 3 Parallel Streams

**Stream A (Geospatial)**: T030 → T032 → T033
**Stream B (Database)**: T031 → T035 → T037
**Stream C (Integration)**: T031 → T034 → T035

### User Story 3 (Phase 5) - 3 Parallel Streams

**Stream A (Photo Validation)**: T036 → T038 → T040
**Stream B (Storage)**: T037 → T039 → T041
**Stream C (API Integration)**: T037 → T042

---

## Independent Test Criteria

### User Story 1 (P1)
1. **Setup**: Create test user and authenticate
2. **Action**: POST /damaged-roads with {title, subdistrict_code, path_points[1], photo_urls[1]}
3. **Expected**: 201 Created response with report ID and confirmation
4. **Validation**: Report persisted in database with all required fields

### User Story 2 (P2)
1. **Setup**: Authenticated user with valid administrative code
2. **Action**: POST /damaged-roads with path_points containing multiple coordinates
3. **Expected**: 201 Created with ordered path stored as PostGIS LineString
4. **Validation**: Path accuracy and coordinate boundary validation working

### User Story 3 (P3)
1. **Setup**: Authenticated user with valid photo URLs
2. **Action**: POST /damaged-roads with photo_urls[] and optional description
3. **Expected**: 201 Created with photo metadata stored and validated
4. **Validation**: Photo accessibility validation and description storage

---

## MVP Scope (User Story 1 Only)

**Minimum Viable Product**: Phase 3 completion (Tasks T001-T028)

**MVP Features**:
- ✅ User authentication via JWT
- ✅ Basic damaged road report creation
- ✅ Required field validation (title, admin code, 1 coordinate, 1 photo)
- ✅ Database persistence
- ✅ Error handling and validation responses
- ✅ Basic API documentation

**MVP Exclusions** (deferred to US2/US3):
- Multi-coordinate path support
- Geospatial boundary validation
- Photo URL accessibility validation
- Detailed descriptions
- Report listing and management

**MVP Timeline**: Approximately 2-3 weeks with 2-3 developers

---

## Constitution Compliance Checklist

- [x] **Hexagonal Architecture**: Core domain independent of external frameworks
- [x] **Go 1.21+ Technology Stack**: All dependencies use Go 1.21+
- [x] **JWT Authentication**: Implemented in T007
- [x] **PostgreSQL + PostGIS**: Database setup in T009, spatial features in T035
- [x] **Input Validation**: Comprehensive validation in T022, T030, T039
- [x] **RESTful API Design**: Resource-based URLs in T020, T047, T049, T050
- [x] **Swagger Annotations**: Added in T021, updated throughout implementation
- [x] **Parameterized Queries**: Repository implementations use sqlx with parameterized queries
- [x] **Testing Deferred**: Test tasks marked as optional per constitution

---

## Risk Mitigation

### Technical Risks
1. **PostGIS Integration**: Mitigated by T031 with dedicated migration
2. **Photo Validation Complexity**: Mitigated by T039 with SSRF protection
3. **Performance at Scale**: Mitigated by T025, T052, T054 with indexing and caching

### Schedule Risks
1. **Parallel Execution**: Maximized with [P] tasks across all phases
2. **Dependencies**: Clear critical path identified for bottleneck management
3. **MVP Focus**: US1 can be delivered independently as working product

---

## Success Metrics

### User Story 1 (MVP)
- ✅ Authenticated users can submit reports
- ✅ All required fields validated
- ✅ Reports persisted with confirmation
- ✅ <200ms response time for report creation

### Full Implementation
- ✅ Multi-coordinate path support
- ✅ Indonesian boundary validation
- ✅ Photo accessibility validation
- ✅ Complete API documentation
- ✅ Performance targets met (1000+ concurrent users)

---

**Next Actions**:
1. Begin with Phase 1 (T001-T006) for project setup
2. Proceed to Phase 2 (T007-T012) for foundational services
3. Focus on Phase 3 (T014-T028) for MVP delivery
4. Continue with User Stories 2 and 3 based on user feedback and priorities