
-- name: ListCommentsByProgramID :many
SELECT c.id, c.program_id, c.user_id, u.name AS user_name, c.content, c.created_at, c.updated_at
FROM comments c
LEFT JOIN "user" u ON c.user_id = u.id
WHERE c.program_id = $1
ORDER BY c.created_at DESC;


-- name: CreateComment :one
INSERT INTO comments (
  program_id,
  user_id,
  content
) VALUES (
  $1, $2, $3
)
RETURNING id, program_id, user_id, content, created_at, updated_at;

-- コメント作成後、user_nameも返すためのクエリ
-- name: CreateCommentWithUserName :one
SELECT c.id, c.program_id, c.user_id, u.name AS user_name, c.content, c.created_at, c.updated_at
FROM comments c
LEFT JOIN "user" u ON c.user_id = u.id
WHERE c.id = (
  SELECT c2.id FROM comments c2 WHERE c2.program_id = $1 AND c2.user_id = $2 AND c2.content = $3 ORDER BY c2.created_at DESC LIMIT 1
);
