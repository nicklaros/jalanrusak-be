package usecases

import "github.com/nicklaros/jalanrusak-be/core/domain/entities"

// GeometryService provides geospatial validation operations for damaged road reports.
// It validates coordinates against Indonesian boundaries and subdistrict centroids.
type GeometryService interface {
	// ValidateCoordinatesInBoundary checks if all coordinates fall within Indonesian national boundaries.
	// Returns error if any coordinate is outside bounds (lat: -11 to 6, lng: 95 to 141).
	ValidateCoordinatesInBoundary(points []entities.Point) error

	// ValidateCoordinatesNearCentroid checks if at least one coordinate from the path
	// falls within the specified radius (in meters) of the subdistrict's centroid.
	// Returns error if subdistrict code not found or all coordinates are too far.
	ValidateCoordinatesNearCentroid(points []entities.Point, subDistrictCode entities.SubDistrictCode, radiusMeters float64) error

	// CalculateDistance computes the Haversine distance in meters between two points.
	// Used for proximity validation and reporting.
	CalculateDistance(point1, point2 entities.Point) float64

	// GetSubDistrictCentroid retrieves the geographic centroid for a given subdistrict code.
	// Returns error if subdistrict not found in the boundary dataset.
	GetSubDistrictCentroid(subDistrictCode entities.SubDistrictCode) (entities.Point, error)
}
