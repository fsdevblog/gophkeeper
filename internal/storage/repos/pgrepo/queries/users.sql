-- name: Users_FindByUsername :one
SELECT * FROM users WHERE username = $1;

-- name: Users_Create :one
INSERT INTO users
    (created_at, updated_at, username, encrypted_password)
VALUES
    (NOW(), NOW(), $1, $2)
RETURNING *;