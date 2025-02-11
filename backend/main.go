package main

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/Junior-Green/heartbeats/database"
	"github.com/Junior-Green/heartbeats/logger"
	"github.com/Junior-Green/heartbeats/server"
	"github.com/Junior-Green/heartbeats/server/ping"
	"github.com/Junior-Green/heartbeats/uds"
	"github.com/Junior-Green/heartbeats/uds/udsserver"
)

const socketEnvKey = "SOCKET_PATH"
const dbEnvKey = "DB_PATH"
const logEnvKey = "LOG_PATH"
const pingInterval = time.Second * 30

// const dsnParams = "/Library/Application Support/HeartBeats/heartbeats.db?cache=shared&mode=memory"

func main() {

	pid, err := strconv.Atoi(os.Getenv("PID"))
	if err != nil {
		logger.Printf("Error getting pid: %v", err)
		os.Exit(1)
	}
	go watchParentProcess(pid)

	socketPath, ok := os.LookupEnv(socketEnvKey)
	if !ok {
		logger.Printf("Could not find socket path environment variable: %v", socketEnvKey)
		os.Exit(1)
	}

	dbPath, ok := os.LookupEnv(dbEnvKey)
	if !ok {
		logger.Printf("Could not find database path environment variable: %v", dbEnvKey)
		os.Exit(1)
	}

	logPath, ok := os.LookupEnv(logEnvKey)
	if ok {
		if f, err := os.Open(logPath); err == nil {
			logger.SetOuput(f)
		} else {
			logger.Debug("Error opening log file. Logs will be written to stdout")
		}
	}

	var dsnParams url.Values = map[string][]string{"cache": {"shared"}, "mode": {"memory"}}

	db, err := database.NewDatabase(strings.Join([]string{dbPath, "?", dsnParams.Encode()}, ""))
	if err != nil {
		logger.Printf("Error initializing connection to database: %v", err)
		os.Exit(1)
	}

	server := udsserver.NewUDSServer()
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

	if err = createPingWorkers(db); err != nil {
		logger.Printf("Error creating ping workers: %v", err)
		os.Exit(1)
	}

	logger.Printf("Listener listening on socket from %s", socketPath)
	conn.Listen()
}

func watchParentProcess(pid int) {
	process, err := os.FindProcess(pid)
	if err != nil {
		logger.Print("Parent process does not exist")
		os.Exit(1)
	}
	for {
		if err = process.Signal(syscall.Signal(0)); err != nil {
			logger.Print("Parent process terminated. Exiting...")
			os.Exit(0)
		}
	}
}

func createPingWorkers(db *database.SqliteDatabase) error {
	servers, err := db.GetAllServers()
	if err != nil {
		return err
	}

	for _, serv := range servers {
		go func(serv server.Server) {
			ch := ping.PingAfter(serv.Host, pingInterval)

			for data := range ch {
				db.UpdateOnlineStatusByHost(serv.Host, data.Throughput.Valid)
				db.AddPingMetricByHost(serv.Host, data)
			}

		}(serv)
	}

	return nil
}
