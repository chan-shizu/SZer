-- name: CreateLike :exec
INSERT INTO likes (user_id, program_id)
VALUES ($1, $2)
ON CONFLICT (user_id, program_id) DO NOTHING;

-- name: DeleteLike :exec
DELETE FROM likes
WHERE user_id = $1 AND program_id = $2;

-- name: CountLikesByProgramID :one
SELECT COUNT(*)::bigint AS like_count
FROM likes
WHERE program_id = $1;

-- name: HasUserLikedProgram :one
SELECT EXISTS(
  SELECT 1
  FROM likes
  WHERE user_id = $1 AND program_id = $2
) AS liked;
