package models

import (
	"time"
)

// DocumentType represents the type of document
type DocumentType string

const (
	DocumentTypeKTPPhoto        DocumentType = "ktp_photo"
	DocumentTypeSelfieWithKTP   DocumentType = "selfie_with_ktp"
	DocumentTypeNPWPPhoto       DocumentType = "npwp_photo"
	DocumentTypeBankStatement   DocumentType = "bank_statement"
	DocumentTypeLandCertificate DocumentType = "land_certificate"
	DocumentTypeBusinessLicense DocumentType = "business_license"
	DocumentTypeInvoiceFile     DocumentType = "invoice_file"
)

// FarmerDocument represents the farmer_documents table in the database
type FarmerDocument struct {
	ID           string       `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	FarmerID     string       `gorm:"type:uuid;not null" json:"farmer_id"`
	DocumentType DocumentType `gorm:"type:document_type;not null" json:"document_type"`
	FileURL      string       `gorm:"type:text;not null" json:"file_url"`
	FileName     *string      `gorm:"type:varchar(255)" json:"file_name,omitempty"`
	FileSize     *int         `json:"file_size,omitempty"`
	MimeType     *string      `gorm:"type:varchar(100)" json:"mime_type,omitempty"`
	UploadedAt   time.Time    `gorm:"default:now()" json:"uploaded_at"`
}

// TableName returns the table name for the FarmerDocument model
func (FarmerDocument) TableName() string {
	return "farmer_documents"
}
