package repositories

import (
	"github.com/ownafarm/ownafarm-backend/internal/models"
	"gorm.io/gorm"
)

// FarmerFilter contains filter options for listing farmers
type FarmerFilter struct {
	Status    []string // Filter by status (pending, under_review, approved, rejected, suspended)
	Page      int      // Current page (1-indexed)
	Limit     int      // Items per page
	SortBy    string   // Field to sort (created_at, full_name, status)
	SortOrder string   // asc or desc
	Search    string   // Search by name, email, or phone
}

// FarmerRepository defines the interface for farmer data access
type FarmerRepository interface {
	Create(farmer *models.Farmer) error
	GetByID(id string) (*models.Farmer, error)
	GetByUserID(userID string) (*models.Farmer, error)
	GetByEmail(email string) (*models.Farmer, error)
	GetByWalletAddress(walletAddress string) (*models.Farmer, error)
	ExistsByEmailOrPhone(email, phone string) (bool, error)
	ExistsByWalletAddress(walletAddress string) (bool, error)
	CreateDocuments(documents []models.FarmerDocument) error
	GetAllWithPagination(filter FarmerFilter) ([]models.Farmer, int64, error)
	Update(farmer *models.Farmer) error
}

type farmerRepository struct {
	db *gorm.DB
}

// NewFarmerRepository creates a new FarmerRepository instance
func NewFarmerRepository(db *gorm.DB) FarmerRepository {
	return &farmerRepository{db: db}
}

// Create creates a new farmer record
func (r *farmerRepository) Create(farmer *models.Farmer) error {
	return r.db.Create(farmer).Error
}

// GetByID retrieves a farmer by ID with documents
func (r *farmerRepository) GetByID(id string) (*models.Farmer, error) {
	var farmer models.Farmer
	if err := r.db.Preload("Documents").First(&farmer, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &farmer, nil
}

// GetByEmail retrieves a farmer by email
func (r *farmerRepository) GetByEmail(email string) (*models.Farmer, error) {
	var farmer models.Farmer
	if err := r.db.First(&farmer, "email = ?", email).Error; err != nil {
		return nil, err
	}
	return &farmer, nil
}

// ExistsByEmailOrPhone checks if a farmer with the given email or phone already exists
func (r *farmerRepository) ExistsByEmailOrPhone(email, phone string) (bool, error) {
	var count int64
	err := r.db.Model(&models.Farmer{}).
		Where("email = ? OR phone_number = ?", email, phone).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// ExistsByWalletAddress checks if a farmer with the given wallet address already exists
func (r *farmerRepository) ExistsByWalletAddress(walletAddress string) (bool, error) {
	var count int64
	err := r.db.Model(&models.Farmer{}).
		Where("wallet_address = ?", walletAddress).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// CreateDocuments creates multiple farmer document records
func (r *farmerRepository) CreateDocuments(documents []models.FarmerDocument) error {
	if len(documents) == 0 {
		return nil
	}
	return r.db.Create(&documents).Error
}

// GetAllWithPagination retrieves farmers with filtering, sorting, and pagination
func (r *farmerRepository) GetAllWithPagination(filter FarmerFilter) ([]models.Farmer, int64, error) {
	var farmers []models.Farmer
	var totalCount int64

	query := r.db.Model(&models.Farmer{})

	// Apply status filter
	if len(filter.Status) > 0 {
		query = query.Where("status IN ?", filter.Status)
	}

	// Apply search filter
	if filter.Search != "" {
		searchPattern := "%" + filter.Search + "%"
		query = query.Where(
			"full_name ILIKE ? OR email ILIKE ? OR phone_number ILIKE ?",
			searchPattern, searchPattern, searchPattern,
		)
	}

	// Get total count before pagination
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	// Apply sorting
	sortBy := filter.SortBy
	if sortBy == "" {
		sortBy = "created_at"
	}
	// Validate sort field to prevent SQL injection
	validSortFields := map[string]bool{
		"created_at": true,
		"full_name":  true,
		"status":     true,
		"email":      true,
		"province":   true,
		"city":       true,
	}
	if !validSortFields[sortBy] {
		sortBy = "created_at"
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

	// Execute query
	if err := query.Find(&farmers).Error; err != nil {
		return nil, 0, err
	}

	return farmers, totalCount, nil
}

// Update updates an existing farmer record
func (r *farmerRepository) Update(farmer *models.Farmer) error {
	return r.db.Save(farmer).Error
}

// GetByUserID retrieves a farmer by user ID
func (r *farmerRepository) GetByUserID(userID string) (*models.Farmer, error) {
	var farmer models.Farmer
	if err := r.db.First(&farmer, "user_id = ?", userID).Error; err != nil {
		return nil, err
	}
	return &farmer, nil
}

// GetByWalletAddress retrieves a farmer by wallet address
func (r *farmerRepository) GetByWalletAddress(walletAddress string) (*models.Farmer, error) {
	var farmer models.Farmer
	if err := r.db.First(&farmer, "wallet_address = ?", walletAddress).Error; err != nil {
		return nil, err
	}
	return &farmer, nil
}
