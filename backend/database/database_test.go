package database

import (
	"os"
	"testing"

	"github.com/Junior-Green/heartbeats/server"
	"github.com/stretchr/testify/assert"
)

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

func TestGetAllServers(t *testing.T) {
	// Create a temporary file to act as the SQLite database
	tempFile, err := os.CreateTemp("", "testdb_*.sqlite")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	// Create a new database instance
	db, err := NewDatabase(tempFile.Name())
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Insert test data
	_, err = db.db.Exec(`INSERT INTO Server (id, host, online, favorite) VALUES 
		('1', 'host1', 1, 0),
		('2', 'host2', 0, 1)`)
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	// Test GetAllServers method
	servers, err := db.GetAllServers()
	if err != nil {
		t.Fatalf("Expected GetAllServers to succeed, got error: %v", err)
	}

	// Verify the results
	expectedServers := []server.Server{
		{Id: "1", Host: "host1", Online: true, Favorite: false},
		{Id: "2", Host: "host2", Online: false, Favorite: true},
	}

	if len(servers) != len(expectedServers) {
		t.Fatalf("Expected %d servers, got %d", len(expectedServers), len(servers))
	}

	for i, server := range servers {
		if server != expectedServers[i] {
			t.Errorf("Expected server %v, got %v", expectedServers[i], server)
		}

		assert.Equal(t, server.Id, expectedServers[i].Id)
		assert.Equal(t, server.Host, expectedServers[i].Host)
		assert.Equal(t, server.Favorite, expectedServers[i].Favorite)
		assert.Equal(t, server.Online, expectedServers[i].Online)
	}
}
func TestNewDatabase(t *testing.T) {
	// Create a temporary file to act as the SQLite database
	tempFile, err := os.CreateTemp("", "testdb_*.sqlite")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	// Test NewDatabase with a valid DSN
	db, err := NewDatabase(tempFile.Name())
	if err != nil {
		t.Fatalf("Expected NewDatabase to succeed, got error: %v", err)
	}
	defer db.Close()

	// Ensure the database connection is established
	if err := db.db.Ping(); err != nil {
		t.Fatalf("Expected database connection to be established, got error: %v", err)
	}

	// Test NewDatabase with an invalid DSN
	_, err = NewDatabase("/invalid/path/to/db.sqlite")
	if err == nil {
		t.Fatal("Expected NewDatabase to fail with invalid DSN, got no error")
	}
}
