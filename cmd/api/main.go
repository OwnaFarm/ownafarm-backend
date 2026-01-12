package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ownafarm/ownafarm-backend/internal/config"
	"github.com/ownafarm/ownafarm-backend/internal/database"
	"github.com/ownafarm/ownafarm-backend/internal/models"
)

func main() {
	// Load config
	cfg := config.LoadConfig()

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

	// 2. Setup router
	router := gin.Default()

	// 3. Routes
	router.GET("/", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"message": "Hello World",
		})
	})

	router.GET("/users", func(ctx *gin.Context) {
		var users []models.User
		database.DB.Find(&users)
		ctx.JSON(http.StatusOK, gin.H{"data": users})
	})

	// 4. Run the server
	err = router.Run(":8080")
	if err != nil {
		panic(err)
	}
}
