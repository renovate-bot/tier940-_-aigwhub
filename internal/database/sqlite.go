package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

func InitSQLite(dbPath string) (*sql.DB, error) {
	// Ensure directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// Open database connection
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Create tables
	if err := createTables(db); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return db, nil
}

func createTables(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS chats (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		provider TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS messages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		chat_id INTEGER NOT NULL,
		role TEXT NOT NULL CHECK(role IN ('user', 'assistant', 'system')),
		content TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (chat_id) REFERENCES chats(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS sessions (
		id TEXT PRIMARY KEY,
		chat_id INTEGER,
		data TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		expires_at DATETIME,
		FOREIGN KEY (chat_id) REFERENCES chats(id) ON DELETE SET NULL
	);

	CREATE INDEX IF NOT EXISTS idx_messages_chat_id ON messages(chat_id);
	CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions(expires_at);
	`

	if _, err := db.Exec(schema); err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	return nil
}