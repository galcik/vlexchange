// Inspired by https://pypi.org/project/testing.postgresql/
package testutils

import (
	"database/sql"
	"fmt"
	"github.com/DATA-DOG/go-txdb"
	"github.com/koron-go/pgctl"
	_ "github.com/lib/pq"
	"io/ioutil"
	"os"
	"path/filepath"
)

type TestingDb struct {
	dir      string
	dsn      string
	txDriver string
}

const minPort = 5000
const maxPort = 5040

var txDriverIdx = 0
var txDsnIdx = 0

func init() {
	if os.Getenv("POSTGRES_HOME") == "" {
		if postgresHome := findPostgresHome(); postgresHome != "" {
			os.Setenv("POSTGRES_HOME", postgresHome)
		}
	}
}

func NewTestingDb() (*TestingDb, error) {
	var tmpDb TestingDb
	if err := tmpDb.open(); err != nil {
		return nil, err
	}
	return &tmpDb, nil
}

func (db *TestingDb) open() error {
	if db.dsn != "" {
		return fmt.Errorf("temporary db already active")
	}

	dir, err := ioutil.TempDir("", "tpg-")
	if err != nil {
		return fmt.Errorf("opening failed: %w", err)
	}
	dir = filepath.Join(dir, "data")

	initDBOptions := pgctl.InitDBOptions{}

	err = pgctl.InitDB(dir, &initDBOptions)
	if err != nil {
		return fmt.Errorf("opening failed: %w", err)
	}

	startOptions := pgctl.StartOptions{}
	for port := minPort; port <= maxPort; port++ {
		startOptions.Port = uint16(port)
		err = pgctl.Start(dir, &startOptions)
		if err == nil {
			break
		}
	}
	if err != nil {
		return fmt.Errorf("opening failed: %w", err)
	}

	err = pgctl.Status(dir)
	if err != nil {
		defer db.Close()
		return fmt.Errorf("opening failed: %w", err)
	}

	db.dsn = pgctl.Name(&initDBOptions, &startOptions)
	return nil
}

func (db *TestingDb) Close() error {
	if db.dsn == "" {
		return nil
	}

	if err := pgctl.Stop(db.dir); err != nil {
		return err
	}
	db.dsn = ""
	db.dir = ""
	db.txDriver = ""
	return nil
}

func (db *TestingDb) Dsn() string {
	return db.dsn
}

func (db *TestingDb) ExecuteSQLFile(filename string) error {
	if db.dsn == "" {
		return fmt.Errorf("temporary db is not open")
	}
	query, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("executing sql failed: %w", err)
	}

	realDb, err := sql.Open("postgres", db.dsn)
	if err != nil {
		return fmt.Errorf("executing sql failed: %w", err)
	}
	defer realDb.Close()

	if _, err := realDb.Exec(string(query)); err != nil {
		return fmt.Errorf("executing sql failed: %w", err)
	}

	return nil
}

func (db *TestingDb) registerSingleTxDriver() (string, error) {
	if db.dsn == "" {
		return "", fmt.Errorf("temporary db is not open")
	}

	txDriverIdx++
	driver := fmt.Sprintf("temporarydbTx%v", txDriverIdx)
	txdb.Register(driver, "postgres", db.dsn)
	return driver, nil
}

func (db *TestingDb) TxDriver() string {
	if len(db.txDriver) == 0 {
		db.txDriver, _ = db.registerSingleTxDriver()
	}

	return db.txDriver
}

func (db *TestingDb) GetTxDsn() string {
	txDsnIdx++
	return fmt.Sprintf("txCon_%v", txDsnIdx)
}

var searchPaths = [...]string{
	"/usr/local/pgsql",
	"/usr/local",
	"/usr/pgsql-*",
	"/usr/lib/postgresql/*",
	"/opt/local/lib/postgresql*",
}

func findPostgresHome() string {
	for _, path := range searchPaths {
		matches, err := filepath.Glob(path)
		if err != nil {
			return ""
		}

		for _, match := range matches {
			postgresFilename := filepath.Join(match, "bin", "postgres")
			if _, err := os.Stat(postgresFilename); os.IsNotExist(err) {
				continue
			}
			return match
		}
	}

	return ""
}
