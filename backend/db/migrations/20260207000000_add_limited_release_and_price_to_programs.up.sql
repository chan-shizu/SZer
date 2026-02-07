ALTER TABLE programs
  ADD COLUMN is_limited_release BOOLEAN NOT NULL DEFAULT false,
  ADD COLUMN price INTEGER NOT NULL DEFAULT 0 CONSTRAINT programs_price_non_negative CHECK (price >= 0);
