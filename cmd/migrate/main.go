package main

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

// usage: go run ./cmd/migrate [up|down|status|create ...] (defaults to "up")
func main() {
	dbstring := os.Getenv("GOOSE_DBSTRING")
	if dbstring == "" {
		log.Fatal("GOOSE_DBSTRING is not set")
	}
	dir := os.Getenv("GOOSE_MIGRATION_DIR")
	if dir == "" {
		dir = "./migrations"
	}

	db, err := sql.Open("pgx", dbstring)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	command := "up"
	if len(os.Args) > 1 {
		command = os.Args[1]
	}

	if err := goose.Run(command, db, dir, os.Args[2:]...); err != nil {
		log.Fatalf("goose %s: %v", command, err)
	}
}
