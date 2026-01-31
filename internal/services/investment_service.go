package services

import (
	"context"
	"errors"
	"log"
	"math"
	"time"

	"github.com/ownafarm/ownafarm-backend/internal/dto/request"
	"github.com/ownafarm/ownafarm-backend/internal/dto/response"
	"github.com/ownafarm/ownafarm-backend/internal/models"
	"github.com/ownafarm/ownafarm-backend/internal/repositories"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

const (
	// WaterXPGain is the XP gained per water action
	WaterXPGain = 5
	// WaterCost is the water points cost per water action
	WaterCost = 10
	// HarvestXPGain is the XP gained per harvest action
	HarvestXPGain = 50
)

var (
	ErrInvestmentNotFound = errors.New("investment not found")
	ErrNotEnoughWater     = errors.New("not enough water points")
	ErrAlreadyHarvested   = errors.New("crop already harvested")
	ErrNotReadyToHarvest  = errors.New("crop not ready to harvest")
)

// InvestmentServiceInterface defines the interface for investment operations
type InvestmentServiceInterface interface {
	SyncInvestments(ctx context.Context, userID, walletAddress string, req *request.SyncInvestmentsRequest) (*response.SyncInvestmentsResponse, error)
	ListCrops(ctx context.Context, userID string, req *request.ListCropsRequest) (*response.ListCropsResponse, error)
	GetCrop(ctx context.Context, userID, cropID string) (*response.CropResponse, error)
	WaterCrop(ctx context.Context, userID, cropID string) (*response.WaterCropResponse, error)
	SyncHarvest(ctx context.Context, userID, walletAddress, cropID string) (*response.SyncHarvestResponse, error)
}

// InvestmentService implements InvestmentServiceInterface
type InvestmentService struct {
	investmentRepo repositories.InvestmentRepository
	invoiceRepo    repositories.InvoiceRepository
	userRepo       repositories.UserRepository
	blockchainSvc  BlockchainService
}

// NewInvestmentService creates a new InvestmentService instance
func NewInvestmentService(
	investmentRepo repositories.InvestmentRepository,
	invoiceRepo repositories.InvoiceRepository,
	userRepo repositories.UserRepository,
	blockchainSvc BlockchainService,
) *InvestmentService {
	return &InvestmentService{
		investmentRepo: investmentRepo,
		invoiceRepo:    invoiceRepo,
		userRepo:       userRepo,
		blockchainSvc:  blockchainSvc,
	}
}

// SyncInvestments syncs investments from blockchain to database
func (s *InvestmentService) SyncInvestments(ctx context.Context, userID, walletAddress string, req *request.SyncInvestmentsRequest) (*response.SyncInvestmentsResponse, error) {
	log.Printf("[SyncInvestments] Starting sync for userID=%s, wallet=%s", userID, walletAddress)

	// Get investment count from blockchain
	count, err := s.blockchainSvc.GetInvestmentCount(ctx, walletAddress)
	if err != nil {
		log.Printf("[SyncInvestments] ERROR getting investment count from blockchain: %v", err)
		return nil, err
	}
	log.Printf("[SyncInvestments] Found %d investments on blockchain for wallet %s", count, walletAddress)

	var newCrops []response.CropResponse
	syncedCount := 0

	// Iterate through all investments
	for i := uint64(0); i < count; i++ {
		log.Printf("[SyncInvestments] Processing investment index %d/%d", i, count)

		// Check if investment already exists in database
		_, err := s.investmentRepo.GetByUserIDAndOnchainID(userID, int64(i))
		if err == nil {
			// Investment already exists, skip
			log.Printf("[SyncInvestments] Investment %d already exists in DB, skipping", i)
			continue
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			// Unexpected error
			log.Printf("[SyncInvestments] ERROR checking existing investment %d: %v", i, err)
			return nil, err
		}

		// Get investment from blockchain
		onchainInv, err := s.blockchainSvc.GetInvestment(ctx, walletAddress, i)
		if err != nil {
			log.Printf("[SyncInvestments] ERROR getting investment %d from blockchain: %v", i, err)
			return nil, err
		}
		log.Printf("[SyncInvestments] Got onchain investment %d: amount=%v, tokenID=%d, claimed=%v",
			i, onchainInv.Amount, onchainInv.TokenID, onchainInv.Claimed)

		// Skip if amount is 0 (invalid investment)
		if onchainInv.Amount == nil || onchainInv.Amount.Sign() == 0 {
			log.Printf("[SyncInvestments] Investment %d has zero amount, skipping", i)
			continue
		}

		// Find invoice by token ID
		invoice, err := s.findInvoiceByTokenID(int64(onchainInv.TokenID))
		if err != nil {
			// Invoice not found in our database, skip this investment
			log.Printf("[SyncInvestments] Invoice with tokenID=%d not found in DB: %v, skipping investment %d", onchainInv.TokenID, err, i)
			continue
		}
		log.Printf("[SyncInvestments] Found invoice %s (name=%s) for tokenID=%d", invoice.ID, invoice.Name, onchainInv.TokenID)

		// Create investment record
		investedAt := time.Unix(int64(onchainInv.InvestedAt), 0)
		onchainID := int64(i)
		amount := decimal.NewFromBigInt(onchainInv.Amount, -18) // Convert from wei to GOLD (18 decimals)

		investment := &models.Investment{
			UserID:              userID,
			InvoiceID:           invoice.ID,
			InvestmentIdOnchain: &onchainID,
			Amount:              amount,
			InvestedAt:          investedAt,
			Status:              models.CropStatusGrowing,
			Progress:            0,
			WaterCount:          0,
			IsHarvested:         onchainInv.Claimed,
		}

		// Calculate initial progress and status
		progress, status := s.calculateProgressAndStatus(investment, invoice)
		investment.Progress = progress
		investment.Status = status

		if onchainInv.Claimed {
			investment.IsHarvested = true
			investment.Status = models.CropStatusHarvested
			harvestedAt := time.Now() // We don't know exact harvest time
			investment.HarvestedAt = &harvestedAt
		}

		log.Printf("[SyncInvestments] Creating investment record: invoiceID=%s, amount=%s, progress=%d, status=%s",
			invoice.ID, amount.String(), progress, status)

		if err := s.investmentRepo.Create(investment); err != nil {
			log.Printf("[SyncInvestments] ERROR creating investment record: %v", err)
			return nil, err
		}
		log.Printf("[SyncInvestments] Successfully created investment record with ID=%s", investment.ID)

		// Update invoice funding totals
		// Error is logged but doesn't fail the sync - totals can be recalculated later
		if err := s.invoiceRepo.UpdateFundingTotals(invoice.ID); err != nil {
			log.Printf("[SyncInvestments] WARNING: failed to update funding totals for invoice %s: %v", invoice.ID, err)
		}

		// Reload with relations for response
		investment, err = s.investmentRepo.GetByIDWithRelations(investment.ID)
		if err != nil {
			log.Printf("[SyncInvestments] ERROR reloading investment with relations: %v", err)
			return nil, err
		}

		newCrops = append(newCrops, s.toCropResponse(investment))
		syncedCount++
	}

	log.Printf("[SyncInvestments] Sync complete. Synced %d new investments for wallet %s", syncedCount, walletAddress)

	return &response.SyncInvestmentsResponse{
		SyncedCount: syncedCount,
		NewCrops:    newCrops,
	}, nil
}

// ListCrops retrieves all crops for a user
func (s *InvestmentService) ListCrops(ctx context.Context, userID string, req *request.ListCropsRequest) (*response.ListCropsResponse, error) {
	// Set defaults
	page := req.Page
	if page < 1 {
		page = 1
	}
	limit := req.Limit
	if limit < 1 {
		limit = 10
	}

	filter := repositories.InvestmentFilter{
		UserID:    userID,
		Status:    req.Status,
		Page:      page,
		Limit:     limit,
		SortBy:    req.SortBy,
		SortOrder: req.SortOrder,
	}

	investments, totalCount, err := s.investmentRepo.GetAllByUserID(filter)
	if err != nil {
		return nil, err
	}

	var crops []response.CropResponse
	for i := range investments {
		// Update progress for active investments
		if investments[i].Status != models.CropStatusHarvested {
			progress, status := s.calculateProgressAndStatus(&investments[i], &investments[i].Invoice)
			investments[i].Progress = progress
			investments[i].Status = status
			// Update in database asynchronously (fire and forget)
			go func(id string, p int, st models.CropStatus) {
				_ = s.investmentRepo.UpdateProgress(id, p, st)
			}(investments[i].ID, progress, status)
		}
		crops = append(crops, s.toCropResponse(&investments[i]))
	}

	return &response.ListCropsResponse{
		Crops:      crops,
		TotalCount: totalCount,
		Page:       page,
		Limit:      limit,
	}, nil
}

// GetCrop retrieves a single crop by ID
func (s *InvestmentService) GetCrop(ctx context.Context, userID, cropID string) (*response.CropResponse, error) {
	investment, err := s.investmentRepo.GetByIDAndUserID(cropID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvestmentNotFound
		}
		return nil, err
	}

	// Update progress if not harvested
	if investment.Status != models.CropStatusHarvested {
		progress, status := s.calculateProgressAndStatus(investment, &investment.Invoice)
		investment.Progress = progress
		investment.Status = status
		_ = s.investmentRepo.UpdateProgress(investment.ID, progress, status)
	}

	resp := s.toCropResponse(investment)
	return &resp, nil
}

// WaterCrop waters a crop (for XP gain, gimmick only)
func (s *InvestmentService) WaterCrop(ctx context.Context, userID, cropID string) (*response.WaterCropResponse, error) {
	// Get investment
	investment, err := s.investmentRepo.GetByIDAndUserID(cropID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvestmentNotFound
		}
		return nil, err
	}

	if investment.IsHarvested {
		return nil, ErrAlreadyHarvested
	}

	// Regenerate water and get fresh user data
	user, err := s.userRepo.RegenerateWater(userID)
	if err != nil {
		return nil, err
	}

	// Check water points
	if user.WaterPoints < WaterCost {
		return nil, ErrNotEnoughWater
	}

	// Increment water count on investment
	if err := s.investmentRepo.IncrementWaterCount(investment.ID); err != nil {
		return nil, err
	}

	// Update user stats: deduct water and add XP
	newWaterPoints := user.WaterPoints - WaterCost
	newXP := user.XP + WaterXPGain

	err = s.userRepo.UpdateGameStats(userID, map[string]interface{}{
		"water_points": newWaterPoints,
		"xp":           newXP,
	})
	if err != nil {
		if errors.Is(err, repositories.ErrNotEnoughWater) {
			// Race condition: another request used the water
			return nil, ErrNotEnoughWater
		}
		return nil, err
	}

	// Reload investment with updated water count
	investment, err = s.investmentRepo.GetByIDAndUserID(cropID, userID)
	if err != nil {
		return nil, err
	}

	return &response.WaterCropResponse{
		Crop:           s.toCropResponse(investment),
		XPGained:       WaterXPGain,
		WaterRemaining: newWaterPoints,
	}, nil
}

// SyncHarvest syncs harvest status from blockchain
func (s *InvestmentService) SyncHarvest(ctx context.Context, userID, walletAddress, cropID string) (*response.SyncHarvestResponse, error) {
	investment, err := s.investmentRepo.GetByIDAndUserID(cropID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvestmentNotFound
		}
		return nil, err
	}

	// Already harvested - return with 0 XP (no double XP)
	if investment.IsHarvested {
		resp := s.toCropResponse(investment)
		return &response.SyncHarvestResponse{
			Crop:     resp,
			XPGained: 0,
		}, nil
	}

	// Check blockchain for harvest status
	if investment.InvestmentIdOnchain == nil {
		return nil, ErrInvestmentNotFound
	}

	onchainInv, err := s.blockchainSvc.GetInvestment(ctx, walletAddress, uint64(*investment.InvestmentIdOnchain))
	if err != nil {
		return nil, err
	}

	xpGained := 0
	if onchainInv.Claimed {
		// Update harvest status
		now := time.Now()
		investment.IsHarvested = true
		investment.Status = models.CropStatusHarvested
		investment.HarvestedAt = &now
		investment.Progress = 100

		// Calculate harvest amount (principal + yield)
		yieldPercent := investment.Invoice.YieldPercent
		harvestAmount := investment.Amount.Mul(decimal.NewFromFloat(1).Add(yieldPercent.Div(decimal.NewFromInt(100))))
		investment.HarvestAmount = &harvestAmount

		if err := s.investmentRepo.Update(investment); err != nil {
			return nil, err
		}

		// Add XP to user profile
		user, err := s.userRepo.GetByID(userID)
		if err != nil {
			return nil, err
		}
		newXP := user.XP + HarvestXPGain
		err = s.userRepo.UpdateGameStats(userID, map[string]interface{}{
			"xp": newXP,
		})
		if err != nil {
			return nil, err
		}
		xpGained = HarvestXPGain
	}

	resp := s.toCropResponse(investment)
	return &response.SyncHarvestResponse{
		Crop:     resp,
		XPGained: xpGained,
	}, nil
}

// calculateProgressAndStatus calculates progress and status based on time
func (s *InvestmentService) calculateProgressAndStatus(investment *models.Investment, invoice *models.Invoice) (int, models.CropStatus) {
	if investment.IsHarvested {
		return 100, models.CropStatusHarvested
	}

	elapsed := time.Since(investment.InvestedAt)
	duration := time.Duration(invoice.DurationDays) * 24 * time.Hour
	progress := int(math.Round(elapsed.Seconds() / duration.Seconds() * 100))

	if progress >= 100 {
		progress = 100
		return progress, models.CropStatusReady
	}

	return progress, models.CropStatusGrowing
}

// findInvoiceByTokenID finds an invoice by its blockchain token ID
func (s *InvestmentService) findInvoiceByTokenID(tokenID int64) (*models.Invoice, error) {
	return s.invoiceRepo.GetByTokenID(tokenID)
}

// toCropResponse converts an Investment model to CropResponse
func (s *InvestmentService) toCropResponse(investment *models.Investment) response.CropResponse {
	invoice := investment.Invoice
	farm := invoice.Farm

	daysLeft := 0
	if investment.Status != models.CropStatusHarvested {
		maturityDate := investment.InvestedAt.Add(time.Duration(invoice.DurationDays) * 24 * time.Hour)
		daysLeft = int(math.Ceil(time.Until(maturityDate).Hours() / 24))
		if daysLeft < 0 {
			daysLeft = 0
		}
	}

	canHarvest := investment.Status == models.CropStatusReady && !investment.IsHarvested

	yieldPercent, _ := invoice.YieldPercent.Float64()
	invested, _ := investment.Amount.Float64()

	var harvestAmount *float64
	if investment.HarvestAmount != nil {
		ha, _ := investment.HarvestAmount.Float64()
		harvestAmount = &ha
	}

	return response.CropResponse{
		ID:            investment.ID,
		Name:          invoice.Name,
		Image:         invoice.ImageURL,
		CCTVImage:     farm.CCTVImageUrl,
		Location:      farm.Location,
		Progress:      investment.Progress,
		DaysLeft:      daysLeft,
		YieldPercent:  yieldPercent,
		Invested:      invested,
		Status:        string(investment.Status),
		PlantedAt:     investment.InvestedAt.Format(time.RFC3339),
		WaterCount:    investment.WaterCount,
		CanHarvest:    canHarvest,
		HarvestAmount: harvestAmount,
	}
}
