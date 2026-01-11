-- =====================
-- ENUMS
-- =====================

CREATE TYPE farmer_status AS ENUM (
    'pending',
    'under_review',
    'approved',
    'rejected',
    'suspended'
);

CREATE TYPE business_type AS ENUM (
    'individual',
    'cv',
    'pt',
    'ud',
    'cooperative'
);

CREATE TYPE document_type AS ENUM (
    'ktp_photo',
    'selfie_with_ktp',
    'npwp_photo',
    'bank_statement',
    'land_certificate',
    'business_license',
    'invoice_file'
);

CREATE TYPE crop_status AS ENUM (
    'growing',
    'ready',
    'harvested'
);

CREATE TYPE transaction_type AS ENUM (
    'purchase',
    'harvest',
    'daily_reward',
    'faucet_claim',
    'withdrawal'
);
