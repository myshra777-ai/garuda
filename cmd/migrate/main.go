package main

import (
	"log"
	"os"

	"github.com/myshra777-ai/garuda/internal/store"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://test:test@localhost:5433/garuda_test?sslmode=disable"
	}

	log.Println("Running migrations...")
	if err := store.Migrate(dbURL, "migrations"); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}
	log.Println("Migrations completed successfully.")
}
