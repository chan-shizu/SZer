-- name: ClearAllData :exec
TRUNCATE TABLE
  comments,
  program_category_tags,
  program_performers,
  programs,
  category_tags,
  performers
RESTART IDENTITY CASCADE;
