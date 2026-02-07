-- Remove is_public column from programs table
ALTER TABLE programs DROP COLUMN IF EXISTS is_public;
