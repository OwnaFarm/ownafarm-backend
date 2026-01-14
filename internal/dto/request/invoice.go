package request

// CreateInvoiceRequest is the request body for creating an invoice
type CreateInvoiceRequest struct {
	FarmID       string  `json:"farm_id" binding:"required,uuid"`
	Name         string  `json:"name" binding:"required,max=200"`
	Description  *string `json:"description,omitempty"`
	ImageKey     *string `json:"image_key,omitempty"` // File key from presigned upload
	TargetFund   float64 `json:"target_fund" binding:"required,gt=0"`
	YieldPercent float64 `json:"yield_percent" binding:"required,gt=0,max=100"`
	DurationDays int     `json:"duration_days" binding:"required,min=1"`
	OfftakerID   *string `json:"offtaker_id,omitempty" binding:"omitempty,max=100"`
}

// PresignInvoiceImageRequest is the request body for getting presigned URL for invoice image
type PresignInvoiceImageRequest struct {
	FileName    string `json:"file_name" binding:"required"`
	ContentType string `json:"content_type" binding:"required,oneof=image/jpeg image/png image/webp"`
}

// ListInvoiceRequest contains query parameters for listing invoices
type ListInvoiceRequest struct {
	FarmID    string   `form:"farm_id"`
	Status    []string `form:"status"` // pending, approved, rejected
	Page      int      `form:"page" binding:"omitempty,min=1"`
	Limit     int      `form:"limit" binding:"omitempty,min=1,max=100"`
	SortBy    string   `form:"sort_by"`    // created_at, name, target_fund, status
	SortOrder string   `form:"sort_order"` // asc, desc
	Search    string   `form:"search"`     // Search in name
}

// RejectInvoiceRequest is the request body for rejecting an invoice
type RejectInvoiceRequest struct {
	Reason *string `json:"reason,omitempty"`
}

// ApproveInvoiceRequest is the request body for approving an invoice
// Contains blockchain data after frontend successfully executes approveInvoice transaction
type ApproveInvoiceRequest struct {
	TokenID        int64  `json:"token_id" binding:"required"`
	ApprovalTxHash string `json:"approval_tx_hash" binding:"required,len=66"`
}
