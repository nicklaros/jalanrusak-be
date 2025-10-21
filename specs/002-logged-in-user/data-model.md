# Data Model: Logged-In Users Report Damaged Roads

**Date**: 2025-10-19
**Branch**: 002-logged-in-user
**Architecture**: Simplified data model with PostGIS path storage and application-layer validation

## Core Entities

### DamagedRoad Report

Represents a single reported occurrence of road damage from a logged-in resident.

```go
type DamagedRoad struct {
    ID              string    `json:"id" db:"id"`
    Title           string    `json:"title" db:"title"`
    SubDistrictCode string    `json:"subdistrict_code" db:"subdistrict_code"`
    Path            Geometry  `json:"path" db:"path"`
    Description     *string   `json:"description,omitempty" db:"description"`
    PhotoURLs       []string  `json:"photo_urls" db:"photo_urls"` // In-memory array; persisted in separate damaged_road_photos table
    AuthorID        string    `json:"author_id" db:"author_id"`
    Status          Status    `json:"status" db:"status"`
    CreatedAt       time.Time `json:"created_at" db:"created_at"`
    UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

// Storage Note: While the domain entity uses PhotoURLs []string for simplicity,
// the repository layer stores photos in a normalized damaged_road_photos table
// with columns: id, damaged_road_id, url, validation_status, validated_at.
// This separation enables photo-level validation tracking without complicating
// the core entity model.
```

**Validation Rules**:
- `Title`: Required, 3-100 characters
- `SubDistrictCode`: Required, format NN.NN.NN.NNNN
- `Path`: Required, PostGIS LINESTRING geometry
- `PhotoURLs`: Required, 1-10 URLs, each must be valid and accessible
- `Description`: Optional, max 500 characters
- `AuthorID`: Required, references existing user
- `Status`: Defaults to `Submitted`

### Point (Coordinate)

Represents a latitude-longitude pair for API input validation.

```go
type Point struct {
    Lat float64 `json:"lat" db:"lat"`
    Lng float64 `json:"lng" db:"lng"`
}
```

**Validation Rules**:
- `Lat`: Between -11 and 6 (Indonesian boundaries)
- `Lng`: Between 95 and 141 (Indonesian boundaries)
- Must be within Indonesian national boundaries
- Should be near selected SubDistrict code centroid (within 200m)

### Geometry (PostGIS Path)

Represents the damaged road path as PostGIS geometry. Stored as LINESTRING in database.

```go
type Geometry struct {
    // Internal PostGIS geometry representation
    // API input comes as []Point, converted to PostGIS geometry
    Type        string    `json:"type"`        // "LineString"
    Coordinates [][][]float64 `json:"coordinates"` // [[lng, lat], [lng, lat], ...]
}
```

### Status (Lifecycle State)

Represents the lifecycle state of a damaged road report.

```go
type Status string

const (
    StatusSubmitted        Status = "submitted"
    StatusUnderVerification Status = "under_verification"
    StatusVerified         Status = "verified"
    StatusPendingResolved  Status = "pending_resolved"
    StatusResolved         Status = "resolved"
    StatusArchived         Status = "archived"
)
```

**Lifecycle Options**:

#### Option 1: Simple Flow (4 states)
1. `Submitted` → `Under Verification`
2. `Under Verification` → `Verified`
3. `Verified` → `Resolved`
4. `Resolved` → `Archived`

#### Option 2: Enhanced Flow (5 states)
1. `Submitted` → `Under Verification`
2. `Under Verification` → `Verified`
3. `Verified` → `Pending Resolved`
4. `Pending Resolved` → `Resolved`
5. `Resolved` → `Archived`

**Rationale for Option 2**: The `Pending Resolved` state provides better tracking for reports that have been verified and scheduled for repair but not yet completed, giving authorities more granular status tracking and citizens clearer expectations about repair timelines.

*Transitions are strictly forward (no backward transitions)*

### Photo Validation Result

Represents the result of photo URL validation.

```go
type PhotoValidationResult struct {
    URL         string    `json:"url"`
    ContentType string    `json:"content_type"`
    Width       int       `json:"width,omitempty"`
    Height      int       `json:"height,omitempty"`
    Size        int64     `json:"size,omitempty"`
    Accessible  bool      `json:"accessible"`
    VerifiedAt  time.Time `json:"verified_at"`
    Error       *string   `json:"error,omitempty"`
}
```

## Relationships

```
User (existing)
  ├── creates ──► DamagedRoad (1:N)
  │                ├── has ──► Geometry (1:1) - PostGIS path
  │                ├── has ──► PhotoURL (1:N, max 10)
  │                └── references ──► SubDistrictCode (validation in app layer)
  │
External APIs (application layer)
  ├── provides SubDistrict validation
  └── provides Indonesian boundary validation
```

## Database Schema

### Primary Tables

#### damaged_roads

```sql
CREATE TABLE damaged_roads (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(100) NOT NULL,
    subdistrict_code VARCHAR(13) NOT NULL,
    path GEOMETRY(LINESTRING, 4326) NOT NULL,
    description TEXT,
    author_id UUID NOT NULL REFERENCES users(id),
    status VARCHAR(50) NOT NULL DEFAULT 'submitted',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    -- Basic data integrity constraints only
    CONSTRAINT valid_title_length CHECK (LENGTH(title) >= 3),
    CONSTRAINT valid_subdistrict_code_format CHECK (subdistrict_code ~ '^\d{2}\.\d{2}\.\d{2}\.\d{4}$'),
    CONSTRAINT valid_path_geometry CHECK (ST_IsValid(path) AND ST_GeometryType(path) = 'ST_LineString')
);

CREATE INDEX idx_damaged_roads_author ON damaged_roads(author_id);
CREATE INDEX idx_damaged_roads_status ON damaged_roads(status);
CREATE INDEX idx_damaged_roads_subdistrict ON damaged_roads(subdistrict_code);
CREATE INDEX idx_damaged_roads_created_at ON damaged_roads(created_at);
CREATE INDEX idx_damaged_roads_path_geom ON damaged_roads USING GIST (path);
```

#### damaged_road_photos

```sql
CREATE TABLE damaged_road_photos (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    road_id UUID NOT NULL REFERENCES damaged_roads(id) ON DELETE CASCADE,
    url TEXT NOT NULL,
    content_type VARCHAR(50),
    file_size BIGINT,
    validation_status VARCHAR(20) DEFAULT 'pending',
    validated_at TIMESTAMP WITH TIME ZONE,
    validation_error TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    CONSTRAINT unique_photo_url UNIQUE(road_id, url)
);

CREATE INDEX idx_damaged_road_photos_road ON damaged_road_photos(road_id);
CREATE INDEX idx_damaged_road_photos_validation ON damaged_road_photos(validation_status);
```

### Validation Strategy

**Database Layer** (Basic Data Integrity):
- NOT NULL constraints
- Foreign key relationships
- Data type validation
- Geometry validity checks
- Basic format validation (regex for SubDistrictCode)

**Application Layer** (Business Validation):
- Indonesian boundary validation (external API or reference data)
- SubDistrictCode existence and location matching validation
- Photo URL accessibility and content validation
- Photo limit enforcement (1-10 photos)
- Coordinate proximity to SubDistrict centroid validation
- Status transition validation

## Domain Services

### Location Validation Service (Application Layer)

```go
type LocationValidator interface {
    ValidateCoordinates(lat, lng float64) error
    ValidateSubDistrictCode(code string) error
    ValidateLocationMatch(lat, lng float64, subDistrictCode string) error
    GetAdministrativeInfo(lat, lng float64) (*AdministrativeInfo, error)
}

// Implementation uses external APIs or cached reference data
type LocationValidatorImpl struct {
    // External API clients or cached reference data
    boundaryAPI     BoundaryAPIClient
    subdistrictAPI   SubDistrictAPIClient
    cache           Cache
}
```

### Photo URL Validation Service (Application Layer)

```go
type PhotoValidator interface {
    ValidateURL(ctx context.Context, url string) (*PhotoValidationResult, error)
    ValidateURLs(ctx context.Context, urls []string) ([]*PhotoValidationResult, error)
    IsURLAllowed(url string) bool
    ValidatePhotoLimit(urls []string) error
}
```

### Geometry Service (Application Layer)

```go
type GeometryService interface {
    ConvertPointsToGeometry(points []Point) (Geometry, error)
    ConvertGeometryToPoints(geometry Geometry) ([]Point, error)
    ValidateGeometry(geometry Geometry) error
    CalculatePathLength(geometry Geometry) (float64, error)
}
```

### Damaged Road Repository (Database Layer)

```go
type DamagedRoadRepository interface {
    Create(ctx context.Context, road *DamagedRoad) error
    GetByID(ctx context.Context, id string) (*DamagedRoad, error)
    GetByAuthor(ctx context.Context, authorID string, limit, offset int) ([]*DamagedRoad, error)
    UpdateStatus(ctx context.Context, id string, status Status) error
    Update(ctx context.Context, road *DamagedRoad) error
    Delete(ctx context.Context, id string) error
    List(ctx context.Context, filters *ListFilters) ([]*DamagedRoad, error)
    FindByGeometry(ctx context.Context, bounds Geometry) ([]*DamagedRoad, error)
}
```

## Value Objects

### SubDistrictCode

```go
type SubDistrictCode string

func (s SubDistrictCode) IsValid() bool {
    matched, _ := regexp.MatchString(`^\d{2}\.\d{2}\.\d{2}\.\d{4}$`, string(s))
    return matched
}

func (s SubDistrictCode) ProvinceCode() string {
    parts := strings.Split(string(s), ".")
    if len(parts) >= 1 {
        return parts[0]
    }
    return ""
}

func (s SubDistrictCode) DistrictCode() string {
    parts := strings.Split(string(s), ".")
    if len(parts) >= 2 {
        return parts[0] + "." + parts[1]
    }
    return ""
}

func (s SubDistrictCode) SubDistrictCode() string {
    parts := strings.Split(string(s), ".")
    if len(parts) >= 3 {
        return parts[0] + "." + parts[1] + "." + parts[2]
    }
    return ""
}
```

### Title

```go
type Title string

func (t Title) IsValid() bool {
    length := len(strings.TrimSpace(string(t)))
    return length >= 3 && length <= 100
}
```

### PhotoURL

```go
type PhotoURL string

func (p PhotoURL) IsValid() bool {
    _, err := url.Parse(string(p))
    return err == nil
}

func (p PhotoURL) IsAllowed() bool {
    parsed, err := url.Parse(string(p))
    if err != nil {
        return false
    }

    allowedHosts := map[string]bool{
        "instagram.com":     true,
        "cdninstagram.com":  true,
        "facebook.com":      true,
        "fbcdn.net":        true,
        "googleusercontent.com": true,
        "storage.googleapis.com": true,
        // Add other allowed domains
    }

    return allowedHosts[parsed.Hostname()]
}
```

## Error Types

```go
type DamagedRoadError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    Field   string `json:"field,omitempty"`
}

var (
    ErrInvalidTitle           = &DamagedRoadError{Code: "INVALID_TITLE", Message: "Title must be 3-100 characters", Field: "title"}
    ErrInvalidSubDistrictCode = &DamagedRoadError{Code: "INVALID_SUBDISTRICT_CODE", Message: "Invalid administrative code format", Field: "subdistrict_code"}
    ErrInvalidCoordinates     = &DamagedRoadError{Code: "INVALID_COORDINATES", Message: "Coordinates outside Indonesian boundaries", Field: "path"}
    ErrPhotoLimitExceeded     = &DamagedRoadError{Code: "PHOTO_LIMIT_EXCEEDED", Message: "Maximum 10 photos allowed", Field: "photo_urls"}
    ErrInvalidPhotoURL        = &DamagedRoadError{Code: "INVALID_PHOTO_URL", Message: "Invalid or inaccessible photo URL", Field: "photo_urls"}
    ErrMissingPhotos          = &DamagedRoadError{Code: "MISSING_PHOTOS", Message: "At least one photo is required", Field: "photo_urls"}
    ErrMissingCoordinates     = &DamagedRoadError{Code: "MISSING_COORDINATES", Message: "At least one coordinate point is required", Field: "path"}
    ErrLocationMismatch       = &DamagedRoadError{Code: "LOCATION_MISMATCH", Message: "Coordinates do not match administrative code", Field: "path"}
    ErrInvalidGeometry        = &DamagedRoadError{Code: "INVALID_GEOMETRY", Message: "Invalid path geometry", Field: "path"}
)
```

## Application Layer Validation

### Business Validation Service

```go
type BusinessValidationService struct {
    locationValidator LocationValidator
    photoValidator     PhotoValidator
    geometryService    GeometryService
}

func (v *BusinessValidationService) ValidateCreateRequest(req *CreateDamagedRoadRequest) error {
    // 1. Basic field validation
    if !Title(req.Title).IsValid() {
        return ErrInvalidTitle
    }

    if !SubDistrictCode(req.SubDistrictCode).IsValid() {
        return ErrInvalidSubDistrictCode
    }

    // 2. Photo validation
    if len(req.PhotoURLs) == 0 {
        return ErrMissingPhotos
    }

    if len(req.PhotoURLs) > 10 {
        return ErrPhotoLimitExceeded
    }

    // 3. Geometry validation
    if len(req.PathPoints) == 0 {
        return ErrMissingCoordinates
    }

    geometry, err := v.geometryService.ConvertPointsToGeometry(req.PathPoints)
    if err != nil {
        return ErrInvalidGeometry
    }

    // 4. Coordinate boundary validation (application layer)
    for _, point := range req.PathPoints {
        if err := v.locationValidator.ValidateCoordinates(point.Lat, point.Lng); err != nil {
            return ErrInvalidCoordinates
        }
    }

    // 5. SubDistrict validation (application layer)
    if err := v.locationValidator.ValidateSubDistrictCode(req.SubDistrictCode); err != nil {
        return ErrInvalidSubDistrictCode
    }

    // 6. Location matching validation (application layer)
    for _, point := range req.PathPoints {
        if err := v.locationValidator.ValidateLocationMatch(point.Lat, point.Lng, req.SubDistrictCode); err != nil {
            return ErrLocationMismatch
        }
    }

    // 7. Photo URL validation (application layer)
    for _, url := range req.PhotoURLs {
        if !PhotoURL(url).IsValid() {
            return ErrInvalidPhotoURL
        }

        if !PhotoURL(url).IsAllowed() {
            return ErrInvalidPhotoURL
        }
    }

    return nil
}
```

## Performance Considerations

### Indexing Strategy

1. **Spatial Indexes**: GIST indexes on path geometry for fast spatial queries
2. **Composite Indexes**: (author_id, created_at) for user-specific queries
3. **Partial Indexes**: For status-based filtering on large datasets

### PostGIS Geometry Benefits

1. **Efficient Storage**: Single field for entire path instead of multiple point rows
2. **Spatial Queries**: Native support for intersection, distance, and containment queries
3. **Indexing**: GIST indexes provide fast spatial queries
4. **Standard Compliance**: Follows OGC Simple Features standard

### Caching Strategy

1. **External API Results**: Cache SubDistrict validation results
2. **Photo Validation Results**: Cache validation results for 1 hour
3. **User Sessions**: Cache user authentication and permissions
4. **Boundary Data**: Cache Indonesian boundary reference data

### Connection Pooling

- **Read Operations**: 20-50 connections for high read throughput
- **Write Operations**: 10-20 connections for report creation
- **Spatial Operations**: Optimized for PostGIS queries