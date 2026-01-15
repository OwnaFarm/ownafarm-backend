package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/ownafarm/ownafarm-backend/internal/dto/request"
	"github.com/ownafarm/ownafarm-backend/internal/dto/response"
	"github.com/ownafarm/ownafarm-backend/internal/models"
	"github.com/ownafarm/ownafarm-backend/internal/repositories"
)

// Errors for FarmerService
var (
	ErrFarmerAlreadyExists        = errors.New("farmer with this email or phone already exists")
	ErrWalletAddressAlreadyExists = errors.New("farmer with this wallet address already exists")
	ErrInvalidWalletAddress       = errors.New("invalid wallet address format")
	ErrInvalidDateFormat          = errors.New("invalid date format, expected YYYY-MM-DD")
	ErrFarmerNotFound             = errors.New("farmer not found")
	ErrDocumentNotFound           = errors.New("document not found")
	ErrInvalidStatusTransition    = errors.New("invalid status transition")
	ErrFarmerAlreadyProcessed     = errors.New("farmer has already been processed")
)

// walletAddressRegex validates Ethereum wallet address format (0x + 40 hex chars)
var walletAddressRegex = regexp.MustCompile(`^0x[0-9a-fA-F]{40}$`)

// FarmerServiceInterface defines the interface for farmer operations
type FarmerServiceInterface interface {
	Register(ctx context.Context, req *request.RegisterFarmerRequest) (*models.Farmer, error)
	GeneratePresignedURLs(ctx context.Context, documentTypes []string) (*response.PresignDocumentsResponse, error)
	GetDocumentDownloadURL(ctx context.Context, farmerID, documentID string) (*response.DocumentDownloadURLResponse, error)
	GetListForAdmin(ctx context.Context, req *request.ListFarmerRequest) (*response.ListFarmerResponse, error)
	GetDetailForAdmin(ctx context.Context, farmerID string) (*response.FarmerDetailResponse, error)
	ApproveFarmer(ctx context.Context, farmerID, adminID, ipAddress, userAgent string) (*response.FarmerStatusUpdateResponse, error)
	RejectFarmer(ctx context.Context, farmerID, adminID string, reason *string, ipAddress, userAgent string) (*response.FarmerStatusUpdateResponse, error)
}

// FarmerService implements FarmerServiceInterface
type FarmerService struct {
	farmerRepo     repositories.FarmerRepository
	storageService StorageService
	auditLogRepo   repositories.AuditLogRepository
}

// NewFarmerService creates a new FarmerService instance
func NewFarmerService(
	farmerRepo repositories.FarmerRepository,
	storageService StorageService,
	auditLogRepo repositories.AuditLogRepository,
) *FarmerService {
	return &FarmerService{
		farmerRepo:     farmerRepo,
		storageService: storageService,
		auditLogRepo:   auditLogRepo,
	}
}

// Register registers a new farmer
func (s *FarmerService) Register(ctx context.Context, req *request.RegisterFarmerRequest) (*models.Farmer, error) {
	// Normalize and validate wallet address
	walletAddress := strings.ToLower(req.PersonalInfo.WalletAddress)
	if !walletAddressRegex.MatchString(walletAddress) {
		return nil, ErrInvalidWalletAddress
	}

	// Check if wallet address already exists
	walletExists, err := s.farmerRepo.ExistsByWalletAddress(walletAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing wallet address: %w", err)
	}
	if walletExists {
		return nil, ErrWalletAddressAlreadyExists
	}

	// Check if farmer already exists by email or phone
	exists, err := s.farmerRepo.ExistsByEmailOrPhone(
		req.PersonalInfo.Email,
		req.PersonalInfo.PhoneNumber,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing farmer: %w", err)
	}
	if exists {
		return nil, ErrFarmerAlreadyExists
	}

	// Parse date of birth
	dob, err := time.Parse("2006-01-02", req.PersonalInfo.DateOfBirth)
	if err != nil {
		return nil, ErrInvalidDateFormat
	}

	// Build farmer model
	farmer := &models.Farmer{
		Status:        models.FarmerStatusPending,
		WalletAddress: walletAddress,

		// Personal Info
		FullName:    req.PersonalInfo.FullName,
		Email:       req.PersonalInfo.Email,
		PhoneNumber: req.PersonalInfo.PhoneNumber,
		IDNumber:    req.PersonalInfo.IDNumber,
		DateOfBirth: dob,
		Address:     req.PersonalInfo.Address,
		Province:    req.PersonalInfo.Province,
		City:        req.PersonalInfo.City,
		District:    req.PersonalInfo.District,
		PostalCode:  req.PersonalInfo.PostalCode,

		// Business Info
		BusinessName:      req.BusinessInfo.BusinessName,
		BusinessType:      models.BusinessType(req.BusinessInfo.BusinessType),
		NPWP:              req.BusinessInfo.NPWP,
		BankName:          req.BusinessInfo.BankName,
		BankAccountNumber: req.BusinessInfo.BankAccountNumber,
		BankAccountName:   req.BusinessInfo.BankAccountName,
		YearsOfExperience: 0,
		CropsExpertise:    pq.StringArray{},
	}

	// Set optional fields
	if req.BusinessInfo.YearsOfExperience != nil {
		farmer.YearsOfExperience = *req.BusinessInfo.YearsOfExperience
	}
	if len(req.BusinessInfo.CropsExpertise) > 0 {
		farmer.CropsExpertise = pq.StringArray(req.BusinessInfo.CropsExpertise)
	}

	// Create farmer record
	if err := s.farmerRepo.Create(farmer); err != nil {
		return nil, fmt.Errorf("failed to create farmer: %w", err)
	}

	// Build and create document records
	// Store only the file_key in FileURL column (not a public URL)
	documents := make([]models.FarmerDocument, 0, len(req.Documents))
	for _, doc := range req.Documents {
		documents = append(documents, models.FarmerDocument{
			FarmerID:     farmer.ID,
			DocumentType: models.DocumentType(doc.DocumentType),
			FileURL:      doc.FileKey, // Store file_key directly, not a public URL
			FileName:     doc.FileName,
			FileSize:     doc.FileSize,
			MimeType:     doc.MimeType,
		})
	}

	if err := s.farmerRepo.CreateDocuments(documents); err != nil {
		return nil, fmt.Errorf("failed to create farmer documents: %w", err)
	}

	farmer.Documents = documents
	return farmer, nil
}

// GeneratePresignedURLs generates presigned URLs for uploading documents
func (s *FarmerService) GeneratePresignedURLs(ctx context.Context, documentTypes []string) (*response.PresignDocumentsResponse, error) {
	urls := make([]response.PresignedURLResponse, 0, len(documentTypes))

	for _, docType := range documentTypes {
		// Generate unique file key
		fileKey := s.generateFileKey(docType)

		// Generate presigned URL (15 minutes expiration)
		uploadURL, err := s.storageService.GetPresignedUploadURL(ctx, fileKey, "application/octet-stream", 15*time.Minute)
		if err != nil {
			return nil, fmt.Errorf("failed to generate presigned URL for %s: %w", docType, err)
		}

		urls = append(urls, response.PresignedURLResponse{
			DocumentType: docType,
			UploadURL:    uploadURL,
			FileKey:      fileKey,
		})
	}

	return &response.PresignDocumentsResponse{URLs: urls}, nil
}

// GetDocumentDownloadURL generates a presigned download URL for a specific document
func (s *FarmerService) GetDocumentDownloadURL(ctx context.Context, farmerID, documentID string) (*response.DocumentDownloadURLResponse, error) {
	// Get farmer with documents
	farmer, err := s.farmerRepo.GetByID(farmerID)
	if err != nil {
		return nil, ErrFarmerNotFound
	}

	// Find the specific document
	var document *models.FarmerDocument
	for i := range farmer.Documents {
		if farmer.Documents[i].ID == documentID {
			document = &farmer.Documents[i]
			break
		}
	}

	if document == nil {
		return nil, ErrDocumentNotFound
	}

	// Generate presigned download URL (1 hour expiration)
	downloadURL, err := s.storageService.GetPresignedDownloadURL(ctx, document.FileURL, 1*time.Hour)
	if err != nil {
		return nil, fmt.Errorf("failed to generate download URL: %w", err)
	}

	return &response.DocumentDownloadURLResponse{
		DocumentID:   document.ID,
		DocumentType: string(document.DocumentType),
		DownloadURL:  downloadURL,
		ExpiresIn:    3600, // 1 hour in seconds
	}, nil
}

// generateFileKey generates a unique file key for a document
func (s *FarmerService) generateFileKey(documentType string) string {
	return fmt.Sprintf("farmers/documents/%s-%s", uuid.New().String(), documentType)
}

// GetListForAdmin retrieves a paginated list of farmers for admin
func (s *FarmerService) GetListForAdmin(ctx context.Context, req *request.ListFarmerRequest) (*response.ListFarmerResponse, error) {
	// Set defaults
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit < 1 {
		req.Limit = 10
	}

	// Build filter
	filter := repositories.FarmerFilter{
		Status:    req.Status,
		Page:      req.Page,
		Limit:     req.Limit,
		SortBy:    req.SortBy,
		SortOrder: req.SortOrder,
		Search:    req.Search,
	}

	// Get farmers from repository
	farmers, totalCount, err := s.farmerRepo.GetAllWithPagination(filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get farmers: %w", err)
	}

	// Map to response
	farmerItems := make([]response.FarmerListItem, 0, len(farmers))
	for _, farmer := range farmers {
		farmerItems = append(farmerItems, response.FarmerListItem{
			ID:                farmer.ID,
			Status:            string(farmer.Status),
			WalletAddress:     farmer.WalletAddress,
			FullName:          farmer.FullName,
			Email:             farmer.Email,
			PhoneNumber:       farmer.PhoneNumber,
			BusinessName:      farmer.BusinessName,
			BusinessType:      string(farmer.BusinessType),
			Province:          farmer.Province,
			City:              farmer.City,
			YearsOfExperience: farmer.YearsOfExperience,
			CreatedAt:         farmer.CreatedAt,
			ReviewedAt:        farmer.ReviewedAt,
		})
	}

	// Calculate total pages
	totalPages := int(totalCount) / req.Limit
	if int(totalCount)%req.Limit > 0 {
		totalPages++
	}

	return &response.ListFarmerResponse{
		Farmers: farmerItems,
		Pagination: response.PaginationMeta{
			Page:       req.Page,
			Limit:      req.Limit,
			TotalItems: totalCount,
			TotalPages: totalPages,
		},
	}, nil
}

// GetDetailForAdmin retrieves farmer detail with documents for admin
func (s *FarmerService) GetDetailForAdmin(ctx context.Context, farmerID string) (*response.FarmerDetailResponse, error) {
	// Get farmer with documents
	farmer, err := s.farmerRepo.GetByID(farmerID)
	if err != nil {
		return nil, ErrFarmerNotFound
	}

	// Generate presigned download URLs for all documents
	documents := make([]response.FarmerDocumentItem, 0, len(farmer.Documents))
	for _, doc := range farmer.Documents {
		downloadURL, err := s.storageService.GetPresignedDownloadURL(ctx, doc.FileURL, 1*time.Hour)
		if err != nil {
			// Log error but continue with empty URL
			downloadURL = ""
		}

		documents = append(documents, response.FarmerDocumentItem{
			ID:           doc.ID,
			DocumentType: string(doc.DocumentType),
			FileName:     doc.FileName,
			DownloadURL:  downloadURL,
			ExpiresIn:    3600, // 1 hour
		})
	}

	// Format dates
	var reviewedAt *string
	if farmer.ReviewedAt != nil {
		formatted := farmer.ReviewedAt.Format(time.RFC3339)
		reviewedAt = &formatted
	}

	// Convert crops expertise
	cropsExpertise := make([]string, len(farmer.CropsExpertise))
	for i, crop := range farmer.CropsExpertise {
		cropsExpertise[i] = crop
	}

	return &response.FarmerDetailResponse{
		ID:                farmer.ID,
		Status:            string(farmer.Status),
		WalletAddress:     farmer.WalletAddress,
		FullName:          farmer.FullName,
		Email:             farmer.Email,
		PhoneNumber:       farmer.PhoneNumber,
		IDNumber:          farmer.IDNumber,
		DateOfBirth:       farmer.DateOfBirth.Format("2006-01-02"),
		Address:           farmer.Address,
		Province:          farmer.Province,
		City:              farmer.City,
		District:          farmer.District,
		PostalCode:        farmer.PostalCode,
		BusinessName:      farmer.BusinessName,
		BusinessType:      string(farmer.BusinessType),
		NPWP:              farmer.NPWP,
		BankName:          farmer.BankName,
		BankAccountNumber: farmer.BankAccountNumber,
		BankAccountName:   farmer.BankAccountName,
		YearsOfExperience: farmer.YearsOfExperience,
		CropsExpertise:    cropsExpertise,
		Documents:         documents,
		ReviewedBy:        farmer.ReviewedBy,
		ReviewedAt:        reviewedAt,
		RejectionReason:   farmer.RejectionReason,
		CreatedAt:         farmer.CreatedAt.Format(time.RFC3339),
	}, nil
}

// ApproveFarmer approves a farmer registration
func (s *FarmerService) ApproveFarmer(ctx context.Context, farmerID, adminID, ipAddress, userAgent string) (*response.FarmerStatusUpdateResponse, error) {
	// Get farmer
	farmer, err := s.farmerRepo.GetByID(farmerID)
	if err != nil {
		return nil, ErrFarmerNotFound
	}

	// Validate status transition (only pending can be approved)
	if farmer.Status != models.FarmerStatusPending {
		return nil, ErrFarmerAlreadyProcessed
	}

	// Store old status for audit
	oldStatus := string(farmer.Status)

	// Update farmer status
	now := time.Now()
	farmer.Status = models.FarmerStatusApproved
	farmer.ReviewedBy = &adminID
	farmer.ReviewedAt = &now

	if err := s.farmerRepo.Update(farmer); err != nil {
		return nil, fmt.Errorf("failed to update farmer: %w", err)
	}

	// Create audit log
	s.createAuditLog(adminID, models.AuditActionApproveFarmer, models.AuditEntityTypeFarmer, farmerID, oldStatus, string(farmer.Status), nil, ipAddress, userAgent)

	return &response.FarmerStatusUpdateResponse{
		FarmerID:   farmer.ID,
		Status:     string(farmer.Status),
		ReviewedBy: adminID,
		ReviewedAt: now,
		Reason:     nil,
	}, nil
}

// RejectFarmer rejects a farmer registration
func (s *FarmerService) RejectFarmer(ctx context.Context, farmerID, adminID string, reason *string, ipAddress, userAgent string) (*response.FarmerStatusUpdateResponse, error) {
	// Get farmer
	farmer, err := s.farmerRepo.GetByID(farmerID)
	if err != nil {
		return nil, ErrFarmerNotFound
	}

	// Validate status transition (only pending can be rejected)
	if farmer.Status != models.FarmerStatusPending {
		return nil, ErrFarmerAlreadyProcessed
	}

	// Store old status for audit
	oldStatus := string(farmer.Status)

	// Update farmer status
	now := time.Now()
	farmer.Status = models.FarmerStatusRejected
	farmer.ReviewedBy = &adminID
	farmer.ReviewedAt = &now
	farmer.RejectionReason = reason

	if err := s.farmerRepo.Update(farmer); err != nil {
		return nil, fmt.Errorf("failed to update farmer: %w", err)
	}

	// Create audit log
	s.createAuditLog(adminID, models.AuditActionRejectFarmer, models.AuditEntityTypeFarmer, farmerID, oldStatus, string(farmer.Status), reason, ipAddress, userAgent)

	return &response.FarmerStatusUpdateResponse{
		FarmerID:   farmer.ID,
		Status:     string(farmer.Status),
		ReviewedBy: adminID,
		ReviewedAt: now,
		Reason:     reason,
	}, nil
}

// createAuditLog creates an audit log entry
func (s *FarmerService) createAuditLog(adminID, action, entityType, entityID, oldStatus, newStatus string, reason *string, ipAddress, userAgent string) {
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
