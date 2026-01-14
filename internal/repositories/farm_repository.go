package repositories

import (
	"github.com/ownafarm/ownafarm-backend/internal/models"
	"gorm.io/gorm"
)

// FarmFilter contains filter options for listing farms
type FarmFilter struct {
	FarmerID  string // Filter by farmer ID
	IsActive  *bool  // Filter by active status
	Page      int    // Current page (1-indexed)
	Limit     int    // Items per page
	SortBy    string // Field to sort (created_at, name)
	SortOrder string // asc or desc
	Search    string // Search by name or location
}

// FarmRepository defines the interface for farm data access
type FarmRepository interface {
	Create(farm *models.Farm) error
	GetByID(id string) (*models.Farm, error)
	GetByIDAndFarmerID(id, farmerID string) (*models.Farm, error)
	GetAllByFarmerID(farmerID string, filter FarmFilter) ([]models.Farm, int64, error)
	Update(farm *models.Farm) error
	SoftDelete(id string) error
	HasActiveInvoices(farmID string) (bool, error)
}

type farmRepository struct {
	db *gorm.DB
}

// NewFarmRepository creates a new FarmRepository instance
func NewFarmRepository(db *gorm.DB) FarmRepository {
	return &farmRepository{db: db}
}

// Create creates a new farm record
func (r *farmRepository) Create(farm *models.Farm) error {
	return r.db.Create(farm).Error
}

// GetByID retrieves a farm by ID
func (r *farmRepository) GetByID(id string) (*models.Farm, error) {
	var farm models.Farm
	if err := r.db.First(&farm, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &farm, nil
}

// GetByIDAndFarmerID retrieves a farm by ID and farmer ID (ownership check)
func (r *farmRepository) GetByIDAndFarmerID(id, farmerID string) (*models.Farm, error) {
	var farm models.Farm
	if err := r.db.First(&farm, "id = ? AND farmer_id = ?", id, farmerID).Error; err != nil {
		return nil, err
	}
	return &farm, nil
}

// GetAllByFarmerID retrieves all farms for a specific farmer with pagination
func (r *farmRepository) GetAllByFarmerID(farmerID string, filter FarmFilter) ([]models.Farm, int64, error) {
	var farms []models.Farm
	var totalCount int64

	query := r.db.Model(&models.Farm{}).Where("farmer_id = ?", farmerID)

	// Apply active status filter
	if filter.IsActive != nil {
		query = query.Where("is_active = ?", *filter.IsActive)
	}

	// Apply search filter
	if filter.Search != "" {
		searchPattern := "%" + filter.Search + "%"
		query = query.Where("name ILIKE ? OR location ILIKE ?", searchPattern, searchPattern)
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
	// Validate sort field
	validSortFields := map[string]bool{
		"created_at": true,
		"name":       true,
		"location":   true,
		"land_area":  true,
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
	if err := query.Find(&farms).Error; err != nil {
		return nil, 0, err
	}

	return farms, totalCount, nil
}

// Update updates an existing farm record
func (r *farmRepository) Update(farm *models.Farm) error {
	return r.db.Save(farm).Error
}

// SoftDelete sets a farm's is_active to false
func (r *farmRepository) SoftDelete(id string) error {
	return r.db.Model(&models.Farm{}).Where("id = ?", id).Update("is_active", false).Error
}

// HasActiveInvoices checks if a farm has any invoices that are not rejected
func (r *farmRepository) HasActiveInvoices(farmID string) (bool, error) {
	var count int64
	err := r.db.Model(&models.Invoice{}).
		Where("farm_id = ? AND status != ?", farmID, models.InvoiceStatusRejected).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
