package main

import (
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/ownafarm/ownafarm-backend/internal/config"
	"github.com/ownafarm/ownafarm-backend/internal/database"
	"github.com/ownafarm/ownafarm-backend/internal/handlers"
	"github.com/ownafarm/ownafarm-backend/internal/middleware"
	"github.com/ownafarm/ownafarm-backend/internal/repositories"
	"github.com/ownafarm/ownafarm-backend/internal/routes"
	"github.com/ownafarm/ownafarm-backend/internal/services"
	"github.com/ownafarm/ownafarm-backend/internal/utils"
)

func main() {
	// Load config
	cfg := config.LoadConfig()

	// Validate JWT secret
	if cfg.JWT.Secret == "" {
		log.Fatal("JWT_SECRET environment variable is required")
	}

	// 1. Connect to database
	err := database.Connect(&cfg.DB)
	if err != nil {
		panic(err)
	}

	// 2. Connect to Valkey
	err = database.ConnectValkey(&cfg.Valkey)
	if err != nil {
		panic(err)
	}
	defer database.CloseValkey()

	// 3. Setup router
	router := gin.Default()

	// 4. Setup CORS - Allow all origins (temporary)
	router.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// 5. Initialize Utils
	jwtUtil := utils.NewJWTUtil(&cfg.JWT)
	adminJwtUtil := utils.NewAdminJWTUtil(&cfg.JWT)

	// 6. Initialize Storage Service
	storageService, err := services.NewR2StorageService(&cfg.R2)
	if err != nil {
		log.Fatal("Failed to initialize R2 storage:", err)
	}

	// 7. Initialize Services
	nonceService := services.NewNonceService(database.Valkey, &cfg.Auth)
	authService := services.NewAuthService(&cfg.Auth)
	rateLimitService := services.NewRateLimitService(database.Valkey)

	// 8. Initialize Repositories
	userRepo := repositories.NewUserRepository(database.DB)
	farmerRepo := repositories.NewFarmerRepository(database.DB)
	adminUserRepo := repositories.NewAdminUserRepository(database.DB)
	auditLogRepo := repositories.NewAuditLogRepository(database.DB)

	// 9. Initialize Farmer Service
	farmerService := services.NewFarmerService(farmerRepo, storageService, auditLogRepo)
	adminAuthService := services.NewAdminAuthService(
		adminUserRepo,
		rateLimitService,
		adminJwtUtil,
		authService,
		database.Valkey,
		cfg.Auth.NonceTTLMinutes,
	)

	// 10. Initialize Handlers
	userHandler := handlers.NewUserHandler(userRepo)
	authHandler := handlers.NewAuthHandler(userRepo, nonceService, authService, jwtUtil)
	farmerHandler := handlers.NewFarmerHandler(farmerService)
	adminAuthHandler := handlers.NewAdminAuthHandler(adminAuthService)

	// 11. Initialize Middleware
	authMiddleware := middleware.NewAuthMiddleware(jwtUtil)
	adminAuthMiddleware := middleware.NewAdminAuthMiddleware(adminJwtUtil)

	// 12. Routes
	routes.SetupRoutes(router, userHandler, authHandler, farmerHandler, adminAuthHandler, authMiddleware, adminAuthMiddleware)

	// 13. Run the server
	err = router.Run(":" + cfg.App.Port)
	if err != nil {
		panic(err)
	}
}
