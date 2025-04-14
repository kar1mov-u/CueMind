-- name: CreateCollection :one
INSERT INTO collections(
    created_at, name, user_id
) VALUES (
    NOW(), $1, $2
)
RETURNING id;


-- name: GetCollectionById :one
SELECT * FROM collections WHERE id=$1 and user_id=$2;

-- name: CheckUserCollectionOwnership :one
SELECT 1 FROM collections WHERE id = $1 AND user_id = $2;

-- name: ListCollections :many
SELECT id, name FROM collections WHERE user_id=$1;

-- name: DeleteCollection :exec
DELETE FROM collections WHERE id=$1 and user_id=$2;

