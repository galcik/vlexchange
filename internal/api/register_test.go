package api

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

type registerTestSuite struct {
	TestServerSuite
}

func (suite *registerTestSuite) TestRegister() {
	recorder := httptest.NewRecorder()

	data, err := json.Marshal(map[string]interface{}{"username": "testuser"})
	suite.Require().NoError(err)

	request, err := http.NewRequest(http.MethodPost, "/register", bytes.NewReader(data))
	suite.Require().NoError(err)

	suite.server.router.ServeHTTP(recorder, request)

	suite.Equal(http.StatusOK, recorder.Code)

	accounts, err := suite.queries.GetAccounts(context.Background())
	suite.Require().NoError(err)
	suite.Equal(1, len(accounts))

	response, err := ioutil.ReadAll(recorder.Body)
	suite.Require().NoError(err)
	var jsonResponse map[string]string
	suite.Require().NoError(json.Unmarshal(response, &jsonResponse))
	token, ok := jsonResponse["token"]
	suite.True(ok)
	suite.Equal(accounts[0].Token, token)
}

func TestRegistration(t *testing.T) {
	suite.Run(t, new(registerTestSuite))
}
