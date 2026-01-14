-- =====================
-- UPDATE INVOICES TABLE FOR STATUS WORKFLOW
-- =====================

-- Create invoice_status enum
CREATE TYPE invoice_status AS ENUM ('pending', 'approved', 'rejected');

-- Add new columns for status workflow
ALTER TABLE invoices 
    ADD COLUMN status invoice_status DEFAULT 'pending',
    ADD COLUMN rejection_reason TEXT,
    ADD COLUMN reviewed_by UUID REFERENCES admin_users(id),
    ADD COLUMN reviewed_at TIMESTAMP;

-- Migrate existing data: is_approved = true -> approved, is_approved = false -> pending
UPDATE invoices SET status = CASE 
    WHEN is_approved = true THEN 'approved'::invoice_status 
    ELSE 'pending'::invoice_status 
END;

-- Set reviewed_at for approved invoices
UPDATE invoices SET reviewed_at = approved_at WHERE is_approved = true;

-- Make status NOT NULL after migration
ALTER TABLE invoices ALTER COLUMN status SET NOT NULL;

-- Drop old is_approved column
ALTER TABLE invoices DROP COLUMN is_approved;

-- Create index on status
CREATE INDEX idx_invoices_status ON invoices(status);

-- Drop old index on is_approved (if exists)
DROP INDEX IF EXISTS idx_invoices_is_approved;

COMMENT ON COLUMN invoices.status IS 'Invoice approval status: pending, approved, rejected';
COMMENT ON COLUMN invoices.rejection_reason IS 'Reason for rejection if status is rejected';
COMMENT ON COLUMN invoices.reviewed_by IS 'Admin who reviewed the invoice';
COMMENT ON COLUMN invoices.reviewed_at IS 'Timestamp when invoice was reviewed';
