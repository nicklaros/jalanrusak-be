package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

// HealthHandler handles health check endpoints
type HealthHandler struct {
	db *sqlx.DB
}

// NewHealthHandler creates a new HealthHandler
func NewHealthHandler(db *sqlx.DB) *HealthHandler {
	return &HealthHandler{db: db}
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string            `json:"status" example:"healthy"`
	Uptime    string            `json:"uptime" example:"1h23m45s"`
	Checks    map[string]string `json:"checks"`
	Timestamp string            `json:"timestamp" example:"2025-10-20T03:55:00Z"`
}

var startTime = time.Now()

// HealthCheck returns the health status of the application
// @Summary Health check
// @Description Returns the health status of the application and its dependencies
// @Tags health
// @Produce json
// @Success 200 {object} HealthResponse "Service is healthy"
// @Failure 503 {object} HealthResponse "Service is unhealthy"
// @Router /health [get]
func (h *HealthHandler) HealthCheck(c *gin.Context) {
	checks := make(map[string]string)
	overallStatus := "healthy"

	// Check database connection
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := h.db.PingContext(ctx); err != nil {
		checks["database"] = "unhealthy: " + err.Error()
		overallStatus = "unhealthy"
	} else {
		checks["database"] = "healthy"
	}

	// Calculate uptime
	uptime := time.Since(startTime).Round(time.Second)

	response := HealthResponse{
		Status:    overallStatus,
		Uptime:    uptime.String(),
		Checks:    checks,
		Timestamp: time.Now().Format(time.RFC3339),
	}

	statusCode := http.StatusOK
	if overallStatus == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, response)
}
