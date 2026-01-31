package handlers

import (
	"errors"
	"net/http"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ownafarm/ownafarm-backend/internal/dto/response"
	"github.com/ownafarm/ownafarm-backend/internal/repositories"
	"gorm.io/gorm"
)

type UserHandler struct {
	repo repositories.UserRepository
}

func NewUserHandler(repo repositories.UserRepository) *UserHandler {
	return &UserHandler{repo: repo}
}

var uuidRegex = regexp.MustCompile("^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$")

func (h *UserHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	// Validate UUID format
	if !uuidRegex.MatchString(id) {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid user ID format",
		})
		return
	}

	// Regenerate water before returning user data
	user, err := h.repo.RegenerateWater(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"status":  "error",
				"message": "User not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Internal server error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   user,
	})
}

// GetWaterBalance returns the current water balance for the authenticated user
func (h *UserHandler) GetWaterBalance(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "error",
			"message": "Unauthorized",
		})
		return
	}

	// Regenerate water and get fresh user data
	user, err := h.repo.RegenerateWater(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"status":  "error",
				"message": "User not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Internal server error",
		})
		return
	}

	// Calculate next regeneration time
	var nextRegenAt *time.Time
	if user.WaterPoints < repositories.MaxWaterPoints && user.LastRegenAt != nil {
		next := user.LastRegenAt.Add(time.Duration(repositories.WaterRegenRateMinutes) * time.Minute)
		nextRegenAt = &next
	}

	resp := response.WaterBalanceResponse{
		WaterPoints:    user.WaterPoints,
		MaxWaterPoints: repositories.MaxWaterPoints,
		NextRegenAt:    nextRegenAt,
		RegenRate:      repositories.WaterRegenRateMinutes,
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   resp,
	})
}

