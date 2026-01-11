-- name: CreateProgramPerformer :exec
INSERT INTO program_performers (
  program_id,
  performer_id
) VALUES (
  $1, $2
);
