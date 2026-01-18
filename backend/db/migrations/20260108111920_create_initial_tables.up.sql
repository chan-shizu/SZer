-- Create initial tables for SZer application

-- App tables
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

-- better-auth tables (generated via better-auth compileMigrations for Postgres)
create table "user" (
  "id" text not null primary key,
  "name" text not null,
  "email" text not null unique,
  "emailVerified" boolean not null,
  "image" text,
  "createdAt" timestamptz default CURRENT_TIMESTAMP not null,
  "updatedAt" timestamptz default CURRENT_TIMESTAMP not null
);

create table "session" (
  "id" text not null primary key,
  "expiresAt" timestamptz not null,
  "token" text not null unique,
  "createdAt" timestamptz default CURRENT_TIMESTAMP not null,
  "updatedAt" timestamptz not null,
  "ipAddress" text,
  "userAgent" text,
  "userId" text not null references "user" ("id") on delete cascade
);

create table "account" (
  "id" text not null primary key,
  "accountId" text not null,
  "providerId" text not null,
  "userId" text not null references "user" ("id") on delete cascade,
  "accessToken" text,
  "refreshToken" text,
  "idToken" text,
  "accessTokenExpiresAt" timestamptz,
  "refreshTokenExpiresAt" timestamptz,
  "scope" text,
  "password" text,
  "createdAt" timestamptz default CURRENT_TIMESTAMP not null,
  "updatedAt" timestamptz not null
);

create table "verification" (
  "id" text not null primary key,
  "identifier" text not null,
  "value" text not null,
  "expiresAt" timestamptz not null,
  "createdAt" timestamptz default CURRENT_TIMESTAMP not null,
  "updatedAt" timestamptz default CURRENT_TIMESTAMP not null
);

create index "session_userId_idx" on "session" ("userId");

create index "account_userId_idx" on "account" ("userId");

create index "verification_identifier_idx" on "verification" ("identifier");

-- Watch histories (user x program)
-- Must be created after better-auth "user" table exists.
CREATE TABLE IF NOT EXISTS watch_histories (
  id BIGSERIAL PRIMARY KEY,
  user_id TEXT NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
  program_id BIGINT NOT NULL REFERENCES programs(id) ON DELETE CASCADE,
  position_seconds INT NOT NULL DEFAULT 0,
  is_completed BOOLEAN NOT NULL DEFAULT FALSE,
  last_watched_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT watch_histories_position_non_negative CHECK (position_seconds >= 0)
);

-- Ensure at most one *incomplete* record per (user_id, program_id)
CREATE UNIQUE INDEX IF NOT EXISTS watch_histories_user_program_incomplete_uq
  ON watch_histories (user_id, program_id)
  WHERE (is_completed = FALSE);

-- For aggregations by program_id (e.g. view_count)
CREATE INDEX IF NOT EXISTS watch_histories_program_id_idx
  ON watch_histories (program_id);

CREATE INDEX IF NOT EXISTS watch_histories_user_last_watched_idx
  ON watch_histories (user_id, last_watched_at DESC);
