package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// CropStatus represents the status of a crop/investment
type CropStatus string

const (
	CropStatusGrowing   CropStatus = "growing"
	CropStatusReady     CropStatus = "ready"
	CropStatusHarvested CropStatus = "harvested"
)

// Investment represents the investments table in the database
type Investment struct {
	ID        string `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID    string `gorm:"type:uuid;not null" json:"user_id"`
	InvoiceID string `gorm:"type:uuid;not null" json:"invoice_id"`

	// Blockchain Reference
	InvestmentIdOnchain *int64 `gorm:"type:bigint" json:"investment_id_onchain,omitempty"`

	// Investment Details
	Amount     decimal.Decimal `gorm:"type:decimal(20,8);not null" json:"amount"`
	InvestedAt time.Time       `gorm:"not null;default:now()" json:"invested_at"`

	// Game State
	Status        CropStatus `gorm:"type:crop_status;default:growing" json:"status"`
	Progress      int        `gorm:"default:0" json:"progress"`
	WaterCount    int        `gorm:"default:0" json:"water_count"`
	LastWateredAt *time.Time `json:"last_watered_at,omitempty"`

	// Harvest
	IsHarvested   bool             `gorm:"default:false" json:"is_harvested"`
	HarvestedAt   *time.Time       `json:"harvested_at,omitempty"`
	HarvestAmount *decimal.Decimal `gorm:"type:decimal(20,8)" json:"harvest_amount,omitempty"`
	HarvestTxHash *string          `gorm:"type:varchar(66)" json:"harvest_tx_hash,omitempty"`

	// TX Reference
	PurchaseTxHash *string `gorm:"type:varchar(66)" json:"purchase_tx_hash,omitempty"`

	// Timestamps
	CreatedAt time.Time `gorm:"default:now()" json:"created_at"`
	UpdatedAt time.Time `gorm:"default:now()" json:"updated_at"`

	// Relations
	User    User    `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Invoice Invoice `gorm:"foreignKey:InvoiceID" json:"invoice,omitempty"`
}

// TableName returns the table name for the Investment model
func (Investment) TableName() string {
	return "investments"
}
