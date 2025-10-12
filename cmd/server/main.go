package main

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/nicklaros/jalanrusak-be/adapters/in/http/handlers"
	"github.com/nicklaros/jalanrusak-be/adapters/in/http/routes"
	"github.com/nicklaros/jalanrusak-be/adapters/out/messaging"
	"github.com/nicklaros/jalanrusak-be/adapters/out/repository/postgres"
	"github.com/nicklaros/jalanrusak-be/adapters/out/security"
	"github.com/nicklaros/jalanrusak-be/config"
	"github.com/nicklaros/jalanrusak-be/core/ports/external"
	"github.com/nicklaros/jalanrusak-be/core/services"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database connection
	db, err := sql.Open("postgres", cfg.Database.URL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test database connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("✓ Connected to database")

	// Initialize repositories (driven adapters)
	userRepo := postgres.NewUserRepository(db)
	refreshTokenRepo := postgres.NewRefreshTokenRepository(db)
	passwordResetTokenRepo := postgres.NewPasswordResetTokenRepository(db)
	authEventLogRepo := postgres.NewAuthEventLogRepository(db)

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

	// Initialize handlers (driving adapters)
	registrationHandler := handlers.NewRegistrationHandler(userService)
	authHandler := handlers.NewAuthHandler(authService, userService, int(cfg.JWT.AccessTokenTTL.Hours()))
	passwordHandler := handlers.NewPasswordHandler(passwordService)

	// Setup Gin router
	router := gin.Default()

	// Configure routes
	routes.SetupRoutes(router, registrationHandler, authHandler, passwordHandler, authService)

	// Start server
	addr := fmt.Sprintf(":%s", cfg.Server.Port)
	log.Printf("🚀 Server starting on %s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
