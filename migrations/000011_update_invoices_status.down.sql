-- =====================
-- REVERT INVOICES TABLE STATUS WORKFLOW
-- =====================

-- Add back is_approved column
ALTER TABLE invoices ADD COLUMN is_approved BOOLEAN DEFAULT false;

-- Migrate data back: approved -> true, else -> false
UPDATE invoices SET is_approved = CASE 
    WHEN status = 'approved' THEN true 
    ELSE false 
END;

-- Drop new columns
ALTER TABLE invoices 
    DROP COLUMN status,
    DROP COLUMN rejection_reason,
    DROP COLUMN reviewed_by,
    DROP COLUMN reviewed_at;

-- Drop index on status
DROP INDEX IF EXISTS idx_invoices_status;

-- Recreate index on is_approved
CREATE INDEX idx_invoices_is_approved ON invoices(is_approved);

-- Drop invoice_status enum
DROP TYPE IF EXISTS invoice_status;
