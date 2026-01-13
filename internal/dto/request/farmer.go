package request

// RegisterFarmerRequest is the main request body for farmer registration
type RegisterFarmerRequest struct {
	PersonalInfo PersonalInfoRequest `json:"personal_info" binding:"required"`
	BusinessInfo BusinessInfoRequest `json:"business_info" binding:"required"`
	Documents    []DocumentRequest   `json:"documents" binding:"required,min=1"`
}

// PersonalInfoRequest contains personal information for farmer registration
type PersonalInfoRequest struct {
	FullName    string `json:"full_name" binding:"required,max=100"`
	Email       string `json:"email" binding:"required,email,max=255"`
	PhoneNumber string `json:"phone_number" binding:"required,max=20"`
	IDNumber    string `json:"id_number" binding:"required,max=20"`
	DateOfBirth string `json:"date_of_birth" binding:"required"` // Format: YYYY-MM-DD
	Address     string `json:"address" binding:"required"`
	Province    string `json:"province" binding:"required,max=100"`
	City        string `json:"city" binding:"required,max=100"`
	District    string `json:"district" binding:"required,max=100"`
	PostalCode  string `json:"postal_code" binding:"required,max=10"`
}

// BusinessInfoRequest contains business information for farmer registration
type BusinessInfoRequest struct {
	BusinessName      *string  `json:"business_name,omitempty" binding:"omitempty,max=200"`
	BusinessType      string   `json:"business_type" binding:"required,oneof=individual cv pt ud cooperative"`
	NPWP              *string  `json:"npwp,omitempty" binding:"omitempty,max=30"`
	BankName          string   `json:"bank_name" binding:"required,max=100"`
	BankAccountNumber string   `json:"bank_account_number" binding:"required,max=30"`
	BankAccountName   string   `json:"bank_account_name" binding:"required,max=100"`
	YearsOfExperience *int     `json:"years_of_experience,omitempty"`
	CropsExpertise    []string `json:"crops_expertise,omitempty"`
}

// DocumentRequest contains document information for farmer registration
type DocumentRequest struct {
	DocumentType string  `json:"document_type" binding:"required,oneof=ktp_photo selfie_with_ktp npwp_photo bank_statement land_certificate business_license invoice_file"`
	FileKey      string  `json:"file_key" binding:"required"`
	FileName     *string `json:"file_name,omitempty"`
	FileSize     *int    `json:"file_size,omitempty"`
	MimeType     *string `json:"mime_type,omitempty"`
}

// PresignDocumentsRequest is the request body for getting presigned URLs
type PresignDocumentsRequest struct {
	DocumentTypes []string `json:"document_types" binding:"required,min=1,dive,oneof=ktp_photo selfie_with_ktp npwp_photo bank_statement land_certificate business_license invoice_file"`
}

// ListFarmerRequest contains query parameters for listing farmers (admin)
type ListFarmerRequest struct {
	Status    []string `form:"status"` // Filter: pending, under_review, approved, rejected, suspended
	Page      int      `form:"page" binding:"omitempty,min=1"`
	Limit     int      `form:"limit" binding:"omitempty,min=1,max=100"`
	SortBy    string   `form:"sort_by"`    // created_at, full_name, status
	SortOrder string   `form:"sort_order"` // asc, desc
	Search    string   `form:"search"`     // Search in name, email, phone
}
