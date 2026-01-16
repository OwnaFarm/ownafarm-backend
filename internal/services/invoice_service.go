package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ownafarm/ownafarm-backend/internal/dto/request"
	"github.com/ownafarm/ownafarm-backend/internal/dto/response"
	"github.com/ownafarm/ownafarm-backend/internal/models"
	"github.com/ownafarm/ownafarm-backend/internal/repositories"
	"github.com/shopspring/decimal"
)

// Errors for InvoiceService
var (
	ErrInvoiceNotFound         = errors.New("invoice not found")
	ErrInvoiceNotOwned         = errors.New("invoice does not belong to this farmer")
	ErrInvoiceAlreadyProcessed = errors.New("invoice has already been processed")
	ErrFarmNotActive           = errors.New("farm is not active")
)

// InvoiceServiceInterface defines the interface for invoice operations
type InvoiceServiceInterface interface {
	Create(ctx context.Context, farmerID string, req *request.CreateInvoiceRequest) (*response.InvoiceResponse, error)
	GetByID(ctx context.Context, farmerID, invoiceID string) (*response.InvoiceResponse, error)
	List(ctx context.Context, farmerID string, req *request.ListInvoiceRequest) (*response.ListInvoiceResponse, error)
	GeneratePresignedImageURL(ctx context.Context, req *request.PresignInvoiceImageRequest) (*response.PresignInvoiceImageResponse, error)

	// Investor operations
	ListAvailableInvoices(ctx context.Context, req *request.ListMarketplaceInvoicesRequest) (*response.ListMarketplaceInvoicesResponse, error)

	// Admin operations
	GetByIDForAdmin(ctx context.Context, invoiceID string) (*response.InvoiceResponse, error)
	ListForAdmin(ctx context.Context, req *request.ListInvoiceRequest) (*response.ListInvoiceAdminResponse, error)
	ApproveInvoice(ctx context.Context, invoiceID, adminID string, req *request.ApproveInvoiceRequest, ipAddress, userAgent string) (*response.InvoiceStatusUpdateResponse, error)
	RejectInvoice(ctx context.Context, invoiceID, adminID string, reason *string, ipAddress, userAgent string) (*response.InvoiceStatusUpdateResponse, error)
}

// InvoiceService implements InvoiceServiceInterface
type InvoiceService struct {
	invoiceRepo    repositories.InvoiceRepository
	farmRepo       repositories.FarmRepository
	storageService StorageService
	auditLogRepo   repositories.AuditLogRepository
}

// NewInvoiceService creates a new InvoiceService instance
func NewInvoiceService(
	invoiceRepo repositories.InvoiceRepository,
	farmRepo repositories.FarmRepository,
	storageService StorageService,
	auditLogRepo repositories.AuditLogRepository,
) *InvoiceService {
	return &InvoiceService{
		invoiceRepo:    invoiceRepo,
		farmRepo:       farmRepo,
		storageService: storageService,
		auditLogRepo:   auditLogRepo,
	}
}

// Create creates a new invoice for a farmer's farm
func (s *InvoiceService) Create(ctx context.Context, farmerID string, req *request.CreateInvoiceRequest) (*response.InvoiceResponse, error) {
	// Verify farm ownership and active status
	farm, err := s.farmRepo.GetByIDAndFarmerID(req.FarmID, farmerID)
	if err != nil {
		return nil, ErrFarmNotFound
	}

	if !farm.IsActive {
		return nil, ErrFarmNotActive
	}

	invoice := &models.Invoice{
		FarmID:       req.FarmID,
		Name:         req.Name,
		Description:  req.Description,
		TargetFund:   decimal.NewFromFloat(req.TargetFund),
		YieldPercent: decimal.NewFromFloat(req.YieldPercent),
		DurationDays: req.DurationDays,
		OfftakerID:   req.OfftakerID,
		Status:       models.InvoiceStatusPending,
		TotalFunded:  decimal.Zero,
		TokenID:      req.TokenID,
	}

	// Set image URL if provided
	if req.ImageKey != nil {
		invoice.ImageURL = req.ImageKey
	}

	if err := s.invoiceRepo.Create(invoice); err != nil {
		return nil, fmt.Errorf("failed to create invoice: %w", err)
	}

	// Set farm for response
	invoice.Farm = *farm

	return s.toInvoiceResponse(invoice), nil
}

// GetByID retrieves an invoice by ID (with ownership check)
func (s *InvoiceService) GetByID(ctx context.Context, farmerID, invoiceID string) (*response.InvoiceResponse, error) {
	invoice, err := s.invoiceRepo.GetByIDAndFarmerID(invoiceID, farmerID)
	if err != nil {
		return nil, ErrInvoiceNotFound
	}

	// Get farm for response
	farm, _ := s.farmRepo.GetByID(invoice.FarmID)
	if farm != nil {
		invoice.Farm = *farm
	}

	return s.toInvoiceResponse(invoice), nil
}

// List retrieves all invoices for a farmer with pagination
func (s *InvoiceService) List(ctx context.Context, farmerID string, req *request.ListInvoiceRequest) (*response.ListInvoiceResponse, error) {
	// Set defaults
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit < 1 {
		req.Limit = 10
	}

	filter := repositories.InvoiceFilter{
		FarmerID:  farmerID,
		FarmID:    req.FarmID,
		Status:    req.Status,
		Page:      req.Page,
		Limit:     req.Limit,
		SortBy:    req.SortBy,
		SortOrder: req.SortOrder,
		Search:    req.Search,
	}

	invoices, totalCount, err := s.invoiceRepo.GetAllByFarmerID(farmerID, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get invoices: %w", err)
	}

	// Map to response
	invoiceItems := make([]response.InvoiceListItem, 0, len(invoices))
	for _, inv := range invoices {
		invoiceItems = append(invoiceItems, response.InvoiceListItem{
			ID:            inv.ID,
			FarmID:        inv.FarmID,
			FarmName:      inv.Farm.Name,
			Name:          inv.Name,
			ImageURL:      inv.ImageURL,
			TargetFund:    inv.TargetFund,
			YieldPercent:  inv.YieldPercent,
			DurationDays:  inv.DurationDays,
			TotalFunded:   inv.TotalFunded,
			IsFullyFunded: inv.IsFullyFunded,
			Status:        string(inv.Status),
			CreatedAt:     inv.CreatedAt,
		})
	}

	// Calculate total pages
	totalPages := int(totalCount) / req.Limit
	if int(totalCount)%req.Limit > 0 {
		totalPages++
	}

	return &response.ListInvoiceResponse{
		Invoices: invoiceItems,
		Pagination: response.PaginationMeta{
			Page:       req.Page,
			Limit:      req.Limit,
			TotalItems: totalCount,
			TotalPages: totalPages,
		},
	}, nil
}

// GeneratePresignedImageURL generates a presigned URL for uploading invoice image
func (s *InvoiceService) GeneratePresignedImageURL(ctx context.Context, req *request.PresignInvoiceImageRequest) (*response.PresignInvoiceImageResponse, error) {
	// Generate unique file key
	fileKey := fmt.Sprintf("invoices/images/%s-%s", uuid.New().String(), req.FileName)

	// Generate presigned URL (15 minutes expiration)
	uploadURL, err := s.storageService.GetPresignedUploadURL(ctx, fileKey, req.ContentType, 15*time.Minute)
	if err != nil {
		return nil, fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return &response.PresignInvoiceImageResponse{
		UploadURL: uploadURL,
		FileKey:   fileKey,
	}, nil
}

// ListAvailableInvoices retrieves approved and available invoices for marketplace
func (s *InvoiceService) ListAvailableInvoices(ctx context.Context, req *request.ListMarketplaceInvoicesRequest) (*response.ListMarketplaceInvoicesResponse, error) {
	// Set defaults
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit < 1 {
		req.Limit = 10
	}

	// Build filter
	filter := repositories.InvoiceFilter{
		IsAvailableForInvestment: true, // Base marketplace filter
		MinTargetFund:            req.MinPrice,
		MaxTargetFund:            req.MaxPrice,
		MinYield:                 req.MinYield,
		MaxYield:                 req.MaxYield,
		MinDuration:              req.MinDuration,
		MaxDuration:              req.MaxDuration,
		MinLandArea:              req.MinLandArea,
		MaxLandArea:              req.MaxLandArea,
		Location:                 req.Location,
		Search:                   req.CropType, // Search in invoice name for crop type
		Page:                     req.Page,
		Limit:                    req.Limit,
		SortBy:                   req.SortBy,
		SortOrder:                req.SortOrder,
	}

	invoices, totalCount, err := s.invoiceRepo.GetAvailableForInvestment(filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get marketplace invoices: %w", err)
	}

	// Map to response
	invoiceItems := make([]response.MarketplaceInvoiceItem, 0, len(invoices))
	for _, inv := range invoices {
		// Calculate funding progress percentage
		fundingProgress := 0.0
		if !inv.TargetFund.IsZero() {
			fundingProgress = inv.TotalFunded.Div(inv.TargetFund).Mul(decimal.NewFromInt(100)).InexactFloat64()
		}

		invoiceItems = append(invoiceItems, response.MarketplaceInvoiceItem{
			ID:              inv.ID,
			Name:            inv.Name,
			Description:     inv.Description,
			ImageURL:        inv.ImageURL,
			TargetFund:      inv.TargetFund,
			TotalFunded:     inv.TotalFunded,
			FundingProgress: fundingProgress,
			YieldPercent:    inv.YieldPercent,
			DurationDays:    inv.DurationDays,
			MaturityDate:    inv.MaturityDate,
			FarmID:          inv.FarmID,
			FarmName:        inv.Farm.Name,
			FarmLocation:    inv.Farm.Location,
			FarmLandArea:    inv.Farm.LandArea,
			FarmCCTVImage:   inv.Farm.CCTVImageUrl,
			FarmerName:      inv.Farm.Farmer.FullName,
			CreatedAt:       inv.CreatedAt,
			ApprovedAt:      inv.ApprovedAt,
		})
	}

	// Calculate total pages
	totalPages := int(totalCount) / req.Limit
	if int(totalCount)%req.Limit > 0 {
		totalPages++
	}

	return &response.ListMarketplaceInvoicesResponse{
		Invoices: invoiceItems,
		Pagination: response.PaginationMeta{
			Page:       req.Page,
			Limit:      req.Limit,
			TotalItems: totalCount,
			TotalPages: totalPages,
		},
	}, nil
}

// GetByIDForAdmin retrieves an invoice by ID (admin - no ownership check)
func (s *InvoiceService) GetByIDForAdmin(ctx context.Context, invoiceID string) (*response.InvoiceResponse, error) {
	invoice, err := s.invoiceRepo.GetByIDWithFarm(invoiceID)
	if err != nil {
		return nil, ErrInvoiceNotFound
	}

	return s.toInvoiceResponse(invoice), nil
}

// ListForAdmin retrieves all invoices with pagination (for admin)
func (s *InvoiceService) ListForAdmin(ctx context.Context, req *request.ListInvoiceRequest) (*response.ListInvoiceAdminResponse, error) {
	// Set defaults
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit < 1 {
		req.Limit = 10
	}

	filter := repositories.InvoiceFilter{
		Status:    req.Status,
		Page:      req.Page,
		Limit:     req.Limit,
		SortBy:    req.SortBy,
		SortOrder: req.SortOrder,
		Search:    req.Search,
	}

	invoices, totalCount, err := s.invoiceRepo.GetAllWithPagination(filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get invoices: %w", err)
	}

	// Map to response
	invoiceItems := make([]response.InvoiceListItemAdmin, 0, len(invoices))
	for _, inv := range invoices {
		invoiceItems = append(invoiceItems, response.InvoiceListItemAdmin{
			ID:            inv.ID,
			FarmID:        inv.FarmID,
			FarmName:      inv.Farm.Name,
			FarmerID:      inv.Farm.FarmerID,
			FarmerName:    inv.Farm.Farmer.FullName,
			Name:          inv.Name,
			ImageURL:      inv.ImageURL,
			TargetFund:    inv.TargetFund,
			YieldPercent:  inv.YieldPercent,
			DurationDays:  inv.DurationDays,
			TotalFunded:   inv.TotalFunded,
			IsFullyFunded: inv.IsFullyFunded,
			Status:        string(inv.Status),
			CreatedAt:     inv.CreatedAt,
			ReviewedAt:    inv.ReviewedAt,
		})
	}

	// Calculate total pages
	totalPages := int(totalCount) / req.Limit
	if int(totalCount)%req.Limit > 0 {
		totalPages++
	}

	return &response.ListInvoiceAdminResponse{
		Invoices: invoiceItems,
		Pagination: response.PaginationMeta{
			Page:       req.Page,
			Limit:      req.Limit,
			TotalItems: totalCount,
			TotalPages: totalPages,
		},
	}, nil
}

// ApproveInvoice approves an invoice with blockchain data
func (s *InvoiceService) ApproveInvoice(ctx context.Context, invoiceID, adminID string, req *request.ApproveInvoiceRequest, ipAddress, userAgent string) (*response.InvoiceStatusUpdateResponse, error) {
	invoice, err := s.invoiceRepo.GetByID(invoiceID)
	if err != nil {
		return nil, ErrInvoiceNotFound
	}

	// Validate status transition (only pending can be approved)
	if invoice.Status != models.InvoiceStatusPending {
		return nil, ErrInvoiceAlreadyProcessed
	}

	// Store old status for audit
	oldStatus := string(invoice.Status)

	// Update invoice status and blockchain data
	now := time.Now()
	invoice.Status = models.InvoiceStatusApproved
	invoice.ReviewedBy = &adminID
	invoice.ReviewedAt = &now
	invoice.ApprovedAt = &now
	invoice.TokenID = &req.TokenID
	invoice.ApprovalTxHash = &req.ApprovalTxHash

	if err := s.invoiceRepo.Update(invoice); err != nil {
		return nil, fmt.Errorf("failed to update invoice: %w", err)
	}

	// Create audit log
	s.createAuditLog(adminID, models.AuditActionApproveInvoice, models.AuditEntityTypeInvoice, invoiceID, oldStatus, string(invoice.Status), nil, ipAddress, userAgent)

	return &response.InvoiceStatusUpdateResponse{
		InvoiceID:      invoice.ID,
		Status:         string(invoice.Status),
		TokenID:        invoice.TokenID,
		ApprovalTxHash: invoice.ApprovalTxHash,
		ReviewedBy:     adminID,
		ReviewedAt:     now,
		Reason:         nil,
	}, nil
}

// RejectInvoice rejects an invoice
func (s *InvoiceService) RejectInvoice(ctx context.Context, invoiceID, adminID string, reason *string, ipAddress, userAgent string) (*response.InvoiceStatusUpdateResponse, error) {
	invoice, err := s.invoiceRepo.GetByID(invoiceID)
	if err != nil {
		return nil, ErrInvoiceNotFound
	}

	// Validate status transition (only pending can be rejected)
	if invoice.Status != models.InvoiceStatusPending {
		return nil, ErrInvoiceAlreadyProcessed
	}

	// Store old status for audit
	oldStatus := string(invoice.Status)

	// Update invoice status
	now := time.Now()
	invoice.Status = models.InvoiceStatusRejected
	invoice.ReviewedBy = &adminID
	invoice.ReviewedAt = &now
	invoice.RejectionReason = reason

	if err := s.invoiceRepo.Update(invoice); err != nil {
		return nil, fmt.Errorf("failed to update invoice: %w", err)
	}

	// Create audit log
	s.createAuditLog(adminID, models.AuditActionRejectInvoice, models.AuditEntityTypeInvoice, invoiceID, oldStatus, string(invoice.Status), reason, ipAddress, userAgent)

	return &response.InvoiceStatusUpdateResponse{
		InvoiceID:  invoice.ID,
		Status:     string(invoice.Status),
		ReviewedBy: adminID,
		ReviewedAt: now,
		Reason:     reason,
	}, nil
}

// createAuditLog creates an audit log entry
func (s *InvoiceService) createAuditLog(adminID, action, entityType, entityID, oldStatus, newStatus string, reason *string, ipAddress, userAgent string) {
	oldValues, _ := json.Marshal(map[string]interface{}{"status": oldStatus})
	newValuesMap := map[string]interface{}{"status": newStatus}
	if reason != nil {
		newValuesMap["rejection_reason"] = *reason
	}
	newValues, _ := json.Marshal(newValuesMap)

	var ipAddr *string
	if ipAddress != "" {
		ipAddr = &ipAddress
	}
	var ua *string
	if userAgent != "" {
		ua = &userAgent
	}

	auditLog := &models.AdminAuditLog{
		AdminID:    adminID,
		Action:     action,
		EntityType: entityType,
		EntityID:   entityID,
		OldValues:  oldValues,
		NewValues:  newValues,
		IPAddress:  ipAddr,
		UserAgent:  ua,
	}

	// Log error but don't fail the main operation
	if err := s.auditLogRepo.Create(auditLog); err != nil {
		fmt.Printf("failed to create audit log: %v\n", err)
	}
}

// toInvoiceResponse converts an Invoice model to InvoiceResponse
func (s *InvoiceService) toInvoiceResponse(invoice *models.Invoice) *response.InvoiceResponse {
	return &response.InvoiceResponse{
		ID:              invoice.ID,
		FarmID:          invoice.FarmID,
		FarmName:        invoice.Farm.Name,
		TokenID:         invoice.TokenID,
		OfftakerID:      invoice.OfftakerID,
		Name:            invoice.Name,
		Description:     invoice.Description,
		ImageURL:        invoice.ImageURL,
		TargetFund:      invoice.TargetFund,
		YieldPercent:    invoice.YieldPercent,
		DurationDays:    invoice.DurationDays,
		TotalFunded:     invoice.TotalFunded,
		IsFullyFunded:   invoice.IsFullyFunded,
		Status:          string(invoice.Status),
		RejectionReason: invoice.RejectionReason,
		ReviewedBy:      invoice.ReviewedBy,
		ReviewedAt:      invoice.ReviewedAt,
		ApprovedAt:      invoice.ApprovedAt,
		FundingDeadline: invoice.FundingDeadline,
		MaturityDate:    invoice.MaturityDate,
		ApprovalTxHash:  invoice.ApprovalTxHash,
		CreatedAt:       invoice.CreatedAt,
		UpdatedAt:       invoice.UpdatedAt,
	}
}
