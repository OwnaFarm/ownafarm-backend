package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// InvoiceStatus represents the status of an invoice
type InvoiceStatus string

const (
	InvoiceStatusPending  InvoiceStatus = "pending"
	InvoiceStatusApproved InvoiceStatus = "approved"
	InvoiceStatusRejected InvoiceStatus = "rejected"
)

// Invoice represents the invoices table in the database
type Invoice struct {
	ID     string `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	FarmID string `gorm:"type:uuid;not null" json:"farm_id"`

	// Blockchain Reference
	TokenID    *int64  `gorm:"type:bigint;unique" json:"token_id,omitempty"`
	OfftakerID *string `gorm:"type:varchar(100)" json:"offtaker_id,omitempty"`

	// Invoice Details
	Name        string  `gorm:"type:varchar(200);not null" json:"name"`
	Description *string `gorm:"type:text" json:"description,omitempty"`
	ImageURL    *string `gorm:"type:text" json:"image_url,omitempty"`

	// Financial
	TargetFund   decimal.Decimal `gorm:"type:decimal(20,8);not null" json:"target_fund"`
	YieldPercent decimal.Decimal `gorm:"type:decimal(5,2);not null" json:"yield_percent"`
	DurationDays int             `gorm:"not null" json:"duration_days"`

	// Funding Status (synced from blockchain)
	TotalFunded   decimal.Decimal `gorm:"type:decimal(20,8);default:0" json:"total_funded"`
	IsFullyFunded bool            `gorm:"default:false" json:"is_fully_funded"`

	// Approval Status
	Status          InvoiceStatus `gorm:"type:invoice_status;not null;default:pending" json:"status"`
	RejectionReason *string       `gorm:"type:text" json:"rejection_reason,omitempty"`
	ReviewedBy      *string       `gorm:"type:uuid" json:"reviewed_by,omitempty"`
	ReviewedAt      *time.Time    `json:"reviewed_at,omitempty"`

	// Dates
	ApprovedAt      *time.Time `json:"approved_at,omitempty"`
	FundingDeadline *time.Time `json:"funding_deadline,omitempty"`
	MaturityDate    *time.Time `json:"maturity_date,omitempty"`

	// Blockchain TX
	ApprovalTxHash *string `gorm:"type:varchar(66)" json:"approval_tx_hash,omitempty"`

	// Timestamps
	CreatedAt time.Time `gorm:"default:now()" json:"created_at"`
	UpdatedAt time.Time `gorm:"default:now()" json:"updated_at"`

	// Relations
	Farm Farm `gorm:"foreignKey:FarmID" json:"farm,omitempty"`
}

// TableName returns the table name for the Invoice model
func (Invoice) TableName() string {
	return "invoices"
}
