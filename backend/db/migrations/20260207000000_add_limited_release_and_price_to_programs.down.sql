ALTER TABLE programs
  DROP COLUMN IF EXISTS price,
  DROP COLUMN IF EXISTS is_limited_release;
