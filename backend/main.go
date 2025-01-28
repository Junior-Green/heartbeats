package main

import (
	"fmt"
	"os"

	"github.com/Junior-Green/heartbeats/database"
	"github.com/Junior-Green/heartbeats/logger"
	"github.com/Junior-Green/heartbeats/uds"
	"github.com/Junior-Green/heartbeats/uds/udsserver"
)

const socketPath = "/var/run/heartbeats.socket"
const dsn = "/Library/Application Support/HeartBeats/heartbeats.db?cache=shared&mode=memory"

func main() {
	//Creates app folder if it doesn't exist
	if _, err := os.Stat(dsn); os.IsNotExist(err) {
		logger.Printf("Error finding database file: %v", err)
		os.Exit(1)
	}

	if _, err := os.Stat("/var/run"); os.IsNotExist(err) {
		logger.Print("Cannot find /var/run")
		os.Exit(1)
	}

	db, err := database.NewDatabase(dsn)
	if err != nil {
		logger.Printf("Error initializing connection to database: %v", err)
		os.Exit(1)
	}

	server := udsserver.UDSServer{}
	//GETS
	server.AddGetHandler("/", handlePing())
	server.AddGetHandler("/server/all", handleGetAllServers(db))
	server.AddGetHandler("/server/host", handleGetServerByHost(db))
	server.AddGetHandler("/metric/host", handleGetMetricsByHost(db))

	//PUTS
	server.AddPutHandler("/server/favorite", handleUpdateFavorite(db))

	//POSTS
	server.AddPostHandler("/server", handleCreateServer(db))

	//DELETES
	server.AddDeleteHandler("/server/host", handleDeleteServerByHost(db))

	conn, err := uds.NewSocketConn(socketPath, server.UDSRequestHandler())
	if err != nil {
		fmt.Printf("Error creating socket connection: %v", err)
	}

	logger.Printf("Listener listening on socket from %s", socketPath)
	conn.Listen()
}
