package db

import (
	"database/sql"
	"log"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

const (
	dbDriver = "postgres"
	dbSource = "postgresql://root:password@localhost:5432/simple_bank?sslmode=disable"
)

var testQueries *Queries

// Will be run before or after any other test functions.
func TestMain(m *testing.M) {
	conn, err := sql.Open(dbDriver, dbSource)
	if err != nil {
		log.Fatalf("Cannot connect to db: %v", err)
	}

	testQueries = New(conn)

	errCode := m.Run() // Run other test functions
	os.Exit(errCode)
}
