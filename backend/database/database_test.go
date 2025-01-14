package database

import (
	"os"
	"testing"
)

func TestNewDatabase(t *testing.T) {
	// Create a temporary file to act as the SQLite database
	tempFile, err := os.CreateTemp("", "testdb_*.sqlite")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	// Test successful database creation
	db, _ := NewDatabase(tempFile.Name())
	if db == nil {
		t.Fatal("Expected non-nil database instance")
	}
	db.Close()

	// Test error handling for invalid database path
	invalidPath := "/invalid/path/to/db.sqlite"
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("Expected os.Exit to be called")
		}
	}()
	// Ensure logger output does not interfere with test
	NewDatabase(invalidPath)
}

func TestClose(t *testing.T) {
	// Create a temporary file to act as the SQLite database
	tempFile, err := os.CreateTemp("", "testdb_*.sqlite")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	// Create a new database instance
	db, _ := NewDatabase(tempFile.Name())

	// Test Close method
	db.Close()

	// Ensure the database is closed
	if err := db.db.Ping(); err == nil {
		t.Fatal("Expected error after closing the database, got nil")
	}
}
