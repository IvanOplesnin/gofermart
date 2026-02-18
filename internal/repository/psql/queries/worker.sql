-- name: ListPending :many
SELECT
    user_id,
    "number",
    "status"      AS order_status,
    uploaded_at
FROM order_numbers
WHERE
    "status" = ANY(sqlc.arg(statuses)::text[])
    AND (next_sync_at IS NULL OR next_sync_at <= $2)
ORDER BY
    next_sync_at NULLS FIRST,
    uploaded_at
LIMIT $1;



-- name: UpdateFromAccrual :exec
UPDATE order_numbers
SET
    "status" = $2,
    next_sync_at = $3
WHERE
    "number" = $1;


-- name: UpdateSyncTime :exec
UPDATE order_numbers
SET next_sync_at = $2
WHERE "number" = $1;



-- name: MarkOrderProcessed :one
UPDATE order_numbers
SET
    "status" = 'PROCESSED',
    accrual = $2,
    next_sync_at = NULL
WHERE
    "number" = $1
    AND user_id = $3
    AND "status" <> 'PROCESSED'
RETURNING user_id, accrual;


-- name: AddToUserBalanceUpsert :exec
INSERT INTO user_balance (user_id, balance, withdrawn)
VALUES ($1, $2, 0)
ON CONFLICT (user_id)
DO UPDATE SET balance = user_balance.balance + EXCLUDED.balance;
