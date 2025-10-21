package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nicklaros/jalanrusak-be/adapters/in/http/dto"
	"github.com/nicklaros/jalanrusak-be/adapters/in/http/middleware"
	"github.com/nicklaros/jalanrusak-be/core/domain/entities"
	"github.com/nicklaros/jalanrusak-be/core/ports/external"
	"github.com/nicklaros/jalanrusak-be/core/ports/usecases"
)

// ValidationHandler handles location validation endpoints
type ValidationHandler struct {
	geometryService usecases.GeometryService
	photoValidator  external.PhotoValidator
}

// NewValidationHandler creates a new ValidationHandler
func NewValidationHandler(geometryService usecases.GeometryService, photoValidator external.PhotoValidator) *ValidationHandler {
	return &ValidationHandler{
		geometryService: geometryService,
		photoValidator:  photoValidator,
	}
}

// ValidateLocation validates coordinates before report submission
// @Summary Validate location coordinates
// @Description Pre-submission validation to check if coordinates fall within Indonesian boundaries and near the specified subdistrict centroid
// @Tags validation
// @Accept json
// @Produce json
// @Param request body dto.ValidateLocationRequest true "Location validation request"
// @Success 200 {object} dto.ValidateLocationResponse "Validation result"
// @Failure 400 {object} dto.ErrorResponse "Invalid request"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /api/v1/validate-location [post]
func (h *ValidationHandler) ValidateLocation(c *gin.Context) {
	var req dto.ValidateLocationRequest
	if !middleware.BindAndValidate(c, &req) {
		return
	}

	// Parse subdistrict code
	subdistrictCode, err := entities.NewSubDistrictCode(req.SubDistrictCode)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid subdistrict code format",
			Message: err.Error(),
		})
		return
	}

	// Convert DTO points to entity points
	points := make([]entities.Point, len(req.PathPoints))
	for i, pointDTO := range req.PathPoints {
		point, err := entities.NewPoint(pointDTO.Lat, pointDTO.Lng)
		if err != nil {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error:   "Invalid coordinates",
				Message: err.Error(),
			})
			return
		}
		points[i] = *point
	}

	response := dto.ValidateLocationResponse{
		Valid:             true,
		Message:           "Coordinates are valid",
		SubDistrictExists: false,
		WithinBoundaries:  false,
		NearCentroid:      false,
	}

	// Check if coordinates are within Indonesian boundaries
	if err := h.geometryService.ValidateCoordinatesInBoundary(points); err != nil {
		response.Valid = false
		response.Message = "Coordinates outside Indonesian boundaries"
		response.WithinBoundaries = false
		c.JSON(http.StatusOK, response)
		return
	}
	response.WithinBoundaries = true

	// Get subdistrict centroid
	centroid, err := h.geometryService.GetSubDistrictCentroid(subdistrictCode)
	if err != nil {
		response.Valid = false
		response.Message = "Subdistrict code not found in boundary dataset"
		response.SubDistrictExists = false
		c.JSON(http.StatusOK, response)
		return
	}
	response.SubDistrictExists = true
	response.CentroidLat = centroid.Lat
	response.CentroidLng = centroid.Lng

	// Calculate minimum distance to centroid
	minDistance := -1.0
	for _, point := range points {
		distance := h.geometryService.CalculateDistance(point, centroid)
		if minDistance < 0 || distance < minDistance {
			minDistance = distance
		}
	}
	response.MinDistanceToCenter = minDistance

	// Check if at least one coordinate is within 200 meters of centroid
	if err := h.geometryService.ValidateCoordinatesNearCentroid(points, subdistrictCode, 200.0); err != nil {
		response.Valid = false
		response.Message = "No coordinate within 200 meters of subdistrict centroid"
		response.NearCentroid = false
		c.JSON(http.StatusOK, response)
		return
	}
	response.NearCentroid = true

	c.JSON(http.StatusOK, response)
}

// ValidatePhotos validates photo URLs with SSRF protection
// @Summary Validate photo URLs
// @Description Pre-submission validation to check if photo URLs are accessible, have valid image content types, and pass SSRF protection checks
// @Tags validation
// @Accept json
// @Produce json
// @Param request body dto.ValidatePhotosRequest true "Photo validation request"
// @Success 200 {object} dto.ValidatePhotosResponse "Validation results"
// @Failure 400 {object} dto.ErrorResponse "Invalid request"
// @Security BearerAuth
// @Router /api/v1/validate-photos [post]
func (h *ValidationHandler) ValidatePhotos(c *gin.Context) {
	var req dto.ValidatePhotosRequest
	if !middleware.BindAndValidate(c, &req) {
		return
	}

	// Validate photo URLs using PhotoValidator
	validationResults := h.photoValidator.ValidateURLs(req.PhotoURLs)

	// Convert external.PhotoValidationResult to dto.PhotoValidationResult
	dtoResults := make([]dto.PhotoValidationResult, len(validationResults))
	allValid := true
	for i, result := range validationResults {
		dtoResults[i] = dto.PhotoValidationResult{
			URL:         result.URL,
			Valid:       result.Valid,
			Error:       result.Error,
			ContentType: result.ContentType,
			SizeBytes:   result.SizeBytes,
		}
		if !result.Valid {
			allValid = false
		}
	}

	response := dto.ValidatePhotosResponse{
		AllValid: allValid,
		Results:  dtoResults,
	}

	c.JSON(http.StatusOK, response)
}
