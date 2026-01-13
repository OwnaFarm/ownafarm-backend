package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ownafarm/ownafarm-backend/internal/utils"
)

const (
	ContextKeyAdminID            = "admin_id"
	ContextKeyAdminWalletAddress = "admin_wallet_address"
	ContextKeyAdminRole          = "admin_role"
	ContextKeyIPAddress          = "ip_address"
	ContextKeyUserAgent          = "user_agent"
)

// AdminAuthMiddleware handles authentication for admin users
type AdminAuthMiddleware struct {
	adminJwtUtil *utils.AdminJWTUtil
}

// NewAdminAuthMiddleware creates a new AdminAuthMiddleware instance
func NewAdminAuthMiddleware(adminJwtUtil *utils.AdminJWTUtil) *AdminAuthMiddleware {
	return &AdminAuthMiddleware{adminJwtUtil: adminJwtUtil}
}

// AdminAuthRequired returns a middleware that requires admin authentication
func (m *AdminAuthMiddleware) AdminAuthRequired() gin.HandlerFunc {
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

		// Validate admin token
		claims, err := m.adminJwtUtil.ValidateToken(tokenString)
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

		// Set admin info in context
		c.Set(ContextKeyAdminID, claims.AdminID)
		c.Set(ContextKeyAdminWalletAddress, claims.WalletAddress)
		c.Set(ContextKeyAdminRole, claims.Role)

		// Set request metadata for audit logging
		c.Set(ContextKeyIPAddress, c.ClientIP())
		c.Set(ContextKeyUserAgent, c.GetHeader("User-Agent"))

		c.Next()
	}
}

// GetAdminID retrieves the admin ID from the context
func GetAdminID(c *gin.Context) (string, bool) {
	adminID, exists := c.Get(ContextKeyAdminID)
	if !exists {
		return "", false
	}
	return adminID.(string), true
}

// GetAdminWalletAddress retrieves the admin wallet address from the context
func GetAdminWalletAddress(c *gin.Context) (string, bool) {
	walletAddress, exists := c.Get(ContextKeyAdminWalletAddress)
	if !exists {
		return "", false
	}
	return walletAddress.(string), true
}

// GetAdminRole retrieves the admin role from the context
func GetAdminRole(c *gin.Context) (string, bool) {
	role, exists := c.Get(ContextKeyAdminRole)
	if !exists {
		return "", false
	}
	return role.(string), true
}

// GetIPAddress retrieves the client IP address from the context
func GetIPAddress(c *gin.Context) (string, bool) {
	ipAddress, exists := c.Get(ContextKeyIPAddress)
	if !exists {
		return "", false
	}
	return ipAddress.(string), true
}

// GetUserAgent retrieves the user agent from the context
func GetUserAgent(c *gin.Context) (string, bool) {
	userAgent, exists := c.Get(ContextKeyUserAgent)
	if !exists {
		return "", false
	}
	return userAgent.(string), true
}
