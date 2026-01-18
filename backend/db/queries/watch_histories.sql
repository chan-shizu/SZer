-- name: UpsertWatchHistory :one
WITH updated AS (
  UPDATE watch_histories AS wh
  SET
    position_seconds = $3,
    is_completed = $4,
    last_watched_at = now(),
    updated_at = now()
  WHERE wh.user_id = $1 AND wh.program_id = $2 AND wh.is_completed = FALSE
  RETURNING wh.id, wh.user_id, wh.program_id, wh.position_seconds, wh.is_completed, wh.last_watched_at, wh.created_at, wh.updated_at
),
inserted_incomplete AS (
  INSERT INTO watch_histories (
    user_id,
    program_id,
    position_seconds,
    is_completed,
    last_watched_at
  )
  SELECT $1, $2, $3, $4, now()
  WHERE $4 = FALSE
  ON CONFLICT (user_id, program_id) WHERE (is_completed = FALSE)
  DO UPDATE SET
    position_seconds = EXCLUDED.position_seconds,
    is_completed = EXCLUDED.is_completed,
    last_watched_at = now(),
    updated_at = now()
  RETURNING id, user_id, program_id, position_seconds, is_completed, last_watched_at, created_at, updated_at
),
inserted_completed AS (
  INSERT INTO watch_histories (
    user_id,
    program_id,
    position_seconds,
    is_completed,
    last_watched_at
  )
  SELECT $1, $2, $3, TRUE, now()
  WHERE $4 = TRUE AND NOT EXISTS (SELECT 1 FROM updated)
  RETURNING id, user_id, program_id, position_seconds, is_completed, last_watched_at, created_at, updated_at
)
SELECT * FROM updated
UNION ALL
SELECT * FROM inserted_incomplete
UNION ALL
SELECT * FROM inserted_completed
LIMIT 1;

-- name: GetIncompleteWatchHistoryByUserAndProgram :one
SELECT id, user_id, program_id, position_seconds, is_completed, last_watched_at, created_at, updated_at
FROM watch_histories
WHERE user_id = $1 AND program_id = $2 AND is_completed = FALSE;

-- name: ListWatchHistoriesByUser :many
SELECT id, user_id, program_id, position_seconds, is_completed, last_watched_at, created_at, updated_at
FROM watch_histories
WHERE user_id = $1
ORDER BY last_watched_at DESC
LIMIT COALESCE(sqlc.narg('limit')::int, 50)
OFFSET COALESCE(sqlc.narg('offset')::int, 0);
