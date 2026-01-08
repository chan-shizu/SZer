
-- Create videos table
CREATE TABLE IF NOT EXISTS videos (
	id BIGSERIAL PRIMARY KEY,
	title TEXT NOT NULL,
	video_path TEXT NOT NULL,
	thumbnail TEXT,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
