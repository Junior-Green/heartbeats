package database

import (
	"database/sql"

	"github.com/Junior-Green/heartbeats/logger"
	"github.com/Junior-Green/heartbeats/server"
	"github.com/Junior-Green/heartbeats/server/ping"
	_ "github.com/mattn/go-sqlite3"
)

type ErrNotFound struct{}

func (e ErrNotFound) Error() string {
	return "Record not found"
}

type ErrDuplicateRow struct{}

func (e ErrDuplicateRow) Error() string {
	return "Record already exists"
}

type sqliteDatabase struct {
	db *sql.DB
}

// GetAllServers retrieves all server records from the database.
// It returns a slice of server.Server and an error if any occurs during the query execution or row scanning.
// The function queries the "Server" table and scans each row into a server.Server struct.
// If an error occurs during the query or scanning process, it logs the error and returns it.
func (db *sqliteDatabase) GetAllServers() ([]server.Server, error) {
	rows, err := db.db.Query("SELECT * FROM Server")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	servers := make([]server.Server, 0)
	for rows.Next() {
		server := server.Server{}

		if err := rows.Scan(&server.Id, &server.Host, &server.Online, &server.Favorite); err != nil {
			logger.Debugf("Error scanning rows: %v", err)
			return nil, err
		}

		servers = append(servers, server)
	}

	if err = rows.Err(); err != nil {
		logger.Debugf("Error scanning rows: %v", err)
		return nil, err
	}

	return servers, nil
}

func (db *sqliteDatabase) GetMetricsByHost(host string) (server.Metrics, error) {
	panic("not implemented")
}

func (db *sqliteDatabase) GetServerByHost(host string) (server.Server, error) {
	panic("not implemented")
}

func (db *sqliteDatabase) AddServer(server server.Server) error {
	panic("not implemented")
}

func (db *sqliteDatabase) DeleteServerByHost(host string) error {
	panic("not implemented")
}

func (db *sqliteDatabase) UpdateOnlineStatusByHost(host string, status bool) error {
	panic("not implemented")
}

func (db *sqliteDatabase) UpdateFavoriteByHost(host string, favorite bool) error {
	panic("not implemented")
}

func (db *sqliteDatabase) AddPingDataByHost(host string, data ping.PingData) error {
	panic("not implemented")
}

// Close closes the database connection.
// It is important to call this method to release any resources held by the database.
func (s *sqliteDatabase) Close() {
	s.db.Close()
}

// init initializes the sqliteDatabase by setting the maximum number of open connections
// to 1 and creating the necessary tables (Server, Marker, and Metric) if they do not
// already exist. It returns an error if there is an issue executing the SQL statements.
func (db *sqliteDatabase) init() error {
	db.db.SetMaxOpenConns(1)

	sqlStmt := `
	CREATE TABLE IF NOT EXISTS Server (
		id CHAR(36) PRIMARY KEY NOT NULL,
		host TEXT UNIQUE NOT NULL, 
		online BOOLEAN NOT NULL, 
		favorite BOOLEAN NOT NULL
	);

	CREATE TABLE IF NOT EXISTS Marker (
		id CHAR(36) PRIMARY KEY,
		latency INTEGER,
		packet_loss REAL,
		throughput REAL,
		dns_resolved INTEGER,
		status_code INTEGER
	);

	CREATE TABLE IF NOT EXISTS Metric (
		id CHAR(36) PRIMARY KEY,
		server CHAR(36) NOT NULL, 
		time TEXT NOT NULL,
		marker CHAR(36) NOT NULL,
		FOREIGN KEY (server) REFERENCES Server(id) ON DELETE CASCADE, 
		FOREIGN KEY (marker) REFERENCES Marker(id)
	);`

	if _, err := db.db.Exec(sqlStmt); err != nil {
		logger.Debugf("Error executing query: %v", err)
		return err
	}

	return nil
}

// NewDatabase creates a new sqliteDatabase instance and establishes a connection to the SQLite database
// specified by the given Data Source Name (DSN). It returns a pointer to the sqliteDatabase and an error
// if any issues occur during the process.
//
// Parameters:
//   - dsn: A string representing the Data Source Name for the SQLite database.
//
// Returns:
//   - *sqliteDatabase: A pointer to the initialized sqliteDatabase instance.
//   - error: An error if there is an issue opening the database, establishing a connection, or initializing the database.
func NewDatabase(dsn string) (*sqliteDatabase, error) {
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		logger.Printf("Error opening database: %v", err)
		return nil, err
	}

	if err = db.Ping(); err != nil {
		logger.Printf("Error establishing connection to database: %v", err)
		return nil, err
	}

	sqlDb := &sqliteDatabase{db: db}
	if err = sqlDb.init(); err != nil {
		logger.Printf("Error initializing database: %v", err)
		return nil, err
	}
	return sqlDb, nil
}
