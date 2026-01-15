package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ownafarm/ownafarm-backend/internal/utils"
)

const (
	AuthorizationHeader = "Authorization"
	BearerPrefix        = "Bearer "
	ContextKeyUserID    = "user_id"
	ContextKeyWallet    = "wallet_address"
)

type AuthMiddleware struct {
	jwtUtil *utils.JWTUtil
}

func NewAuthMiddleware(jwtUtil *utils.JWTUtil) *AuthMiddleware {
	return &AuthMiddleware{jwtUtil: jwtUtil}
}

func (m *AuthMiddleware) AuthRequired() gin.HandlerFunc {
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

		// Validate token
		claims, err := m.jwtUtil.ValidateToken(tokenString)
		if err != nil {
			if errors.Is(err, utils.ErrExpiredToken) {
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

		// Set user info in context
		c.Set(ContextKeyUserID, claims.UserID)
		c.Set(ContextKeyWallet, claims.WalletAddress)

		c.Next()
	}
}

// GetUserID retrieves the user ID from the context
func GetUserID(c *gin.Context) (string, bool) {
	userID, exists := c.Get(ContextKeyUserID)
	if !exists {
		return "", false
	}
	return userID.(string), true
}

// GetWalletAddress retrieves the wallet address from the context
func GetWalletAddress(c *gin.Context) (string, bool) {
	wallet, exists := c.Get(ContextKeyWallet)
	if !exists {
		return "", false
	}
	return wallet.(string), true
}
