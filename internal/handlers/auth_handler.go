package handlers

import (
	"errors"
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"
	"github.com/ownafarm/ownafarm-backend/internal/models"
	"github.com/ownafarm/ownafarm-backend/internal/repositories"
	"github.com/ownafarm/ownafarm-backend/internal/services"
	"github.com/ownafarm/ownafarm-backend/internal/utils"
	"gorm.io/gorm"
)

var walletAddressRegex = regexp.MustCompile("^0x[0-9a-fA-F]{40}$")

type AuthHandler struct {
	userRepo     repositories.UserRepository
	nonceService services.NonceServiceInterface
	authService  services.AuthServiceInterface
	jwtUtil      *utils.JWTUtil
}

func NewAuthHandler(
	userRepo repositories.UserRepository,
	nonceService services.NonceServiceInterface,
	authService services.AuthServiceInterface,
	jwtUtil *utils.JWTUtil,
) *AuthHandler {
	return &AuthHandler{
		userRepo:     userRepo,
		nonceService: nonceService,
		authService:  authService,
		jwtUtil:      jwtUtil,
	}
}

type GetNonceRequest struct {
	WalletAddress string `form:"wallet_address" binding:"required"`
}

type GetNonceResponse struct {
	Nonce   string `json:"nonce"`
	Message string `json:"message"`
}

func (h *AuthHandler) GetNonce(c *gin.Context) {
	var req GetNonceRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "wallet_address is required",
		})
		return
	}

	// Validate wallet address format
	if !walletAddressRegex.MatchString(req.WalletAddress) {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid wallet address format",
		})
		return
	}

	// Generate nonce
	nonce, err := h.nonceService.GenerateNonce(c.Request.Context(), req.WalletAddress)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to generate nonce",
		})
		return
	}

	// Build sign message
	message := h.nonceService.BuildSignMessage(nonce)

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": GetNonceResponse{
			Nonce:   nonce,
			Message: message,
		},
	})
}

type LoginRequest struct {
	WalletAddress string `json:"wallet_address" binding:"required"`
	Signature     string `json:"signature" binding:"required"`
	Nonce         string `json:"nonce" binding:"required"`
}

type LoginResponse struct {
	Token string       `json:"token"`
	User  *models.User `json:"user"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid request body",
		})
		return
	}

	// Validate wallet address format
	if !walletAddressRegex.MatchString(req.WalletAddress) {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid wallet address format",
		})
		return
	}

	// Validate nonce
	if err := h.nonceService.ValidateAndDeleteNonce(c.Request.Context(), req.WalletAddress, req.Nonce); err != nil {
		if errors.Is(err, services.ErrNonceNotFound) || errors.Is(err, services.ErrNonceMismatch) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"status":  "error",
				"message": "Invalid or expired nonce",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to validate nonce",
		})
		return
	}

	// Build message and verify signature
	message := h.nonceService.BuildSignMessage(req.Nonce)
	if err := h.authService.VerifySignature(req.WalletAddress, req.Signature, message); err != nil {
		if errors.Is(err, services.ErrInvalidSignature) ||
			errors.Is(err, services.ErrInvalidWalletFormat) ||
			errors.Is(err, services.ErrSignatureMismatch) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"status":  "error",
				"message": "Invalid signature",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to verify signature",
		})
		return
	}

	// Normalize wallet address
	normalizedAddress := h.authService.NormalizeWalletAddress(req.WalletAddress)

	// Get or create a user
	user, err := h.userRepo.GetByWalletAddress(normalizedAddress)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Create a new user
			user = &models.User{
				WalletAddress: normalizedAddress,
			}
			if err := h.userRepo.Create(user); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"status":  "error",
					"message": "Failed to create user",
				})
				return
			}
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"message": "Failed to get user",
			})
			return
		}
	}

	// Update last login
	if err := h.userRepo.UpdateLastLogin(user.ID); err != nil {
		// Log error but don't fail the login
		// TODO: Add proper logging
	}

	// Generate JWT token
	token, err := h.jwtUtil.GenerateToken(user.ID, user.WalletAddress)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to generate token",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": LoginResponse{
			Token: token,
			User:  user,
		},
	})
}
