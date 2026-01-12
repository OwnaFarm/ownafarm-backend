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
	authMiddleware *middleware.AuthMiddleware,
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

	// Protected routes
	protected := router.Group("/")
	protected.Use(authMiddleware.AuthRequired())
	{
		protected.GET("/users/:id", userHandler.GetByID)
	}
}
