package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ownafarm/ownafarm-backend/internal/dto/request"
	"github.com/ownafarm/ownafarm-backend/internal/dto/response"
	"github.com/ownafarm/ownafarm-backend/internal/services"
)

type AdminAuthHandler struct {
	adminAuthService *services.AdminAuthService
}

func NewAdminAuthHandler(adminAuthService *services.AdminAuthService) *AdminAuthHandler {
	return &AdminAuthHandler{
		adminAuthService: adminAuthService,
	}
}

// GetNonce handles nonce generation for admin wallet authentication
// @Summary Get admin nonce
// @Description Get a nonce for admin wallet signature authentication
// @Tags Admin Auth
// @Accept json
// @Produce json
// @Param wallet_address query string true "Admin wallet address"
// @Success 200 {object} response.AdminNonceResponse
// @Failure 400 {object} response.AdminLoginErrorResponse
// @Failure 500 {object} response.AdminLoginErrorResponse
// @Router /admin/auth/nonce [get]
func (h *AdminAuthHandler) GetNonce(c *gin.Context) {
	walletAddress := c.Query("wallet_address")
	if walletAddress == "" {
		c.JSON(http.StatusBadRequest, response.AdminLoginErrorResponse{
			Error: "wallet_address query parameter is required",
		})
		return
	}

	// Validate wallet address format
	if !walletAddressRegex.MatchString(walletAddress) {
		c.JSON(http.StatusBadRequest, response.AdminLoginErrorResponse{
			Error: "Invalid wallet address format",
		})
		return
	}

	result, err := h.adminAuthService.GetNonce(c.Request.Context(), walletAddress)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.AdminLoginErrorResponse{
			Error: "Failed to generate nonce",
		})
		return
	}

	c.JSON(http.StatusOK, response.AdminNonceResponse{
		Nonce:       result.Nonce,
		SignMessage: result.SignMessage,
	})
}

// Login handles admin login with wallet signature
// @Summary Admin login
// @Description Authenticate admin user with wallet signature
// @Tags Admin Auth
// @Accept json
// @Produce json
// @Param request body request.AdminWalletLoginRequest true "Admin wallet login credentials"
// @Success 200 {object} response.AdminLoginResponse
// @Failure 400 {object} response.AdminLoginErrorResponse
// @Failure 401 {object} response.AdminLoginErrorResponse
// @Failure 429 {object} response.AdminLoginErrorResponse
// @Failure 500 {object} response.AdminLoginErrorResponse
// @Router /admin/auth/login [post]
func (h *AdminAuthHandler) Login(c *gin.Context) {
	var req request.AdminWalletLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.AdminLoginErrorResponse{
			Error: "Invalid request: " + err.Error(),
		})
		return
	}

	result, rateLimitInfo, err := h.adminAuthService.Login(c.Request.Context(), req.WalletAddress, req.Signature, req.Nonce)
	if err != nil {
		// Handle rate limit exceeded
		if errors.Is(err, services.ErrRateLimitExceeded) {
			c.JSON(http.StatusTooManyRequests, response.AdminLoginErrorResponse{
				Error:             "Too many login attempts. Please try again later.",
				RetryAfterSeconds: &rateLimitInfo.RetryAfter,
			})
			return
		}

		// Handle invalid credentials (signature mismatch, invalid nonce)
		if errors.Is(err, services.ErrInvalidCredentials) {
			resp := response.AdminLoginErrorResponse{
				Error: "Invalid signature or nonce",
			}
			if rateLimitInfo != nil {
				resp.RemainingAttempts = &rateLimitInfo.Remaining
			}
			c.JSON(http.StatusUnauthorized, resp)
			return
		}

		// Handle admin not found
		if errors.Is(err, services.ErrAdminNotFound) {
			resp := response.AdminLoginErrorResponse{
				Error: "Admin account not found for this wallet address",
			}
			if rateLimitInfo != nil {
				resp.RemainingAttempts = &rateLimitInfo.Remaining
			}
			c.JSON(http.StatusUnauthorized, resp)
			return
		}

		// Handle inactive account
		if errors.Is(err, services.ErrAccountInactive) {
			c.JSON(http.StatusUnauthorized, response.AdminLoginErrorResponse{
				Error: "Account is inactive. Please contact support.",
			})
			return
		}

		// Handle other errors
		c.JSON(http.StatusInternalServerError, response.AdminLoginErrorResponse{
			Error: "Internal server error",
		})
		return
	}

	c.JSON(http.StatusOK, response.AdminLoginResponse{
		Token: result.Token,
		Admin: response.AdminInfo{
			ID:            result.Admin.ID,
			WalletAddress: result.Admin.WalletAddress,
			Role:          result.Admin.Role,
		},
	})
}
