package models

import (
	"time"

	"github.com/lib/pq"
)

// FarmerStatus represents the status of a farmer registration
type FarmerStatus string

const (
	FarmerStatusPending     FarmerStatus = "pending"
	FarmerStatusUnderReview FarmerStatus = "under_review"
	FarmerStatusApproved    FarmerStatus = "approved"
	FarmerStatusRejected    FarmerStatus = "rejected"
	FarmerStatusSuspended   FarmerStatus = "suspended"
)

// BusinessType represents the type of business
type BusinessType string

const (
	BusinessTypeIndividual  BusinessType = "individual"
	BusinessTypeCV          BusinessType = "cv"
	BusinessTypePT          BusinessType = "pt"
	BusinessTypeUD          BusinessType = "ud"
	BusinessTypeCooperative BusinessType = "cooperative"
)

// Farmer represents the farmers table in the database
type Farmer struct {
	ID            string       `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID        *string      `gorm:"type:uuid" json:"user_id,omitempty"`
	Status        FarmerStatus `gorm:"type:farmer_status;default:pending" json:"status"`
	WalletAddress string       `gorm:"type:varchar(42);not null;uniqueIndex" json:"wallet_address"`

	// Step 1: Personal Info
	FullName    string    `gorm:"type:varchar(100);not null" json:"full_name"`
	Email       string    `gorm:"type:varchar(255);not null" json:"email"`
	PhoneNumber string    `gorm:"type:varchar(20);not null" json:"phone_number"`
	IDNumber    string    `gorm:"type:varchar(20);not null" json:"id_number"`
	DateOfBirth time.Time `gorm:"type:date;not null" json:"date_of_birth"`
	Address     string    `gorm:"type:text;not null" json:"address"`
	Province    string    `gorm:"type:varchar(100);not null" json:"province"`
	City        string    `gorm:"type:varchar(100);not null" json:"city"`
	District    string    `gorm:"type:varchar(100);not null" json:"district"`
	PostalCode  string    `gorm:"type:varchar(10);not null" json:"postal_code"`

	// Step 2: Business Info
	BusinessName      *string        `gorm:"type:varchar(200)" json:"business_name,omitempty"`
	BusinessType      BusinessType   `gorm:"type:business_type;not null" json:"business_type"`
	NPWP              *string        `gorm:"type:varchar(30)" json:"npwp,omitempty"`
	BankName          string         `gorm:"type:varchar(100);not null" json:"bank_name"`
	BankAccountNumber string         `gorm:"type:varchar(30);not null" json:"bank_account_number"`
	BankAccountName   string         `gorm:"type:varchar(100);not null" json:"bank_account_name"`
	YearsOfExperience int            `gorm:"default:0" json:"years_of_experience"`
	CropsExpertise    pq.StringArray `gorm:"type:text[]" json:"crops_expertise"`

	// Admin
	ReviewedBy      *string    `gorm:"type:uuid" json:"reviewed_by,omitempty"`
	ReviewedAt      *time.Time `json:"reviewed_at,omitempty"`
	RejectionReason *string    `gorm:"type:text" json:"rejection_reason,omitempty"`

	// Timestamps
	CreatedAt time.Time `gorm:"default:now()" json:"created_at"`
	UpdatedAt time.Time `gorm:"default:now()" json:"updated_at"`

	// Relations
	Documents []FarmerDocument `gorm:"foreignKey:FarmerID" json:"documents,omitempty"`
}

// TableName returns the table name for the Farmer model
func (Farmer) TableName() string {
	return "farmers"
}
