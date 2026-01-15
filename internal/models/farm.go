package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// Farm represents the farms table in the database
type Farm struct {
	ID        string `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	FarmerID  string `gorm:"type:uuid;not null" json:"farmer_id"`

	// Basic Info
	Name        string  `gorm:"type:varchar(200);not null" json:"name"`
	Description *string `gorm:"type:text" json:"description,omitempty"`
	Location    string  `gorm:"type:varchar(255);not null" json:"location"`

	// Coordinates
	Latitude  *decimal.Decimal `gorm:"type:decimal(10,8)" json:"latitude,omitempty"`
	Longitude *decimal.Decimal `gorm:"type:decimal(11,8)" json:"longitude,omitempty"`

	// Land Info
	LandArea *decimal.Decimal `gorm:"type:decimal(10,2)" json:"land_area,omitempty"` // in hectares

	// CCTV Monitoring
	CCTVUrl         *string    `gorm:"type:text;column:cctv_url" json:"cctv_url,omitempty"`
	CCTVImageUrl    *string    `gorm:"type:text;column:cctv_image_url" json:"cctv_image_url,omitempty"`
	CCTVLastUpdated *time.Time `gorm:"column:cctv_last_updated" json:"cctv_last_updated,omitempty"`

	// Status
	IsActive bool `gorm:"default:true" json:"is_active"`

	// Timestamps
	CreatedAt time.Time `gorm:"default:now()" json:"created_at"`
	UpdatedAt time.Time `gorm:"default:now()" json:"updated_at"`

	// Relations
	Farmer   Farmer    `gorm:"foreignKey:FarmerID" json:"farmer,omitempty"`
	Invoices []Invoice `gorm:"foreignKey:FarmID" json:"invoices,omitempty"`
}

// TableName returns the table name for the Farm model
func (Farm) TableName() string {
	return "farms"
}
