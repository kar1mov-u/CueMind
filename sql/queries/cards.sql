-- name: GetCardsFomCollection :many
SELECT * FROM cards WHERE collection_id=$1;