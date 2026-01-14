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
	farmHandler *handlers.FarmHandler,
	invoiceHandler *handlers.InvoiceHandler,
	investmentHandler *handlers.InvestmentHandler,
	authMiddleware *middleware.AuthMiddleware,
	adminAuthMiddleware *middleware.AdminAuthMiddleware,
	farmerAuthMiddleware *middleware.FarmerAuthMiddleware,
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

		// Marketplace routes (investor can browse available invoices)
		protected.GET("/marketplace/invoices", invoiceHandler.ListMarketplace)

		// Crop/Investment routes
		protected.POST("/crops/sync", investmentHandler.SyncInvestments)
		protected.GET("/crops", investmentHandler.ListCrops)
		protected.GET("/crops/:id", investmentHandler.GetCrop)
		protected.POST("/crops/:id/water", investmentHandler.WaterCrop)
		protected.POST("/crops/:id/harvest/sync", investmentHandler.SyncHarvest)
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
		// Farmer management
		admin.GET("/farmers", farmerHandler.GetListForAdmin)
		admin.PATCH("/farmers/:id/approve", farmerHandler.ApproveFarmer)
		admin.PATCH("/farmers/:id/reject", farmerHandler.RejectFarmer)

		// Invoice management
		admin.GET("/invoices", invoiceHandler.ListForAdmin)
		admin.GET("/invoices/:id", invoiceHandler.GetByIDForAdmin)
		admin.PATCH("/invoices/:id/approve", invoiceHandler.ApproveInvoice)
		admin.PATCH("/invoices/:id/reject", invoiceHandler.RejectInvoice)
	}

	// Farmer routes (requires farmer auth - approved farmers only)
	farmer := router.Group("/farmer")
	farmer.Use(farmerAuthMiddleware.FarmerAuthRequired())
	{
		// Farm management
		farmer.POST("/farms", farmHandler.Create)
		farmer.GET("/farms", farmHandler.List)
		farmer.GET("/farms/:id", farmHandler.GetByID)
		farmer.PUT("/farms/:id", farmHandler.Update)
		farmer.DELETE("/farms/:id", farmHandler.Delete)

		// Invoice management
		farmer.POST("/invoices", invoiceHandler.Create)
		farmer.GET("/invoices", invoiceHandler.List)
		farmer.GET("/invoices/:id", invoiceHandler.GetByID)
		farmer.POST("/invoices/image/presign", invoiceHandler.GetPresignedImageURL)
	}
}
