
-- Drop tables created in 20260108111920_create_videos_table.up.sql

-- Dependent tables first
DROP TABLE IF EXISTS video_category_tags;
DROP TABLE IF EXISTS video_performers;
DROP TABLE IF EXISTS comments;

-- Base tables
DROP TABLE IF EXISTS performers;
DROP TABLE IF EXISTS category_tags;
DROP TABLE IF EXISTS videos;
