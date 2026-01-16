-- Drop index
DROP INDEX IF EXISTS idx_farmers_wallet_address;

-- Drop unique constraint
ALTER TABLE farmers DROP CONSTRAINT IF EXISTS uq_farmers_wallet_address;

-- Drop column
ALTER TABLE farmers DROP COLUMN IF EXISTS wallet_address;
