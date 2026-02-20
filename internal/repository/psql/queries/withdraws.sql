-- name: EnsureBalanceRow :exec
INSERT INTO user_balance (user_id, balance, withdrawn)
VALUES ($1, 0, 0)
ON CONFLICT (user_id) DO NOTHING;


-- name: LockBalanceRow :one
SELECT balance, withdrawn
FROM user_balance
WHERE user_id = $1
FOR UPDATE;


-- name: WithdrawIfEnough :one
UPDATE user_balance
SET
  balance   = balance - sqlc.arg(summa),
  withdrawn = withdrawn + sqlc.arg(summa)
WHERE user_id = $1
  AND balance >= sqlc.arg(summa)
RETURNING balance, withdrawn;


-- name: AddWithdrawal :exec
INSERT INTO withdraws (user_id, order_number, summa, processed_at)
VALUES ($1, $2, $3, $4);


-- name: ListWithdraws :many
SELECT id, user_id, order_number, summa, processed_at
FROM withdraws
WHERE user_id = $1
ORDER BY processed_at DESC;


-- name: BalnceByUserID :one
SELECT id, user_id, balance, withdrawn
FROM user_balance
WHERE user_id = $1
LIMIT 1;