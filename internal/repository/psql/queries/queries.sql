-- name: AddUser :one
INSERT INTO users ("login", password_hash)
VALUES ($1, $2)
RETURNING id;


-- name: GetUserByLogin :one
SELECT id, "login", password_hash
FROM users
WHERE "login" = $1
LIMIT 1;


-- name: GetUserByID :one
SELECT id
FROM users
WHERE id = $1
LIMIT 1;


-- name: AddOrder :exec
INSERT INTO order_numbers (user_id, "number", "status", uploaded_at)
VALUES ($1, $2, $3, $4);


-- name: GetOrderByNumber :one
SELECT id, user_id, "number", "status", uploaded_at
FROM order_numbers
WHERE "number" = $1
LIMIT 1;


-- name: GetOrdersByUserID :many
SELECT id, user_id, "number", "status", accrual, uploaded_at
FROM order_numbers
WHERE user_id = $1
ORDER BY uploaded_at DESC;