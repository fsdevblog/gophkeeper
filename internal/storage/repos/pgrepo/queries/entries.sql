-- name: Entries_GetAllByUserID :many
SELECT * FROM entries WHERE user_id = $1 OFFSET @_offset LIMIT @_limit;

-- name: Entries_GetCountByUserID :one
SELECT count(*) FROM entries WHERE user_id = $1;

-- name: Entries_Create :one
INSERT INTO entries
    (type, user_id, device_id, title)
    VALUES
    ($1, $2, $3, $4)
RETURNING *;