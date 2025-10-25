package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/nicklaros/jalanrusak-be/core/domain/entities"
	"github.com/nicklaros/jalanrusak-be/core/domain/errors"
	"github.com/nicklaros/jalanrusak-be/core/ports/external"
)

// DamagedRoadRepository implements the damaged road repository using PostgreSQL
type DamagedRoadRepository struct {
	db *sqlx.DB
}

// NewDamagedRoadRepository creates a new PostgreSQL damaged road repository
func NewDamagedRoadRepository(db *sqlx.DB) external.DamagedRoadRepository {
	return &DamagedRoadRepository{db: db}
}

// damagedRoadRow represents the database row structure
type damagedRoadRow struct {
	ID              uuid.UUID      `db:"id"`
	Title           string         `db:"title"`
	SubDistrictCode string         `db:"subdistrict_code"`
	Path            string         `db:"path"` // PostGIS geometry as text
	Description     sql.NullString `db:"description"`
	PhotoURLs       pq.StringArray `db:"photo_urls"`
	AuthorID        uuid.UUID      `db:"author_id"`
	Status          string         `db:"status"`
	CreatedAt       sql.NullTime   `db:"created_at"`
	UpdatedAt       sql.NullTime   `db:"updated_at"`
}

// toEntity converts a database row to an entity
func (row *damagedRoadRow) toEntity() (*entities.DamagedRoad, error) {
	// Parse geometry from PostGIS text format
	var geometry entities.Geometry
	if err := json.Unmarshal([]byte(row.Path), &geometry); err != nil {
		return nil, fmt.Errorf("failed to parse geometry: %w", err)
	}

	title, err := entities.NewTitle(row.Title)
	if err != nil {
		return nil, fmt.Errorf("invalid title: %w", err)
	}

	subdistrictCode, err := entities.NewSubDistrictCode(row.SubDistrictCode)
	if err != nil {
		return nil, fmt.Errorf("invalid subdistrict code: %w", err)
	}

	var description *entities.Description
	if row.Description.Valid {
		desc, err := entities.NewDescription(row.Description.String)
		if err != nil {
			return nil, fmt.Errorf("invalid description: %w", err)
		}
		description = &desc
	}

	road := &entities.DamagedRoad{
		ID:              row.ID,
		Title:           title,
		SubDistrictCode: subdistrictCode,
		Path:            geometry,
		Description:     description,
		PhotoURLs:       row.PhotoURLs,
		AuthorID:        row.AuthorID,
		Status:          entities.Status(row.Status),
		CreatedAt:       row.CreatedAt.Time,
		UpdatedAt:       row.UpdatedAt.Time,
	}

	return road, nil
}

// Create creates a new damaged road report
func (r *DamagedRoadRepository) Create(ctx context.Context, road *entities.DamagedRoad) error {
	// Convert geometry to GeoJSON for PostGIS
	geometryJSON, err := json.Marshal(road.Path)
	if err != nil {
		return errors.NewDatabaseError("marshal geometry", err)
	}

	var description sql.NullString
	if road.Description != nil {
		description = sql.NullString{String: road.Description.String(), Valid: true}
	}

	// Start a transaction
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return errors.NewDatabaseError("begin transaction", err)
	}
	defer tx.Rollback()

	// Insert the damaged road (without photo_urls column)
	roadQuery := `
		INSERT INTO damaged_roads (
			id, title, subdistrict_code, path, description, author_id, status, created_at, updated_at
		) VALUES (
			$1, $2, $3, ST_GeomFromGeoJSON($4), $5, $6, $7, $8, $9
		)
	`

	_, err = tx.ExecContext(ctx, roadQuery,
		road.ID,
		road.Title.String(),
		road.SubDistrictCode.String(),
		string(geometryJSON),
		description,
		road.AuthorID,
		road.Status.String(),
		road.CreatedAt,
		road.UpdatedAt,
	)

	if err != nil {
		return errors.NewDatabaseError("create damaged road", err)
	}

	// Insert photos into damaged_road_photos table
	if len(road.PhotoURLs) > 0 {
		photoQuery := `
			INSERT INTO damaged_road_photos (road_id, url, validation_status)
			VALUES ($1, $2, 'pending')
		`
		for _, photoURL := range road.PhotoURLs {
			_, err = tx.ExecContext(ctx, photoQuery, road.ID, photoURL)
			if err != nil {
				return errors.NewDatabaseError("insert damaged road photo", err)
			}
		}
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		return errors.NewDatabaseError("commit transaction", err)
	}

	return nil
}

// FindByID retrieves a damaged road report by ID
func (r *DamagedRoadRepository) FindByID(ctx context.Context, id uuid.UUID) (*entities.DamagedRoad, error) {
	query := `
		SELECT 
			id, title, subdistrict_code, 
			ST_AsGeoJSON(path) as path,
			description, 
			ARRAY(SELECT url FROM damaged_road_photos WHERE road_id = $1) as photo_urls,
			author_id, status, created_at, updated_at
		FROM damaged_roads
		WHERE id = $1
	`

	var row damagedRoadRow
	err := r.db.GetContext(ctx, &row, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, errors.NewDatabaseError("find damaged road by id", err)
	}

	return row.toEntity()
}

// FindByAuthor retrieves damaged road reports by author with pagination
func (r *DamagedRoadRepository) FindByAuthor(
	ctx context.Context,
	authorID uuid.UUID,
	limit, offset int,
) ([]*entities.DamagedRoad, int, error) {
	// Get total count
	var total int
	countQuery := `SELECT COUNT(*) FROM damaged_roads WHERE author_id = $1`
	if err := r.db.GetContext(ctx, &total, countQuery, authorID); err != nil {
		return nil, 0, errors.NewDatabaseError("count reports by author", err)
	}

	// Get paginated results
	query := `
		SELECT 
			dr.id, dr.title, dr.subdistrict_code,
			ST_AsGeoJSON(dr.path) as path,
			dr.description,
			ARRAY(SELECT url FROM damaged_road_photos WHERE road_id = dr.id) as photo_urls,
			dr.author_id, dr.status, dr.created_at, dr.updated_at
		FROM damaged_roads dr
		WHERE dr.author_id = $1
		ORDER BY dr.created_at DESC
		LIMIT $2 OFFSET $3
	`

	var rows []damagedRoadRow
	if err := r.db.SelectContext(ctx, &rows, query, authorID, limit, offset); err != nil {
		return nil, 0, errors.NewDatabaseError("find reports by author", err)
	}

	roads := make([]*entities.DamagedRoad, 0, len(rows))
	for _, row := range rows {
		road, err := row.toEntity()
		if err != nil {
			return nil, 0, fmt.Errorf("failed to convert row to entity: %w", err)
		}
		roads = append(roads, road)
	}

	return roads, total, nil
}

// List retrieves damaged road reports with filters and pagination
func (r *DamagedRoadRepository) List(
	ctx context.Context,
	filters *entities.DamagedRoadFilters,
) ([]*entities.DamagedRoad, int, error) {
	// Build query with filters
	baseQuery := `
		SELECT 
			dr.id, dr.title, dr.subdistrict_code,
			ST_AsGeoJSON(dr.path) as path,
			dr.description,
			ARRAY(SELECT url FROM damaged_road_photos WHERE road_id = dr.id) as photo_urls,
			dr.author_id, dr.status, dr.created_at, dr.updated_at
		FROM damaged_roads dr
		WHERE 1=1
	`

	countQuery := `SELECT COUNT(*) FROM damaged_roads WHERE 1=1`

	args := []interface{}{}
	argPos := 1

	// Apply filters
	if filters.Status != nil {
		baseQuery += fmt.Sprintf(" AND dr.status = $%d", argPos)
		countQuery += fmt.Sprintf(" AND status = $%d", argPos)
		args = append(args, filters.Status.String())
		argPos++
	}

	if filters.SubDistrictCode != nil {
		baseQuery += fmt.Sprintf(" AND dr.subdistrict_code = $%d", argPos)
		countQuery += fmt.Sprintf(" AND subdistrict_code = $%d", argPos)
		args = append(args, *filters.SubDistrictCode)
		argPos++
	}

	if filters.AuthorID != nil {
		baseQuery += fmt.Sprintf(" AND dr.author_id = $%d", argPos)
		countQuery += fmt.Sprintf(" AND author_id = $%d", argPos)
		args = append(args, *filters.AuthorID)
		argPos++
	}

	// Get total count
	var total int
	if err := r.db.GetContext(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, errors.NewDatabaseError("count reports", err)
	}

	// Add ordering and pagination
	baseQuery += fmt.Sprintf(" ORDER BY dr.created_at DESC LIMIT $%d OFFSET $%d", argPos, argPos+1)
	args = append(args, filters.Limit, filters.Offset)

	// Execute query
	var rows []damagedRoadRow
	if err := r.db.SelectContext(ctx, &rows, baseQuery, args...); err != nil {
		return nil, 0, errors.NewDatabaseError("list reports", err)
	}

	roads := make([]*entities.DamagedRoad, 0, len(rows))
	for _, row := range rows {
		road, err := row.toEntity()
		if err != nil {
			return nil, 0, fmt.Errorf("failed to convert row to entity: %w", err)
		}
		roads = append(roads, road)
	}

	return roads, total, nil
}

// UpdateStatus updates the status of a damaged road report
func (r *DamagedRoadRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status entities.Status) error {
	query := `
		UPDATE damaged_roads
		SET status = $1, updated_at = NOW()
		WHERE id = $2
	`

	result, err := r.db.ExecContext(ctx, query, status.String(), id)
	if err != nil {
		return errors.NewDatabaseError("update status", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return errors.NewDatabaseError("check rows affected", err)
	}

	if rows == 0 {
		return errors.ErrRecordNotFound
	}

	return nil
}

// Update updates an existing damaged road report
func (r *DamagedRoadRepository) Update(ctx context.Context, road *entities.DamagedRoad) error {
	geometryJSON, err := json.Marshal(road.Path)
	if err != nil {
		return errors.NewDatabaseError("marshal geometry", err)
	}

	var description sql.NullString
	if road.Description != nil {
		description = sql.NullString{String: road.Description.String(), Valid: true}
	}

	// Start a transaction
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return errors.NewDatabaseError("begin transaction", err)
	}
	defer tx.Rollback()

	// Update the damaged road (without photo_urls column)
	roadQuery := `
		UPDATE damaged_roads
		SET title = $1, subdistrict_code = $2, path = ST_GeomFromGeoJSON($3), 
		    description = $4, status = $5, updated_at = $6
		WHERE id = $7
	`

	result, err := tx.ExecContext(ctx, roadQuery,
		road.Title.String(),
		road.SubDistrictCode.String(),
		string(geometryJSON),
		description,
		road.Status.String(),
		road.UpdatedAt,
		road.ID,
	)

	if err != nil {
		return errors.NewDatabaseError("update damaged road", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return errors.NewDatabaseError("check rows affected", err)
	}

	if rows == 0 {
		return errors.ErrRecordNotFound
	}

	// Delete existing photos
	deletePhotosQuery := `DELETE FROM damaged_road_photos WHERE road_id = $1`
	_, err = tx.ExecContext(ctx, deletePhotosQuery, road.ID)
	if err != nil {
		return errors.NewDatabaseError("delete existing photos", err)
	}

	// Insert new photos
	if len(road.PhotoURLs) > 0 {
		photoQuery := `
			INSERT INTO damaged_road_photos (road_id, url, validation_status)
			VALUES ($1, $2, 'pending')
		`
		for _, photoURL := range road.PhotoURLs {
			_, err = tx.ExecContext(ctx, photoQuery, road.ID, photoURL)
			if err != nil {
				return errors.NewDatabaseError("insert damaged road photo", err)
			}
		}
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		return errors.NewDatabaseError("commit transaction", err)
	}

	return nil
}

// Delete deletes a damaged road report by ID
func (r *DamagedRoadRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM damaged_roads WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return errors.NewDatabaseError("delete damaged road", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return errors.NewDatabaseError("check rows affected", err)
	}

	if rows == 0 {
		return errors.ErrRecordNotFound
	}

	return nil
}

// FindByGeometry finds damaged road reports within a geographic boundary
func (r *DamagedRoadRepository) FindByGeometry(
	ctx context.Context,
	bounds entities.Geometry,
) ([]*entities.DamagedRoad, error) {
	geometryJSON, err := json.Marshal(bounds)
	if err != nil {
		return nil, errors.NewDatabaseError("marshal bounds geometry", err)
	}

	query := `
		SELECT 
			dr.id, dr.title, dr.subdistrict_code,
			ST_AsGeoJSON(dr.path) as path,
			dr.description,
			ARRAY(SELECT url FROM damaged_road_photos WHERE road_id = dr.id) as photo_urls,
			dr.author_id, dr.status, dr.created_at, dr.updated_at
		FROM damaged_roads dr
		WHERE ST_Intersects(dr.path, ST_GeomFromGeoJSON($1))
		ORDER BY dr.created_at DESC
	`

	var rows []damagedRoadRow
	if err := r.db.SelectContext(ctx, &rows, query, string(geometryJSON)); err != nil {
		return nil, errors.NewDatabaseError("find by geometry", err)
	}

	roads := make([]*entities.DamagedRoad, 0, len(rows))
	for _, row := range rows {
		road, err := row.toEntity()
		if err != nil {
			return nil, fmt.Errorf("failed to convert row to entity: %w", err)
		}
		roads = append(roads, road)
	}

	return roads, nil
}
