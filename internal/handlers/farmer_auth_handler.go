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

var farmerWalletAddressRegex = regexp.MustCompile("^0x[0-9a-fA-F]{40}$")

// FarmerAuthHandler handles farmer authentication endpoints
type FarmerAuthHandler struct {
	farmerRepo         repositories.FarmerRepository
	farmerNonceService services.FarmerNonceServiceInterface
	authService        services.AuthServiceInterface
	farmerJwtUtil      *utils.FarmerJWTUtil
}

// NewFarmerAuthHandler creates a new FarmerAuthHandler instance
func NewFarmerAuthHandler(
	farmerRepo repositories.FarmerRepository,
	farmerNonceService services.FarmerNonceServiceInterface,
	authService services.AuthServiceInterface,
	farmerJwtUtil *utils.FarmerJWTUtil,
) *FarmerAuthHandler {
	return &FarmerAuthHandler{
		farmerRepo:         farmerRepo,
		farmerNonceService: farmerNonceService,
		authService:        authService,
		farmerJwtUtil:      farmerJwtUtil,
	}
}

// FarmerNonceResponse represents the nonce response for farmer auth
type FarmerNonceResponse struct {
	Nonce   string `json:"nonce"`
	Message string `json:"message"`
}

// FarmerLoginRequest represents the login request for farmer auth
type FarmerLoginRequest struct {
	WalletAddress string `json:"wallet_address" binding:"required"`
	Signature     string `json:"signature" binding:"required"`
	Nonce         string `json:"nonce" binding:"required"`
}

// FarmerLoginResponse represents the login response for farmer auth
type FarmerLoginResponse struct {
	Token  string                  `json:"token"`
	Farmer FarmerLoginResponseData `json:"farmer"`
}

// FarmerLoginResponseData contains farmer data in login response
type FarmerLoginResponseData struct {
	ID            string `json:"id"`
	WalletAddress string `json:"wallet_address"`
	FullName      string `json:"full_name"`
	Email         string `json:"email"`
}

// GetNonce handles nonce generation for farmer wallet authentication
// @Summary Get farmer nonce
// @Description Get a nonce for farmer wallet signature authentication
// @Tags Farmer Auth
// @Accept json
// @Produce json
// @Param wallet_address query string true "Farmer wallet address"
// @Success 200 {object} FarmerNonceResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /farmer/auth/nonce [get]
func (h *FarmerAuthHandler) GetNonce(c *gin.Context) {
	walletAddress := c.Query("wallet_address")
	if walletAddress == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "wallet_address query parameter is required",
		})
		return
	}

	// Validate wallet address format
	if !farmerWalletAddressRegex.MatchString(walletAddress) {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid wallet address format",
		})
		return
	}

	// Generate nonce
	nonce, err := h.farmerNonceService.GenerateNonce(c.Request.Context(), walletAddress)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to generate nonce",
		})
		return
	}

	// Build sign message
	message := h.farmerNonceService.BuildSignMessage(nonce)

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": FarmerNonceResponse{
			Nonce:   nonce,
			Message: message,
		},
	})
}

// Login handles farmer login with wallet signature
// @Summary Farmer login
// @Description Authenticate farmer with wallet signature (only approved farmers can login)
// @Tags Farmer Auth
// @Accept json
// @Produce json
// @Param request body FarmerLoginRequest true "Farmer wallet login credentials"
// @Success 200 {object} FarmerLoginResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /farmer/auth/login [post]
func (h *FarmerAuthHandler) Login(c *gin.Context) {
	var req FarmerLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid request body",
		})
		return
	}

	// Validate wallet address format
	if !farmerWalletAddressRegex.MatchString(req.WalletAddress) {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid wallet address format",
		})
		return
	}

	// Validate nonce
	if err := h.farmerNonceService.ValidateAndDeleteNonce(c.Request.Context(), req.WalletAddress, req.Nonce); err != nil {
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
	message := h.farmerNonceService.BuildSignMessage(req.Nonce)
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

	// Get farmer by wallet address
	farmer, err := h.farmerRepo.GetByWalletAddress(normalizedAddress)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"status":  "error",
				"message": "Farmer account not found. Please register first.",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to get farmer",
		})
		return
	}

	// Check if farmer is approved
	if farmer.Status != models.FarmerStatusApproved {
		c.JSON(http.StatusForbidden, gin.H{
			"status":  "error",
			"message": "Farmer account is not approved",
			"data": gin.H{
				"current_status": farmer.Status,
			},
		})
		return
	}

	// Generate JWT token
	token, err := h.farmerJwtUtil.GenerateToken(farmer.ID, farmer.WalletAddress)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to generate token",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": FarmerLoginResponse{
			Token: token,
			Farmer: FarmerLoginResponseData{
				ID:            farmer.ID,
				WalletAddress: farmer.WalletAddress,
				FullName:      farmer.FullName,
				Email:         farmer.Email,
			},
		},
	})
}
