-- name: CreateAuthUser :one
INSERT INTO "user" (
  id,
  name,
  email,
  "emailVerified",
  image,
  points,
  "createdAt",
  "updatedAt"
) VALUES (
  $1, $2, $3, $4, $5, $6, now(), now()
)
RETURNING id, name, email, "emailVerified", image, points, "createdAt", "updatedAt";


-- name: CreateCredentialAccount :one
INSERT INTO "account" (
  id,
  "accountId",
  "providerId",
  "userId",
  password,
  "createdAt",
  "updatedAt"
) VALUES (
  $1, $2, 'credential', $3, $4, now(), now()
)
RETURNING id, "accountId", "providerId", "userId", password, "createdAt", "updatedAt";
