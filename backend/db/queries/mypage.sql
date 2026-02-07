-- name: ListWatchingProgramsByUser :many
SELECT
  p.id AS program_id,
  p.title,
  p.thumbnail_path,
  p.view_count,
  p.is_limited_release,
  p.price,
  COALESCE((SELECT COUNT(*) FROM likes l WHERE l.program_id = p.id), 0)::bigint AS like_count,
  COALESCE(
    jsonb_agg(DISTINCT jsonb_build_object(
      'id', ct.id,
      'name', ct.name
    )) FILTER (WHERE ct.id IS NOT NULL),
    '[]'::jsonb
  ) AS category_tags
FROM watch_histories wh
JOIN programs p ON p.id = wh.program_id
-- 視聴回数はprogramsテーブルのview_countを参照
LEFT JOIN program_category_tags pct ON p.id = pct.program_id
LEFT JOIN category_tags ct ON pct.tag_id = ct.id
WHERE wh.user_id = $1 AND wh.is_completed = FALSE AND p.is_public = true
GROUP BY
  p.id,
  p.title,
  p.thumbnail_path,
  p.view_count,
  p.is_limited_release,
  p.price,
  wh.last_watched_at
ORDER BY wh.last_watched_at DESC
LIMIT COALESCE(sqlc.narg('limit')::int, 50)
OFFSET COALESCE(sqlc.narg('offset')::int, 0);

-- name: ListLikedProgramsByUser :many
SELECT
  p.id AS program_id,
  p.title,
  p.thumbnail_path,
  p.view_count,
  p.is_limited_release,
  p.price,
  COALESCE((SELECT COUNT(*) FROM likes l WHERE l.program_id = p.id), 0)::bigint AS like_count,
  COALESCE(
    jsonb_agg(DISTINCT jsonb_build_object(
      'id', ct.id,
      'name', ct.name
    )) FILTER (WHERE ct.id IS NOT NULL),
    '[]'::jsonb
  ) AS category_tags
FROM likes lk
JOIN programs p ON p.id = lk.program_id
-- 視聴回数はprogramsテーブルのview_countを参照
LEFT JOIN program_category_tags pct ON p.id = pct.program_id
LEFT JOIN category_tags ct ON pct.tag_id = ct.id
WHERE lk.user_id = $1 AND p.is_public = true
GROUP BY
  p.id,
  p.title,
  p.thumbnail_path,
  p.view_count,
  p.is_limited_release,
  p.price,
  lk.created_at
ORDER BY lk.created_at DESC
LIMIT COALESCE(sqlc.narg('limit')::int, 50)
OFFSET COALESCE(sqlc.narg('offset')::int, 0);
