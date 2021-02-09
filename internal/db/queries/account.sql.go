// Code generated by sqlc. DO NOT EDIT.
// source: account.sql

package queries

import (
	"context"
)

const createAccount = `-- name: CreateAccount :one
INSERT INTO account (username, token)
VALUES ($1, $2)
RETURNING id, username, token, usd_amount, btc_amount
`

type CreateAccountParams struct {
	Username string
	Token    string
}

func (q *Queries) CreateAccount(ctx context.Context, arg CreateAccountParams) (Account, error) {
	row := q.db.QueryRowContext(ctx, createAccount, arg.Username, arg.Token)
	var i Account
	err := row.Scan(
		&i.ID,
		&i.Username,
		&i.Token,
		&i.UsdAmount,
		&i.BtcAmount,
	)
	return i, err
}

const getAccountById = `-- name: GetAccountById :one
SELECT id, username, token, usd_amount, btc_amount
FROM account
WHERE id = $1
LIMIT 1
`

func (q *Queries) GetAccountById(ctx context.Context, id int32) (Account, error) {
	row := q.db.QueryRowContext(ctx, getAccountById, id)
	var i Account
	err := row.Scan(
		&i.ID,
		&i.Username,
		&i.Token,
		&i.UsdAmount,
		&i.BtcAmount,
	)
	return i, err
}

const getAccountByToken = `-- name: GetAccountByToken :one
SELECT id, username, token, usd_amount, btc_amount
FROM account
WHERE token = $1
LIMIT 1
`

func (q *Queries) GetAccountByToken(ctx context.Context, token string) (Account, error) {
	row := q.db.QueryRowContext(ctx, getAccountByToken, token)
	var i Account
	err := row.Scan(
		&i.ID,
		&i.Username,
		&i.Token,
		&i.UsdAmount,
		&i.BtcAmount,
	)
	return i, err
}

const transferAmounts = `-- name: TransferAmounts :execrows
UPDATE account
SET btc_amount = btc_amount + $2,
    usd_amount = usd_amount + $3
WHERE id = $1
  AND btc_amount + $2 >= 0
  AND usd_amount + $3 >= 0
`

type TransferAmountsParams struct {
	ID        int32
	BtcAmount int64
	UsdAmount int64
}

func (q *Queries) TransferAmounts(ctx context.Context, arg TransferAmountsParams) (int64, error) {
	result, err := q.db.ExecContext(ctx, transferAmounts, arg.ID, arg.BtcAmount, arg.UsdAmount)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
