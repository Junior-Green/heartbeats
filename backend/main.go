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
const dbPath = "/Library/Application Support/HeartBeats/heartbeats.db"

func main() {
	_, err := database.NewDatabase(dbPath)
	if err != nil {
		logger.Printf("Error initializing connection to database: %v", err)
		os.Exit(1)
	}

	server := udsserver.UDSServer{}
	//GETS
	server.AddGetHandler("/", handlePing())
	//PUTS

	//POSTS

	//DELETES

	conn, err := uds.NewSocketConn(socketPath, server.UDSRequestHandler())
	if err != nil {
		fmt.Printf("Error creating socket connection: %v", err)
	}
	conn.Listen()
}
