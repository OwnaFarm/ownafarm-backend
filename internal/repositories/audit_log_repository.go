package repositories

import (
	"github.com/ownafarm/ownafarm-backend/internal/models"
	"gorm.io/gorm"
)

// AuditLogRepository defines the interface for audit log data access
type AuditLogRepository interface {
	Create(log *models.AdminAuditLog) error
}

type auditLogRepository struct {
	db *gorm.DB
}

// NewAuditLogRepository creates a new AuditLogRepository instance
func NewAuditLogRepository(db *gorm.DB) AuditLogRepository {
	return &auditLogRepository{db: db}
}

// Create creates a new audit log record
func (r *auditLogRepository) Create(log *models.AdminAuditLog) error {
	return r.db.Create(log).Error
}
