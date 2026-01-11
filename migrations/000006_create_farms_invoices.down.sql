DROP INDEX IF EXISTS idx_invoices_is_fully_funded;
DROP INDEX IF EXISTS idx_invoices_is_approved;
DROP INDEX IF EXISTS idx_invoices_farm_id;
DROP INDEX IF EXISTS idx_invoices_token_id;
DROP TABLE IF EXISTS invoices;

DROP INDEX IF EXISTS idx_farms_is_active;
DROP INDEX IF EXISTS idx_farms_farmer_id;
DROP TABLE IF EXISTS farms;
