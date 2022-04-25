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
		log.Fatalf("Cannot load config: %v", err)
	}

	conn, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatalf("Cannot connect to db: %v", err)
	}

	store := db.NewStore(conn)
	server := api.NewServer(store)

	if err := server.Start(config.ServerAddress); err != nil {
		log.Fatalf("Cannot start server: %v", err)
	}
}
