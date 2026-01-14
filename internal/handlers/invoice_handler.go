package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ownafarm/ownafarm-backend/internal/dto/request"
	"github.com/ownafarm/ownafarm-backend/internal/middleware"
	"github.com/ownafarm/ownafarm-backend/internal/services"
)

// InvoiceHandler handles invoice-related HTTP requests
type InvoiceHandler struct {
	invoiceService services.InvoiceServiceInterface
}

// NewInvoiceHandler creates a new InvoiceHandler instance
func NewInvoiceHandler(invoiceService services.InvoiceServiceInterface) *InvoiceHandler {
	return &InvoiceHandler{
		invoiceService: invoiceService,
	}
}

// Create handles invoice creation
// POST /farmer/invoices
func (h *InvoiceHandler) Create(c *gin.Context) {
	farmerID, exists := middleware.GetFarmerID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "error",
			"message": "Farmer not authenticated",
		})
		return
	}

	var req request.CreateInvoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	resp, err := h.invoiceService.Create(c.Request.Context(), farmerID, &req)
	if err != nil {
		if errors.Is(err, services.ErrFarmNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"status":  "error",
				"message": "Farm not found",
			})
			return
		}
		if errors.Is(err, services.ErrFarmNotActive) {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  "error",
				"message": "Farm is not active",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to create invoice",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status": "success",
		"data": gin.H{
			"invoice": resp,
		},
	})
}

// GetByID handles getting an invoice by ID
// GET /farmer/invoices/:id
func (h *InvoiceHandler) GetByID(c *gin.Context) {
	farmerID, exists := middleware.GetFarmerID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "error",
			"message": "Farmer not authenticated",
		})
		return
	}

	invoiceID := c.Param("id")
	if invoiceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invoice ID is required",
		})
		return
	}

	resp, err := h.invoiceService.GetByID(c.Request.Context(), farmerID, invoiceID)
	if err != nil {
		if errors.Is(err, services.ErrInvoiceNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"status":  "error",
				"message": "Invoice not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to get invoice",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"invoice": resp,
		},
	})
}

// List handles listing invoices for the authenticated farmer
// GET /farmer/invoices
func (h *InvoiceHandler) List(c *gin.Context) {
	farmerID, exists := middleware.GetFarmerID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "error",
			"message": "Farmer not authenticated",
		})
		return
	}

	var req request.ListInvoiceRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid query parameters",
			"details": err.Error(),
		})
		return
	}

	resp, err := h.invoiceService.List(c.Request.Context(), farmerID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to list invoices",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   resp,
	})
}

// GetPresignedImageURL generates presigned URL for invoice image upload
// POST /farmer/invoices/image/presign
func (h *InvoiceHandler) GetPresignedImageURL(c *gin.Context) {
	// Verify farmer is authenticated (middleware already checks this)
	_, exists := middleware.GetFarmerID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "error",
			"message": "Farmer not authenticated",
		})
		return
	}

	var req request.PresignInvoiceImageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	resp, err := h.invoiceService.GeneratePresignedImageURL(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to generate presigned URL",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   resp,
	})
}

// --- Admin Handlers ---

// GetByIDForAdmin handles getting an invoice by ID (admin)
// GET /admin/invoices/:id
func (h *InvoiceHandler) GetByIDForAdmin(c *gin.Context) {
	invoiceID := c.Param("id")
	if invoiceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invoice ID is required",
		})
		return
	}

	resp, err := h.invoiceService.GetByIDForAdmin(c.Request.Context(), invoiceID)
	if err != nil {
		if errors.Is(err, services.ErrInvoiceNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"status":  "error",
				"message": "Invoice not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to get invoice",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"invoice": resp,
		},
	})
}

// ListForAdmin handles listing invoices for admin with filtering and pagination
// GET /admin/invoices
func (h *InvoiceHandler) ListForAdmin(c *gin.Context) {
	var req request.ListInvoiceRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid query parameters",
			"details": err.Error(),
		})
		return
	}

	resp, err := h.invoiceService.ListForAdmin(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to list invoices",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   resp,
	})
}

// ApproveInvoice handles invoice approval by admin
// PATCH /admin/invoices/:id/approve
func (h *InvoiceHandler) ApproveInvoice(c *gin.Context) {
	adminID, exists := middleware.GetAdminID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "error",
			"message": "Admin not authenticated",
		})
		return
	}

	invoiceID := c.Param("id")
	if invoiceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invoice ID is required",
		})
		return
	}

	ipAddress, _ := middleware.GetIPAddress(c)
	userAgent, _ := middleware.GetUserAgent(c)

	resp, err := h.invoiceService.ApproveInvoice(c.Request.Context(), invoiceID, adminID, ipAddress, userAgent)
	if err != nil {
		if errors.Is(err, services.ErrInvoiceNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"status":  "error",
				"message": "Invoice not found",
			})
			return
		}
		if errors.Is(err, services.ErrInvoiceAlreadyProcessed) {
			c.JSON(http.StatusConflict, gin.H{
				"status":  "error",
				"message": "Invoice has already been processed",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to approve invoice",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   resp,
	})
}

// RejectInvoice handles invoice rejection by admin
// PATCH /admin/invoices/:id/reject
func (h *InvoiceHandler) RejectInvoice(c *gin.Context) {
	adminID, exists := middleware.GetAdminID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "error",
			"message": "Admin not authenticated",
		})
		return
	}

	invoiceID := c.Param("id")
	if invoiceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invoice ID is required",
		})
		return
	}

	var req request.RejectInvoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Allow empty body (reason is optional)
		req = request.RejectInvoiceRequest{}
	}

	ipAddress, _ := middleware.GetIPAddress(c)
	userAgent, _ := middleware.GetUserAgent(c)

	resp, err := h.invoiceService.RejectInvoice(c.Request.Context(), invoiceID, adminID, req.Reason, ipAddress, userAgent)
	if err != nil {
		if errors.Is(err, services.ErrInvoiceNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"status":  "error",
				"message": "Invoice not found",
			})
			return
		}
		if errors.Is(err, services.ErrInvoiceAlreadyProcessed) {
			c.JSON(http.StatusConflict, gin.H{
				"status":  "error",
				"message": "Invoice has already been processed",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to reject invoice",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   resp,
	})
}
