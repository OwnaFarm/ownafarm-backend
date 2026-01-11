-- =====================
-- GAME MECHANICS
-- =====================

-- Daily Rewards Table
CREATE TABLE daily_rewards (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    
    reward_date DATE NOT NULL,
    gold_amount DECIMAL(20, 8) NOT NULL,
    streak_day INT DEFAULT 1,
    claimed_at TIMESTAMP DEFAULT now()
);

COMMENT ON COLUMN daily_rewards.streak_day IS 'Consecutive login day';

-- Indexes
CREATE UNIQUE INDEX idx_daily_rewards_user_date ON daily_rewards(user_id, reward_date);

-- Achievements Table
CREATE TABLE achievements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    code VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    icon VARCHAR(50),
    xp_reward INT DEFAULT 0,
    gold_reward DECIMAL(20, 8) DEFAULT 0,
    
    -- Requirements
    requirement_type VARCHAR(50),
    requirement_value INT,
    
    created_at TIMESTAMP DEFAULT now()
);

COMMENT ON COLUMN achievements.requirement_type IS 'e.g., level, harvest_count, investment_total';

-- User Achievements Table
CREATE TABLE user_achievements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    achievement_id UUID NOT NULL REFERENCES achievements(id),
    
    unlocked_at TIMESTAMP DEFAULT now()
);

-- Indexes
CREATE UNIQUE INDEX idx_user_achievements_user_achievement ON user_achievements(user_id, achievement_id);

-- Leaderboard Snapshots Table
CREATE TABLE leaderboard_snapshots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    
    snapshot_date DATE NOT NULL,
    rank INT NOT NULL,
    total_xp INT NOT NULL,
    total_gold_earned DECIMAL(20, 8),
    total_harvests INT,
    level INT
);

-- Indexes
CREATE INDEX idx_leaderboard_date_rank ON leaderboard_snapshots(snapshot_date, rank);
CREATE INDEX idx_leaderboard_user_date ON leaderboard_snapshots(user_id, snapshot_date);
