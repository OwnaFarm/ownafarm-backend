package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ownafarm/ownafarm-backend/internal/utils"
)

const (
	ContextKeyFarmerID     = "farmer_id"
	ContextKeyFarmerWallet = "farmer_wallet_address"
)

// FarmerAuthMiddleware provides authentication for approved farmers
type FarmerAuthMiddleware struct {
	jwtUtil *utils.FarmerJWTUtil
}

// NewFarmerAuthMiddleware creates a new FarmerAuthMiddleware instance
func NewFarmerAuthMiddleware(jwtUtil *utils.FarmerJWTUtil) *FarmerAuthMiddleware {
	return &FarmerAuthMiddleware{
		jwtUtil: jwtUtil,
	}
}

// FarmerAuthRequired ensures the request has a valid farmer JWT token
func (m *FarmerAuthMiddleware) FarmerAuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get Authorization header
		authHeader := c.GetHeader(AuthorizationHeader)
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"status":  "error",
				"message": "Authorization header is required",
			})
			return
		}

		// Check Bearer prefix
		if !strings.HasPrefix(authHeader, BearerPrefix) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"status":  "error",
				"message": "Invalid authorization header format",
			})
			return
		}

		// Extract token
		tokenString := strings.TrimPrefix(authHeader, BearerPrefix)
		if tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"status":  "error",
				"message": "Token is required",
			})
			return
		}

		// Validate farmer token
		claims, err := m.jwtUtil.ValidateToken(tokenString)
		if err != nil {
			if errors.Is(err, utils.ErrExpiredFarmerToken) {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"status":  "error",
					"message": "Token has expired",
				})
				return
			}
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"status":  "error",
				"message": "Invalid token",
			})
			return
		}

		// Set farmer info in context
		c.Set(ContextKeyFarmerID, claims.FarmerID)
		c.Set(ContextKeyFarmerWallet, claims.WalletAddress)

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

// GetFarmerWalletAddress retrieves the farmer wallet address from the context
func GetFarmerWalletAddress(c *gin.Context) (string, bool) {
	wallet, exists := c.Get(ContextKeyFarmerWallet)
	if !exists {
		return "", false
	}
	return wallet.(string), true
}
