package repositories

import (
	"time"

	"github.com/ownafarm/ownafarm-backend/internal/models"
	"gorm.io/gorm"
)

// InvestmentFilter contains filter options for listing investments
type InvestmentFilter struct {
	UserID    string // Filter by user ID (required)
	Status    string // Filter by status (growing, ready, harvested)
	Page      int    // Current page (1-indexed)
	Limit     int    // Items per page
	SortBy    string // Field to sort (invested_at, progress, status)
	SortOrder string // asc or desc
}

// InvestmentRepository defines the interface for investment data access
type InvestmentRepository interface {
	Create(investment *models.Investment) error
	GetByID(id string) (*models.Investment, error)
	GetByIDWithRelations(id string) (*models.Investment, error)
	GetByIDAndUserID(id, userID string) (*models.Investment, error)
	GetByUserIDAndOnchainID(userID string, onchainID int64) (*models.Investment, error)
	GetAllByUserID(filter InvestmentFilter) ([]models.Investment, int64, error)
	Update(investment *models.Investment) error
	UpdateProgress(id string, progress int, status models.CropStatus) error
	IncrementWaterCount(id string) error
}

type investmentRepository struct {
	db *gorm.DB
}

// NewInvestmentRepository creates a new InvestmentRepository instance
func NewInvestmentRepository(db *gorm.DB) InvestmentRepository {
	return &investmentRepository{db: db}
}

// Create creates a new investment record
func (r *investmentRepository) Create(investment *models.Investment) error {
	return r.db.Create(investment).Error
}

// GetByID retrieves an investment by ID
func (r *investmentRepository) GetByID(id string) (*models.Investment, error) {
	var investment models.Investment
	if err := r.db.First(&investment, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &investment, nil
}

// GetByIDWithRelations retrieves an investment by ID with invoice and farm relations
func (r *investmentRepository) GetByIDWithRelations(id string) (*models.Investment, error) {
	var investment models.Investment
	if err := r.db.
		Preload("Invoice").
		Preload("Invoice.Farm").
		First(&investment, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &investment, nil
}

// GetByIDAndUserID retrieves an investment by ID with ownership check
func (r *investmentRepository) GetByIDAndUserID(id, userID string) (*models.Investment, error) {
	var investment models.Investment
	if err := r.db.
		Preload("Invoice").
		Preload("Invoice.Farm").
		Where("id = ? AND user_id = ?", id, userID).
		First(&investment).Error; err != nil {
		return nil, err
	}
	return &investment, nil
}

// GetByUserIDAndOnchainID retrieves an investment by user ID and onchain ID
func (r *investmentRepository) GetByUserIDAndOnchainID(userID string, onchainID int64) (*models.Investment, error) {
	var investment models.Investment
	if err := r.db.
		Where("user_id = ? AND investment_id_onchain = ?", userID, onchainID).
		First(&investment).Error; err != nil {
		return nil, err
	}
	return &investment, nil
}

// GetAllByUserID retrieves all investments for a user with filtering and pagination
func (r *investmentRepository) GetAllByUserID(filter InvestmentFilter) ([]models.Investment, int64, error) {
	var investments []models.Investment
	var totalCount int64

	query := r.db.Model(&models.Investment{}).
		Where("user_id = ?", filter.UserID)

	// Apply status filter
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}

	// Get total count before pagination
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	// Apply sorting
	sortBy := filter.SortBy
	if sortBy == "" {
		sortBy = "invested_at"
	}
	// Validate sort field
	validSortFields := map[string]bool{
		"invested_at": true,
		"progress":    true,
		"status":      true,
		"created_at":  true,
		"amount":      true,
	}
	if !validSortFields[sortBy] {
		sortBy = "invested_at"
	}

	sortOrder := filter.SortOrder
	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "desc"
	}

	query = query.Order(sortBy + " " + sortOrder)

	// Apply pagination
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Limit < 1 {
		filter.Limit = 10
	}
	if filter.Limit > 100 {
		filter.Limit = 100
	}

	offset := (filter.Page - 1) * filter.Limit
	query = query.Offset(offset).Limit(filter.Limit)

	// Execute query with preload
	if err := query.
		Preload("Invoice").
		Preload("Invoice.Farm").
		Find(&investments).Error; err != nil {
		return nil, 0, err
	}

	return investments, totalCount, nil
}

// Update updates an existing investment record
func (r *investmentRepository) Update(investment *models.Investment) error {
	return r.db.Save(investment).Error
}

// UpdateProgress updates the progress and status of an investment
func (r *investmentRepository) UpdateProgress(id string, progress int, status models.CropStatus) error {
	return r.db.Model(&models.Investment{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"progress":   progress,
			"status":     status,
			"updated_at": time.Now(),
		}).Error
}

// IncrementWaterCount increments the water count for an investment
func (r *investmentRepository) IncrementWaterCount(id string) error {
	now := time.Now()
	return r.db.Model(&models.Investment{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"water_count":     gorm.Expr("water_count + 1"),
			"last_watered_at": now,
			"updated_at":      now,
		}).Error
}
