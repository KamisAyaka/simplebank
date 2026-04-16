package main

import (
	"database/sql"
	"log"

	"github.com/KamisAyaka/simplebank/api"
	db "github.com/KamisAyaka/simplebank/db/sqlc"
	"github.com/KamisAyaka/simplebank/util"
	_ "github.com/lib/pq"
)

func main() {
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal("cannot load config:", err)
	}
	conn, err := sql.Open(config.DBDriver, config.DBSource)
	conn.SetMaxOpenConns(20)
	conn.SetMaxIdleConns(20)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	store := db.NewStore(conn)
	server := api.NewServer(store)

	err = server.Start(config.ServerAddress)
	if err != nil {
		log.Fatal("cannot start server:", err)
	}
}
