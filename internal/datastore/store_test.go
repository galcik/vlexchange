package datastore

import (
	"database/sql"
	"github.com/galcik/vlexchange/internal/currency"
	"github.com/galcik/vlexchange/internal/datastore/queries"
	"github.com/stretchr/testify/suite"
	"testing"
)

type TestStoreSuite struct {
	suite.Suite
	dbHelper *dbHelper
	db       *sql.DB
	store    Store
}

func (suite *TestStoreSuite) BeforeTest(suiteName, testName string) {
	var err error
	suite.db, err = sql.Open(temporaryDb.TxDriver(), temporaryDb.GetTxDsn())
	suite.Nil(err)
	suite.store, err = NewStore(suite.db)
	suite.Nil(err)

	suite.dbHelper = newDBHelper(suite.T(), suite.db)
}

func (suite *TestStoreSuite) AfterTest(suiteName, testName string) {
	suite.db.Close()
}

func (suite *TestStoreSuite) TestGetAccountByToken() {
	testAccount1 := suite.dbHelper.createAccount(queries.Account{Username: "tester1", Token: "111111"})
	testAccount2 := suite.dbHelper.createAccount(queries.Account{Username: "tester2", Token: "222222"})

	testCases := []struct {
		token    string
		expected *queries.Account
	}{
		{
			"111111", testAccount1,
		},
		{
			"222222", testAccount2,
		},
		{
			"333333", nil,
		},
	}

	for i := range testCases {
		tc := testCases[i]
		account, err := suite.store.GetAccountByToken(tc.token)
		suite.NoError(err)
		if tc.expected != nil {
			suite.Equal(*tc.expected, *account)
		} else {
			suite.Nil(account)
		}
	}
}

func (suite *TestStoreSuite) TestGetAccount() {
	testAccount1 := suite.dbHelper.createAccount(queries.Account{Username: "tester1", Token: "111111"})
	testAccount2 := suite.dbHelper.createAccount(queries.Account{Username: "tester2", Token: "222222"})

	testCases := []struct {
		accountId int32
		expected  *queries.Account
	}{
		{
			testAccount1.ID, testAccount1,
		},
		{
			testAccount2.ID, testAccount2,
		},
		{
			testAccount1.ID + testAccount2.ID + 1, nil,
		},
	}

	for i := range testCases {
		tc := testCases[i]
		account, err := suite.store.GetAccount(tc.accountId)
		suite.NoError(err)
		if tc.expected != nil {
			suite.Equal(*tc.expected, *account)
		} else {
			suite.Nil(account)
		}
	}
}

func (suite *TestStoreSuite) TestComplexScenario() {
	userA := suite.dbHelper.createAccount(queries.Account{Username: "A", Token: "AA"})
	userB := suite.dbHelper.createAccount(queries.Account{Username: "B", Token: "BB"})
	userC := suite.dbHelper.createAccount(queries.Account{Username: "C", Token: "CC"})
	userD := suite.dbHelper.createAccount(queries.Account{Username: "D", Token: "DD"})

	var success bool
	var err error

	success, err = suite.store.DepositAccount(userA.ID, currency.NewBTC(1), currency.USD(0))
	suite.Equal(true, success)
	suite.Nil(err)

	success, err = suite.store.DepositAccount(userB.ID, currency.NewBTC(10), currency.USD(0))
	suite.Equal(true, success)
	suite.Nil(err)

	success, err = suite.store.DepositAccount(userC.ID, currency.BTC(0), currency.NewUSD(250_000))
	suite.Equal(true, success)
	suite.Nil(err)

	success, err = suite.store.DepositAccount(userD.ID, currency.BTC(0), currency.NewUSD(300_000))
	suite.Equal(true, success)
	suite.Nil(err)

	var affectedOrderIds []int32
	order, affectedOrderIds, err := suite.store.CreateStandingOrder(CreateStandingOrderParams{
		AccountID:  userA.ID,
		OrderType:  queries.OrderTypeSell,
		Quantity:   currency.NewBTC(10),
		LimitPrice: currency.NewUSD(10_000),
	})
	suite.Nil(err)
	suite.NotNil(order)
	suite.Equal(queries.OrderStateCancelled, order.State)
	suite.Equal(len(affectedOrderIds), 1)

	success, err = suite.store.DepositAccount(userA.ID, currency.NewBTC(9), currency.USD(0))
	suite.Equal(true, success)
	suite.Nil(err)

	order1, affectedOrderIds, err := suite.store.CreateStandingOrder(CreateStandingOrderParams{
		AccountID:  userA.ID,
		OrderType:  queries.OrderTypeSell,
		Quantity:   currency.NewBTC(10),
		LimitPrice: currency.NewUSD(10_000),
	})
	suite.Nil(err)
	suite.NotNil(order1)
	suite.Equal(queries.OrderStateLive, order1.State)
	suite.ElementsMatch([]int32{order1.ID}, affectedOrderIds)

	order2, affectedOrderIds, err := suite.store.CreateStandingOrder(CreateStandingOrderParams{
		AccountID:  userB.ID,
		OrderType:  queries.OrderTypeSell,
		Quantity:   currency.NewBTC(10),
		LimitPrice: currency.NewUSD(20_000),
	})
	suite.Nil(err)
	suite.NotNil(order2)
	suite.Equal(queries.OrderStateLive, order2.State)
	suite.ElementsMatch([]int32{order2.ID}, affectedOrderIds)

	orderResult, affectedOrderIds, err := suite.store.ExecuteMarketOrder(CreateMarketOrderParams{
		AccountID: userC.ID,
		OrderType: queries.OrderTypeBuy,
		Quantity:  currency.NewBTC(15),
	})
	suite.Nil(err)
	suite.NotNil(order)
	suite.Equal(currency.NewBTC(15), orderResult.Quantity)
	suite.Equal(currency.NewUSD(200_000), orderResult.Price)
	suite.ElementsMatch([]int32{order1.ID, order2.ID}, affectedOrderIds)
	userC = suite.dbHelper.getAccounts()[userC.ID]
	suite.Equal(currency.NewBTC(15), currency.BTC(userC.BtcAmount))
	suite.Equal(currency.NewUSD(50_000), currency.USD(userC.UsdAmount))
	orders := suite.dbHelper.getStandingOrders()
	suite.Equal(3, len(orders))
	order1 = orders[order1.ID]
	suite.Equal(queries.OrderStateFulfilled, order1.State)
	suite.Equal(currency.NewBTC(10), currency.BTC(order1.FilledQuantity))
	suite.Equal(currency.NewBTC(0), currency.BTC(order1.Quantity))
	suite.Equal(currency.NewUSD(100_000), currency.USD(order1.FilledPrice))
	order2 = orders[order2.ID]
	suite.Equal(queries.OrderStateLive, order2.State)
	suite.Equal(currency.NewBTC(5), currency.BTC(order2.FilledQuantity))
	suite.Equal(currency.NewBTC(5), currency.BTC(order2.Quantity))
	suite.Equal(currency.NewUSD(100_000), currency.USD(order2.FilledPrice))

	order3, affectedOrderIds, err := suite.store.CreateStandingOrder(CreateStandingOrderParams{
		AccountID:  userD.ID,
		OrderType:  queries.OrderTypeBuy,
		Quantity:   currency.NewBTC(20),
		LimitPrice: currency.NewUSD(10_000),
	})
	suite.Nil(err)
	suite.NotNil(order3)
	suite.Equal(queries.OrderStateLive, order3.State)
	suite.ElementsMatch([]int32{order3.ID}, affectedOrderIds)
	orders = suite.dbHelper.getStandingOrders()
	suite.Equal(4, len(orders))

	order4, affectedOrderIds, err := suite.store.CreateStandingOrder(CreateStandingOrderParams{
		AccountID:  userD.ID,
		OrderType:  queries.OrderTypeBuy,
		Quantity:   currency.NewBTC(10),
		LimitPrice: currency.NewUSD(25_000),
	})
	suite.Nil(err)
	suite.NotNil(order4)
	suite.Equal(queries.OrderStateCancelled, order4.State)
	suite.ElementsMatch([]int32{order4.ID}, affectedOrderIds)
	orders = suite.dbHelper.getStandingOrders()
	suite.Equal(5, len(orders))

	err = suite.store.DeleteStandingOrder(order3.ID)
	suite.Nil(err)
	orders = suite.dbHelper.getStandingOrders()
	suite.Equal(4, len(orders))

	order5, affectedOrderIds, err := suite.store.CreateStandingOrder(CreateStandingOrderParams{
		AccountID:  userD.ID,
		OrderType:  queries.OrderTypeBuy,
		Quantity:   currency.NewBTC(10),
		LimitPrice: currency.NewUSD(25_000),
	})
	suite.Nil(err)
	suite.NotNil(order5)
	suite.Equal(queries.OrderStateLive, order5.State)
	suite.Equal(currency.NewBTC(5), currency.BTC(order5.Quantity))
	suite.Equal(currency.NewBTC(5), currency.BTC(order5.FilledQuantity))
	suite.Equal(currency.NewUSD(100_000), currency.USD(order5.FilledPrice))

	suite.ElementsMatch([]int32{order5.ID, order2.ID}, affectedOrderIds)
	orders = suite.dbHelper.getStandingOrders()
	suite.Equal(5, len(orders))
	suite.Equal(queries.OrderStateFulfilled, orders[order2.ID].State)
}

func TestStoreTestSuite(t *testing.T) {
	suite.Run(t, new(TestStoreSuite))
}
