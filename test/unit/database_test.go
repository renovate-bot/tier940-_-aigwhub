package unit

import (
	"os"
	"path/filepath"
	"testing"

	"ai-gateway-hub/internal/database"
	"ai-gateway-hub/internal/utils"
)

func TestSQLiteDatabase(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "sqlite_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Setup path manager
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)
	utils.InitPathManager()

	t.Run("InitSQLite_CreatesDatabase", func(t *testing.T) {
		dbPath := "./test_data/test.db"
		
		db, err := database.InitSQLite(dbPath)
		if err != nil {
			t.Fatalf("InitSQLite failed: %v", err)
		}
		defer db.Close()
		
		// Check if database file was created
		expectedPath := filepath.Join(tempDir, dbPath)
		if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
			t.Errorf("Database file was not created at %s", expectedPath)
		}
		
		// Test database connection
		err = db.Ping()
		if err != nil {
			t.Errorf("Database ping failed: %v", err)
		}
	})

	t.Run("InitSQLite_CreatesDirectories", func(t *testing.T) {
		dbPath := "./deep/nested/path/test.db"
		
		db, err := database.InitSQLite(dbPath)
		if err != nil {
			t.Fatalf("InitSQLite failed: %v", err)
		}
		defer db.Close()
		
		// Check if directory was created
		expectedDir := filepath.Join(tempDir, "deep/nested/path")
		if _, err := os.Stat(expectedDir); os.IsNotExist(err) {
			t.Errorf("Database directory was not created at %s", expectedDir)
		}
	})

	t.Run("InitSQLite_CreatesTables", func(t *testing.T) {
		dbPath := "./tables_test.db"
		
		db, err := database.InitSQLite(dbPath)
		if err != nil {
			t.Fatalf("InitSQLite failed: %v", err)
		}
		defer db.Close()
		
		// Check if tables were created
		tables := []string{"chats", "messages", "sessions"}
		for _, table := range tables {
			var name string
			query := "SELECT name FROM sqlite_master WHERE type='table' AND name=?"
			err := db.QueryRow(query, table).Scan(&name)
			if err != nil {
				t.Errorf("Table %s was not created: %v", table, err)
			}
		}
	})

	t.Run("InitSQLite_TablesHaveCorrectSchema", func(t *testing.T) {
		dbPath := "./schema_test.db"
		
		db, err := database.InitSQLite(dbPath)
		if err != nil {
			t.Fatalf("InitSQLite failed: %v", err)
		}
		defer db.Close()
		
		// Test chats table schema
		_, err = db.Exec(`INSERT INTO chats (title, provider, created_at, updated_at) 
						  VALUES ('test', 'claude', datetime('now'), datetime('now'))`)
		if err != nil {
			t.Errorf("Failed to insert into chats table: %v", err)
		}
		
		// Test messages table schema
		_, err = db.Exec(`INSERT INTO messages (chat_id, role, content, created_at) 
						  VALUES (1, 'user', 'test message', datetime('now'))`)
		if err != nil {
			t.Errorf("Failed to insert into messages table: %v", err)
		}
		
		// Test sessions table schema
		_, err = db.Exec(`INSERT INTO sessions (id, data, expires_at) 
						  VALUES ('test-session', '{}', datetime('now'))`)
		if err != nil {
			t.Errorf("Failed to insert into sessions table: %v", err)
		}
	})

	t.Run("InitSQLite_InvalidPath", func(t *testing.T) {
		// Try to create database in a path that can't be created
		dbPath := "/root/cannot_create/test.db"
		
		_, err := database.InitSQLite(dbPath)
		if err == nil {
			t.Error("Expected error for invalid database path, got nil")
		}
	})

	t.Run("InitSQLite_ExistingDatabase", func(t *testing.T) {
		dbPath := "./existing_test.db"
		
		// Create database first time
		db1, err := database.InitSQLite(dbPath)
		if err != nil {
			t.Fatalf("First InitSQLite failed: %v", err)
		}
		
		// Insert test data
		_, err = db1.Exec(`INSERT INTO chats (title, provider, created_at, updated_at) 
						   VALUES ('existing', 'claude', datetime('now'), datetime('now'))`)
		if err != nil {
			t.Fatalf("Failed to insert test data: %v", err)
		}
		db1.Close()
		
		// Open database second time
		db2, err := database.InitSQLite(dbPath)
		if err != nil {
			t.Fatalf("Second InitSQLite failed: %v", err)
		}
		defer db2.Close()
		
		// Check that existing data is still there
		var title string
		err = db2.QueryRow("SELECT title FROM chats WHERE title = 'existing'").Scan(&title)
		if err != nil {
			t.Errorf("Failed to read existing data: %v", err)
		}
		if title != "existing" {
			t.Errorf("Expected title 'existing', got '%s'", title)
		}
	})
}

func TestRedisConnection(t *testing.T) {
	t.Run("InitRedis", func(t *testing.T) {
		// Test with a non-existent Redis server
		client := database.InitRedis("localhost:9999")
		if client == nil {
			t.Fatal("InitRedis returned nil client")
		}
		
		// The client should be created even if connection fails
		// Actual connection errors will occur when operations are performed
		defer client.Close()
	})

	t.Run("InitRedis_DefaultAddress", func(t *testing.T) {
		client := database.InitRedis("localhost:6379")
		if client == nil {
			t.Fatal("InitRedis returned nil client")
		}
		defer client.Close()
		
		// Test basic operations (will fail if Redis not available, but shouldn't panic)
		ctx := utils.NewContext()
		_, err := client.Ping(ctx).Result()
		
		// We don't assert success/failure because Redis might not be running
		// but we test that it doesn't panic
		t.Logf("Redis ping result: %v", err)
	})
}