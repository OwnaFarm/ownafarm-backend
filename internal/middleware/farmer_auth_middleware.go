package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ownafarm/ownafarm-backend/internal/models"
	"github.com/ownafarm/ownafarm-backend/internal/repositories"
)

const (
	ContextKeyFarmerID = "farmer_id"
)

// FarmerAuthMiddleware provides authentication for approved farmers
type FarmerAuthMiddleware struct {
	authMiddleware *AuthMiddleware
	farmerRepo     repositories.FarmerRepository
}

// NewFarmerAuthMiddleware creates a new FarmerAuthMiddleware instance
func NewFarmerAuthMiddleware(authMiddleware *AuthMiddleware, farmerRepo repositories.FarmerRepository) *FarmerAuthMiddleware {
	return &FarmerAuthMiddleware{
		authMiddleware: authMiddleware,
		farmerRepo:     farmerRepo,
	}
}

// FarmerAuthRequired ensures the user is authenticated and has an approved farmer profile
func (m *FarmerAuthMiddleware) FarmerAuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		// First, run the base auth middleware
		m.authMiddleware.AuthRequired()(c)
		if c.IsAborted() {
			return
		}

		// Get user ID from context (set by base auth middleware)
		userID, exists := GetUserID(c)
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"status":  "error",
				"message": "User not authenticated",
			})
			return
		}

		// Get farmer by user ID
		farmer, err := m.farmerRepo.GetByUserID(userID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"status":  "error",
				"message": "No farmer profile found for this user",
			})
			return
		}

		// Check if farmer is approved
		if farmer.Status != models.FarmerStatusApproved {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"status":  "error",
				"message": "Farmer profile is not approved",
				"data": gin.H{
					"current_status": farmer.Status,
				},
			})
			return
		}

		// Set farmer ID in context
		c.Set(ContextKeyFarmerID, farmer.ID)

		c.Next()
	}
}

// GetFarmerID retrieves the farmer ID from the context
func GetFarmerID(c *gin.Context) (string, bool) {
	farmerID, exists := c.Get(ContextKeyFarmerID)
	if !exists {
		return "", false
	}
	return farmerID.(string), true
}
