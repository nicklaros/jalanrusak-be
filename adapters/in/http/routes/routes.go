package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/nicklaros/jalanrusak-be/adapters/in/http/handlers"
	"github.com/nicklaros/jalanrusak-be/adapters/in/http/middleware"
	"github.com/nicklaros/jalanrusak-be/core/ports/usecases"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// SetupRoutes configures all HTTP routes
func SetupRoutes(
	router *gin.Engine,
	registrationHandler *handlers.RegistrationHandler,
	authHandler *handlers.AuthHandler,
	passwordHandler *handlers.PasswordHandler,
	reportHandler *handlers.ReportHandler,
	validationHandler *handlers.ValidationHandler,
	healthHandler *handlers.HealthHandler,
	authService usecases.AuthService,
) {
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Health check (public, no rate limit)
	router.GET("/health", healthHandler.HealthCheck)

	// API v1 routes
	apiV1 := router.Group("/api/v1")
	{
		// Auth routes (public)
		auth := apiV1.Group("/auth")
		{
			auth.POST("/register", registrationHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)

			// Password reset (public)
			auth.POST("/password/reset-request", passwordHandler.RequestPasswordReset)
			auth.POST("/password/reset-confirm", passwordHandler.ResetPassword)
		}

		// Protected routes (require authentication)
		protected := apiV1.Group("")
		protected.Use(middleware.AuthMiddleware(authService))
		{
			protected.POST("/auth/logout", authHandler.Logout)
			protected.POST("/auth/password/change", passwordHandler.ChangePassword)

			// Validation endpoints
			protected.POST("/validate-location", validationHandler.ValidateLocation)
			protected.POST("/validate-photos", validationHandler.ValidatePhotos)

			// Damaged road report routes
			protected.POST("/damaged-roads", reportHandler.CreateReport)
			protected.GET("/damaged-roads", reportHandler.ListReports)
			protected.GET("/damaged-roads/:id", reportHandler.GetReport)
			protected.PATCH("/damaged-roads/:id/status", reportHandler.UpdateReportStatus)
		}
	}
}
