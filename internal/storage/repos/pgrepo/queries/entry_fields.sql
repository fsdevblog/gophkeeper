-- name: EntryFields_BatchCreate :batchone
INSERT INTO entry_fields
    (entry_id, key, value, is_private)
VALUES
    ($1, $2, $3, $4)
RETURNING *;

-- name: EntryFields_GetFieldsByEntryIDs :many
SELECT * FROM entry_fields WHERE entry_id = ANY(@ids::uuid[]);
