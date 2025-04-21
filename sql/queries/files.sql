-- name: CreateFile :one
INSERT INTO files(
    collection_id, user_id, file_name, file_key
) VALUES (
    $1, $2, $3, $4
)
RETURNING id;

-- name: GetFilesForCollection :many
SELECT * FROM files WHERE collection_id=$1 and user_id = $2;