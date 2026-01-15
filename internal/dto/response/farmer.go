package response

import "time"

// RegisterFarmerResponse is the response for farmer registration
type RegisterFarmerResponse struct {
	FarmerID string `json:"farmer_id"`
	Status   string `json:"status"`
}

// PresignedURLResponse represents a single presigned URL response
type PresignedURLResponse struct {
	DocumentType string `json:"document_type"`
	UploadURL    string `json:"upload_url"`
	FileKey      string `json:"file_key"`
}

// PresignDocumentsResponse is the response for presigned URLs request
type PresignDocumentsResponse struct {
	URLs []PresignedURLResponse `json:"urls"`
}

// DocumentDownloadURLResponse is the response for document download URL request
type DocumentDownloadURLResponse struct {
	DocumentID   string `json:"document_id"`
	DocumentType string `json:"document_type"`
	DownloadURL  string `json:"download_url"`
	ExpiresIn    int    `json:"expires_in"` // seconds
}

// ListFarmerResponse is the response for listing farmers (admin)
type ListFarmerResponse struct {
	Farmers    []FarmerListItem `json:"farmers"`
	Pagination PaginationMeta   `json:"pagination"`
}

// FarmerListItem represents a farmer in the list response
type FarmerListItem struct {
	ID                string     `json:"id"`
	Status            string     `json:"status"`
	FullName          string     `json:"full_name"`
	Email             string     `json:"email"`
	PhoneNumber       string     `json:"phone_number"`
	BusinessName      *string    `json:"business_name,omitempty"`
	BusinessType      string     `json:"business_type"`
	Province          string     `json:"province"`
	City              string     `json:"city"`
	YearsOfExperience int        `json:"years_of_experience"`
	CreatedAt         time.Time  `json:"created_at"`
	ReviewedAt        *time.Time `json:"reviewed_at,omitempty"`
}

// PaginationMeta contains pagination metadata
type PaginationMeta struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	TotalItems int64 `json:"total_items"`
	TotalPages int   `json:"total_pages"`
}

// FarmerDetailResponse is the response for farmer detail (admin)
type FarmerDetailResponse struct {
	ID                string               `json:"id"`
	Status            string               `json:"status"`
	FullName          string               `json:"full_name"`
	Email             string               `json:"email"`
	PhoneNumber       string               `json:"phone_number"`
	IDNumber          string               `json:"id_number"`
	DateOfBirth       string               `json:"date_of_birth"`
	Address           string               `json:"address"`
	Province          string               `json:"province"`
	City              string               `json:"city"`
	District          string               `json:"district"`
	PostalCode        string               `json:"postal_code"`
	BusinessName      *string              `json:"business_name,omitempty"`
	BusinessType      string               `json:"business_type"`
	NPWP              *string              `json:"npwp,omitempty"`
	BankName          string               `json:"bank_name"`
	BankAccountNumber string               `json:"bank_account_number"`
	BankAccountName   string               `json:"bank_account_name"`
	YearsOfExperience int                  `json:"years_of_experience"`
	CropsExpertise    []string             `json:"crops_expertise"`
	Documents         []FarmerDocumentItem `json:"documents"`
	ReviewedBy        *string              `json:"reviewed_by,omitempty"`
	ReviewedAt        *string              `json:"reviewed_at,omitempty"`
	RejectionReason   *string              `json:"rejection_reason,omitempty"`
	CreatedAt         string               `json:"created_at"`
}

// FarmerDocumentItem represents a document in the farmer detail response
type FarmerDocumentItem struct {
	ID           string  `json:"id"`
	DocumentType string  `json:"document_type"`
	FileName     *string `json:"file_name,omitempty"`
	DownloadURL  string  `json:"download_url"`
	ExpiresIn    int     `json:"expires_in"` // seconds
}
