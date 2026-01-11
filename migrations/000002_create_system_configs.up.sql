-- =====================
-- SYSTEM CONFIG
-- =====================

CREATE TABLE level_configs (
    level INT PRIMARY KEY,
    xp_required INT NOT NULL,
    water_capacity INT DEFAULT 100,
    daily_reward_multiplier DECIMAL(3, 2) DEFAULT 1.00
);

COMMENT ON COLUMN level_configs.xp_required IS 'Total XP needed to reach this level';
COMMENT ON COLUMN level_configs.water_capacity IS 'Max water points at this level';

CREATE TABLE system_configs (
    key VARCHAR(100) PRIMARY KEY,
    value TEXT NOT NULL,
    description TEXT,
    updated_at TIMESTAMP DEFAULT now()
);
