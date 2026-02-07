-- Add is_public column to programs table
ALTER TABLE programs ADD COLUMN IF NOT EXISTS is_public boolean NOT NULL DEFAULT true;
