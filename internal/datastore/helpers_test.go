package datastore

import (
	"context"
	"database/sql"
	"github.com/galcik/vlexchange/internal/datastore/queries"
	tq "github.com/galcik/vlexchange/internal/datastore/testqueries"
	"github.com/stretchr/testify/require"
	"testing"
)

type dbHelper struct {
	t       *testing.T
	db      *sql.DB
	context context.Context
	queries *tq.Queries
}

func newDBHelper(t *testing.T, db *sql.DB) *dbHelper {
	return &dbHelper{t, db, context.Background(), tq.New(db)}
}

func (helper *dbHelper) createAccount(accountSpec queries.Account) *tq.Account {
	createdAccount, err := helper.queries.CreateAccount(helper.context, tq.CreateAccountParams{
		Username:  accountSpec.Username,
		Token:     accountSpec.Token,
		UsdAmount: accountSpec.UsdAmount,
		BtcAmount: accountSpec.BtcAmount,
	})
	require.NoError(helper.t, err, "unable to create test account")
	require.Equal(helper.t, accountSpec.Username, createdAccount.Username)
	require.Equal(helper.t, accountSpec.Token, createdAccount.Token)
	require.Equal(helper.t, accountSpec.UsdAmount, createdAccount.UsdAmount)
	require.Equal(helper.t, accountSpec.BtcAmount, createdAccount.BtcAmount)
	return &createdAccount
}

func (helper *dbHelper) getAccounts() map[int32]*tq.Account {
	accounts, err := helper.queries.GetAccounts(helper.context)
	require.NoError(helper.t, err, "unable to get accounts")
	result := make(map[int32]*tq.Account, len(accounts))
	for i := range accounts {
		result[accounts[i].ID] = &accounts[i]
	}
	return result
}

func (helper *dbHelper) createStandingOrder(orderSpec queries.StandingOrder) *tq.StandingOrder {
	createdOrder, err := helper.queries.CreateStandingOrder(helper.context,
		tq.CreateStandingOrderParams{
			AccountID:         orderSpec.AccountID,
			Type:              tq.OrderType(orderSpec.Type),
			State:             tq.OrderState(orderSpec.State),
			Quantity:          orderSpec.Quantity,
			FilledQuantity:    orderSpec.FilledQuantity,
			FilledPrice:       orderSpec.FilledPrice,
			LimitPrice:        orderSpec.LimitPrice,
			ReservedUsdAmount: orderSpec.ReservedUsdAmount,
			ReservedBtcAmount: orderSpec.ReservedBtcAmount,
			WebhookUrl:        orderSpec.WebhookUrl,
		})
	require.NoError(helper.t, err, "unable to create test account")
	return &createdOrder
}

func (helper *dbHelper) getStandingOrders() map[int32]*tq.StandingOrder {
	standingOrders, err := helper.queries.GetStandingOrders(helper.context)
	require.NoError(helper.t, err, "unable to get standing orders")
	result := make(map[int32]*tq.StandingOrder, len(standingOrders))
	for i := range standingOrders {
		result[standingOrders[i].ID] = &standingOrders[i]
	}
	return result
}
