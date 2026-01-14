package repositories

import (
	"github.com/ownafarm/ownafarm-backend/internal/models"
	"gorm.io/gorm"
)

// InvoiceFilter contains filter options for listing invoices
type InvoiceFilter struct {
	FarmID                    string   // Filter by farm ID
	FarmerID                  string   // Filter by farmer ID (via farm relation)
	Status                    []string // Filter by status (pending, approved, rejected)
	Page                      int      // Current page (1-indexed)
	Limit                     int      // Items per page
	SortBy                    string   // Field to sort (created_at, name, target_fund, yield_percent, duration_days)
	SortOrder                 string   // asc or desc
	Search                    string   // Search by name
	IsAvailableForInvestment  bool     // Filter for marketplace (approved + not fully funded)
	MinTargetFund             *float64 // Min target fund filter
	MaxTargetFund             *float64 // Max target fund filter
	MinYield                  *float64 // Min yield percent filter
	MaxYield                  *float64 // Max yield percent filter
	MinDuration               *int     // Min duration days filter
	MaxDuration               *int     // Max duration days filter
	MinLandArea               *float64 // Min farm land area filter
	MaxLandArea               *float64 // Max farm land area filter
	Location                  string   // Filter by farm location (ILIKE search)
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
	GetAvailableForInvestment(filter InvoiceFilter) ([]models.Invoice, int64, error)
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

// GetAvailableForInvestment retrieves approved and available invoices for marketplace
func (r *invoiceRepository) GetAvailableForInvestment(filter InvoiceFilter) ([]models.Invoice, int64, error) {
	var invoices []models.Invoice
	var totalCount int64

	// Base query with JOIN to farms for location and land area filters
	query := r.db.Model(&models.Invoice{}).
		Joins("LEFT JOIN farms ON farms.id = invoices.farm_id")

	// Marketplace base filters: approved and not fully funded
	query = query.Where("invoices.status = ?", "approved").
		Where("invoices.is_fully_funded = ?", false)

	// Apply range filters for target_fund
	if filter.MinTargetFund != nil {
		query = query.Where("invoices.target_fund >= ?", *filter.MinTargetFund)
	}
	if filter.MaxTargetFund != nil {
		query = query.Where("invoices.target_fund <= ?", *filter.MaxTargetFund)
	}

	// Apply range filters for yield_percent
	if filter.MinYield != nil {
		query = query.Where("invoices.yield_percent >= ?", *filter.MinYield)
	}
	if filter.MaxYield != nil {
		query = query.Where("invoices.yield_percent <= ?", *filter.MaxYield)
	}

	// Apply range filters for duration_days
	if filter.MinDuration != nil {
		query = query.Where("invoices.duration_days >= ?", *filter.MinDuration)
	}
	if filter.MaxDuration != nil {
		query = query.Where("invoices.duration_days <= ?", *filter.MaxDuration)
	}

	// Apply range filters for farm land_area
	if filter.MinLandArea != nil {
		query = query.Where("farms.land_area >= ?", *filter.MinLandArea)
	}
	if filter.MaxLandArea != nil {
		query = query.Where("farms.land_area <= ?", *filter.MaxLandArea)
	}

	// Apply location filter (farm location ILIKE search)
	if filter.Location != "" {
		locationPattern := "%" + filter.Location + "%"
		query = query.Where("farms.location ILIKE ?", locationPattern)
	}

	// Apply search filter (invoice name ILIKE search)
	if filter.Search != "" {
		searchPattern := "%" + filter.Search + "%"
		query = query.Where("invoices.name ILIKE ?", searchPattern)
	}

	// Get total count before pagination
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	// Apply sorting with validation
	sortBy := filter.SortBy
	if sortBy == "" {
		sortBy = "invoices.yield_percent" // Default: highest yield first
	} else {
		sortBy = "invoices." + sortBy
	}

	// Validate sort field
	validSortFields := map[string]bool{
		"invoices.created_at":    true,
		"invoices.name":          true,
		"invoices.target_fund":   true,
		"invoices.yield_percent": true,
		"invoices.duration_days": true,
	}
	if !validSortFields[sortBy] {
		sortBy = "invoices.yield_percent"
	}

	sortOrder := filter.SortOrder
	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "desc" // Default: descending (highest first)
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

	// Execute query with preloads
	if err := query.Preload("Farm").Preload("Farm.Farmer").Find(&invoices).Error; err != nil {
		return nil, 0, err
	}

	return invoices, totalCount, nil
}
