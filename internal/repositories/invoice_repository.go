package repositories

import (
	"github.com/ownafarm/ownafarm-backend/internal/models"
	"gorm.io/gorm"
)

// InvoiceFilter contains filter options for listing invoices
type InvoiceFilter struct {
	FarmID    string   // Filter by farm ID
	FarmerID  string   // Filter by farmer ID (via farm relation)
	Status    []string // Filter by status (pending, approved, rejected)
	Page      int      // Current page (1-indexed)
	Limit     int      // Items per page
	SortBy    string   // Field to sort (created_at, name, target_fund)
	SortOrder string   // asc or desc
	Search    string   // Search by name
}

// InvoiceRepository defines the interface for invoice data access
type InvoiceRepository interface {
	Create(invoice *models.Invoice) error
	GetByID(id string) (*models.Invoice, error)
	GetByIDWithFarm(id string) (*models.Invoice, error)
	GetByIDAndFarmerID(id, farmerID string) (*models.Invoice, error)
	GetByTokenID(tokenID int64) (*models.Invoice, error)
	GetAllByFarmerID(farmerID string, filter InvoiceFilter) ([]models.Invoice, int64, error)
	GetAllWithPagination(filter InvoiceFilter) ([]models.Invoice, int64, error)
	Update(invoice *models.Invoice) error
}

type invoiceRepository struct {
	db *gorm.DB
}

// NewInvoiceRepository creates a new InvoiceRepository instance
func NewInvoiceRepository(db *gorm.DB) InvoiceRepository {
	return &invoiceRepository{db: db}
}

// Create creates a new invoice record
func (r *invoiceRepository) Create(invoice *models.Invoice) error {
	return r.db.Create(invoice).Error
}

// GetByID retrieves an invoice by ID
func (r *invoiceRepository) GetByID(id string) (*models.Invoice, error) {
	var invoice models.Invoice
	if err := r.db.First(&invoice, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &invoice, nil
}

// GetByIDWithFarm retrieves an invoice by ID with farm relation
func (r *invoiceRepository) GetByIDWithFarm(id string) (*models.Invoice, error) {
	var invoice models.Invoice
	if err := r.db.Preload("Farm").First(&invoice, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &invoice, nil
}

// GetByIDAndFarmerID retrieves an invoice by ID and farmer ID (ownership check via farm)
func (r *invoiceRepository) GetByIDAndFarmerID(id, farmerID string) (*models.Invoice, error) {
	var invoice models.Invoice
	if err := r.db.
		Joins("JOIN farms ON farms.id = invoices.farm_id").
		Where("invoices.id = ? AND farms.farmer_id = ?", id, farmerID).
		First(&invoice).Error; err != nil {
		return nil, err
	}
	return &invoice, nil
}

// GetAllByFarmerID retrieves all invoices for a specific farmer with pagination
func (r *invoiceRepository) GetAllByFarmerID(farmerID string, filter InvoiceFilter) ([]models.Invoice, int64, error) {
	var invoices []models.Invoice
	var totalCount int64

	query := r.db.Model(&models.Invoice{}).
		Joins("JOIN farms ON farms.id = invoices.farm_id").
		Where("farms.farmer_id = ?", farmerID)

	// Apply farm filter
	if filter.FarmID != "" {
		query = query.Where("invoices.farm_id = ?", filter.FarmID)
	}

	// Apply status filter
	if len(filter.Status) > 0 {
		query = query.Where("invoices.status IN ?", filter.Status)
	}

	// Apply search filter
	if filter.Search != "" {
		searchPattern := "%" + filter.Search + "%"
		query = query.Where("invoices.name ILIKE ?", searchPattern)
	}

	// Get total count before pagination
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	// Apply sorting
	sortBy := filter.SortBy
	if sortBy == "" {
		sortBy = "invoices.created_at"
	} else {
		sortBy = "invoices." + sortBy
	}
	// Validate sort field
	validSortFields := map[string]bool{
		"invoices.created_at":  true,
		"invoices.name":        true,
		"invoices.target_fund": true,
		"invoices.status":      true,
	}
	if !validSortFields[sortBy] {
		sortBy = "invoices.created_at"
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
	if err := query.Preload("Farm").Find(&invoices).Error; err != nil {
		return nil, 0, err
	}

	return invoices, totalCount, nil
}

// GetAllWithPagination retrieves all invoices with filtering and pagination (for admin)
func (r *invoiceRepository) GetAllWithPagination(filter InvoiceFilter) ([]models.Invoice, int64, error) {
	var invoices []models.Invoice
	var totalCount int64

	query := r.db.Model(&models.Invoice{})

	// Apply status filter
	if len(filter.Status) > 0 {
		query = query.Where("status IN ?", filter.Status)
	}

	// Apply search filter
	if filter.Search != "" {
		searchPattern := "%" + filter.Search + "%"
		query = query.Where("name ILIKE ?", searchPattern)
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
		"created_at":  true,
		"name":        true,
		"target_fund": true,
		"status":      true,
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

	// Execute query with preload
	if err := query.Preload("Farm").Preload("Farm.Farmer").Find(&invoices).Error; err != nil {
		return nil, 0, err
	}

	return invoices, totalCount, nil
}

// GetByTokenID retrieves an invoice by its blockchain token ID
func (r *invoiceRepository) GetByTokenID(tokenID int64) (*models.Invoice, error) {
	var invoice models.Invoice
	if err := r.db.Preload("Farm").First(&invoice, "token_id = ?", tokenID).Error; err != nil {
		return nil, err
	}
	return &invoice, nil
}

// Update updates an existing invoice record
func (r *invoiceRepository) Update(invoice *models.Invoice) error {
	return r.db.Save(invoice).Error
}
