-- name: CreateCategoryTag :one
INSERT INTO category_tags (
  name
) VALUES (
  $1
)
RETURNING id, name, created_at, updated_at;
