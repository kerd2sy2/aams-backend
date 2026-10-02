-- Migration: Convert vehicles plate_number unique index to partial index
-- This allows soft-deleted records to have the same plate number as active records

-- Step 1: Hard-delete all soft-deleted (deleted_at IS NOT NULL) vehicles
DELETE FROM vehicles WHERE deleted_at IS NOT NULL;

-- Step 2: Drop the current full unique index
DROP INDEX IF EXISTS idx_vehicles_plate_number;

-- Step 3: Create a partial unique index (only active records)
CREATE UNIQUE INDEX idx_vehicles_plate_number 
ON vehicles (plate_number) 
WHERE deleted_at IS NULL;
