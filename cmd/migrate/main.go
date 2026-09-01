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
	dir := os.Getenv("GOOSE_MIGRATION_DIR")
	if dir == "" {
		dir = "./migrations"
	}

	command := "up"
	if len(os.Args) > 1 {
		command = os.Args[1]
	}

	// "create" only writes a new migration file, no DB connection needed.
	if command == "create" {
		if err := goose.Run(command, nil, dir, os.Args[2:]...); err != nil {
			log.Fatalf("goose %s: %v", command, err)
		}
		return
	}

	dbstring := os.Getenv("GOOSE_DBSTRING")
	if dbstring == "" {
		log.Fatal("GOOSE_DBSTRING is not set")
	}

	db, err := sql.Open("pgx", dbstring)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	if err := goose.Run(command, db, dir, os.Args[2:]...); err != nil {
		log.Fatalf("goose %s: %v", command, err)
	}
}
