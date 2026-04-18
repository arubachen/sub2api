ALTER TABLE promo_codes
  ADD COLUMN IF NOT EXISTS allowed_email_suffixes JSONB NOT NULL DEFAULT '[]'::jsonb;

COMMENT ON COLUMN promo_codes.allowed_email_suffixes IS 'Allowed email suffixes for promo registration usage; empty array means unrestricted';
