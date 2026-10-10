-- Fix: Add missing columns to investigations table (approved_by, rejected_by, etc.)
ALTER TABLE investigations ADD COLUMN IF NOT EXISTS approved_by char(36);
ALTER TABLE investigations ADD COLUMN IF NOT EXISTS rejected_by char(36);
ALTER TABLE investigations ADD COLUMN IF NOT EXISTS approved_by_name varchar(100);
ALTER TABLE investigations ADD COLUMN IF NOT EXISTS approved_by_username varchar(50);
ALTER TABLE investigations ADD COLUMN IF NOT EXISTS rejected_by_name varchar(100);
ALTER TABLE investigations ADD COLUMN IF NOT EXISTS rejected_by_username varchar(50);
ALTER TABLE investigations ADD COLUMN IF NOT EXISTS approved_at timestamp with time zone;
ALTER TABLE investigations ADD COLUMN IF NOT EXISTS rejected_at timestamp with time zone;
ALTER TABLE investigations ADD COLUMN IF NOT EXISTS updated_at timestamp with time zone;
ALTER TABLE investigations ADD COLUMN IF NOT EXISTS deleted_at timestamp with time zone;
