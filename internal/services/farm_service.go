package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/ownafarm/ownafarm-backend/internal/dto/request"
	"github.com/ownafarm/ownafarm-backend/internal/dto/response"
	"github.com/ownafarm/ownafarm-backend/internal/models"
	"github.com/ownafarm/ownafarm-backend/internal/repositories"
	"github.com/shopspring/decimal"
)

// Errors for FarmService
var (
	ErrFarmNotFound           = errors.New("farm not found")
	ErrFarmNotOwned           = errors.New("farm does not belong to this farmer")
	ErrFarmHasActiveInvoices  = errors.New("cannot delete farm with active invoices")
)

// FarmServiceInterface defines the interface for farm operations
type FarmServiceInterface interface {
	Create(ctx context.Context, farmerID string, req *request.CreateFarmRequest) (*response.FarmResponse, error)
	GetByID(ctx context.Context, farmerID, farmID string) (*response.FarmResponse, error)
	List(ctx context.Context, farmerID string, req *request.ListFarmRequest) (*response.ListFarmResponse, error)
	Update(ctx context.Context, farmerID, farmID string, req *request.UpdateFarmRequest) (*response.FarmResponse, error)
	Delete(ctx context.Context, farmerID, farmID string) error
}

// FarmService implements FarmServiceInterface
type FarmService struct {
	farmRepo repositories.FarmRepository
}

// NewFarmService creates a new FarmService instance
func NewFarmService(farmRepo repositories.FarmRepository) *FarmService {
	return &FarmService{
		farmRepo: farmRepo,
	}
}

// Create creates a new farm for a farmer
func (s *FarmService) Create(ctx context.Context, farmerID string, req *request.CreateFarmRequest) (*response.FarmResponse, error) {
	farm := &models.Farm{
		FarmerID:    farmerID,
		Name:        req.Name,
		Description: req.Description,
		Location:    req.Location,
		IsActive:    true,
	}

	// Set optional coordinates
	if req.Latitude != nil {
		lat := decimal.NewFromFloat(*req.Latitude)
		farm.Latitude = &lat
	}
	if req.Longitude != nil {
		lng := decimal.NewFromFloat(*req.Longitude)
		farm.Longitude = &lng
	}
	if req.LandArea != nil {
		area := decimal.NewFromFloat(*req.LandArea)
		farm.LandArea = &area
	}

	if err := s.farmRepo.Create(farm); err != nil {
		return nil, fmt.Errorf("failed to create farm: %w", err)
	}

	return s.toFarmResponse(farm), nil
}

// GetByID retrieves a farm by ID (with ownership check)
func (s *FarmService) GetByID(ctx context.Context, farmerID, farmID string) (*response.FarmResponse, error) {
	farm, err := s.farmRepo.GetByIDAndFarmerID(farmID, farmerID)
	if err != nil {
		return nil, ErrFarmNotFound
	}

	return s.toFarmResponse(farm), nil
}

// List retrieves all farms for a farmer with pagination
func (s *FarmService) List(ctx context.Context, farmerID string, req *request.ListFarmRequest) (*response.ListFarmResponse, error) {
	// Set defaults
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit < 1 {
		req.Limit = 10
	}

	filter := repositories.FarmFilter{
		FarmerID:  farmerID,
		IsActive:  req.IsActive,
		Page:      req.Page,
		Limit:     req.Limit,
		SortBy:    req.SortBy,
		SortOrder: req.SortOrder,
		Search:    req.Search,
	}

	farms, totalCount, err := s.farmRepo.GetAllByFarmerID(farmerID, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get farms: %w", err)
	}

	// Map to response
	farmItems := make([]response.FarmListItem, 0, len(farms))
	for _, farm := range farms {
		farmItems = append(farmItems, response.FarmListItem{
			ID:        farm.ID,
			Name:      farm.Name,
			Location:  farm.Location,
			LandArea:  farm.LandArea,
			IsActive:  farm.IsActive,
			CreatedAt: farm.CreatedAt,
		})
	}

	// Calculate total pages
	totalPages := int(totalCount) / req.Limit
	if int(totalCount)%req.Limit > 0 {
		totalPages++
	}

	return &response.ListFarmResponse{
		Farms: farmItems,
		Pagination: response.PaginationMeta{
			Page:       req.Page,
			Limit:      req.Limit,
			TotalItems: totalCount,
			TotalPages: totalPages,
		},
	}, nil
}

// Update updates an existing farm
func (s *FarmService) Update(ctx context.Context, farmerID, farmID string, req *request.UpdateFarmRequest) (*response.FarmResponse, error) {
	farm, err := s.farmRepo.GetByIDAndFarmerID(farmID, farmerID)
	if err != nil {
		return nil, ErrFarmNotFound
	}

	// Update fields if provided
	if req.Name != nil {
		farm.Name = *req.Name
	}
	if req.Description != nil {
		farm.Description = req.Description
	}
	if req.Location != nil {
		farm.Location = *req.Location
	}
	if req.Latitude != nil {
		lat := decimal.NewFromFloat(*req.Latitude)
		farm.Latitude = &lat
	}
	if req.Longitude != nil {
		lng := decimal.NewFromFloat(*req.Longitude)
		farm.Longitude = &lng
	}
	if req.LandArea != nil {
		area := decimal.NewFromFloat(*req.LandArea)
		farm.LandArea = &area
	}
	if req.CCTVUrl != nil {
		farm.CCTVUrl = req.CCTVUrl
	}
	if req.CCTVImageUrl != nil {
		farm.CCTVImageUrl = req.CCTVImageUrl
	}

	if err := s.farmRepo.Update(farm); err != nil {
		return nil, fmt.Errorf("failed to update farm: %w", err)
	}

	return s.toFarmResponse(farm), nil
}

// Delete soft deletes a farm
func (s *FarmService) Delete(ctx context.Context, farmerID, farmID string) error {
	// Verify ownership
	_, err := s.farmRepo.GetByIDAndFarmerID(farmID, farmerID)
	if err != nil {
		return ErrFarmNotFound
	}

	// Soft delete (set is_active = false)
	if err := s.farmRepo.SoftDelete(farmID); err != nil {
		return fmt.Errorf("failed to delete farm: %w", err)
	}

	return nil
}

// toFarmResponse converts a Farm model to FarmResponse
func (s *FarmService) toFarmResponse(farm *models.Farm) *response.FarmResponse {
	return &response.FarmResponse{
		ID:              farm.ID,
		FarmerID:        farm.FarmerID,
		Name:            farm.Name,
		Description:     farm.Description,
		Location:        farm.Location,
		Latitude:        farm.Latitude,
		Longitude:       farm.Longitude,
		LandArea:        farm.LandArea,
		CCTVUrl:         farm.CCTVUrl,
		CCTVImageUrl:    farm.CCTVImageUrl,
		CCTVLastUpdated: farm.CCTVLastUpdated,
		IsActive:        farm.IsActive,
		CreatedAt:       farm.CreatedAt,
		UpdatedAt:       farm.UpdatedAt,
	}
}
