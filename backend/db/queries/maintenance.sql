-- name: ClearAllData :exec
TRUNCATE TABLE
  likes,
  watch_histories,
  "verification",
  "account",
  "session",
  "user",
  comments,
  program_category_tags,
  program_performers,
  programs,
  category_tags,
  performers
RESTART IDENTITY CASCADE;
