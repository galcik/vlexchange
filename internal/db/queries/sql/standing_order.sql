-- name: CreateStandingOrder :one
INSERT INTO standing_order (account_id, type, state, quantity, limit_price, reserved_btc_amount, reserved_usd_amount,
                            webhook_url)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING *;

-- name: GetStandingOrder :one
SELECT *
FROM standing_order
WHERE id = $1 LIMIT 1;

-- name: GetStandingOrders :many
SELECT *
FROM standing_order
WHERE id IN (@order_ids::integer[]);

-- name: DeleteStandingOrder :exec
DELETE
FROM standing_order
WHERE id = $1;

-- name: GetReservedAmounts :one
SELECT SUM(reserved_usd_amount) as usd_amount, SUM(reserved_btc_amount) as btc_amount
FROM standing_order
WHERE account_id = $1
  AND state = 'live';

-- name: GetBestBuyer :one
SELECT *
FROM standing_order
WHERE state = 'live'
  AND type = 'buy'
  AND limit_price >= $1
ORDER BY limit_price DESC LIMIT 1;

-- name: GetBestSeller :one
SELECT *
FROM standing_order
WHERE state = 'live'
  AND type = 'sell'
  AND limit_price <= $1
ORDER BY limit_price ASC LIMIT 1;

-- name: SatisfyOrder :one
UPDATE standing_order
SET quantity            = quantity - $2,
    filled_quantity     = filled_quantity + $2,
    filled_price        = filled_price + $3,
    state               = CASE
                              WHEN quantity - $2 = 0 THEN 'fulfilled'
                              ELSE state
        END,
    reserved_usd_amount = reserved_usd_amount - $4,
    reserved_btc_amount = reserved_btc_amount - $5
WHERE id = $1
  AND quantity - $2 >= 0 RETURNING *;

