package api

import (
	"bytes"
	"context"
	"encoding/json"
	cmMocks "github.com/galcik/vlexchange/internal/coinmarket/mocks"
	"github.com/galcik/vlexchange/internal/currency"
	"github.com/galcik/vlexchange/internal/datastore/testqueries"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

type PostBalanceTestCase struct {
	request         map[string]interface{}
	expectedUsd     currency.USD
	expectedBtc     currency.BTC
	expectedSuccess bool
}

type PostBalanceTestSuite struct {
	TestServerSuite
	testCase PostBalanceTestCase
}

func (suite *PostBalanceTestSuite) TestPostBalance() {
	_, err := suite.queries.CreateAccount(context.Background(), testqueries.CreateAccountParams{
		Username:  "TestUser",
		Token:     "111222",
		UsdAmount: currency.NewUSD(40_000).Internal(),
		BtcAmount: currency.NewBTC(1.5).Internal(),
	})
	suite.Require().NoError(err)

	data, err := json.Marshal(suite.testCase.request)
	suite.Require().NoError(err)

	request, err := http.NewRequest(http.MethodPost, "/balance", bytes.NewReader(data))
	suite.Require().NotNil(request)
	suite.Require().NoError(err)
	request.Header.Set("X-Token", "111222")

	recorder := httptest.NewRecorder()
	suite.server.router.ServeHTTP(recorder, request)

	suite.Equal(http.StatusOK, recorder.Code)

	response, err := ioutil.ReadAll(recorder.Body)
	suite.Require().NoError(err)
	var jsonResponse map[string]bool
	suite.Require().NoError(json.Unmarshal(response, &jsonResponse))
	success, ok := jsonResponse["success"]
	suite.True(ok)
	suite.Equal(suite.testCase.expectedSuccess, success)

	accounts, err := suite.queries.GetAccounts(context.Background())
	suite.Require().NoError(err)
	suite.Equal(1, len(accounts))
	suite.Equal(suite.testCase.expectedUsd.Internal(), accounts[0].UsdAmount)
	suite.Equal(suite.testCase.expectedBtc.Internal(), accounts[0].BtcAmount)
}

func TestPostBalance(t *testing.T) {
	testCases := []PostBalanceTestCase{
		{
			map[string]interface{}{"currency": "usd", "topupAmount": "0"},
			currency.NewUSD(40_000),
			currency.NewBTC(1.5),
			true,
		},
		{
			map[string]interface{}{"currency": "usd", "topupAmount": "100.5"},
			currency.NewUSD(40_100.5),
			currency.NewBTC(1.5),
			true,
		},
		{
			map[string]interface{}{"currency": "btc", "topupAmount": "1.5"},
			currency.NewUSD(40_000),
			currency.NewBTC(3),
			true,
		},
	}
	for _, testCase := range testCases {
		testSuite := new(PostBalanceTestSuite)
		testSuite.testCase = testCase
		suite.Run(t, testSuite)
	}
}

type GetBalanceTestSuite struct {
	TestServerSuite
}

func (suite *GetBalanceTestSuite) TestGetBalance() {
	_, err := suite.queries.CreateAccount(context.Background(), testqueries.CreateAccountParams{
		Username:  "TestUser",
		Token:     "111222",
		UsdAmount: currency.NewUSD(40_000).Internal(),
		BtcAmount: currency.NewBTC(1.5).Internal(),
	})
	suite.Require().NoError(err)

	request, err := http.NewRequest(http.MethodGet, "/balance", http.NoBody)
	suite.Require().NotNil(request)
	suite.Require().NoError(err)
	request.Header.Set("X-Token", "111222")

	recorder := httptest.NewRecorder()
	cmServiceMock := &cmMocks.CoinmarketService{}
	cmServiceMock.On("GetBTCPriceInUSD", mock.Anything).Return(float64(10000), nil)
	suite.server.coinmarketService = cmServiceMock

	suite.server.router.ServeHTTP(recorder, request)

	suite.Equal(http.StatusOK, recorder.Code)

	response, err := ioutil.ReadAll(recorder.Body)
	suite.Require().NoError(err)
	expectedResponse, err := json.Marshal(map[string]string{
		"BTC":           "1.50000000",
		"USD":           "40000.00",
		"USDEquivalent": "15000.00",
	})
	suite.Require().NoError(err)
	suite.JSONEq(string(response), string(expectedResponse))
}

func TestGetBalance(t *testing.T) {
	suite.Run(t, new(GetBalanceTestSuite))
}
