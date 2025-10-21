# Quickstart Guide: Logged-In Users Report Damaged Roads

**Purpose**: Development setup and implementation guide for the damaged road reporting feature
**Branch**: 002-logged-in-user
**Updated**: 2025-10-19

## Overview

This feature enables logged-in users to submit damaged road reports with photos, location data, and administrative information. The system validates Indonesian boundaries, administrative codes, and enforces photo limits (1-10 photos per report).

## Architecture Summary

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   HTTP API      │    │   Domain Logic   │    │   Data Layer    │
│   (Gin)         │◄──►│   (Hexagonal)    │◄──►│   (PostgreSQL)  │
│                 │    │                  │    │                 │
│ - /damaged-roads│    │ - DamagedRoad    │    │ - damaged_roads │
│ - /validation   │    │ - Validation     │    │ - boundaries    │
│ - /auth         │    │ - Business Rules │    │ - users         │
└─────────────────┘    └──────────────────┘    └─────────────────┘
```

## Prerequisites

### Development Environment
- **Go 1.21+** installed
- **PostgreSQL 14+** with PostGIS extension
- **Git** for version control
- **Make** for build automation

### Required Tools
```bash
# Install Go dependencies
go mod download

# Install migration tool
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Install Swagger generator (optional)
go install github.com/swaggo/swag/cmd/swag@latest
```

## Database Setup

### 1. PostgreSQL with PostGIS
```sql
-- Create database
CREATE DATABASE jalanrusak;

-- Connect to database and enable PostGIS
\c jalanrusak;
CREATE EXTENSION IF NOT EXISTS postgis;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
```

### 2. Run Migrations
```bash
# Create migration files (if not exists)
migrate create -ext sql -dir migrations create_damaged_roads_table
migrate create -ext sql -dir migrations create_administrative_boundaries_table

# Run migrations
migrate -path migrations -database "postgres://user:pass@localhost/jalanrusak?sslmode=disable" up
```

### 3. Load Administrative Data
Download Indonesian administrative boundaries from data.go.id and load using:
```bash
# Use provided scripts (to be created)
./scripts/load-admin-boundaries.sh path/to/boundary-data.shp
```

## Project Structure

```
jalanrusak-be/
├── cmd/
│   └── server/
│       └── main.go                 # Application entry point
├── internal/
│   ├── core/
│   │   ├── domain/
│   │   │   ├── entities/           # Business entities
│   │   │   │   ├── damaged_road.go
│   │   │   │   ├── point.go
│   │   │   │   └── admin_boundary.go
│   │   │   ├── errors/             # Domain errors
│   │   │   └── value_objects/      # Value objects
│   │   ├── ports/
│   │   │   ├── usecases/           # Application ports
│   │   │   └── external/           # External service ports
│   │   └── services/               # Business logic
│   │       ├── damaged_road_service.go
│   │       ├── location_validator.go
│   │       └── photo_validator.go
│   └── adapters/
│       ├── in/
│       │   └── http/
│       │       ├── handlers/       # HTTP handlers
│       │       ├── middleware/     # Auth, logging, CORS
│       │       └── routes/         # Route definitions
│       └── out/
│           ├── repository/
│           │   └── postgres/       # Database adapters
│           └── services/           # External service adapters
├── migrations/                     # Database migrations
├── config/                         # Configuration
├── docs/                          # Generated Swagger docs
└── scripts/                       # Utility scripts
```

## Implementation Steps

### Phase 1: Core Domain (1-2 days)

1. **Create Domain Entities**
```go
// internal/core/domain/entities/damaged_road.go
type DamagedRoad struct {
    ID              string
    Title           string
    SubDistrictCode string
    PathPoints      []Point
    Description     *string
    PhotoURLs       []string
    AuthorID        string
    Status          Status
    CreatedAt       time.Time
    UpdatedAt       time.Time
}
```

2. **Define Ports and Interfaces**
```go
// internal/core/ports/usecases/damaged_road.go
type DamagedRoadUseCase interface {
    CreateReport(ctx context.Context, req *CreateReportRequest) (*DamagedRoad, error)
    GetReport(ctx context.Context, id string) (*DamagedRoad, error)
    ListReports(ctx context.Context, filters *ListFilters) ([]*DamagedRoad, error)
    UpdateStatus(ctx context.Context, id string, status Status) error
}
```

3. **Implement Business Logic**
```go
// internal/core/services/damaged_road_service.go
type DamagedRoadService struct {
    repo DamagedRoadRepository
    locationValidator LocationValidator
    photoValidator PhotoValidator
}
```

### Phase 2: HTTP Layer (2-3 days)

1. **Create HTTP Handlers**
```go
// internal/adapters/in/http/handlers/damaged_road.go
type DamagedRoadHandler struct {
    useCase usecases.DamagedRoadUseCase
}

// @Summary Create damaged road report
// @Description Create a new damaged road report
// @Tags damaged-roads
// @Accept json
// @Produce json
// @Param request body CreateDamagedRoadRequest true "Report data"
// @Success 201 {object} DamagedRoadResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /damaged-roads [post]
func (h *DamagedRoadHandler) CreateReport(c *gin.Context) {
    // Implementation
}
```

2. **Setup Middleware**
```go
// internal/adapters/in/http/middleware/auth.go
func JWTAuthMiddleware(jwtSecret string) gin.HandlerFunc {
    return func(c *gin.Context) {
        // JWT validation logic
    }
}
```

3. **Configure Routes**
```go
// internal/adapters/in/http/routes/routes.go
func SetupRoutes(router *gin.Engine, handlers *Handlers) {
    api := router.Group("/api/v1")
    api.Use(middleware.JWTAuthMiddleware(config.JWTSecret))

    damagedRoads := api.Group("/damaged-roads")
    {
        damagedRoads.POST("", handlers.DamagedRoad.CreateReport)
        damagedRoads.GET("", handlers.DamagedRoad.ListReports)
        damagedRoads.GET("/:id", handlers.DamagedRoad.GetReport)
        damagedRoads.PATCH("/:id/status", handlers.DamagedRoad.UpdateStatus)
    }

    validation := api.Group("/validation")
    {
        validation.POST("/photo-urls", handlers.Validation.ValidatePhotoURLs)
        validation.GET("/subdistrict-codes/:code", handlers.Validation.ValidateSubDistrictCode)
        validation.POST("/coordinates", handlers.Validation.ValidateCoordinates)
    }
}
```

### Phase 3: Data Layer (2-3 days)

1. **Create Repository Implementation**
```go
// internal/adapters/out/repository/postgres/damaged_road.go
type DamagedRoadRepository struct {
    db *sqlx.DB
}

func (r *DamagedRoadRepository) Create(ctx context.Context, road *DamagedRoad) error {
    query := `
        INSERT INTO damaged_roads (id, title, kemendagri_code, description, author_id, status)
        VALUES ($1, $2, $3, $4, $5, $6)
        RETURNING created_at, updated_at`

    // Implementation with proper error handling
}
```

2. **Implement Validation Services**
```go
// internal/core/services/location_validator.go
type LocationValidator struct {
    db *sqlx.DB
}

func (v *LocationValidator) ValidateCoordinates(lat, lng float64) error {
    query := `
        SELECT COUNT(*) FROM administrative_boundaries
        WHERE ST_Within(ST_MakePoint($1, $2), geometry)`

    var count int
    err := v.db.Get(&count, query, lng, lat)
    if err != nil {
        return err
    }

    if count == 0 {
        return errors.New("coordinates outside Indonesian boundaries")
    }

    return nil
}
```

## Configuration

### Environment Variables
```bash
# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=jalanrusak
DB_PASSWORD=your_password
DB_NAME=jalanrusak

# JWT
JWT_SECRET=your_jwt_secret_key
JWT_EXPIRATION=24h

# Server
SERVER_PORT=8080
SERVER_HOST=localhost

# Photo Validation
PHOTO_VALIDATION_TIMEOUT=30s
PHOTO_VALIDATION_MAX_CONCURRENCY=5
PHOTO_VALIDATION_CACHE_TTL=1h

# Rate Limiting
RATE_LIMIT_REQUESTS_PER_MINUTE=100
```

### Config Structure
```go
// config/config.go
type Config struct {
    Database DatabaseConfig
    JWT      JWTConfig
    Server   ServerConfig
    PhotoValidation PhotoValidationConfig
    RateLimit RateLimitConfig
}
```

## Testing (Optional)

### Unit Tests
```go
// internal/core/services/damaged_road_service_test.go
func TestDamagedRoadService_CreateReport(t *testing.T) {
    // Test business logic
}
```

### Integration Tests
```go
// internal/adapters/out/repository/postgres/damaged_road_test.go
func TestDamagedRoadRepository_Create(t *testing.T) {
    // Test database operations
}
```

### API Tests
```go
// tests/api/damaged_road_test.go
func TestDamagedRoadAPI_CreateReport(t *testing.T) {
    // Test HTTP endpoints
}
```

## Development Workflow

### 1. Setup Development Environment
```bash
# Clone repository
git clone <repository-url>
cd jalanrusak-be

# Checkout feature branch
git checkout 002-logged-in-user

# Install dependencies
go mod download

# Setup environment
cp .env.example .env
# Edit .env with your configuration
```

### 2. Run Local Development
```bash
# Start database
docker-compose up -d postgres

# Run migrations
make migrate-up

# Start server
make run

# Run tests (optional)
make test
```

### 3. API Development
```bash
# Generate Swagger documentation
make docs

# Run API tests
make test-api

# Validate endpoints
curl -X GET http://localhost:8080/api/v1/damaged-roads \
  -H "Authorization: Bearer <jwt-token>"
```

## Performance Considerations

### Database Indexing
```sql
-- Spatial indexing for coordinates
CREATE INDEX idx_damaged_road_points_location
ON damaged_road_points USING GIST (ST_Point(lng, lat));

-- Composite index for user queries
CREATE INDEX idx_damaged_roads_author_created
ON damaged_roads(author_id, created_at DESC);
```

### Caching Strategy
```go
// Cache administrative boundary data
// Cache photo validation results
// Cache user sessions
```

### Connection Pooling
```go
// Configure database connection pool
db.SetMaxOpenConns(25)
db.SetMaxIdleConns(25)
db.SetConnMaxLifetime(5 * time.Minute)
```

## Security Checklist

- [ ] JWT authentication implemented
- [ ] Input validation at all boundaries
- [ ] SQL injection prevention (parameterized queries)
- [ ] SSRF protection for photo URL validation
- [ ] Rate limiting implemented
- [ ] CORS configured properly
- [ ] Environment variables for secrets
- [ ] HTTPS in production

## Deployment

### Build for Production
```bash
# Build binary
make build-prod

# Create Docker image
docker build -t jalanrusak-be .

# Run with Docker Compose
docker-compose -f docker-compose.prod.yml up -d
```

### Health Checks
```go
// GET /health
{
    "status": "ok",
    "timestamp": "2025-10-19T10:00:00Z",
    "database": "connected",
    "version": "1.0.0"
}
```

## Monitoring

### Metrics to Track
- API response times (P95 < 2s)
- Error rates (< 1%)
- Database query performance
- Photo validation success rates
- User authentication success

### Logging
```go
// Structured logging with context
logger.Info("Report created",
    "report_id", report.ID,
    "user_id", report.AuthorID,
    "kemendagri_code", report.KemendagriCode,
)
```

## Troubleshooting

### Common Issues
1. **PostGIS Extension Not Found**: `CREATE EXTENSION postgis;`
2. **JWT Token Invalid**: Check JWT_SECRET and token expiration
3. **Photo Validation Failing**: Check network connectivity and URL accessibility
4. **Coordinates Validation Failing**: Ensure administrative boundary data is loaded

### Debug Commands
```bash
# Check database connection
make db-connect

# View recent logs
make logs

# Run health check
curl http://localhost:8080/health
```

## Next Steps

1. **Implement core domain entities**
2. **Setup database schema and migrations**
3. **Create HTTP handlers with Swagger annotations**
4. **Implement validation services**
5. **Add authentication middleware**
6. **Write tests (optional)**
7. **Deploy to staging environment**
8. **Performance testing and optimization**

## References

- [Feature Specification](./spec.md)
- [Data Model](./data-model.md)
- [API Contracts](./contracts/openapi.yaml)
- [Research Findings](./research.md)
- [JalanRusak Backend Constitution](../../.specify/memory/constitution.md)