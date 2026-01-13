package models

import (
	"time"
)

// AdminUser represents the admin_users table in the database
type AdminUser struct {
	ID            string     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	WalletAddress string     `gorm:"type:varchar(42);uniqueIndex;not null" json:"wallet_address"`
	Role          string     `gorm:"type:varchar(50);default:admin" json:"role"`
	IsActive      bool       `gorm:"default:true" json:"is_active"`
	LastLoginAt   *time.Time `json:"last_login_at,omitempty"`
	CreatedAt     time.Time  `gorm:"default:now()" json:"created_at"`
	UpdatedAt     time.Time  `gorm:"default:now()" json:"updated_at"`
}

// TableName returns the table name for the AdminUser model
func (AdminUser) TableName() string {
	return "admin_users"
}
