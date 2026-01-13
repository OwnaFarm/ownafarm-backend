package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ownafarm/ownafarm-backend/internal/dto/request"
	"github.com/ownafarm/ownafarm-backend/internal/dto/response"
	"github.com/ownafarm/ownafarm-backend/internal/middleware"
	"github.com/ownafarm/ownafarm-backend/internal/services"
)

// FarmerHandler handles farmer-related HTTP requests
type FarmerHandler struct {
	farmerService services.FarmerServiceInterface
}

// NewFarmerHandler creates a new FarmerHandler instance
func NewFarmerHandler(farmerService services.FarmerServiceInterface) *FarmerHandler {
	return &FarmerHandler{
		farmerService: farmerService,
	}
}

// Register handles farmer registration
// POST /farmers/register
func (h *FarmerHandler) Register(c *gin.Context) {
	var req request.RegisterFarmerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	farmer, err := h.farmerService.Register(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, services.ErrFarmerAlreadyExists) {
			c.JSON(http.StatusConflict, gin.H{
				"status":  "error",
				"message": "Farmer with this email or phone number already exists",
			})
			return
		}
		if errors.Is(err, services.ErrInvalidDateFormat) {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  "error",
				"message": "Invalid date format for date_of_birth, expected YYYY-MM-DD",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to register farmer",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status": "success",
		"data": response.RegisterFarmerResponse{
			FarmerID: farmer.ID,
			Status:   string(farmer.Status),
		},
	})
}

// GetPresignedURLs handles presigned URL generation for document uploads
// POST /farmers/documents/presign
func (h *FarmerHandler) GetPresignedURLs(c *gin.Context) {
	var req request.PresignDocumentsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	resp, err := h.farmerService.GeneratePresignedURLs(c.Request.Context(), req.DocumentTypes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to generate presigned URLs",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   resp,
	})
}

// GetDocumentDownloadURL generates a presigned download URL for a specific document
// GET /farmers/:farmer_id/documents/:document_id/download
func (h *FarmerHandler) GetDocumentDownloadURL(c *gin.Context) {
	farmerID := c.Param("farmer_id")
	documentID := c.Param("document_id")

	if farmerID == "" || documentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "farmer_id and document_id are required",
		})
		return
	}

	resp, err := h.farmerService.GetDocumentDownloadURL(c.Request.Context(), farmerID, documentID)
	if err != nil {
		if errors.Is(err, services.ErrFarmerNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"status":  "error",
				"message": "Farmer not found",
			})
			return
		}
		if errors.Is(err, services.ErrDocumentNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"status":  "error",
				"message": "Document not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to generate download URL",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   resp,
	})
}

// GetListForAdmin handles listing farmers for admin with filtering and pagination
// GET /admin/farmers
func (h *FarmerHandler) GetListForAdmin(c *gin.Context) {
	var req request.ListFarmerRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid query parameters",
			"details": err.Error(),
		})
		return
	}

	resp, err := h.farmerService.GetListForAdmin(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to get farmers",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   resp,
	})
}

// ApproveFarmer handles approving a farmer registration
// PATCH /admin/farmers/:id/approve
func (h *FarmerHandler) ApproveFarmer(c *gin.Context) {
	farmerID := c.Param("id")
	if farmerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Farmer ID is required",
		})
		return
	}

	// Get admin info from context
	adminID, exists := middleware.GetAdminID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "error",
			"message": "Admin ID not found in context",
		})
		return
	}

	ipAddress, _ := middleware.GetIPAddress(c)
	userAgent, _ := middleware.GetUserAgent(c)

	resp, err := h.farmerService.ApproveFarmer(c.Request.Context(), farmerID, adminID, ipAddress, userAgent)
	if err != nil {
		if errors.Is(err, services.ErrFarmerNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"status":  "error",
				"message": "Farmer not found",
			})
			return
		}
		if errors.Is(err, services.ErrFarmerAlreadyProcessed) {
			c.JSON(http.StatusConflict, gin.H{
				"status":  "error",
				"message": "Farmer has already been processed (not in pending status)",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to approve farmer",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   resp,
	})
}

// RejectFarmer handles rejecting a farmer registration
// PATCH /admin/farmers/:id/reject
func (h *FarmerHandler) RejectFarmer(c *gin.Context) {
	farmerID := c.Param("id")
	if farmerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Farmer ID is required",
		})
		return
	}

	// Parse request body (optional reason)
	var req request.RejectFarmerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// If body is empty or invalid, continue with nil reason
		req.Reason = nil
	}

	// Get admin info from context
	adminID, exists := middleware.GetAdminID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "error",
			"message": "Admin ID not found in context",
		})
		return
	}

	ipAddress, _ := middleware.GetIPAddress(c)
	userAgent, _ := middleware.GetUserAgent(c)

	resp, err := h.farmerService.RejectFarmer(c.Request.Context(), farmerID, adminID, req.Reason, ipAddress, userAgent)
	if err != nil {
		if errors.Is(err, services.ErrFarmerNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"status":  "error",
				"message": "Farmer not found",
			})
			return
		}
		if errors.Is(err, services.ErrFarmerAlreadyProcessed) {
			c.JSON(http.StatusConflict, gin.H{
				"status":  "error",
				"message": "Farmer has already been processed (not in pending status)",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to reject farmer",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   resp,
	})
}
