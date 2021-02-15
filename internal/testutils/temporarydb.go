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

type TemporaryDb struct {
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
		os.Setenv("POSTGRES_HOME", "/usr/lib/postgresql/13")
	}
}

func NewTemporaryDb() (*TemporaryDb, error) {
	var tmpDb TemporaryDb
	if err := tmpDb.open(); err != nil {
		return nil, err
	}
	return &tmpDb, nil
}

func (db *TemporaryDb) open() error {
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

func (db *TemporaryDb) Close() error {
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

func (db *TemporaryDb) Dsn() string {
	return db.dsn
}

func (db *TemporaryDb) ExecuteSQLFile(filename string) error {
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

func (db *TemporaryDb) registerSingleTxDriver() (string, error) {
	if db.dsn == "" {
		return "", fmt.Errorf("temporary db is not open")
	}

	txDriverIdx++
	driver := fmt.Sprintf("temporarydbTx%v", txDriverIdx)
	txdb.Register(driver, "postgres", db.dsn)
	return driver, nil
}

func (db *TemporaryDb) TxDriver() string {
	if len(db.txDriver) == 0 {
		db.txDriver, _ = db.registerSingleTxDriver()
	}

	return db.txDriver
}

func (db *TemporaryDb) GetTxDsn() string {
	txDsnIdx++
	return fmt.Sprintf("txCon_%v", txDsnIdx)
}
