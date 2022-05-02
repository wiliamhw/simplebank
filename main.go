package main

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
	"github.com/wiliamhw/simplebank/api"
	db "github.com/wiliamhw/simplebank/db/sqlc"
	"github.com/wiliamhw/simplebank/util"
)

func main() {
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatalf("cannot load config: %v", err)
	}

	conn, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatalf("cannot connect to db: %v", err)
	}

	store := db.NewStore(conn)
	server, err := api.NewServer(config, store)
	if err != nil {
		log.Fatalf("cannot create server: %v", err)
	}

	if err := server.Start(config.ServerAddress); err != nil {
		log.Fatalf("cannot start server: %v", err)
	}
}
