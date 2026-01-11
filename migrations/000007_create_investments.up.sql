-- =====================
-- INVESTMENT & CROPS
-- =====================

CREATE TABLE investments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    invoice_id UUID NOT NULL REFERENCES invoices(id),
    
    -- Blockchain Reference
    investment_id_onchain BIGINT,
    
    -- Investment Details
    amount DECIMAL(20, 8) NOT NULL,
    invested_at TIMESTAMP NOT NULL DEFAULT now(),
    
    -- Game State
    status crop_status DEFAULT 'growing',
    progress INT DEFAULT 0,
    water_count INT DEFAULT 0,
    last_watered_at TIMESTAMP,
    
    -- Harvest
    is_harvested BOOLEAN DEFAULT false,
    harvested_at TIMESTAMP,
    harvest_amount DECIMAL(20, 8),
    harvest_tx_hash VARCHAR(66),
    
    -- TX Reference
    purchase_tx_hash VARCHAR(66),
    
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now()
);

COMMENT ON COLUMN investments.investment_id_onchain IS 'Investment ID on smart contract';
COMMENT ON COLUMN investments.amount IS 'Amount invested in GOLD';
COMMENT ON COLUMN investments.progress IS 'Progress 0-100%';
COMMENT ON COLUMN investments.water_count IS 'Times watered';
COMMENT ON COLUMN investments.harvest_amount IS 'Amount received after harvest';

-- Indexes
CREATE INDEX idx_investments_user_id ON investments(user_id);
CREATE INDEX idx_investments_invoice_id ON investments(invoice_id);
CREATE INDEX idx_investments_status ON investments(status);
CREATE INDEX idx_investments_user_status ON investments(user_id, status);
