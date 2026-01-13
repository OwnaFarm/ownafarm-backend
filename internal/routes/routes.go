package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ownafarm/ownafarm-backend/internal/handlers"
	"github.com/ownafarm/ownafarm-backend/internal/middleware"
)

func SetupRoutes(
	router *gin.Engine,
	userHandler *handlers.UserHandler,
	authHandler *handlers.AuthHandler,
	farmerHandler *handlers.FarmerHandler,
	adminAuthHandler *handlers.AdminAuthHandler,
	authMiddleware *middleware.AuthMiddleware,
	adminAuthMiddleware *middleware.AdminAuthMiddleware,
) {
	router.GET("/", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"message": "Hello World",
		})
	})

	// Auth routes (public)
	auth := router.Group("/auth")
	{
		auth.GET("/nonce", authHandler.GetNonce)
		auth.POST("/login", authHandler.Login)
	}

	// Farmer routes (public)
	farmers := router.Group("/farmers")
	{
		farmers.POST("/register", farmerHandler.Register)
		farmers.POST("/documents/presign", farmerHandler.GetPresignedURLs)
	}

	// Protected routes (investor auth)
	protected := router.Group("/")
	protected.Use(authMiddleware.AuthRequired())
	{
		protected.GET("/users/:id", userHandler.GetByID)
		// Farmer document download (requires auth)
		protected.GET("/farmers/:farmer_id/documents/:document_id/download", farmerHandler.GetDocumentDownloadURL)
	}

	// Admin auth routes (public)
	adminAuth := router.Group("/admin/auth")
	{
		adminAuth.GET("/nonce", adminAuthHandler.GetNonce)
		adminAuth.POST("/login", adminAuthHandler.Login)
	}

	// Admin routes (requires admin auth)
	admin := router.Group("/admin")
	admin.Use(adminAuthMiddleware.AdminAuthRequired())
	{
		admin.GET("/farmers", farmerHandler.GetListForAdmin)
		admin.PATCH("/farmers/:id/approve", farmerHandler.ApproveFarmer)
		admin.PATCH("/farmers/:id/reject", farmerHandler.RejectFarmer)
	}
}
