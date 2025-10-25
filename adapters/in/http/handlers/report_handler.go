package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nicklaros/jalanrusak-be/adapters/in/http/dto"
	"github.com/nicklaros/jalanrusak-be/adapters/in/http/middleware"
	"github.com/nicklaros/jalanrusak-be/core/domain/entities"
	domainerrors "github.com/nicklaros/jalanrusak-be/core/domain/errors"
	"github.com/nicklaros/jalanrusak-be/core/ports/usecases"
)

// ReportHandler handles HTTP requests for damaged road reports
type ReportHandler struct {
	reportService usecases.ReportService
}

// NewReportHandler creates a new report handler
func NewReportHandler(reportService usecases.ReportService) *ReportHandler {
	return &ReportHandler{
		reportService: reportService,
	}
}

// CreateReport godoc
// @Summary Create a new damaged road report
// @Description Logged-in users can submit a new damaged road report with title, location coordinates, photos, and optional description
// @Tags Damaged Roads
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateDamagedRoadRequest true "Create damaged road request"
// @Success 201 {object} dto.DamagedRoadResponse "Report created successfully"
// @Failure 400 {object} dto.ErrorResponse "Bad request - validation errors"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized - authentication required"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /damaged-roads [post]
func (h *ReportHandler) CreateReport(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "unauthorized",
			Message: "User authentication required",
		})
		return
	}

	authorID, err := uuid.Parse(userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "internal_error",
			Message: "Invalid user ID format: " + err.Error(),
		})
		return
	}

	// Bind and validate request
	var req dto.CreateDamagedRoadRequest
	if !middleware.BindAndValidate(c, &req) {
		return
	}

	// Convert DTO to entities
	title, subdistrictCode, points, description, err := req.ToEntity()
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "validation_error",
			Message: err.Error(),
		})
		return
	}

	// Create the report
	road, err := h.reportService.CreateReport(
		c.Request.Context(),
		title,
		subdistrictCode,
		points,
		req.PhotoURLs,
		authorID,
		description,
	)

	if err != nil {
		// Handle validation errors
		var validationErr *domainerrors.ValidationError
		if errors.As(err, &validationErr) {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error:   "validation_error",
				Message: validationErr.Error(),
			})
			return
		}

		// Handle other errors
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to create report",
		})
		return
	}

	// Return created report
	response := dto.FromDamagedRoad(road)
	c.JSON(http.StatusCreated, response)
}

// GetReport godoc
// @Summary Get a specific damaged road report
// @Description Retrieve detailed information about a specific damaged road report
// @Tags Damaged Roads
// @Produce json
// @Security BearerAuth
// @Param id path string true "Report ID" format(uuid)
// @Success 200 {object} dto.DamagedRoadResponse "Report details"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 404 {object} dto.ErrorResponse "Report not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /damaged-roads/{id} [get]
func (h *ReportHandler) GetReport(c *gin.Context) {
	// Parse report ID from URL
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid report ID format",
		})
		return
	}

	// Get the report
	road, err := h.reportService.GetReport(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, domainerrors.ErrReportNotFound) {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error:   "not_found",
				Message: "Report not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to retrieve report",
		})
		return
	}

	// Return report
	response := dto.FromDamagedRoad(road)
	c.JSON(http.StatusOK, response)
}

// ListReports godoc
// @Summary List damaged road reports
// @Description Get paginated list of damaged road reports with optional filters
// @Tags Damaged Roads
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20) maximum(100)
// @Param status query string false "Filter by status"
// @Param subdistrict_code query string false "Filter by subdistrict code"
// @Success 200 {object} dto.DamagedRoadListResponse "List of reports"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /damaged-roads [get]
func (h *ReportHandler) ListReports(c *gin.Context) {
	// Parse pagination parameters
	page := 1
	if pageParam := c.Query("page"); pageParam != "" {
		if _, err := fmt.Sscanf(pageParam, "%d", &page); err != nil || page < 1 {
			page = 1
		}
	}

	limit := 20
	if limitParam := c.Query("limit"); limitParam != "" {
		if _, err := fmt.Sscanf(limitParam, "%d", &limit); err != nil || limit < 1 || limit > 100 {
			limit = 20
		}
	}

	offset := (page - 1) * limit

	// Build filters
	filters := entities.NewDamagedRoadFilters()
	filters.Limit = limit
	filters.Offset = offset

	// Status filter
	if statusParam := c.Query("status"); statusParam != "" {
		status := entities.Status(statusParam)
		if status.IsValid() {
			filters.Status = &status
		}
	}

	// Subdistrict code filter
	if subdistrictParam := c.Query("subdistrict_code"); subdistrictParam != "" {
		filters.SubDistrictCode = &subdistrictParam
	}

	// Get reports
	roads, total, err := h.reportService.ListReports(c.Request.Context(), filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to retrieve reports",
		})
		return
	}

	// Convert to DTOs
	responses := make([]dto.DamagedRoadResponse, len(roads))
	for i, road := range roads {
		responses[i] = dto.FromDamagedRoad(road)
	}

	// Return paginated response
	c.JSON(http.StatusOK, dto.DamagedRoadListResponse{
		Data: responses,
		Pagination: dto.PaginationMeta{
			Total:  total,
			Limit:  limit,
			Offset: offset,
			Page:   page,
		},
	})
}

// UpdateReportStatus godoc
// @Summary Update report status
// @Description Update the status of a damaged road report (for administrators/verificators)
// @Tags Damaged Roads
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Report ID" format(uuid)
// @Param request body dto.UpdateStatusRequest true "Update status request"
// @Success 200 {object} dto.DamagedRoadResponse "Status updated successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid status transition"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 403 {object} dto.ErrorResponse "Forbidden"
// @Failure 404 {object} dto.ErrorResponse "Report not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /damaged-roads/{id}/status [patch]
func (h *ReportHandler) UpdateReportStatus(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "unauthorized",
			Message: "User authentication required",
		})
		return
	}

	requesterID, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "internal_error",
			Message: "Invalid user ID format",
		})
		return
	}

	// Parse report ID
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid report ID format",
		})
		return
	}

	// Bind and validate request
	var req dto.UpdateStatusRequest
	if !middleware.BindAndValidate(c, &req) {
		return
	}

	// Validate status
	newStatus := entities.Status(req.Status)
	if !newStatus.IsValid() {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_status",
			Message: "Invalid status value",
		})
		return
	}

	// Update status
	road, err := h.reportService.UpdateReportStatus(c.Request.Context(), id, newStatus, requesterID)
	if err != nil {
		if errors.Is(err, domainerrors.ErrReportNotFound) {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error:   "not_found",
				Message: "Report not found",
			})
			return
		}

		var validationErr *domainerrors.ValidationError
		if errors.As(err, &validationErr) {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error:   "invalid_transition",
				Message: validationErr.Error(),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to update status",
		})
		return
	}

	// Return updated report
	response := dto.FromDamagedRoad(road)
	c.JSON(http.StatusOK, response)
}
