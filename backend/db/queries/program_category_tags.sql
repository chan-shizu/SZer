-- name: CreateProgramCategoryTag :exec
INSERT INTO program_category_tags (
  program_id,
  tag_id
) VALUES (
  $1, $2
);
