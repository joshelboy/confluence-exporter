package db

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/marcboeker/go-duckdb"
)

type Page struct {
	Title string
	Body  string
	Link  string
	UID   string
}

// InitDB initializes the DuckDB database and creates the pages table if it doesn't exist
func InitDB(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("duckdb", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}

	// Create table if it doesn't exist
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS pages (
			title VARCHAR,
			body VARCHAR,
			link VARCHAR,
			uid VARCHAR
		)
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to create table: %v", err)
	}

	return db, nil
}

// InsertPage inserts a page into the database
func InsertPage(db *sql.DB, page Page) error {
	_, err := db.Exec(`
		INSERT INTO pages (title, body, link, uid)
		VALUES (?, ?, ?, ?)
	`, page.Title, page.Body, page.Link, page.UID)

	if err != nil {
		return fmt.Errorf("failed to insert page: %v", err)
	}

	return nil
}

// CloseDB closes the database connection
func CloseDB(db *sql.DB) {
	if db != nil {
		if err := db.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}
} 