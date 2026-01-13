package models

import (
	"encoding/json"
	"time"
)

// AdminAuditLog represents the admin_audit_logs table in the database
type AdminAuditLog struct {
	ID         string          `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	AdminID    string          `gorm:"type:uuid;not null" json:"admin_id"`
	Action     string          `gorm:"type:varchar(100);not null" json:"action"`
	EntityType string          `gorm:"type:varchar(50);not null" json:"entity_type"`
	EntityID   string          `gorm:"type:uuid;not null" json:"entity_id"`
	OldValues  json.RawMessage `gorm:"type:jsonb" json:"old_values,omitempty"`
	NewValues  json.RawMessage `gorm:"type:jsonb" json:"new_values,omitempty"`
	IPAddress  *string         `gorm:"type:inet" json:"ip_address,omitempty"`
	UserAgent  *string         `gorm:"type:text" json:"user_agent,omitempty"`
	CreatedAt  time.Time       `gorm:"default:now()" json:"created_at"`
}

// TableName returns the table name for the AdminAuditLog model
func (AdminAuditLog) TableName() string {
	return "admin_audit_logs"
}

// Audit log action constants
const (
	AuditActionApproveFarmer = "approve_farmer"
	AuditActionRejectFarmer  = "reject_farmer"
)

// Audit log entity type constants
const (
	AuditEntityTypeFarmer = "farmer"
)
