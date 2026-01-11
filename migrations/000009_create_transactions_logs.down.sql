DROP INDEX IF EXISTS idx_water_logs_created_at;
DROP INDEX IF EXISTS idx_water_logs_investment_id;
DROP INDEX IF EXISTS idx_water_logs_user_id;
DROP TABLE IF EXISTS water_logs;

DROP INDEX IF EXISTS idx_xp_logs_created_at;
DROP INDEX IF EXISTS idx_xp_logs_user_id;
DROP TABLE IF EXISTS xp_logs;

DROP INDEX IF EXISTS idx_gold_transactions_tx_hash;
DROP INDEX IF EXISTS idx_gold_transactions_created_at;
DROP INDEX IF EXISTS idx_gold_transactions_type;
DROP INDEX IF EXISTS idx_gold_transactions_user_id;
DROP TABLE IF EXISTS gold_transactions;
