package database

import (
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/Junior-Green/heartbeats/server"
	"github.com/Junior-Green/heartbeats/server/ping"
	"github.com/google/uuid"
	"github.com/guregu/null/v5"
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
	expectedServers := []server.Server{
		{Id: uuid.NewString(), Host: "host1", Online: true, Favorite: false},
		{Id: uuid.NewString(), Host: "host2", Online: false, Favorite: true},
	}

	err = db.AddServer(expectedServers[0])
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}
	err = db.AddServer(expectedServers[1])
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	// Test GetAllServers method
	servers, err := db.GetAllServers()
	if err != nil {
		t.Fatalf("Expected GetAllServers to succeed, got error: %v", err)
	}

	// Verify the results
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
func TestAddServer(t *testing.T) {
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

	// Define test cases
	tests := []struct {
		name      string
		server    server.Server
		wantErr   bool
		errorType error
	}{
		{
			name: "Invalid Id",
			server: server.Server{
				Id:       "1",
				Host:     "host1",
				Online:   true,
				Favorite: false,
			},
			wantErr:   true,
			errorType: ErrCheckConstraint{"Server.id"},
		},
		{
			name: "Valid server",
			server: server.Server{
				Id:       "8e6e69dd-7c15-465b-86c8-ea216eb8c7a4",
				Host:     "host2",
				Online:   true,
				Favorite: false,
			},
			wantErr:   false,
			errorType: nil,
		},
		{
			name: "Duplicate server",
			server: server.Server{
				Id:       "5d881406-6cd9-488e-8c04-319e13c6ee6e",
				Host:     "host2",
				Online:   true,
				Favorite: false,
			},
			wantErr:   true,
			errorType: ErrUniqueConstraint{},
		},
		{
			name: "Duplicate key",
			server: server.Server{
				Id:       "8e6e69dd-7c15-465b-86c8-ea216eb8c7a4",
				Host:     "host3",
				Online:   true,
				Favorite: false,
			},
			wantErr:   true,
			errorType: ErrUniqueConstraint{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := db.AddServer(tt.server)
			t.Logf("Error creating server: %v", err)

			if tt.wantErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.errorType)
			} else {
				var expectedId string
				err = db.db.QueryRow("SELECT id FROM Server WHERE id = ?", tt.server.Id).Scan(&expectedId)
				assert.Nil(t, err)
				assert.Equal(t, expectedId, tt.server.Id)
			}
		})
	}
}
func TestDeleteServerByHost(t *testing.T) {
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

	// Insert test data using AddServer function
	err = db.AddServer(server.Server{Id: uuid.NewString(), Host: "host1", Online: true, Favorite: false})
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}
	err = db.AddServer(server.Server{Id: uuid.NewString(), Host: "host2", Online: false, Favorite: true})
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	// Define test cases
	tests := []struct {
		name    string
		host    string
		wantErr bool
		errType error
	}{
		{
			name:    "Delete existing server",
			host:    "host1",
			wantErr: false,
			errType: nil,
		},
		{
			name:    "Delete non-existing server",
			host:    "nonexistent",
			wantErr: true,
			errType: ErrNotFound{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := db.DeleteServerByHost(tt.host)
			if tt.wantErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.errType)
			} else {
				assert.NoError(t, err)
				var count int
				err = db.db.QueryRow("SELECT COUNT(*) FROM Server WHERE host = ?", tt.host).Scan(&count)
				assert.NoError(t, err)
				assert.Equal(t, 0, count)
			}
		})
	}
}
func TestGetServerByHost(t *testing.T) {
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

	// Insert test data using AddServer function
	err = db.AddServer(server.Server{Id: uuid.NewString(), Host: "host1", Online: true, Favorite: false})
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}
	err = db.AddServer(server.Server{Id: uuid.NewString(), Host: "host2", Online: false, Favorite: true})
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	// Define test cases
	tests := []struct {
		name     string
		host     string
		wantErr  bool
		errType  error
		expected server.Server
	}{
		{
			name:     "Get existing server",
			host:     "host1",
			wantErr:  false,
			errType:  nil,
			expected: server.Server{Host: "host1", Online: true, Favorite: false},
		},
		{
			name:     "Get non-existing server",
			host:     "nonexistent",
			wantErr:  true,
			errType:  sql.ErrNoRows,
			expected: server.Server{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server, err := db.GetServerByHost(tt.host)
			if tt.wantErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.errType)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected.Host, server.Host)
				assert.Equal(t, tt.expected.Online, server.Online)
				assert.Equal(t, tt.expected.Favorite, server.Favorite)
			}
		})
	}
}
func TestUpdateOnlineStatusByHost(t *testing.T) {
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

	// Insert test data using AddServer function
	err = db.AddServer(server.Server{Id: uuid.NewString(), Host: "host1", Online: true, Favorite: false})
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}
	err = db.AddServer(server.Server{Id: uuid.NewString(), Host: "host2", Online: false, Favorite: true})
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	// Define test cases
	tests := []struct {
		name    string
		host    string
		online  bool
		wantErr bool
		errType error
	}{
		{
			name:    "Update existing server online status to true",
			host:    "host2",
			online:  true,
			wantErr: false,
			errType: nil,
		},
		{
			name:    "Update existing server online status to false",
			host:    "host1",
			online:  false,
			wantErr: false,
			errType: nil,
		},
		{
			name:    "Update non-existing server online status",
			host:    "nonexistent",
			online:  true,
			wantErr: true,
			errType: ErrNotFound{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := db.UpdateOnlineStatusByHost(tt.host, tt.online)
			if tt.wantErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.errType)
			} else {
				assert.NoError(t, err)
				var online bool
				err = db.db.QueryRow("SELECT online FROM Server WHERE host = ?", tt.host).Scan(&online)
				assert.NoError(t, err)
				assert.Equal(t, tt.online, online)
			}
		})
	}
}
func TestUpdateFavoriteByHost(t *testing.T) {
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

	// Insert test data using AddServer function
	err = db.AddServer(server.Server{Id: uuid.NewString(), Host: "host1", Online: true, Favorite: false})
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}
	err = db.AddServer(server.Server{Id: uuid.NewString(), Host: "host2", Online: false, Favorite: true})
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	// Define test cases
	tests := []struct {
		name     string
		host     string
		favorite bool
		wantErr  bool
		errType  error
	}{
		{
			name:     "Update existing server favorite status to true",
			host:     "host1",
			favorite: true,
			wantErr:  false,
			errType:  nil,
		},
		{
			name:     "Update existing server favorite status to false",
			host:     "host2",
			favorite: false,
			wantErr:  false,
			errType:  nil,
		},
		{
			name:     "Update non-existing server favorite status",
			host:     "nonexistent",
			favorite: true,
			wantErr:  true,
			errType:  ErrNotFound{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := db.UpdateFavoriteByHost(tt.host, tt.favorite)
			if tt.wantErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.errType)
			} else {
				assert.NoError(t, err)
				var favorite bool
				err = db.db.QueryRow("SELECT favorite FROM Server WHERE host = ?", tt.host).Scan(&favorite)
				assert.NoError(t, err)
				assert.Equal(t, tt.favorite, favorite)
			}
		})
	}
}
func TestAddPingMetricByHost(t *testing.T) {
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

	// Insert test server data
	server := server.Server{Id: uuid.NewString(), Host: "host1", Online: true, Favorite: false}
	err = db.AddServer(server)
	if err != nil {
		t.Fatalf("Failed to insert test server data: %v", err)
	}

	// Define test cases
	tests := []struct {
		name    string
		host    string
		data    ping.PingData
		wantErr bool
		errType error
	}{
		{
			name: "Add new ping metric",
			host: "host1",
			data: ping.PingData{
				Latency:        null.IntFrom(100),
				PacketLoss:     null.FloatFrom(0.0),
				Throughput:     null.FloatFrom(100.0),
				DnsResolveTime: null.IntFrom(50),
				Rtt:            null.IntFrom(200),
				StatusCode:     null.IntFrom(200),
				Date:           time.Now(),
			},
			wantErr: false,
			errType: nil,
		},
		{
			name: "Add ping metric for non-existing server",
			host: "nonexistent",
			data: ping.PingData{
				Latency:        null.IntFrom(100),
				PacketLoss:     null.FloatFrom(0.0),
				Throughput:     null.FloatFrom(100.0),
				DnsResolveTime: null.IntFrom(50),
				Rtt:            null.IntFrom(200),
				StatusCode:     null.IntFrom(200),
				Date:           time.Now(),
			},
			wantErr: true,
			errType: sql.ErrNoRows,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := db.AddPingMetricByHost(tt.host, tt.data)
			if tt.wantErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.errType)
			} else {
				assert.NoError(t, err)

				// Verify the metric was added
				var count int
				err = db.db.QueryRow("SELECT COUNT(*) FROM Metric WHERE server_id = ?", server.Id).Scan(&count)
				assert.NoError(t, err)
				assert.Equal(t, 1, count)
			}
		})
	}
}
func TestAddMetric(t *testing.T) {
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

	// Insert test server data
	server := server.Server{Id: uuid.NewString(), Host: "host1", Online: true, Favorite: false}
	err = db.AddServer(server)
	if err != nil {
		t.Fatalf("Failed to insert test server data: %v", err)
	}

	// Insert test marker data
	markerId := uuid.NewString()
	query := `INSERT INTO Marker VALUES (?, ?, ?, ?, ?, ?, ?)`
	_, err = db.db.Exec(query, markerId, 100, 0.0, 100.0, 50, 200, 200)
	if err != nil {
		t.Fatalf("Failed to insert test marker data: %v", err)
	}

	// Define test cases
	tests := []struct {
		name      string
		timestamp time.Time
		serverId  string
		markerId  string
		wantErr   bool
		errType   error
	}{
		{
			name:      "Add valid metric",
			timestamp: time.Now(),
			serverId:  server.Id,
			markerId:  markerId,
			wantErr:   false,
			errType:   nil,
		},
		{
			name:      "Add metric with invalid serverId",
			timestamp: time.Now(),
			serverId:  "invalid-server-id",
			markerId:  markerId,
			wantErr:   true,
			errType:   fmt.Errorf("Error adding marker: %v", sql.ErrNoRows),
		},
		{
			name:      "Add metric with invalid markerId",
			timestamp: time.Now(),
			serverId:  server.Id,
			markerId:  "invalid-marker-id",
			wantErr:   true,
			errType:   fmt.Errorf("Error adding marker: %v", sql.ErrNoRows),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx, err := db.db.Begin()
			if err != nil {
				t.Fatalf("Failed to begin transaction: %v", err)
			}
			defer tx.Rollback()

			err = db.addMetric(tx, tt.timestamp, tt.serverId, tt.markerId)
			if tt.wantErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.errType)
			} else {
				assert.NoError(t, err)

				// Verify the metric was added
				var count int
				err = db.db.QueryRow("SELECT COUNT(*) FROM Metric WHERE server_id = ? AND marker_id = ?", tt.serverId, tt.markerId).Scan(&count)
				assert.NoError(t, err)
				assert.Equal(t, 1, count)
			}
		})
	}
}
