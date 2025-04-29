-- name: GetFilesForCollection :many
SELECT * FROM files WHERE collection_id=$1 and user_id = $2;

-- name: DeleteFile :exec
DELETE FROM files WHERE id=$1;

-- name: CraeteFileEntry :one
INSERT INTO files(
    collection_id, user_id
) VALUES (
    $1, $2
) 
RETURNING id;

-- name: CompeleteFileDetails :exec
UPDATE files SET file_name=$1,format =$2 WHERE id=$3 and collection_id=$4 and user_id=$5;


-- name: ProcessedCheck :one
SELECT processed FROM files WHERE id = $1;

-- name: Processed :exec
UPDATE files SET processed = $1 WHERE id = $2;

-- name: DeleteAllFiles :exec
DELETE FROM files WHERE collection_id=$1 and user_id=$2;