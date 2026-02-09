-- name: AddUser :one
INSERT INTO users ("login", password_hash)
VALUES ($1, $2)
RETURNING id;