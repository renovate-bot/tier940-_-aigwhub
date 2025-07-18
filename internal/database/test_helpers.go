package database

import (
	"database/sql"
	"os"
)

// InitTestDB creates an in-memory SQLite database for testing
func InitTestDB() (*sql.DB, error) {
	// Use in-memory database for tests with foreign key support
	db, err := sql.Open("sqlite3", ":memory:?_foreign_keys=on")
	if err != nil {
		return nil, err
	}

	// Enable foreign key constraints
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		db.Close()
		return nil, err
	}

	// Create tables
	if err := createTables(db); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

// InitTestDBWithFile creates a file-based SQLite database for testing
func InitTestDBWithFile(dbPath string) (*sql.DB, error) {
	// Remove existing test database if it exists
	os.Remove(dbPath)

	return InitSQLite(dbPath)
}