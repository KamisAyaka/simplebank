package db

import (
	"database/sql"
	"log"
	"os"
	"testing"

	"github.com/KamisAyaka/simplebank/util"
	_ "github.com/lib/pq"
)

var testQueries *Queries
var testDB *sql.DB

func TestMain(m *testing.M) {
	config, err := util.LoadConfig("../..")
	if err != nil {
		log.Fatal("cannot load config:", err)
	}
	testDB, err = sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}
	testDB.SetMaxOpenConns(20)
	testDB.SetMaxIdleConns(20)
	if err = testDB.Ping(); err != nil {
		log.Fatal("cannot ping db:", err)
	}
	testQueries = New(testDB)
	code := m.Run()
	_ = testDB.Close()
	os.Exit(code)
}
