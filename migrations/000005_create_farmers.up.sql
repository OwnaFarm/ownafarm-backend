-- =====================
-- FARMER DATA
-- =====================

CREATE TABLE farmers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    status farmer_status DEFAULT 'pending',
    
    -- Step 1: Personal Info
    full_name VARCHAR(100) NOT NULL,
    email VARCHAR(255) NOT NULL,
    phone_number VARCHAR(20) NOT NULL,
    id_number VARCHAR(20) NOT NULL,
    date_of_birth DATE NOT NULL,
    address TEXT NOT NULL,
    province VARCHAR(100) NOT NULL,
    city VARCHAR(100) NOT NULL,
    district VARCHAR(100) NOT NULL,
    postal_code VARCHAR(10) NOT NULL,
    
    -- Step 2: Business Info
    business_name VARCHAR(200),
    business_type business_type NOT NULL,
    npwp VARCHAR(30),
    bank_name VARCHAR(100) NOT NULL,
    bank_account_number VARCHAR(30) NOT NULL,
    bank_account_name VARCHAR(100) NOT NULL,
    years_of_experience INT DEFAULT 0,
    crops_expertise TEXT[],
    
    -- Admin
    reviewed_by UUID REFERENCES admin_users(id),
    reviewed_at TIMESTAMP,
    rejection_reason TEXT,
    
    -- Timestamps
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now()
);

COMMENT ON COLUMN farmers.user_id IS 'Optional: if farmer also plays the game';
COMMENT ON COLUMN farmers.id_number IS 'KTP Number';
COMMENT ON COLUMN farmers.npwp IS 'Tax ID Number';
COMMENT ON COLUMN farmers.crops_expertise IS 'Array of crop types farmer specializes in';

-- Indexes
CREATE INDEX idx_farmers_email ON farmers(email);
CREATE INDEX idx_farmers_phone_number ON farmers(phone_number);
CREATE INDEX idx_farmers_status ON farmers(status);

-- Farmer Documents Table
CREATE TABLE farmer_documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    farmer_id UUID NOT NULL REFERENCES farmers(id),
    document_type document_type NOT NULL,
    file_url TEXT NOT NULL,
    file_name VARCHAR(255),
    file_size INT,
    mime_type VARCHAR(100),
    uploaded_at TIMESTAMP DEFAULT now()
);

COMMENT ON COLUMN farmer_documents.file_url IS 'Cloud Storage URL (S3/GCS)';
COMMENT ON COLUMN farmer_documents.file_size IS 'File size in bytes';

-- Indexes
CREATE INDEX idx_farmer_documents_farmer_type ON farmer_documents(farmer_id, document_type);
