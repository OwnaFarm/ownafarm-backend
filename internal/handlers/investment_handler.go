package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ownafarm/ownafarm-backend/internal/dto/request"
	"github.com/ownafarm/ownafarm-backend/internal/services"
)

// InvestmentHandler handles investment-related HTTP requests
type InvestmentHandler struct {
	investmentService services.InvestmentServiceInterface
}

// NewInvestmentHandler creates a new InvestmentHandler instance
func NewInvestmentHandler(investmentService services.InvestmentServiceInterface) *InvestmentHandler {
	return &InvestmentHandler{investmentService: investmentService}
}

// SyncInvestments syncs investments from blockchain to database
// POST /crops/sync
func (h *InvestmentHandler) SyncInvestments(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	walletAddress, exists := c.Get("wallet_address")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Wallet address not found"})
		return
	}

	var req request.SyncInvestmentsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Optional body, so we ignore binding errors
		req = request.SyncInvestmentsRequest{}
	}

	resp, err := h.investmentService.SyncInvestments(c.Request.Context(), userID.(string), walletAddress.(string), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ListCrops lists all crops for the authenticated user
// GET /crops
func (h *InvestmentHandler) ListCrops(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req request.ListCropsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.investmentService.ListCrops(c.Request.Context(), userID.(string), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetCrop retrieves a single crop by ID
// GET /crops/:id
func (h *InvestmentHandler) GetCrop(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	cropID := c.Param("id")
	if cropID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Crop ID is required"})
		return
	}

	resp, err := h.investmentService.GetCrop(c.Request.Context(), userID.(string), cropID)
	if err != nil {
		if errors.Is(err, services.ErrInvestmentNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Crop not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// WaterCrop waters a crop (game mechanic for XP gain)
// POST /crops/:id/water
func (h *InvestmentHandler) WaterCrop(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	cropID := c.Param("id")
	if cropID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Crop ID is required"})
		return
	}

	resp, err := h.investmentService.WaterCrop(c.Request.Context(), userID.(string), cropID)
	if err != nil {
		if errors.Is(err, services.ErrInvestmentNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Crop not found"})
			return
		}
		if errors.Is(err, services.ErrNotEnoughWater) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Not enough water points"})
			return
		}
		if errors.Is(err, services.ErrAlreadyHarvested) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Crop already harvested"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// SyncHarvest syncs harvest status from blockchain
// POST /crops/:id/harvest/sync
func (h *InvestmentHandler) SyncHarvest(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	walletAddress, exists := c.Get("wallet_address")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Wallet address not found"})
		return
	}

	cropID := c.Param("id")
	if cropID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Crop ID is required"})
		return
	}

	resp, err := h.investmentService.SyncHarvest(c.Request.Context(), userID.(string), walletAddress.(string), cropID)
	if err != nil {
		if errors.Is(err, services.ErrInvestmentNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Crop not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}
