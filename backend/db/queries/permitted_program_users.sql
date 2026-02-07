-- name: IsUserPermittedForProgram :one
SELECT EXISTS (
  SELECT 1 FROM permitted_program_users
  WHERE user_id = $1 AND program_id = $2
) AS is_permitted;

-- name: AddPermittedProgramUser :exec
INSERT INTO permitted_program_users (user_id, program_id)
VALUES ($1, $2)
ON CONFLICT (user_id, program_id) DO NOTHING;

-- name: RemovePermittedProgramUser :exec
DELETE FROM permitted_program_users
WHERE user_id = $1 AND program_id = $2;
