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
  p.created_at,
  p.updated_at;

-- name: GetPrograms :many
SELECT
  p.id AS program_id,
  p.title,
  p.thumbnail_path,
  COALESCE(
    jsonb_agg(DISTINCT jsonb_build_object(
      'id', ct.id,
      'name', ct.name
    )) FILTER (WHERE ct.id IS NOT NULL),
    '[]'::jsonb
  ) AS category_tags
FROM programs p
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
  p.thumbnail_path;