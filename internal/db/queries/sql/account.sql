-- name: GetAccountByToken :one
SELECT *
FROM account
WHERE token = $1
LIMIT 1;

-- name: GetAccountById :one
SELECT *
FROM account
WHERE id = $1
LIMIT 1;

-- name: CreateAccount :one
INSERT INTO account (username, token)
VALUES ($1, $2)
RETURNING *;

-- name: TransferAmounts :execrows
UPDATE account
SET btc_amount = btc_amount + $2,
    usd_amount = usd_amount + $3
WHERE id = $1
  AND btc_amount + $2 >= 0
  AND usd_amount + $3 >= 0;
