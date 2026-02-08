-- name: CreateRequest :one
INSERT INTO requests (user_id, content, name, contact, note)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;
