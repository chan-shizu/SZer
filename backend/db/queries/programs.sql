-- name: CreateProgram :one
INSERT INTO programs (
  title,
  video_path,
  thumbnail_path,
  description
) VALUES (
  $1, $2, $3, $4
)
RETURNING id, title, video_path, thumbnail_path, description, created_at, updated_at;

-- name: GetProgramByID :one
SELECT
  p.id AS program_id,
  p.title,
  p.video_path,
  p.thumbnail_path,
  p.description,
  COALESCE(wc.view_count, 0)::bigint AS view_count,
  p.created_at AS program_created_at,
  p.updated_at AS program_updated_at,
  COALESCE(
    jsonb_agg(DISTINCT jsonb_build_object(
      'id', ct.id,
      'name', ct.name
    )) FILTER (WHERE ct.id IS NOT NULL),
    '[]'::jsonb
  ) AS category_tags,
  COALESCE(
    jsonb_agg(DISTINCT jsonb_build_object(
      'id', pe.id,
      'first_name', pe.first_name,
      'last_name', pe.last_name,
      'first_name_kana', pe.first_name_kana,
      'last_name_kana', pe.last_name_kana,
      'image_path', pe.image_path
    )) FILTER (WHERE pe.id IS NOT NULL),
    '[]'::jsonb
  ) AS performers
FROM programs p
LEFT JOIN (
  SELECT program_id, COUNT(*)::bigint AS view_count
  FROM watch_histories
  GROUP BY program_id
) wc ON wc.program_id = p.id
LEFT JOIN program_category_tags pct ON p.id = pct.program_id
LEFT JOIN category_tags ct ON pct.tag_id = ct.id
LEFT JOIN program_performers pp ON p.id = pp.program_id
LEFT JOIN performers pe ON pp.performer_id = pe.id
WHERE p.id = $1
GROUP BY
  p.id,
  p.title,
  p.video_path,
  p.thumbnail_path,
  p.description,
  wc.view_count,
  p.created_at,
  p.updated_at;

-- name: GetProgramDetailsByID :one
SELECT
  p.id AS program_id,
  p.title,
  p.video_path,
  p.thumbnail_path,
  p.description,
  COALESCE(wc.view_count, 0)::bigint AS view_count,
  COALESCE((SELECT COUNT(*) FROM likes l WHERE l.program_id = p.id), 0)::bigint AS like_count,
  EXISTS(
    SELECT 1
    FROM likes l
    WHERE l.program_id = p.id AND l.user_id = $2
  ) AS liked,
  p.created_at AS program_created_at,
  p.updated_at AS program_updated_at,
  COALESCE(
    jsonb_agg(DISTINCT jsonb_build_object(
      'id', ct.id,
      'name', ct.name
    )) FILTER (WHERE ct.id IS NOT NULL),
    '[]'::jsonb
  ) AS category_tags,
  COALESCE(
    jsonb_agg(DISTINCT jsonb_build_object(
      'id', pe.id,
      'first_name', pe.first_name,
      'last_name', pe.last_name,
      'first_name_kana', pe.first_name_kana,
      'last_name_kana', pe.last_name_kana,
      'image_path', pe.image_path
    )) FILTER (WHERE pe.id IS NOT NULL),
    '[]'::jsonb
  ) AS performers
FROM programs p
LEFT JOIN (
  SELECT program_id, COUNT(*)::bigint AS view_count
  FROM watch_histories
  GROUP BY program_id
) wc ON wc.program_id = p.id
LEFT JOIN program_category_tags pct ON p.id = pct.program_id
LEFT JOIN category_tags ct ON pct.tag_id = ct.id
LEFT JOIN program_performers pp ON p.id = pp.program_id
LEFT JOIN performers pe ON pp.performer_id = pe.id
WHERE p.id = $1
GROUP BY
  p.id,
  p.title,
  p.video_path,
  p.thumbnail_path,
  p.description,
  wc.view_count,
  p.created_at,
  p.updated_at;

-- name: GetPrograms :many
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
FROM programs p
LEFT JOIN (
  SELECT program_id, COUNT(*)::bigint AS view_count
  FROM watch_histories
  GROUP BY program_id
) wc ON wc.program_id = p.id
LEFT JOIN program_category_tags pct ON p.id = pct.program_id
LEFT JOIN category_tags ct ON pct.tag_id = ct.id
WHERE
  (sqlc.narg('title')::text IS NULL OR p.title ILIKE '%' || sqlc.narg('title')::text || '%')
  AND (
    sqlc.narg('tag_ids')::bigint[] IS NULL
    OR p.id IN (
      SELECT pct2.program_id
      FROM program_category_tags pct2
      WHERE pct2.tag_id = ANY(sqlc.narg('tag_ids')::bigint[])
      GROUP BY pct2.program_id
      HAVING COUNT(DISTINCT pct2.tag_id) = array_length(sqlc.narg('tag_ids')::bigint[], 1)
    )
  )
GROUP BY
  p.id,
  p.title,
  p.thumbnail_path,
  wc.view_count;

-- name: GetTopPrograms :many
SELECT
  p.id AS program_id,
  p.title,
  p.thumbnail_path,
  COALESCE(wc.view_count, 0)::bigint AS view_count,
  COALESCE((SELECT COUNT(*) FROM likes l WHERE l.program_id = p.id), 0)::bigint AS like_count
FROM programs p
LEFT JOIN (
  SELECT program_id, COUNT(*)::bigint AS view_count
  FROM watch_histories
  GROUP BY program_id
) wc ON wc.program_id = p.id
ORDER BY p.created_at DESC
LIMIT 7;

-- name: GetTopLikedPrograms :many
WITH params AS (
  SELECT COALESCE(sqlc.narg('limit')::int, 7)::int AS n
),
top_likes AS (
  SELECT
    l.program_id,
    COUNT(*)::bigint AS like_count
  FROM likes l
  GROUP BY l.program_id
  ORDER BY like_count DESC
  LIMIT (SELECT n FROM params)
),
fallback AS (
  SELECT
    p.id AS program_id,
    0::bigint AS like_count
  FROM programs p
  WHERE p.id NOT IN (SELECT program_id FROM top_likes)
  ORDER BY p.created_at DESC
  LIMIT GREATEST((SELECT n FROM params) - (SELECT COUNT(*) FROM top_likes), 0)
),
selected AS (
  SELECT program_id, like_count FROM top_likes
  UNION ALL
  SELECT program_id, like_count FROM fallback
),
view_counts AS (
  SELECT
    wh.program_id,
    COUNT(*)::bigint AS view_count
  FROM watch_histories wh
  WHERE wh.program_id IN (SELECT program_id FROM selected)
  GROUP BY wh.program_id
)
SELECT
  p.id AS program_id,
  p.title,
  p.thumbnail_path,
  COALESCE(vc.view_count, 0)::bigint AS view_count,
  s.like_count
FROM selected s
JOIN programs p ON p.id = s.program_id
LEFT JOIN view_counts vc ON vc.program_id = p.id
ORDER BY s.like_count DESC, p.created_at DESC;

-- name: GetTopViewedPrograms :many
WITH top_view_counts AS (
  SELECT
    wh.program_id,
    COUNT(*)::bigint AS view_count
  FROM watch_histories wh
  GROUP BY wh.program_id
  ORDER BY view_count DESC
  LIMIT COALESCE(sqlc.narg('limit')::int, 7)
),
likes_count AS (
  SELECT
    l.program_id,
    COUNT(*)::bigint AS like_count
  FROM likes l
  WHERE l.program_id IN (SELECT program_id FROM top_view_counts)
  GROUP BY l.program_id
)
SELECT
  p.id AS program_id,
  p.title,
  p.thumbnail_path,
  tvc.view_count,
  COALESCE(lc.like_count, 0)::bigint AS like_count
FROM top_view_counts tvc
JOIN programs p ON p.id = tvc.program_id
LEFT JOIN likes_count lc ON lc.program_id = p.id
ORDER BY tvc.view_count DESC, p.created_at DESC;

-- name: ExistsProgram :one
SELECT EXISTS(
  SELECT 1
  FROM programs
  WHERE id = $1
) AS exists;