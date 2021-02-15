package datastore

import (
	"github.com/galcik/vlexchange/internal/testutils"
	"os"
	"testing"
)

var temporaryDb *testutils.TemporaryDb

func TestMain(m *testing.M) {
	os.Exit(runWithTemporaryDb(m))
}

func runWithTemporaryDb(m *testing.M) int {
	var err error
	temporaryDb, err = testutils.NewTemporaryDb()
	if err != nil {
		panic(err)
	}
	defer temporaryDb.Close()

	if err := temporaryDb.ExecuteSQLFile("./queries/sql/schema.sql"); err != nil {
		panic(err)
	}

	if temporaryDb.TxDriver() == "" {
		panic("missing tx driver for temporary db")
	}

	return m.Run()
}
