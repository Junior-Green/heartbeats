package database

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/Junior-Green/heartbeats/logger"
	"github.com/Junior-Green/heartbeats/server"
	"github.com/Junior-Green/heartbeats/server/ping"
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

type ErrNotFound struct {
	Err error
}

func (e ErrNotFound) Error() string {
	return "Record not found"
}

type ErrDuplicateRow struct {
	Err error
}

func (e ErrDuplicateRow) Error() string {
	return "Record already exists"
}

type ErrCheckConstraint struct {
	Err error
}

func (e ErrCheckConstraint) Error() string {
	return fmt.Sprintf("Check constraint violation: %v", e.Err)
}

type ErrUniqueConstraint struct {
	Err error
}

func (e ErrUniqueConstraint) Error() string {
	return "Row already exists"
}

type ErrForeignConstraint struct {
	Err error
}

func (e ErrForeignConstraint) Error() string {
	return fmt.Sprintf("Foreign constraint violation: %v: ", e.Err)
}

type SqliteDatabase struct {
	db *sql.DB
}

// GetAllServers retrieves all server records from the database.
// It returns a slice of server.Server and an error if any occurs during the query execution or row scanning.
// The function queries the "Server" table and scans each row into a server.Server struct.
// If an error occurs during the query or scanning process, it logs the error and returns it.
func (db *SqliteDatabase) GetAllServers() ([]server.Server, error) {
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

// GetMetricsByHost retrieves the metrics for a given host from the database.
// It first fetches the server information using the host, then queries the metrics
// associated with the server. The metrics include datetime, latency, packet loss,
// throughput, DNS resolved time, RTT, and status code.
//
// Parameters:
//   - host: The hostname for which metrics are to be retrieved.
//
// Returns:
//   - server.Metrics: A struct containing the metrics data for the specified host.
//   - error: An error object if any error occurs during the process.
//
// The function performs the following steps:
//  1. Fetches the server information using the provided host.
//  2. Executes a SQL query to retrieve the metrics associated with the server.
//  3. Scans the query results and populates the metrics struct.
//  4. Returns the populated metrics struct or an error if any occurs.
func (db *SqliteDatabase) GetMetricsByHost(host string) (server.Metrics, error) {
	s, err := db.GetServerByHost(host)
	if err != nil {
		return server.Metrics{}, err
	}

	query := `
	SELECT datetime, latency, packet_loss, throughput, dns_resolved, rtt, status_code
	FROM Metric 
	INNER JOIN Marker ON Metric.marker_id=Marker.id
	WHERE server_id=?;
	`
	rows, err := db.db.Query(query, s.Id)
	if err != nil {
		return server.Metrics{}, err
	}
	defer rows.Close()

	metrics := server.Metrics{}
	for rows.Next() {
		var (
			dateStr  string
			pingData ping.PingData
		)

		err := rows.Scan(&dateStr, &pingData.Latency, &pingData.PacketLoss, &pingData.Throughput, &pingData.DnsResolveTime, &pingData.Rtt, &pingData.StatusCode)
		if err != nil {
			logger.Debugf("Error scanning rows: %v", err)
			return server.Metrics{}, err
		}

		pingData.Date, err = time.Parse(time.DateTime, dateStr)
		if err != nil {
			return server.Metrics{}, err
		}

		metrics.DnsResolved = append(metrics.DnsResolved, server.DnsResolvedMarker{Date: pingData.Date, DnsResolved: pingData.DnsResolveTime})
		metrics.Latency = append(metrics.Latency, server.LatencyMarker{Date: pingData.Date, Latency: pingData.Latency})
		metrics.PacketLoss = append(metrics.PacketLoss, server.PacketLossMarker{Date: pingData.Date, PacketLoss: pingData.PacketLoss})
		metrics.Rtt = append(metrics.Rtt, server.RttMarker{Date: pingData.Date, Rtt: pingData.Rtt})
		metrics.StatusCode = append(metrics.StatusCode, server.StatusCodeMarker{Date: pingData.Date, StatusCode: pingData.StatusCode})
		metrics.Throughput = append(metrics.Throughput, server.ThroughputMarker{Date: pingData.Date, Throughput: pingData.Throughput})
	}

	if err = rows.Err(); err != nil {
		logger.Debugf("Error scanning rows: %v", err)
		return server.Metrics{}, err
	}

	return metrics, nil
}

// GetServerByHost retrieves a server from the database by its host.
// It executes a SQL query to find the server with the specified host.
// If the server is found, it scans the result into a server.Server struct and returns it.
// If the server is not found, it returns an ErrNotFound error.
// If there is an error scanning the result, it returns an error with a descriptive message.
//
// Parameters:
//
//	host - the host of the server to retrieve
//
// Returns:
//
//	server.Server - the server with the specified host
//	error - an error if the server is not found or if there is an issue scanning the result
func (db *SqliteDatabase) GetServerByHost(host string) (server.Server, error) {
	query := "SELECT * FROM Server WHERE host = ?"

	row := db.db.QueryRow(query, host)
	if row == nil {
		panic("unexpected behavior from package database/sql")
	}

	s := server.Server{}
	if err := row.Scan(&s.Id, &s.Host, &s.Online, &s.Favorite); err != nil {
		return server.Server{}, err
	}

	return s, nil
}

// AddServer inserts a new server record into the Server table in the SQLite database.
// It takes a server.Server object as an argument and returns an error if the insertion fails.
//
// Parameters:
//   - server: The server.Server object containing the server details to be added.
//
// Returns:
//   - error: An error object if the insertion fails, otherwise nil.
func (db *SqliteDatabase) AddServer(server server.Server) error {
	query := "INSERT INTO Server VALUES (?,?,?,?)"

	_, err := db.db.Exec(query, server.Id, server.Host, server.Online, server.Favorite)
	if err != nil {
		logger.Debugf("Error adding server: %v", err)
		msg := err.Error()

		switch strings.Fields(msg)[0] {
		case "CHECK":
			return ErrCheckConstraint{err}
		case "UNIQUE":
			return ErrUniqueConstraint{err}
		default:
			return fmt.Errorf("Error adding server: %s", msg)
		}
	}

	return nil
}

// DeleteServerByHost deletes a server entry from the database based on the provided host.
// It executes a DELETE SQL statement to remove the server with the matching host.
// If the deletion is successful but no rows are affected, it returns an ErrNotFound error.
// If there is an error during the execution of the SQL statement, it logs the error and returns it.
//
// Parameters:
//   - host: The host of the server to be deleted.
//
// Returns:
//   - error: An error if the deletion fails or no rows are affected, otherwise nil.
func (db *SqliteDatabase) DeleteServerByHost(host string) error {
	query := "DELETE FROM Server WHERE host = ?"

	res, err := db.db.Exec(query, host)
	if err != nil {
		logger.Debugf("Error deleting server: %v", err)
		return err
	}
	if num, err := res.RowsAffected(); err == nil && num == 0 {
		return ErrNotFound{err}
	}

	return nil
}

// UpdateOnlineStatusByHost updates the online status of a server identified by its host.
// It takes the host string and a boolean indicating the online status.
// If the update is successful, it returns nil. If the server is not found, it returns an ErrNotFound.
// If there is an error during the update, it logs the error and returns it.
func (db *SqliteDatabase) UpdateOnlineStatusByHost(host string, online bool) error {
	query := `
	UPDATE Server 
	SET online = ?
	WHERE host = ?`

	res, err := db.db.Exec(query, online, host)
	if err != nil {
		logger.Debugf("Error updating server online status: %v", err)
		return err
	}
	if num, err := res.RowsAffected(); err == nil && num == 0 {
		return ErrNotFound{err}
	}

	return nil
}

// UpdateFavoriteByHost updates the favorite status of a server identified by its host.
// It sets the favorite field to the provided boolean value.
// If the server with the specified host is not found, it returns an ErrNotFound error.
//
// Parameters:
//   - host: The host of the server to update.
//   - favorite: The new favorite status to set.
//
// Returns:
//   - error: An error if the update operation fails or if the server is not found.
func (db *SqliteDatabase) UpdateFavoriteByHost(host string, favorite bool) error {
	query := `
	UPDATE Server 
	SET favorite = ?
	WHERE host = ?`

	res, err := db.db.Exec(query, favorite, host)
	if err != nil {
		logger.Debugf("Error updating server favorite value: %v", err)
		return err
	}
	if num, err := res.RowsAffected(); err == nil && num == 0 {
		return ErrNotFound{err}
	}

	return nil
}

// AddPingMetricByHost adds a ping metric for a given host to the database.
// It retrieves the server by host, gets or creates a marker ID, and adds the metric within a transaction.
//
// Parameters:
//   - host: The hostname for which the ping metric is being added.
//   - data: The ping data to be added.
//
// Returns:
//   - error: An error if any issues occur during the process, otherwise nil.
func (db *SqliteDatabase) AddPingMetricByHost(host string, data ping.PingData) error {
	server, err := db.GetServerByHost(host)
	if err != nil {
		return ErrNotFound{err}
	}

	markerId, err := db.getMarkerId(data)
	if reflect.TypeOf(err) != reflect.TypeOf(ErrNotFound{}) {
		logger.Debugf("Error retrieving marker id: %v", err)
		return err
	}

	tx, err := db.db.Begin()
	if err != nil {
		return fmt.Errorf("Error starting database transaction: %v", err)
	}
	defer tx.Commit()

	if markerId == "" {
		newId, err := db.addMarker(tx, data)
		if err != nil {
			tx.Rollback()
			return err
		}
		markerId = newId
	}

	if err := db.addMetric(tx, data.Date, server.Id, markerId); err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

// addMetric inserts a new metric record into the Metric table within a transaction.
// It takes the following parameters:
// - tx: the transaction within which the metric will be added.
// - t: the timestamp of the metric.
// - serverId: the ID of the server associated with the metric.
// - markerId: the ID of the marker associated with the metric.
// It returns an error if the insertion fails due to a constraint violation or any other reason.
func (db *SqliteDatabase) addMetric(tx *sql.Tx, t time.Time, serverId, markerId string) error {
	query := `INSERT INTO Metric VALUES (?, ?, ?, ?)`

	_, err := tx.Exec(query, uuid.NewString(), t.Format(time.DateTime), serverId, markerId)
	if err != nil {
		switch strings.Fields(err.Error())[0] {
		case "FOREIGN":
			return ErrForeignConstraint{err}
		default:
			return fmt.Errorf("Error adding marker: %v", err)
		}
	}
	return nil
}

// addMarker inserts a new marker into the Marker table within a transaction.
// It takes a sql.Tx object and ping.PingData as input and returns the generated marker ID or an error.
//
// Parameters:
// - tx: A sql.Tx object representing the current transaction.
// - data: A ping.PingData object containing the marker data to be inserted.
//
// Returns:
// - string: The generated marker ID.
// - error: An error if the insertion fails, including specific errors for check and unique constraints.
func (db *SqliteDatabase) addMarker(tx *sql.Tx, data ping.PingData) (string, error) {
	query := `INSERT INTO Marker VALUES (?, ?, ?, ?, ?, ?, ?)`

	markerId := uuid.NewString()
	_, err := tx.Exec(query, markerId, data.Latency, data.PacketLoss, data.Throughput, data.DnsResolveTime, data.Rtt, data.StatusCode)
	if err != nil {
		msg := err.Error()

		switch strings.Fields(msg)[0] {
		case "CHECK":
			return "", ErrCheckConstraint{err}
		case "UNIQUE":
			return "", ErrUniqueConstraint{err}
		default:
			return "", fmt.Errorf("Error adding marker: %s", msg)
		}
	}

	return markerId, nil
}

// getMarkerId retrieves the marker ID from the database based on the provided ping data.
// It constructs and executes a SQL query to find a matching record in the Marker table
// using the latency, packet loss, throughput, DNS resolve time, and status code from the ping data.
// If a matching record is found, the marker ID is returned. If no matching record is found or an error occurs,
// an empty string and the error are returned.
//
// Parameters:
//   - data: ping.PingData containing the latency, packet loss, throughput, DNS resolve time, and status code.
//
// Returns:
//   - string: The marker ID if a matching record is found, otherwise an empty string.
//   - error: An error if the query execution or row scanning fails, otherwise nil.
func (db *SqliteDatabase) getMarkerId(data ping.PingData) (string, error) {
	query := `
	SELECT id FROM Marker WHERE
	latency = ? AND packet_loss = ? AND throughput = ? AND dns_resolved = ? AND status_code = ?;
	`

	row := db.db.QueryRow(query, data.Latency, data.PacketLoss, data.Throughput, data.DnsResolveTime, data.StatusCode)
	if row == nil {
		panic("unexpected behavior from package database/sql")
	}

	var marker_id string
	if err := row.Scan(&marker_id); err != nil {
		return "", ErrNotFound{err}
	}

	return marker_id, nil
}

// Close closes the database connection.
// It is important to call this method to release any resources held by the database.
func (s *SqliteDatabase) Close() {
	s.db.Close()
}

// init initializes the SQLite database by setting the maximum number of open connections
// to 1 and creating the necessary tables (Server, Marker, and Metric) if they do not
// already exist. It returns an error if the table creation query fails.
//
// The Server table stores information about servers, including their ID, host, online
// status, and whether they are marked as favorite.
//
// The Marker table stores various metrics related to server performance, such as latency,
// packet loss, throughput, DNS resolution time, round-trip time (RTT), and status code.
//
// The Metric table links servers and markers, recording the datetime of the metric,
// the server ID, and the marker ID. It also enforces foreign key constraints to ensure
// referential integrity.
func (db *SqliteDatabase) init() error {
	db.db.SetMaxOpenConns(1)

	query := `
	CREATE TABLE IF NOT EXISTS Server (
		id CHAR(36) PRIMARY KEY NOT NULL CHECK (LENGTH(id) = 36),
		host TEXT UNIQUE NOT NULL, 
		online BOOLEAN NOT NULL, 
		favorite BOOLEAN NOT NULL
	);

	CREATE TABLE IF NOT EXISTS Marker (
		id CHAR(36) PRIMARY KEY CHECK (LENGTH(id) = 36),
		latency INTEGER CHECK (latency >= 0),
		packet_loss REAL CHECK (packet_loss >= 0 AND packet_loss <= 100),
		throughput REAL CHECK (throughput >= 0),
		dns_resolved INTEGER CHECK (dns_resolved >= 0),
		rtt INTEGER CHECK (rtt >= 0),
		status_code INTEGER CHECK (status_code >= 0),
		UNIQUE (latency, packet_loss, throughput, dns_resolved, rtt, status_code)
	);

	CREATE TABLE IF NOT EXISTS Metric (
		id CHAR(36) PRIMARY KEY CHECK (LENGTH(id) = 36),
		datetime CHAR(19) NOT NULL CHECK (LENGTH(datetime) = 19),
		server_id CHAR(36) NOT NULL,
		marker_id CHAR(36) NOT NULL,
		FOREIGN KEY (server_id) REFERENCES Server(id) ON DELETE CASCADE, 
		FOREIGN KEY (marker_id) REFERENCES Marker(id) ON DELETE CASCADE
	);`

	if _, err := db.db.Exec(query); err != nil {
		logger.Debugf("Error executing query: %v", err)
		return err
	}

	return nil
}

// NewDatabase creates a new SqliteDatabase instance by opening a connection to the SQLite database
// specified by the given Data Source Name (DSN). It ensures that foreign key enforcement is enabled
// and initializes the database.
//
// Parameters:
//   - dsn: The Data Source Name (DSN) string for connecting to the SQLite database.
//
// Returns:
//   - *SqliteDatabase: A pointer to the initialized SqliteDatabase instance.
//   - error: An error if there was an issue opening the database, establishing the connection,
//     enabling foreign key enforcement, or initializing the database.
func NewDatabase(dsn string) (*SqliteDatabase, error) {
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		logger.Printf("Error opening database: %v", err)
		return nil, err
	}

	if err = db.Ping(); err != nil {
		logger.Printf("Error establishing connection to database: %v", err)
		return nil, err
	}

	//Enable foreign key enforcement
	if _, err := db.Exec("PRAGMA foreign_keys = ON;"); err != nil {
		logger.Printf("Error setting foreign key enforcement: %v", err)
		return nil, err
	}

	sqlDb := &SqliteDatabase{db: db}
	if err = sqlDb.init(); err != nil {
		logger.Printf("Error initializing database: %v", err)
		return nil, err
	}
	return sqlDb, nil
}
