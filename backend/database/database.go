package database

import (
	"database/sql"

	"github.com/Junior-Green/heartbeats/logger"
	"github.com/Junior-Green/heartbeats/server"
	_ "github.com/mattn/go-sqlite3"
)

type sqliteDatabase struct {
	db *sql.DB
}

func (db *sqliteDatabase) GetAllServers() (server.Server, error) {
	return server.Server{}, nil
}

func (s *sqliteDatabase) Close() {
	s.db.Close()
}

// NewDatabase creates a new sqliteDatabase instance by opening a connection to the SQLite database
// specified by dbPath. If there is an error opening or pinging the database, the function logs the
// error and exits the program.
//
// Parameters:
//   - dbPath: The file path to the SQLite database.
//
// Returns:
//   - A pointer to an sqliteDatabase instance.
func NewDatabase(dbPath string) (*sqliteDatabase, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		logger.Printf("Error opening database: %v", err)
		return nil, err
	}

	if err = db.Ping(); err != nil {
		logger.Printf("Error establishing connection to database: %v", err)
		return nil, err
	}

	return &sqliteDatabase{db: db}, nil
}
