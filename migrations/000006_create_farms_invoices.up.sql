-- =====================
-- FARM & INVOICE DATA
-- =====================

CREATE TABLE farms (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    farmer_id UUID NOT NULL REFERENCES farmers(id),
    
    name VARCHAR(200) NOT NULL,
    description TEXT,
    location VARCHAR(255) NOT NULL,
    latitude DECIMAL(10, 8),
    longitude DECIMAL(11, 8),
    land_area DECIMAL(10, 2),
    
    -- CCTV Monitoring
    cctv_url TEXT,
    cctv_image_url TEXT,
    cctv_last_updated TIMESTAMP,
    
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now()
);

COMMENT ON COLUMN farms.land_area IS 'Area in hectares';
COMMENT ON COLUMN farms.cctv_url IS 'Live CCTV stream URL';
COMMENT ON COLUMN farms.cctv_image_url IS 'Latest CCTV snapshot';

-- Indexes
CREATE INDEX idx_farms_farmer_id ON farms(farmer_id);
CREATE INDEX idx_farms_is_active ON farms(is_active);

-- Invoices Table
CREATE TABLE invoices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    farm_id UUID NOT NULL REFERENCES farms(id),
    
    -- Blockchain Reference
    token_id BIGINT UNIQUE,
    offtaker_id VARCHAR(100),
    
    -- Invoice Details
    name VARCHAR(200) NOT NULL,
    description TEXT,
    image_url TEXT,
    
    -- Financial
    target_fund DECIMAL(20, 8) NOT NULL,
    yield_percent DECIMAL(5, 2) NOT NULL,
    duration_days INT NOT NULL,
    
    -- Status (synced from blockchain)
    total_funded DECIMAL(20, 8) DEFAULT 0,
    is_fully_funded BOOLEAN DEFAULT false,
    is_approved BOOLEAN DEFAULT false,
    
    -- Dates
    approved_at TIMESTAMP,
    funding_deadline TIMESTAMP,
    maturity_date TIMESTAMP,
    
    -- Blockchain TX
    approval_tx_hash VARCHAR(66),
    
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now()
);

COMMENT ON COLUMN invoices.token_id IS 'NFT Token ID on smart contract';
COMMENT ON COLUMN invoices.offtaker_id IS 'Offtaker identifier';
COMMENT ON COLUMN invoices.name IS 'Crop name, e.g., "Cabai Indofood"';
COMMENT ON COLUMN invoices.image_url IS 'Crop/seed image for game';
COMMENT ON COLUMN invoices.target_fund IS 'Target funding in GOLD';
COMMENT ON COLUMN invoices.yield_percent IS 'Return percentage, e.g., 18.00';
COMMENT ON COLUMN invoices.duration_days IS 'Duration until harvest/maturity';

-- Indexes
CREATE UNIQUE INDEX idx_invoices_token_id ON invoices(token_id);
CREATE INDEX idx_invoices_farm_id ON invoices(farm_id);
CREATE INDEX idx_invoices_is_approved ON invoices(is_approved);
CREATE INDEX idx_invoices_is_fully_funded ON invoices(is_fully_funded);
