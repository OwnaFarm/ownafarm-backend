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

	// 6. Initialize Services
	nonceService := services.NewNonceService(database.Valkey, &cfg.Auth)
	authService := services.NewAuthService(&cfg.Auth)

	// 7. Initialize Repository
	userRepo := repositories.NewUserRepository(database.DB)

	// 8. Initialize Handlers
	userHandler := handlers.NewUserHandler(userRepo)
	authHandler := handlers.NewAuthHandler(userRepo, nonceService, authService, jwtUtil)

	// 9. Initialize Storage Service
	storageService, err := services.NewR2StorageService(&cfg.R2)
	if err != nil {
		log.Fatal("Failed to initialize R2 storage:", err)
	}
	_ = storageService // TODO: Wire to handlers when needed

	// 10. Initialize Middleware
	authMiddleware := middleware.NewAuthMiddleware(jwtUtil)

	// 11. Routes
	routes.SetupRoutes(router, userHandler, authHandler, authMiddleware)

	// 12. Run the server
	err = router.Run(":" + cfg.App.Port)
	if err != nil {
		panic(err)
	}
}
