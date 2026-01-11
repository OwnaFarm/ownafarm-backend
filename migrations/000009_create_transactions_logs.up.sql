-- =====================
-- TRANSACTIONS & LOGS
-- =====================

-- Gold Transactions Table
CREATE TABLE gold_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    
    transaction_type transaction_type NOT NULL,
    amount DECIMAL(20, 8) NOT NULL,
    balance_after DECIMAL(20, 8),
    
    -- Reference
    reference_id UUID,
    reference_type VARCHAR(50),
    
    -- Blockchain
    tx_hash VARCHAR(66),
    block_number BIGINT,
    
    description TEXT,
    created_at TIMESTAMP DEFAULT now()
);

COMMENT ON COLUMN gold_transactions.amount IS 'Positive for credit, negative for debit';
COMMENT ON COLUMN gold_transactions.balance_after IS 'Balance after transaction (from contract)';
COMMENT ON COLUMN gold_transactions.reference_id IS 'Related investment_id, daily_reward_id, etc.';

-- Indexes
CREATE INDEX idx_gold_transactions_user_id ON gold_transactions(user_id);
CREATE INDEX idx_gold_transactions_type ON gold_transactions(transaction_type);
CREATE INDEX idx_gold_transactions_created_at ON gold_transactions(created_at);
CREATE INDEX idx_gold_transactions_tx_hash ON gold_transactions(tx_hash);

-- XP Logs Table
CREATE TABLE xp_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    
    xp_gained INT NOT NULL,
    source VARCHAR(50) NOT NULL,
    source_id UUID,
    
    level_before INT,
    level_after INT,
    
    created_at TIMESTAMP DEFAULT now()
);

COMMENT ON COLUMN xp_logs.source IS 'e.g., watering, harvest, daily_login, achievement';
COMMENT ON COLUMN xp_logs.source_id IS 'Reference to related entity';

-- Indexes
CREATE INDEX idx_xp_logs_user_id ON xp_logs(user_id);
CREATE INDEX idx_xp_logs_created_at ON xp_logs(created_at);

-- Water Logs Table
CREATE TABLE water_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    investment_id UUID NOT NULL REFERENCES investments(id),
    
    water_spent INT NOT NULL DEFAULT 1,
    xp_gained INT DEFAULT 0,
    progress_added INT DEFAULT 0,
    
    created_at TIMESTAMP DEFAULT now()
);

COMMENT ON COLUMN water_logs.progress_added IS 'Progress percentage added';

-- Indexes
CREATE INDEX idx_water_logs_user_id ON water_logs(user_id);
CREATE INDEX idx_water_logs_investment_id ON water_logs(investment_id);
CREATE INDEX idx_water_logs_created_at ON water_logs(created_at);
