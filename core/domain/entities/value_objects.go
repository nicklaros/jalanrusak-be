package entities

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/nicklaros/jalanrusak-be/core/domain/errors"
)

// Point represents a geographic coordinate point (latitude, longitude)
type Point struct {
	Lat float64 `json:"lat" db:"lat"`
	Lng float64 `json:"lng" db:"lng"`
}

// NewPoint creates a new Point with validation
func NewPoint(lat, lng float64) (*Point, error) {
	p := &Point{Lat: lat, Lng: lng}
	if err := p.Validate(); err != nil {
		return nil, err
	}
	return p, nil
}

// Validate validates the point coordinates
func (p *Point) Validate() error {
	if p.Lat < -11 || p.Lat > 6 {
		return errors.NewValidationError("lat", "latitude must be between -11 and 6 (Indonesian boundaries)", errors.ErrCoordinatesOutOfBounds)
	}
	if p.Lng < 95 || p.Lng > 141 {
		return errors.NewValidationError("lng", "longitude must be between 95 and 141 (Indonesian boundaries)", errors.ErrCoordinatesOutOfBounds)
	}
	return nil
}

// Geometry represents a PostGIS geometry object (LineString for paths)
type Geometry struct {
	Type        string      `json:"type" db:"type"`               // "LineString"
	Coordinates [][]float64 `json:"coordinates" db:"coordinates"` // [[lng, lat], [lng, lat], ...]
}

// NewGeometry creates a new Geometry from coordinate pairs
func NewGeometry(coordinates [][]float64) (*Geometry, error) {
	g := &Geometry{
		Type:        "LineString",
		Coordinates: coordinates,
	}
	if err := g.Validate(); err != nil {
		return nil, err
	}
	return g, nil
}

// NewGeometryFromPoints creates a Geometry from Point objects
func NewGeometryFromPoints(points []Point) (*Geometry, error) {
	if len(points) == 0 {
		return nil, errors.NewValidationError("points", "at least 1 point required", errors.ErrInvalidPath)
	}
	if len(points) > 100 {
		return nil, errors.NewValidationError("points", "cannot have more than 100 points", errors.ErrTooManyPathPoints)
	}

	coordinates := make([][]float64, len(points))
	for i, p := range points {
		if err := p.Validate(); err != nil {
			return nil, fmt.Errorf("invalid point at index %d: %w", i, err)
		}
		coordinates[i] = []float64{p.Lng, p.Lat} // GeoJSON format: [longitude, latitude]
	}

	return NewGeometry(coordinates)
}

// Validate validates the geometry
func (g *Geometry) Validate() error {
	if g.Type != "LineString" {
		return errors.NewValidationError("type", "geometry type must be LineString", errors.ErrInvalidGeometry)
	}
	if len(g.Coordinates) < 1 {
		return errors.NewValidationError("coordinates", "at least 1 coordinate pair required", errors.ErrInvalidPath)
	}
	if len(g.Coordinates) > 100 {
		return errors.NewValidationError("coordinates", "cannot have more than 100 coordinate pairs", errors.ErrTooManyPathPoints)
	}

	for i, coord := range g.Coordinates {
		if len(coord) != 2 {
			return errors.NewValidationError("coordinates", fmt.Sprintf("coordinate at index %d must have exactly 2 values", i), errors.ErrInvalidGeometry)
		}
		lng, lat := coord[0], coord[1]
		if lat < -11 || lat > 6 {
			return errors.NewValidationError("coordinates", fmt.Sprintf("latitude at index %d must be between -11 and 6", i), errors.ErrCoordinatesOutOfBounds)
		}
		if lng < 95 || lng > 141 {
			return errors.NewValidationError("coordinates", fmt.Sprintf("longitude at index %d must be between 95 and 141", i), errors.ErrCoordinatesOutOfBounds)
		}
	}

	return nil
}

// ToPoints converts Geometry coordinates to Point objects
func (g *Geometry) ToPoints() []Point {
	points := make([]Point, len(g.Coordinates))
	for i, coord := range g.Coordinates {
		points[i] = Point{
			Lng: coord[0],
			Lat: coord[1],
		}
	}
	return points
}

// SubDistrictCode represents an Indonesian administrative code (Kemendagri format)
// Format: NN.NN.NN.NNNN (Province.District.Subdistrict.Village)
type SubDistrictCode string

var subdistrictCodeRegex = regexp.MustCompile(`^\d{2}\.\d{2}\.\d{2}\.\d{4}$`)

// NewSubDistrictCode creates a new SubDistrictCode with validation
func NewSubDistrictCode(code string) (SubDistrictCode, error) {
	s := SubDistrictCode(code)
	if err := s.Validate(); err != nil {
		return "", err
	}
	return s, nil
}

// Validate validates the subdistrict code format
func (s SubDistrictCode) Validate() error {
	if !subdistrictCodeRegex.MatchString(string(s)) {
		return errors.NewValidationError("subdistrict_code", "must match format NN.NN.NN.NNNN", errors.ErrInvalidSubDistrictCode)
	}
	return nil
}

// ProvinceCode returns the province code (first 2 digits)
func (s SubDistrictCode) ProvinceCode() string {
	parts := strings.Split(string(s), ".")
	if len(parts) >= 1 {
		return parts[0]
	}
	return ""
}

// DistrictCode returns the district code (Province.District)
func (s SubDistrictCode) DistrictCode() string {
	parts := strings.Split(string(s), ".")
	if len(parts) >= 2 {
		return parts[0] + "." + parts[1]
	}
	return ""
}

// SubDistrictLevel returns the subdistrict level code (Province.District.Subdistrict)
func (s SubDistrictCode) SubDistrictLevel() string {
	parts := strings.Split(string(s), ".")
	if len(parts) >= 3 {
		return parts[0] + "." + parts[1] + "." + parts[2]
	}
	return ""
}

// VillageCode returns the full village code
func (s SubDistrictCode) VillageCode() string {
	return string(s)
}

// String returns the string representation
func (s SubDistrictCode) String() string {
	return string(s)
}

// Title represents a report title with validation
type Title string

// NewTitle creates a new Title with validation
func NewTitle(title string) (Title, error) {
	t := Title(title)
	if err := t.Validate(); err != nil {
		return "", err
	}
	return t, nil
}

// Validate validates the title
func (t Title) Validate() error {
	length := len(string(t))
	if length < 3 {
		return errors.NewValidationError("title", "must be at least 3 characters", errors.ErrInvalidTitle)
	}
	if length > 100 {
		return errors.NewValidationError("title", "cannot exceed 100 characters", errors.ErrInvalidTitle)
	}
	if strings.TrimSpace(string(t)) == "" {
		return errors.NewValidationError("title", "cannot be empty or whitespace only", errors.ErrInvalidTitle)
	}
	return nil
}

// String returns the string representation
func (t Title) String() string {
	return string(t)
}

// Description represents an optional report description with validation
type Description string

// NewDescription creates a new Description with validation
func NewDescription(desc string) (Description, error) {
	d := Description(desc)
	if err := d.Validate(); err != nil {
		return "", err
	}
	return d, nil
}

// Validate validates the description
func (d Description) Validate() error {
	if len(string(d)) > 500 {
		return errors.NewValidationError("description", "cannot exceed 500 characters", errors.ErrInvalidDescription)
	}
	return nil
}

// String returns the string representation
func (d Description) String() string {
	return string(d)
}

// IsEmpty checks if the description is empty
func (d Description) IsEmpty() bool {
	return strings.TrimSpace(string(d)) == ""
}
