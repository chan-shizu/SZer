-- name: ListWatchingProgramsByUser :many
SELECT
  p.id AS program_id,
  p.title,
  p.thumbnail_path,
  COALESCE(wc.view_count, 0)::bigint AS view_count,
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
LEFT JOIN (
  SELECT program_id, COUNT(*)::bigint AS view_count
  FROM watch_histories
  GROUP BY program_id
) wc ON wc.program_id = p.id
LEFT JOIN program_category_tags pct ON p.id = pct.program_id
LEFT JOIN category_tags ct ON pct.tag_id = ct.id
WHERE wh.user_id = $1 AND wh.is_completed = FALSE
GROUP BY
  p.id,
  p.title,
  p.thumbnail_path,
  wc.view_count,
  wh.last_watched_at
ORDER BY wh.last_watched_at DESC
LIMIT COALESCE(sqlc.narg('limit')::int, 50)
OFFSET COALESCE(sqlc.narg('offset')::int, 0);

-- name: ListLikedProgramsByUser :many
SELECT
  p.id AS program_id,
  p.title,
  p.thumbnail_path,
  COALESCE(wc.view_count, 0)::bigint AS view_count,
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
LEFT JOIN (
  SELECT program_id, COUNT(*)::bigint AS view_count
  FROM watch_histories
  GROUP BY program_id
) wc ON wc.program_id = p.id
LEFT JOIN program_category_tags pct ON p.id = pct.program_id
LEFT JOIN category_tags ct ON pct.tag_id = ct.id
WHERE lk.user_id = $1
GROUP BY
  p.id,
  p.title,
  p.thumbnail_path,
  wc.view_count,
  lk.created_at
ORDER BY lk.created_at DESC
LIMIT COALESCE(sqlc.narg('limit')::int, 50)
OFFSET COALESCE(sqlc.narg('offset')::int, 0);
