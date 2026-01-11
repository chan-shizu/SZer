-- name: CreatePerformer :one
INSERT INTO performers (
  first_name,
  last_name,
  first_name_kana,
  last_name_kana,
  image_path
) VALUES (
  $1, $2, $3, $4, $5
)
RETURNING id, first_name, last_name, first_name_kana, last_name_kana, image_path, created_at, updated_at;
