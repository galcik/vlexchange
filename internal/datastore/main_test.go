package datastore

import (
	"database/sql"
	"github.com/galcik/vlexchange/internal/testutils"
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

	if err := testingDb.ExecuteSQLFile("./schema/schema.sql"); err != nil {
		panic(err)
	}

	if testingDb.TxDriver() == "" {
		panic("missing tx driver for temporary db")
	}

	return m.Run()
}

type TestSuite struct {
	suite.Suite
	dbHelper *dbHelper
	db       *sql.DB
	store    Store
}

func (suite *TestSuite) BeforeTest(suiteName, testName string) {
	var err error
	suite.db, err = sql.Open(testingDb.TxDriver(), testingDb.GetTxDsn())
	suite.Require().NoError(err)
	suite.store, err = NewStore(suite.db)
	suite.Require().NoError(err)

	suite.dbHelper = newDBHelper(suite.T(), suite.db)
}

func (suite *TestSuite) AfterTest(suiteName, testName string) {
	suite.db.Close()
}
