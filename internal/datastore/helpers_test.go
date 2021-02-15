package datastore

import (
	"context"
	"database/sql"
	"github.com/galcik/vlexchange/internal/datastore/queries"
	"github.com/stretchr/testify/require"
	"testing"
)

type dbHelper struct {
	t       *testing.T
	db      *sql.DB
	context context.Context
	querier queries.Querier
}

func newDBHelper(t *testing.T, db *sql.DB) *dbHelper {
	return &dbHelper{t, db, context.Background(), queries.New(db)}
}

func (helper *dbHelper) createAccount(accountSpec queries.Account) *queries.Account {
	createdAccount, err := helper.querier.InsertAccountForTest(helper.context, queries.InsertAccountForTestParams{
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

func (helper *dbHelper) getAccounts() map[int32]*queries.Account {
	accounts, err := helper.querier.SelectAccountsForTest(helper.context)
	require.NoError(helper.t, err, "unable to get accounts")
	result := make(map[int32]*queries.Account, len(accounts))
	for i := range accounts {
		result[accounts[i].ID] = &accounts[i]
	}
	return result
}

func (helper *dbHelper) createStandingOrder(orderSpec queries.StandingOrder) *queries.StandingOrder {
	createdOrder, err := helper.querier.InsertStandingOrderForTest(helper.context,
		queries.InsertStandingOrderForTestParams{
			AccountID:         orderSpec.AccountID,
			Type:              orderSpec.Type,
			State:             orderSpec.State,
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

func (helper *dbHelper) getStandingOrders() map[int32]*queries.StandingOrder {
	standingOrders, err := helper.querier.SelectStandingOrdersForTest(helper.context)
	require.NoError(helper.t, err, "unable to get standing orders")
	result := make(map[int32]*queries.StandingOrder, len(standingOrders))
	for i := range standingOrders {
		result[standingOrders[i].ID] = &standingOrders[i]
	}
	return result
}
