-- Add wallet_address column to farmers table
ALTER TABLE farmers ADD COLUMN wallet_address VARCHAR(42) NOT NULL;

-- Add unique constraint
ALTER TABLE farmers ADD CONSTRAINT uq_farmers_wallet_address UNIQUE (wallet_address);

-- Add index for lookup
CREATE INDEX idx_farmers_wallet_address ON farmers(wallet_address);

COMMENT ON COLUMN farmers.wallet_address IS 'Ethereum wallet address for farmer login (0x + 40 hex chars)';
