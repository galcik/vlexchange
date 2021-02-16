-- name: GetAccounts :many
SELECT *
FROM account;

-- name: CreateAccount :one
INSERT INTO account (username, token, usd_amount, btc_amount)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetStandingOrders :many
SELECT *
FROM standing_order;

-- name: CreateStandingOrder :one
INSERT INTO standing_order (account_id, type, state, quantity, filled_quantity, filled_price, limit_price,
                            reserved_btc_amount, reserved_usd_amount,
                            webhook_url)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING *;



