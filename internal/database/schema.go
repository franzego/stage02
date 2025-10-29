package database

import (
	"database/sql"
	_ "embed"
	"log"
)

//go:embed schema.sql
var schemaSQL string

func RunMigrations(db *sql.DB) error {
	log.Println("Running database migrations...")

	_, err := db.Exec(schemaSQL)
	if err != nil {
		return err
	}

	log.Println("✓ Migrations completed successfully")
	return nil
}
