package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/nicklaros/jalanrusak-be/core/domain/entities"
	"github.com/nicklaros/jalanrusak-be/core/domain/errors"
	"github.com/nicklaros/jalanrusak-be/core/ports/external"
	"github.com/nicklaros/jalanrusak-be/core/ports/usecases"
	"github.com/nicklaros/jalanrusak-be/pkg/logger"
)

// ReportServiceImpl implements the ReportService use case
type ReportServiceImpl struct {
	repo           external.DamagedRoadRepository
	geometrySvc    usecases.GeometryService
	photoValidator external.PhotoValidator
}

// NewReportService creates a new ReportService implementation
func NewReportService(repo external.DamagedRoadRepository, geometrySvc usecases.GeometryService, photoValidator external.PhotoValidator) usecases.ReportService {
	return &ReportServiceImpl{
		repo:           repo,
		geometrySvc:    geometrySvc,
		photoValidator: photoValidator,
	}
}

// CreateReport creates a new damaged road report
func (s *ReportServiceImpl) CreateReport(
	ctx context.Context,
	title entities.Title,
	subdistrictCode entities.SubDistrictCode,
	pathPoints []entities.Point,
	photoURLs []string,
	authorID uuid.UUID,
	description *entities.Description,
) (*entities.DamagedRoad, error) {
	logger.InfoContext(ctx, "Creating new damaged road report", map[string]interface{}{
		"author_id":        authorID.String(),
		"title":            title.String(),
		"subdistrict_code": subdistrictCode.String(),
		"path_points":      len(pathPoints),
		"photo_urls":       len(photoURLs),
	})

	// Validate photo URLs with SSRF protection (FR-004)
	photoResults := s.photoValidator.ValidateURLs(photoURLs)
	var invalidPhotos []string
	for _, result := range photoResults {
		if !result.Valid {
			invalidPhotos = append(invalidPhotos, fmt.Sprintf("%s: %s", result.URL, result.Error))
		}
	}
	if len(invalidPhotos) > 0 {
		logger.WarnContext(ctx, "Invalid photo URLs detected", map[string]interface{}{
			"invalid_count": len(invalidPhotos),
			"errors":        invalidPhotos,
		})
		return nil, fmt.Errorf("%w: %v", errors.ErrInvalidPhotoURLs, strings.Join(invalidPhotos, "; "))
	}

	// Validate coordinates are within Indonesian boundaries (FR-005)
	if err := s.geometrySvc.ValidateCoordinatesInBoundary(pathPoints); err != nil {
		logger.WarnContext(ctx, "Coordinates outside Indonesian boundaries", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, err
	}

	// Validate coordinates are near subdistrict centroid (FR-006)
	// At least one coordinate must be within 200 meters per spec
	// if err := s.geometrySvc.ValidateCoordinatesNearCentroid(pathPoints, subdistrictCode, 200.0); err != nil {
	// 	logger.WarnContext(ctx, "Coordinates do not match subdistrict location", map[string]interface{}{
	// 		"error":            err.Error(),
	// 		"subdistrict_code": subdistrictCode.String(),
	// 	})
	// 	return nil, err
	// }

	// Convert path points to geometry
	geometry, err := entities.NewGeometryFromPoints(pathPoints)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to convert path points to geometry", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("invalid path points: %w", err)
	}

	// Create the damaged road entity
	road, err := entities.NewDamagedRoad(
		title,
		subdistrictCode,
		*geometry,
		photoURLs,
		authorID,
		description,
	)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to create damaged road entity", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("failed to create report: %w", err)
	}

	// Save to repository
	if err := s.repo.Create(ctx, road); err != nil {
		logger.ErrorContext(ctx, "Failed to save damaged road report", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("failed to save report: %w", err)
	}

	logger.InfoContext(ctx, "Successfully created damaged road report", map[string]interface{}{
		"report_id": road.ID.String(),
	})

	return road, nil
}

// GetReport retrieves a damaged road report by ID
func (s *ReportServiceImpl) GetReport(ctx context.Context, id uuid.UUID) (*entities.DamagedRoad, error) {
	logger.DebugContext(ctx, "Retrieving damaged road report", map[string]interface{}{
		"report_id": id.String(),
	})

	road, err := s.repo.FindByID(ctx, id)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to retrieve damaged road report", map[string]interface{}{
			"report_id": id.String(),
			"error":     err.Error(),
		})
		return nil, fmt.Errorf("failed to get report: %w", err)
	}

	if road == nil {
		return nil, errors.ErrReportNotFound
	}

	return road, nil
}

// ListReportsByAuthor retrieves all reports created by a specific author
func (s *ReportServiceImpl) ListReportsByAuthor(
	ctx context.Context,
	authorID uuid.UUID,
	limit, offset int,
) ([]*entities.DamagedRoad, int, error) {
	logger.DebugContext(ctx, "Listing reports by author", map[string]interface{}{
		"author_id": authorID.String(),
		"limit":     limit,
		"offset":    offset,
	})

	// Set default pagination values
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	roads, total, err := s.repo.FindByAuthor(ctx, authorID, limit, offset)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to list reports by author", map[string]interface{}{
			"author_id": authorID.String(),
			"error":     err.Error(),
		})
		return nil, 0, fmt.Errorf("failed to list reports: %w", err)
	}

	return roads, total, nil
}

// ListReports retrieves damaged road reports with filters
func (s *ReportServiceImpl) ListReports(
	ctx context.Context,
	filters *entities.DamagedRoadFilters,
) ([]*entities.DamagedRoad, int, error) {
	logger.DebugContext(ctx, "Listing reports with filters", map[string]interface{}{
		"limit":  filters.Limit,
		"offset": filters.Offset,
	})

	// Set default pagination values
	if filters.Limit <= 0 || filters.Limit > 100 {
		filters.Limit = 20
	}
	if filters.Offset < 0 {
		filters.Offset = 0
	}

	roads, total, err := s.repo.List(ctx, filters)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to list reports", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, 0, fmt.Errorf("failed to list reports: %w", err)
	}

	return roads, total, nil
}

// UpdateReportStatus updates the status of a damaged road report
func (s *ReportServiceImpl) UpdateReportStatus(
	ctx context.Context,
	id uuid.UUID,
	newStatus entities.Status,
	requesterID uuid.UUID,
) (*entities.DamagedRoad, error) {
	logger.InfoContext(ctx, "Updating report status", map[string]interface{}{
		"report_id":    id.String(),
		"new_status":   newStatus.String(),
		"requester_id": requesterID.String(),
	})

	// Get the existing report
	road, err := s.repo.FindByID(ctx, id)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to retrieve report for status update", map[string]interface{}{
			"report_id": id.String(),
			"error":     err.Error(),
		})
		return nil, fmt.Errorf("failed to get report: %w", err)
	}

	if road == nil {
		return nil, errors.ErrReportNotFound
	}

	// Update the status (entity validates transition)
	if err := road.UpdateStatus(newStatus); err != nil {
		logger.WarnContext(ctx, "Invalid status transition attempted", map[string]interface{}{
			"report_id":   id.String(),
			"from_status": road.Status.String(),
			"to_status":   newStatus.String(),
			"error":       err.Error(),
		})
		return nil, err
	}

	// Save the updated status
	if err := s.repo.UpdateStatus(ctx, id, newStatus); err != nil {
		logger.ErrorContext(ctx, "Failed to save status update", map[string]interface{}{
			"report_id": id.String(),
			"error":     err.Error(),
		})
		return nil, fmt.Errorf("failed to update status: %w", err)
	}

	logger.InfoContext(ctx, "Successfully updated report status", map[string]interface{}{
		"report_id":  id.String(),
		"new_status": newStatus.String(),
	})

	return road, nil
}

// DeleteReport deletes a damaged road report
func (s *ReportServiceImpl) DeleteReport(ctx context.Context, id uuid.UUID, requesterID uuid.UUID) error {
	logger.InfoContext(ctx, "Deleting damaged road report", map[string]interface{}{
		"report_id":    id.String(),
		"requester_id": requesterID.String(),
	})

	// Get the existing report to check authorization
	road, err := s.repo.FindByID(ctx, id)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to retrieve report for deletion", map[string]interface{}{
			"report_id": id.String(),
			"error":     err.Error(),
		})
		return fmt.Errorf("failed to get report: %w", err)
	}

	if road == nil {
		return errors.ErrReportNotFound
	}

	// Check if requester is authorized to delete
	if !road.CanBeEditedBy(requesterID) {
		logger.WarnContext(ctx, "Unauthorized deletion attempt", map[string]interface{}{
			"report_id":    id.String(),
			"requester_id": requesterID.String(),
			"author_id":    road.AuthorID.String(),
		})
		return errors.ErrUnauthorizedAccess
	}

	// Delete the report
	if err := s.repo.Delete(ctx, id); err != nil {
		logger.ErrorContext(ctx, "Failed to delete report", map[string]interface{}{
			"report_id": id.String(),
			"error":     err.Error(),
		})
		return fmt.Errorf("failed to delete report: %w", err)
	}

	logger.InfoContext(ctx, "Successfully deleted damaged road report", map[string]interface{}{
		"report_id": id.String(),
	})

	return nil
}
