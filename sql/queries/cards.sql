-- name: GetCardsFomCollection :many
SELECT * FROM cards WHERE collection_id=$1;

-- name: GetCard :one
SELECT cards.*
FROM cards
JOIN collections ON cards.collection_id = collections.id
WHERE cards.id = $1 AND collections.user_id = $2;

-- name: CreateCard :one
INSERT INTO cards(
    front, back, created_at, collection_id
) VALUES 
    ($1, $2, NOW(), $3)
RETURNING id;


-- name: DeleteCard :exec
DELETE FROM cards WHERE id=$1 and collection_id=$2;

-- name: DeleteAllCards :exec
DELETE FROM cards WHERE collection_id=$1;