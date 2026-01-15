package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ownafarm/ownafarm-backend/internal/dto/request"
	"github.com/ownafarm/ownafarm-backend/internal/middleware"
	"github.com/ownafarm/ownafarm-backend/internal/services"
)

// FarmHandler handles farm-related HTTP requests
type FarmHandler struct {
	farmService services.FarmServiceInterface
}

// NewFarmHandler creates a new FarmHandler instance
func NewFarmHandler(farmService services.FarmServiceInterface) *FarmHandler {
	return &FarmHandler{
		farmService: farmService,
	}
}

// Create handles farm creation
// POST /farmer/farms
func (h *FarmHandler) Create(c *gin.Context) {
	farmerID, exists := middleware.GetFarmerID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "error",
			"message": "Farmer not authenticated",
		})
		return
	}

	var req request.CreateFarmRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	resp, err := h.farmService.Create(c.Request.Context(), farmerID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to create farm",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status": "success",
		"data": gin.H{
			"farm": resp,
		},
	})
}

// GetByID handles getting a farm by ID
// GET /farmer/farms/:id
func (h *FarmHandler) GetByID(c *gin.Context) {
	farmerID, exists := middleware.GetFarmerID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "error",
			"message": "Farmer not authenticated",
		})
		return
	}

	farmID := c.Param("id")
	if farmID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Farm ID is required",
		})
		return
	}

	resp, err := h.farmService.GetByID(c.Request.Context(), farmerID, farmID)
	if err != nil {
		if errors.Is(err, services.ErrFarmNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"status":  "error",
				"message": "Farm not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to get farm",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"farm": resp,
		},
	})
}

// List handles listing farms for the authenticated farmer
// GET /farmer/farms
func (h *FarmHandler) List(c *gin.Context) {
	farmerID, exists := middleware.GetFarmerID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "error",
			"message": "Farmer not authenticated",
		})
		return
	}

	var req request.ListFarmRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid query parameters",
			"details": err.Error(),
		})
		return
	}

	resp, err := h.farmService.List(c.Request.Context(), farmerID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to list farms",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   resp,
	})
}

// Update handles updating a farm
// PUT /farmer/farms/:id
func (h *FarmHandler) Update(c *gin.Context) {
	farmerID, exists := middleware.GetFarmerID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "error",
			"message": "Farmer not authenticated",
		})
		return
	}

	farmID := c.Param("id")
	if farmID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Farm ID is required",
		})
		return
	}

	var req request.UpdateFarmRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	resp, err := h.farmService.Update(c.Request.Context(), farmerID, farmID, &req)
	if err != nil {
		if errors.Is(err, services.ErrFarmNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"status":  "error",
				"message": "Farm not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to update farm",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"farm": resp,
		},
	})
}

// Delete handles soft deleting a farm
// DELETE /farmer/farms/:id
func (h *FarmHandler) Delete(c *gin.Context) {
	farmerID, exists := middleware.GetFarmerID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "error",
			"message": "Farmer not authenticated",
		})
		return
	}

	farmID := c.Param("id")
	if farmID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Farm ID is required",
		})
		return
	}

	err := h.farmService.Delete(c.Request.Context(), farmerID, farmID)
	if err != nil {
		if errors.Is(err, services.ErrFarmNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"status":  "error",
				"message": "Farm not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to delete farm",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Farm deleted successfully",
	})
}
