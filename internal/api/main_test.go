package api

import (
	"database/sql"
	"github.com/galcik/vlexchange/internal/datastore"
	"github.com/galcik/vlexchange/internal/datastore/testqueries"
	"github.com/galcik/vlexchange/internal/testutils"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"os"
	"testing"
)

var testingDb *testutils.TestingDb

func TestMain(m *testing.M) {
	os.Exit(runWithTemporaryDb(m))
}

func runWithTemporaryDb(m *testing.M) int {
	var err error
	testingDb, err = testutils.NewTestingDb()
	if err != nil {
		panic(err)
	}
	defer testingDb.Close()

	if err := testingDb.ExecuteSQLFile("../datastore/schema/schema.sql"); err != nil {
		panic(err)
	}

	if testingDb.TxDriver() == "" {
		panic("missing tx driver for temporary db")
	}

	return m.Run()
}

func newTestServer(t *testing.T, store datastore.Store) *Server {
	server, err := NewServer(store)
	require.NoError(t, err)

	return server
}

type TestServerSuite struct {
	suite.Suite
	db      *sql.DB
	server  *Server
	store   datastore.Store
	queries *testqueries.Queries
}

func (suite *TestServerSuite) BeforeTest(suiteName, testName string) {
	var err error
	suite.db, err = sql.Open(testingDb.TxDriver(), testingDb.GetTxDsn())
	suite.Require().Nil(err)
	suite.store, err = datastore.NewStore(suite.db)
	suite.Require().Nil(err)
	suite.queries = testqueries.New(suite.db)
	suite.server = newTestServer(suite.T(), suite.store)
	suite.Require().Nil(err)
}

func (suite *TestServerSuite) AfterTest(suiteName, testName string) {
	suite.db.Close()
}
