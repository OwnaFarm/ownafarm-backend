package response

import (
	"time"

	"github.com/shopspring/decimal"
)

// InvoiceResponse is the response for a single invoice
type InvoiceResponse struct {
	ID              string          `json:"id"`
	FarmID          string          `json:"farm_id"`
	FarmName        string          `json:"farm_name"`
	TokenID         *int64          `json:"token_id,omitempty"`
	OfftakerID      *string         `json:"offtaker_id,omitempty"`
	Name            string          `json:"name"`
	Description     *string         `json:"description,omitempty"`
	ImageURL        *string         `json:"image_url,omitempty"`
	TargetFund      decimal.Decimal `json:"target_fund"`
	YieldPercent    decimal.Decimal `json:"yield_percent"`
	DurationDays    int             `json:"duration_days"`
	TotalFunded     decimal.Decimal `json:"total_funded"`
	IsFullyFunded   bool            `json:"is_fully_funded"`
	Status          string          `json:"status"`
	RejectionReason *string         `json:"rejection_reason,omitempty"`
	ReviewedBy      *string         `json:"reviewed_by,omitempty"`
	ReviewedAt      *time.Time      `json:"reviewed_at,omitempty"`
	ApprovedAt      *time.Time      `json:"approved_at,omitempty"`
	FundingDeadline *time.Time      `json:"funding_deadline,omitempty"`
	MaturityDate    *time.Time      `json:"maturity_date,omitempty"`
	ApprovalTxHash  *string         `json:"approval_tx_hash,omitempty"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

// InvoiceListItem represents an invoice in the list response
type InvoiceListItem struct {
	ID            string          `json:"id"`
	FarmID        string          `json:"farm_id"`
	FarmName      string          `json:"farm_name"`
	Name          string          `json:"name"`
	ImageURL      *string         `json:"image_url,omitempty"`
	TargetFund    decimal.Decimal `json:"target_fund"`
	YieldPercent  decimal.Decimal `json:"yield_percent"`
	DurationDays  int             `json:"duration_days"`
	TotalFunded   decimal.Decimal `json:"total_funded"`
	IsFullyFunded bool            `json:"is_fully_funded"`
	Status        string          `json:"status"`
	CreatedAt     time.Time       `json:"created_at"`
}

// InvoiceListItemAdmin represents an invoice in the admin list response (includes farmer info)
type InvoiceListItemAdmin struct {
	ID            string          `json:"id"`
	FarmID        string          `json:"farm_id"`
	FarmName      string          `json:"farm_name"`
	FarmerID      string          `json:"farmer_id"`
	FarmerName    string          `json:"farmer_name"`
	Name          string          `json:"name"`
	ImageURL      *string         `json:"image_url,omitempty"`
	TargetFund    decimal.Decimal `json:"target_fund"`
	YieldPercent  decimal.Decimal `json:"yield_percent"`
	DurationDays  int             `json:"duration_days"`
	TotalFunded   decimal.Decimal `json:"total_funded"`
	IsFullyFunded bool            `json:"is_fully_funded"`
	Status        string          `json:"status"`
	CreatedAt     time.Time       `json:"created_at"`
	ReviewedAt    *time.Time      `json:"reviewed_at,omitempty"`
}

// ListInvoiceResponse is the response for listing invoices (farmer)
type ListInvoiceResponse struct {
	Invoices   []InvoiceListItem `json:"invoices"`
	Pagination PaginationMeta    `json:"pagination"`
}

// ListInvoiceAdminResponse is the response for listing invoices (admin)
type ListInvoiceAdminResponse struct {
	Invoices   []InvoiceListItemAdmin `json:"invoices"`
	Pagination PaginationMeta         `json:"pagination"`
}

// CreateInvoiceResponse is the response for creating an invoice
type CreateInvoiceResponse struct {
	Invoice InvoiceResponse `json:"invoice"`
}

// PresignInvoiceImageResponse is the response for presigned URL for invoice image
type PresignInvoiceImageResponse struct {
	UploadURL string `json:"upload_url"`
	FileKey   string `json:"file_key"`
}

// InvoiceStatusUpdateResponse is the response for approve/reject invoice
type InvoiceStatusUpdateResponse struct {
	InvoiceID      string    `json:"invoice_id"`
	Status         string    `json:"status"`
	TokenID        *int64    `json:"token_id,omitempty"`
	ApprovalTxHash *string   `json:"approval_tx_hash,omitempty"`
	ReviewedBy     string    `json:"reviewed_by"`
	ReviewedAt     time.Time `json:"reviewed_at"`
	Reason         *string   `json:"reason,omitempty"`
}
