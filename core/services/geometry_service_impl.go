package services

import (
	"fmt"
	"math"

	"github.com/nicklaros/jalanrusak-be/core/domain/entities"
	"github.com/nicklaros/jalanrusak-be/core/domain/errors"
	"github.com/nicklaros/jalanrusak-be/core/ports/external"
	"github.com/nicklaros/jalanrusak-be/core/ports/usecases"
)

// geometryServiceImpl implements GeometryService for geospatial validation operations.
type geometryServiceImpl struct {
	boundaryRepo external.BoundaryRepository
}

// NewGeometryService creates a new GeometryService instance with the provided boundary repository.
func NewGeometryService(boundaryRepo external.BoundaryRepository) usecases.GeometryService {
	return &geometryServiceImpl{
		boundaryRepo: boundaryRepo,
	}
}

// ValidateCoordinatesInBoundary checks if all coordinates fall within Indonesian national boundaries.
// Indonesian bounds: latitude -11 to 6, longitude 95 to 141.
func (s *geometryServiceImpl) ValidateCoordinatesInBoundary(points []entities.Point) error {
	const (
		minLat = -11.0
		maxLat = 6.0
		minLng = 95.0
		maxLng = 141.0
	)

	for i, point := range points {
		if point.Lat < minLat || point.Lat > maxLat {
			return fmt.Errorf("%w: coordinate %d latitude %.6f is outside Indonesian bounds [%.1f, %.1f]",
				errors.ErrCoordinatesOutOfBounds, i, point.Lat, minLat, maxLat)
		}
		if point.Lng < minLng || point.Lng > maxLng {
			return fmt.Errorf("%w: coordinate %d longitude %.6f is outside Indonesian bounds [%.1f, %.1f]",
				errors.ErrCoordinatesOutOfBounds, i, point.Lng, minLng, maxLng)
		}
	}

	return nil
}

// ValidateCoordinatesNearCentroid checks if at least one coordinate from the path
// falls within the specified radius (in meters) of the subdistrict's centroid.
// Implements FR-006 requirement: "at least one coordinate must fall within 200 meters of centroid".
func (s *geometryServiceImpl) ValidateCoordinatesNearCentroid(points []entities.Point, subDistrictCode entities.SubDistrictCode, radiusMeters float64) error {
	// Retrieve centroid from repository
	centroid, err := s.boundaryRepo.GetCentroid(subDistrictCode)
	if err != nil {
		return fmt.Errorf("%w: %v", errors.ErrSubDistrictNotFound, err)
	}

	// Check if any point falls within the radius
	for _, point := range points {
		distance := s.CalculateDistance(point, centroid)
		if distance <= radiusMeters {
			return nil // At least one point is within radius - validation passes
		}
	}

	// All points are too far from centroid
	return fmt.Errorf("%w: no coordinate falls within %.0f meters of subdistrict %s centroid (%.6f, %.6f)",
		errors.ErrLocationNotInBoundary, radiusMeters, string(subDistrictCode), centroid.Lat, centroid.Lng)
}

// CalculateDistance computes the Haversine distance in meters between two geographic points.
// Haversine formula accounts for Earth's curvature and provides accurate results for small distances.
func (s *geometryServiceImpl) CalculateDistance(point1, point2 entities.Point) float64 {
	const earthRadiusMeters = 6371000.0 // Earth's mean radius in meters

	// Convert degrees to radians
	lat1Rad := degreesToRadians(point1.Lat)
	lat2Rad := degreesToRadians(point2.Lat)
	deltaLatRad := degreesToRadians(point2.Lat - point1.Lat)
	deltaLngRad := degreesToRadians(point2.Lng - point1.Lng)

	// Haversine formula
	a := math.Sin(deltaLatRad/2)*math.Sin(deltaLatRad/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(deltaLngRad/2)*math.Sin(deltaLngRad/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadiusMeters * c
}

// GetSubDistrictCentroid retrieves the geographic centroid for a given subdistrict code.
func (s *geometryServiceImpl) GetSubDistrictCentroid(subDistrictCode entities.SubDistrictCode) (entities.Point, error) {
	centroid, err := s.boundaryRepo.GetCentroid(subDistrictCode)
	if err != nil {
		return entities.Point{}, fmt.Errorf("%w: %v", errors.ErrSubDistrictNotFound, err)
	}
	return centroid, nil
}

// degreesToRadians converts degrees to radians for trigonometric calculations.
func degreesToRadians(degrees float64) float64 {
	return degrees * math.Pi / 180.0
}
