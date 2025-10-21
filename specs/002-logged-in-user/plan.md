# Implementation Plan: Logged-In Users Report Damaged Roads

**Branch**: `002-logged-in-user` | **Date**: 2025-10-20 | **Spec**: [specs/002-logged-in-user/spec.md](specs/002-logged-in-user/spec.md)
**Input**: Feature specification from `/specs/002-logged-in-user/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

Enable authenticated users to submit damaged road reports with title, Indonesian administrative area codes (Kemendagri), coordinate paths, photo URLs, and optional descriptions. Implementation follows hexagonal architecture with Go 1.21+, PostgreSQL + PostGIS for geospatial validation, JWT authentication, and comprehensive input validation at all boundaries.

## Technical Context

<!--
  ACTION REQUIRED: Replace the content in this section with the technical details
  for the project. The structure here is presented in advisory capacity to guide
  the iteration process.
-->

**Language/Version**: Go 1.21+ (per constitution)
**Primary Dependencies**: Gin Framework, golang-jwt/jwt, testify, golang-migrate, orb (geospatial), sqlx (database)
**Storage**: PostgreSQL with PostGIS extension for geospatial validation
**Testing**: Testify framework (optional per constitution)
**Target Platform**: Linux server
**Project Type**: Backend API service
**Performance Goals**: <200ms p95 response time, support 1000+ concurrent users
**Constraints**: Hexagonal architecture mandatory, JWT authentication, role-based authorization, input validation at all boundaries
**Scale/Scope**: Initial target 10k users, 1k+ reports/month

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Required Principles (MUST PASS)

✅ **Hexagonal Architecture**: Core domain must be independent of external frameworks
✅ **Go 1.21+ Technology Stack**: Mandated by constitution
✅ **JWT Authentication**: Required for protected endpoints
✅ **PostgreSQL**: Primary database with PostGIS for geospatial
✅ **Input Validation**: Must validate at all boundaries
✅ **RESTful API Design**: Resource-based URLs with proper HTTP verbs
✅ **Swagger Annotations**: Every endpoint must have up-to-date documentation
✅ **Parameterized Queries**: SQL injection prevention mandatory

### Quality Gates

- ✅ Architecture follows hexagonal pattern
- ✅ Authentication and authorization implemented
- ✅ Database migrations included
- ⚠️ Testing deferred (constitution allows deferred testing)
- ⚠️ Geospatial validation complexity justified (Indonesian boundaries)

## Project Structure

### Documentation (this feature)

```
specs/[###-feature]/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```
backend/ (Go hexagonal architecture)
├── cmd/server/
│   └── main.go
├── core/
│   ├── domain/entities/
│   │   ├── damaged_road.go
│   │   ├── user.go
│   │   └── value_objects.go
│   ├── domain/errors/
│   │   └── errors.go
│   ├── ports/usecases/
│   │   ├── report_service.go
│   │   └── user_service.go
│   ├── ports/external/
│   │   └── repository.go
│   └── services/
│       ├── report_service_impl.go
│       └── user_service_impl.go
├── adapters/
│   ├── in/http/
│   │   ├── handlers/
│   │   │   ├── report_handler.go
│   │   │   └── auth_handler.go
│   │   ├── middleware/
│   │   │   ├── auth.go
│   │   │   └── validation.go
│   │   └── routes/
│   │       └── routes.go
│   └── out/repository/postgres/
│       ├── damaged_road_repo.go
│       └── user_repo.go
├── migrations/
│   └── 001_create_damaged_roads.up.sql
├── docs/
│   └── swagger.yaml
└── tests/ (optional - deferred per constitution)
    ├── integration/
    └── unit/
```

**Structure Decision**: Hexagonal backend architecture following constitution requirements. Core domain contains business logic, adapters handle external interfaces, PostgreSQL with PostGIS for geospatial data.

## Complexity Tracking

*Fill ONLY if Constitution Check has violations that must be justified*

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| [e.g., 4th project] | [current need] | [why 3 projects insufficient] |
| [e.g., Repository pattern] | [specific problem] | [why direct DB access insufficient] |
