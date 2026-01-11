-- name: CreateComment :one
INSERT INTO comments (
  program_id,
  content
) VALUES (
  $1, $2
)
RETURNING id, program_id, content, created_at, updated_at;
