-- Fix: Add missing columns to otp_requests table (branch_id, branch_name, employee_name)
ALTER TABLE otp_requests ADD COLUMN IF NOT EXISTS branch_id char(36);
ALTER TABLE otp_requests ADD COLUMN IF NOT EXISTS branch_name varchar(100);
ALTER TABLE otp_requests ADD COLUMN IF NOT EXISTS employee_name varchar(150);
