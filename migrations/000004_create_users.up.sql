-- =====================
-- USER & AUTH
-- =====================

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    wallet_address VARCHAR(42) UNIQUE NOT NULL,
    name VARCHAR(100),
    email VARCHAR(255) UNIQUE,
    avatar VARCHAR(50),
    
    -- Game Stats
    level INT DEFAULT 1,
    xp INT DEFAULT 0,
    water_points INT DEFAULT 100,
    last_regen_at TIMESTAMP DEFAULT now(),
    
    -- Timestamps
    last_login_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now()
);

COMMENT ON COLUMN users.wallet_address IS 'Ethereum wallet address';
COMMENT ON COLUMN users.avatar IS 'Avatar emoji or image key';
COMMENT ON COLUMN users.water_points IS 'Water points for watering crops';
COMMENT ON COLUMN users.last_regen_at IS 'Last water regen';

-- Indexes
CREATE UNIQUE INDEX idx_users_wallet_address ON users(wallet_address);
