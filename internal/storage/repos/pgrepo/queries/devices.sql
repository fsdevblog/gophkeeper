-- name: Devices_Create :one
INSERT INTO devices
    (type, user_id, client_uuid, platform, platform_version, state_version, app_version, last_active_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;
