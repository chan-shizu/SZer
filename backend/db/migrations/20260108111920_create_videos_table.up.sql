
-- Create programs table
CREATE TABLE IF NOT EXISTS programs (
	id BIGSERIAL PRIMARY KEY,
	title TEXT NOT NULL,
	video_path TEXT NOT NULL,
	thumbnail_path TEXT,
    description TEXT,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS category_tags (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS program_category_tags (
    program_id BIGINT NOT NULL REFERENCES programs(id) ON DELETE CASCADE,
    tag_id BIGINT NOT NULL REFERENCES category_tags(id) ON DELETE CASCADE,
    PRIMARY KEY (program_id, tag_id)
);

CREATE TABLE IF NOT EXISTS performers (
    id BIGSERIAL PRIMARY KEY,
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    first_name_kana TEXT NOT NULL,
    last_name_kana TEXT NOT NULL,
    image_path TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS program_performers (
    program_id BIGINT NOT NULL REFERENCES programs(id) ON DELETE CASCADE,
    performer_id BIGINT NOT NULL REFERENCES performers(id) ON DELETE CASCADE,
    PRIMARY KEY (program_id, performer_id)
);

CREATE TABLE IF NOT EXISTS comments (
    id BIGSERIAL PRIMARY KEY,
    program_id BIGINT NOT NULL REFERENCES programs(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);