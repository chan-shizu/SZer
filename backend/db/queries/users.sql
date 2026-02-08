-- name: AddPointsToUser :one
UPDATE "user"
SET points = points + $2,
    "updatedAt" = now()
WHERE id = $1
RETURNING points;

-- name: GetUserPoints :one
SELECT points
FROM "user"
WHERE id = $1;

-- name: DeductPointsFromUser :one
UPDATE "user"
SET points = points - $2,
    "updatedAt" = now()
WHERE id = $1 AND points >= $2
RETURNING points;
