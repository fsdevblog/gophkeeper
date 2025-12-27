-- name: Devices_Create :one
INSERT INTO devices
    (type, user_id, client_uuid, platform, platform_version, state_version, app_version)
VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (user_id, client_uuid)
    DO UPDATE SET
                  updated_at = NOW(),
                  last_active_at = NOW()
RETURNING *;
