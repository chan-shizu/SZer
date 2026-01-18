-- Rollback for 20260108111920_create_initial_tables.up.sql

-- better-auth (indexes then tables)
DROP INDEX IF EXISTS "verification_identifier_idx";
DROP INDEX IF EXISTS "account_userId_idx";
DROP INDEX IF EXISTS "session_userId_idx";

-- watch histories depends on "user" and programs
DROP INDEX IF EXISTS watch_histories_user_program_incomplete_uq;
DROP INDEX IF EXISTS watch_histories_program_id_idx;
DROP INDEX IF EXISTS watch_histories_user_last_watched_idx;
DROP TABLE IF EXISTS watch_histories;

DROP TABLE IF EXISTS "verification";
DROP TABLE IF EXISTS "account";
DROP TABLE IF EXISTS "session";
DROP TABLE IF EXISTS "user";

-- App tables (dependent tables first)
DROP TABLE IF EXISTS program_performers;
DROP TABLE IF EXISTS program_category_tags;
DROP TABLE IF EXISTS comments;

DROP TABLE IF EXISTS performers;
DROP TABLE IF EXISTS category_tags;
DROP TABLE IF EXISTS programs;
