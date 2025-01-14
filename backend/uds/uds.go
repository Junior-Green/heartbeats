package uds

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/Junior-Green/heartbeats/logger"
)

const timeout = time.Second * 3
const retry = 5

type action string
type status int
type UDSHandler func(UDSRequest) UDSResponse

const (
	SUCCESS status = iota
	BADREQUEST
	NOTFOUND
	ERROR
)

const (
	GET    action = "GET"
	PUT    action = "PUT"
	POST   action = "POST"
	DELETE action = "DELETE"
)

type UDSRequest struct {
	Action   action `json:"action"`
	Resource string `json:"resource"`
	Payload  []byte `json:"payload"`
}

type UDSResponse struct {
	Status  status `json:"status"`
	Payload []byte `json:"data"`
}

type socketConn struct {
	socketPath string
	handler    func(UDSRequest) UDSResponse
	listener   net.Listener
}

func (s *socketConn) Listen() {
	defer s.listener.Close()

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			logger.Print("Error accepting connection:", err)
			continue
		}

		go s.handleRequest(conn)
	}
}

func (s *socketConn) handleRequest(c net.Conn) {
	defer c.Close()
	c.SetDeadline(time.Now().Add(timeout))

	buf := make([]byte, 0, 1024)
	if _, err := c.Read(buf); err != nil {
		logger.Printf("Error reading request: %v", err)
	}

	var req UDSRequest
	if err := json.Unmarshal(buf, &req); err != nil {
		logger.Printf("Error decoding request: %v", err)
		return
	}

	resp := s.handler(req)
	bytes, err := json.Marshal(resp)
	if err != nil {
		logger.Printf("Error marshalling response: %v", err)
		return
	}

	if _, err := c.Write(bytes); err != nil {
		logger.Printf("Error writing request: %v", err)
	}
}

func NewSocketConn(socketPath string, handler UDSHandler) (*socketConn, error) {
	socket := &socketConn{
		socketPath: socketPath,
		handler:    handler,
	}

	if _, err := os.Stat(socket.socketPath); err == nil {
		os.Remove(socket.socketPath)
	}

	for i := 0; i < retry; i++ {
		// Attempt to start the listener
		listener, err := net.Listen("unix", socket.socketPath)
		if err != nil {
			fmt.Printf("Error starting listener: %v\nRetrying...\n", err)
			continue
		}
		socket.listener = listener
		break
	}

	if socket.listener == nil {
		panic("ERROR: Could not start listener")
	}

	return socket, nil
}
