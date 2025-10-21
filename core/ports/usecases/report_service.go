package usecases

import (
	"context"

	"github.com/google/uuid"
	"github.com/nicklaros/jalanrusak-be/core/domain/entities"
)

// ReportService defines the use case interface for damaged road report operations
type ReportService interface {
	// CreateReport creates a new damaged road report
	// Returns the created report or an error if validation fails
	CreateReport(
		ctx context.Context,
		title entities.Title,
		subdistrictCode entities.SubDistrictCode,
		pathPoints []entities.Point,
		photoURLs []string,
		authorID uuid.UUID,
		description *entities.Description,
	) (*entities.DamagedRoad, error)

	// GetReport retrieves a damaged road report by ID
	GetReport(ctx context.Context, id uuid.UUID) (*entities.DamagedRoad, error)

	// ListReportsByAuthor retrieves all reports created by a specific author
	ListReportsByAuthor(
		ctx context.Context,
		authorID uuid.UUID,
		limit, offset int,
	) ([]*entities.DamagedRoad, int, error)

	// ListReports retrieves damaged road reports with filters
	ListReports(
		ctx context.Context,
		filters *entities.DamagedRoadFilters,
	) ([]*entities.DamagedRoad, int, error)

	// UpdateReportStatus updates the status of a damaged road report
	// Only authorized users (verificators/admins) can update status
	UpdateReportStatus(
		ctx context.Context,
		id uuid.UUID,
		newStatus entities.Status,
		requesterID uuid.UUID,
	) (*entities.DamagedRoad, error)

	// DeleteReport deletes a damaged road report
	// Only the author can delete their own report
	DeleteReport(ctx context.Context, id uuid.UUID, requesterID uuid.UUID) error
}
