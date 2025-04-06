-- name: CreateUser :one
INSERT INTO users (
    created_at, updated_at, username, email, password
) VALUES (
    NOW(), NOW(), $1, $2, $3
)
RETURNING *;

-- name: GetUser :one
SELECT * FROM users WHERE email=$1 OR username=$1;