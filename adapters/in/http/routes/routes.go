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
	authService usecases.AuthService,
) {
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API v1 routes
	api := router.Group("/api")
	{
		// Auth routes (public)
		auth := api.Group("/auth")
		{
			auth.POST("/register", registrationHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)

			// Password reset (public)
			auth.POST("/password/reset-request", passwordHandler.RequestPasswordReset)
			auth.POST("/password/reset-confirm", passwordHandler.ResetPassword)
		}

		// Protected routes (require authentication)
		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware(authService))
		{
			protected.POST("/auth/logout", authHandler.Logout)
			protected.POST("/auth/password/change", passwordHandler.ChangePassword)
			// Add more protected routes here
		}
	}
}
