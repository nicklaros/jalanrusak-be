package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/nicklaros/jalanrusak-be/core/domain/entities"
	"github.com/nicklaros/jalanrusak-be/core/domain/errors"
	"github.com/nicklaros/jalanrusak-be/core/ports/external"
)

// boundaryRepository implements external.BoundaryRepository using PostgreSQL.
type boundaryRepository struct {
	db *sqlx.DB
}

// NewBoundaryRepository creates a new PostgreSQL boundary repository.
func NewBoundaryRepository(db *sqlx.DB) external.BoundaryRepository {
	return &boundaryRepository{db: db}
}

// GetCentroid retrieves the geographic centroid for a given subdistrict code.
func (r *boundaryRepository) GetCentroid(subDistrictCode entities.SubDistrictCode) (entities.Point, error) {
	ctx := context.Background()

	var result struct {
		Lat float64 `db:"centroid_lat"`
		Lng float64 `db:"centroid_lng"`
	}
	query := `
		SELECT centroid_lat, centroid_lng
		FROM subdistrict_centroids
		WHERE subdistrict_code = $1
	`

	err := r.db.GetContext(ctx, &result, query, string(subDistrictCode))
	if err != nil {
		if err == sql.ErrNoRows {
			return entities.Point{}, fmt.Errorf("%w: subdistrict code %s not found in boundary dataset",
				errors.ErrSubDistrictNotFound, string(subDistrictCode))
		}
		return entities.Point{}, fmt.Errorf("failed to retrieve centroid for %s: %w", string(subDistrictCode), err)
	}

	centroid := entities.Point{
		Lat: result.Lat,
		Lng: result.Lng,
	}

	return centroid, nil
}

// CheckSubDistrictExists verifies if a subdistrict code exists in the official dataset.
func (r *boundaryRepository) CheckSubDistrictExists(subDistrictCode entities.SubDistrictCode) (bool, error) {
	ctx := context.Background()

	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM subdistrict_centroids WHERE subdistrict_code = $1)`

	err := r.db.GetContext(ctx, &exists, query, string(subDistrictCode))
	if err != nil {
		return false, fmt.Errorf("failed to check subdistrict existence for %s: %w", string(subDistrictCode), err)
	}

	return exists, nil
}

// StoreCentroid stores centroid data for a subdistrict (for data seeding/updates).
func (r *boundaryRepository) StoreCentroid(subDistrictCode entities.SubDistrictCode, centroid entities.Point) error {
	ctx := context.Background()

	query := `
		INSERT INTO subdistrict_centroids (subdistrict_code, centroid_lat, centroid_lng, name)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (subdistrict_code) 
		DO UPDATE SET 
			centroid_lat = EXCLUDED.centroid_lat,
			centroid_lng = EXCLUDED.centroid_lng,
			updated_at = CURRENT_TIMESTAMP
	`

	// Extract name from subdistrict code for basic reference (can be enhanced with proper name lookup)
	name := fmt.Sprintf("Subdistrict %s", string(subDistrictCode))

	_, err := r.db.ExecContext(ctx, query, string(subDistrictCode), centroid.Lat, centroid.Lng, name)
	if err != nil {
		return fmt.Errorf("failed to store centroid for %s: %w", string(subDistrictCode), err)
	}

	return nil
}
