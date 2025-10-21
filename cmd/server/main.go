package main

import (
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/nicklaros/jalanrusak-be/adapters/in/http/handlers"
	"github.com/nicklaros/jalanrusak-be/adapters/in/http/middleware"
	"github.com/nicklaros/jalanrusak-be/adapters/in/http/routes"
	"github.com/nicklaros/jalanrusak-be/adapters/out/messaging"
	"github.com/nicklaros/jalanrusak-be/adapters/out/repository/postgres"
	"github.com/nicklaros/jalanrusak-be/adapters/out/security"
	outServices "github.com/nicklaros/jalanrusak-be/adapters/out/services"
	"github.com/nicklaros/jalanrusak-be/config"
	"github.com/nicklaros/jalanrusak-be/core/ports/external"
	"github.com/nicklaros/jalanrusak-be/core/services"
	docs "github.com/nicklaros/jalanrusak-be/docs"
	"github.com/ulule/limiter/v3"
)

// @title Jalanrusak API
// @version 1.0
// @description API documentation for the Jalanrusak backend service.
// @Schemes http
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and the JWT token.
func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database connection with PostGIS support
	dbConfig := postgres.ConnectionConfig{
		Host:            cfg.Database.Host,
		Port:            cfg.Database.Port,
		User:            cfg.Database.User,
		Password:        cfg.Database.Password,
		DBName:          cfg.Database.DBName,
		SSLMode:         cfg.Database.SSLMode,
		MaxOpenConns:    cfg.Database.MaxOpenConns,
		MaxIdleConns:    cfg.Database.MaxIdleConns,
		ConnMaxLifetime: cfg.Database.ConnMaxLifetime,
	}

	db, err := postgres.NewConnection(dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer postgres.Close(db)
	log.Println("✓ Connected to database with PostGIS support")

	// Initialize repositories (driven adapters)
	userRepo := postgres.NewUserRepository(db.DB)
	refreshTokenRepo := postgres.NewRefreshTokenRepository(db.DB)
	passwordResetTokenRepo := postgres.NewPasswordResetTokenRepository(db.DB)
	authEventLogRepo := postgres.NewAuthEventLogRepository(db.DB)
	damagedRoadRepo := postgres.NewDamagedRoadRepository(db)

	// Initialize security adapters
	passwordHasher := security.NewBcryptHasher(12) // cost 12 for production
	tokenGenerator := security.NewJWTTokenGenerator(cfg.JWT.Secret, int(cfg.JWT.AccessTokenTTL.Hours()))

	// Initialize messaging adapters
	var emailService external.EmailService
	if cfg.Email.ServiceType == "smtp" {
		// TODO: Implement SMTP email service
		log.Println("⚠️  SMTP email service not yet implemented, falling back to console")
		emailService = messaging.NewConsoleEmailService()
	} else {
		emailService = messaging.NewConsoleEmailService()
	}

	// Initialize services (core business logic)
	userService := services.NewUserService(userRepo, passwordHasher, authEventLogRepo)
	authService := services.NewAuthService(
		userRepo,
		refreshTokenRepo,
		passwordHasher,
		tokenGenerator,
		authEventLogRepo,
		int(cfg.JWT.RefreshTokenTTL.Hours()/24), // convert to days
	)
	passwordService := services.NewPasswordService(
		userRepo,
		passwordResetTokenRepo,
		passwordHasher,
		tokenGenerator,
		emailService,
		authEventLogRepo,
	)

	// Initialize boundary repository and geometry service
	boundaryRepo := postgres.NewBoundaryRepository(db)
	geometryService := services.NewGeometryService(boundaryRepo)

	// Initialize photo validator with SSRF protection
	photoValidator := outServices.NewPhotoValidator()

	// Initialize report service with geometry and photo validation
	reportService := services.NewReportService(damagedRoadRepo, geometryService, photoValidator)

	// Initialize handlers (driving adapters)
	registrationHandler := handlers.NewRegistrationHandler(userService)
	authHandler := handlers.NewAuthHandler(authService, userService, int(cfg.JWT.AccessTokenTTL.Hours()))
	passwordHandler := handlers.NewPasswordHandler(passwordService)
	reportHandler := handlers.NewReportHandler(reportService)
	validationHandler := handlers.NewValidationHandler(geometryService, photoValidator)
	healthHandler := handlers.NewHealthHandler(db)

	// Setup Gin router without default middleware
	router := gin.New()

	// Add custom middleware
	router.Use(gin.Recovery())                        // Panic recovery
	router.Use(middleware.RequestIDMiddleware())      // Request ID tracking
	router.Use(middleware.RequestLoggingMiddleware()) // Structured logging

	// Configure CORS
	router.Use(middleware.CORSMiddleware())

	// Apply rate limiting to API routes
	router.Use(middleware.RateLimitMiddleware(limiter.Rate{
		Period: 1 * time.Minute,
		Limit:  100, // 100 requests per minute per IP
	}))

	docs.SwaggerInfo.BasePath = "/api/v1"
	docs.SwaggerInfo.Host = fmt.Sprintf("localhost:%s", cfg.Server.Port)
	docs.SwaggerInfo.Schemes = []string{"http"}

	// Configure routes
	routes.SetupRoutes(router, registrationHandler, authHandler, passwordHandler, reportHandler, validationHandler, healthHandler, authService)

	// Start server
	addr := fmt.Sprintf(":%s", cfg.Server.Port)
	log.Printf("🚀 Server starting on %s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
