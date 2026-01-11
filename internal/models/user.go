package models

import (
	"time"
)

// User represents the users table in the database
type User struct {
	ID            string  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	WalletAddress string  `gorm:"type:varchar(42);uniqueIndex:idx_users_wallet_address;not null" json:"wallet_address"`
	Name          *string `gorm:"type:varchar(100)" json:"name,omitempty"`
	Email         *string `gorm:"type:varchar(255);unique" json:"email,omitempty"`
	Avatar        *string `gorm:"type:varchar(50)" json:"avatar,omitempty"`

	// Game Stats
	Level       int        `gorm:"default:1" json:"level"`
	XP          int        `gorm:"column:xp;default:0" json:"xp"`
	WaterPoints int        `gorm:"default:100" json:"water_points"`
	LastRegenAt *time.Time `gorm:"default:now()" json:"last_regen_at,omitempty"`

	// Timestamps
	LastLoginAt *time.Time `gorm:"" json:"last_login_at,omitempty"`
	CreatedAt   time.Time  `gorm:"default:now()" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"default:now()" json:"updated_at"`
}

// TableName returns the table name for the User model
func (User) TableName() string {
	return "users"
}
