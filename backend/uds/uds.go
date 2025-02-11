package uds

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"syscall"

	"golang.org/x/sys/unix"

	"github.com/Junior-Green/heartbeats/logger"
)

const retry = 5
const bufferSize = 1024

type action string
type status int
type UDSHandler func(UDSRequest, *UDSResponse)

const (
	Success status = iota
	BadRequest
	NotFound
	Internal
	Duplicate
)

const (
	GET    action = "GET"
	PUT    action = "PUT"
	POST   action = "POST"
	DELETE action = "DELETE"
)

type Payload struct {
	Data []byte `json:"data"`
}

// UDSRequest represents a request in the UDS (Unix Domain Socket) communication.
// It contains the action to be performed, the resource being targeted, and any
// additional payload data required for the action.
//
// Fields:
// - Action: The action to be performed, represented by the `action` type.
// - Resource: The resource being targeted by the action.
// - Payload: Additional data required for the action, represented as a byte slice.
type UDSRequest struct {
	Id       string  `json:"id"`
	Action   action  `json:"action"`
	Resource string  `json:"resource"`
	Payload  Payload `json:"payload"`
}

// UDSResponse represents the response structure for UDS (Unix Domain Socket) communication.
// It contains the status of the response and the payload data.
//
// Fields:
// - Status: The status of the response, represented by the `status` type.
// - Payload: The payload data of the response, represented as a byte slice and serialized as "data" in JSON.
type UDSResponse struct {
	Id      string `json:"id"`
	Status  status `json:"status"`
	Payload []byte `json:"payload"`
}

// socketConn represents a Unix Domain Socket (UDS) connection.
// It contains the socket path, a handler function to process UDS requests,
// and a listener to accept incoming connections.
type socketConn struct {
	socketPath string
	handler    UDSHandler
	listener   *net.UnixListener
}

// Listen starts the socket connection listener. It continuously accepts new
// connections and handles each connection in a separate goroutine. If an error
// occurs while accepting a connection, it logs the error and continues to
// accept new connections. The listener is closed when the function returns.
func (s *socketConn) Listen() {
	defer s.listener.Close()

	for {
		conn, err := s.listener.AcceptUnix()
		conn.SetReadBuffer(bufferSize)
		conn.SetWriteBuffer(bufferSize)
		setBlockingMode(conn, true)

		if err != nil {
			logger.Print("Error accepting connection:", err)
			continue
		}
		logger.Print("Client connection accepted")
		go s.handleRequest(conn)
	}
}

// setBlockingMode sets the blocking mode of a Unix domain socket connection.
// If blocking is true, the connection is set to blocking mode. If blocking is false,
// the connection is set to non-blocking mode.
//
// Parameters:
//   - conn: The Unix domain socket connection to modify.
//   - blocking: A boolean indicating whether to set the connection to blocking mode.
//
// Returns:
//   - error: An error if the operation fails, otherwise nil.
func setBlockingMode(conn *net.UnixConn, blocking bool) error {
	// Get the raw file descriptor
	rawConn, err := conn.SyscallConn()
	if err != nil {
		return fmt.Errorf("failed to get raw connection: %v", err)
	}

	var controlErr error
	err = rawConn.Control(func(fd uintptr) {
		// Get the current flags

		flags, err := unix.FcntlInt(fd, syscall.F_GETFL, 0)
		if err != nil {
			controlErr = fmt.Errorf("failed to get flags: %v", err)
			return
		}

		// Clear the non-blocking flag
		if blocking {
			flags &= ^unix.O_NONBLOCK
		} else {
			flags |= unix.O_NONBLOCK
		}

		// Set the new flags
		_, err = unix.FcntlInt(fd, syscall.F_SETFL, flags)
		if err != nil {
			controlErr = fmt.Errorf("failed to set flags: %v", err)
			return
		}
	})

	if err != nil {
		return fmt.Errorf("failed to control raw connection: %v", err)
	}
	if controlErr != nil {
		return controlErr
	}

	return nil
}

// handleRequest handles incoming requests on the socket connection.
// It reads the request, unmarshals it into a UDSRequest struct, processes it using the handler,
// and then marshals and writes the response back to the connection.
// If any error occurs during reading, unmarshalling, marshalling, or writing, it logs the error.
//
// Parameters:
//
//	c (net.Conn): The network connection to read from and write to.
//
// Note:
//
//	The connection is closed at the end of the function.
func (s *socketConn) handleRequest(c net.Conn) {
	defer s.listener.Close()
	defer c.Close()

	for {
		buf := make([]byte, 0, bufferSize)
		numBytes, err := c.Read(buf)
		if err != nil {
			logger.Printf("Error reading request: %v", err)
			continue
		} else if numBytes == 0 {
			logger.Printf("Empty buffer received")
			continue
		}

		var req UDSRequest
		if err := json.Unmarshal(buf, &req); err != nil {
			logger.Printf("Error decoding request: %v", err)
			logger.Printf("This is not valid JSON: %s", buf)
			continue
		}

		resp := &UDSResponse{Id: req.Id, Status: Success}
		s.handler(req, resp)

		bytes, err := json.Marshal(resp)
		if err != nil {
			logger.Printf("Error marshalling response: %v", err)
			continue
		}

		if _, err := c.Write(bytes); err != nil {
			logger.Printf("Error writing request: %v", err)
			continue
		}
	}
}

// NewSocketConn creates a new Unix Domain Socket (UDS) connection with the specified socket path and handler.
// It attempts to remove any existing socket file at the specified path before starting a new listener.
// The function retries starting the listener a predefined number of times before panicking if unsuccessful.
//
// Parameters:
//   - socketPath: The file system path where the UDS socket will be created.
//   - handler: An implementation of the UDSHandler interface to handle incoming connections.
//
// Returns:
//   - *socketConn: A pointer to the created socketConn instance.
//   - error: An error if the socket connection could not be created.
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
		listener, err := net.ListenUnix("unix", &net.UnixAddr{Name: socketPath, Net: "unix"})
		if err != nil {
			fmt.Printf("Error starting listener: %v\nRetrying...\n", err)
			continue
		}
		socket.listener = listener
		break
	}

	if socket.listener == nil {
		logger.Debug("ERROR: Could not start listener")
		return nil, fmt.Errorf("ERROR: Could not start listener")
	}

	return socket, nil
}

// Ok sets the status of the given UDSResponse to Success and assigns the provided payload to it.
// It returns the modified UDSResponse.
//
// Parameters:
//   - resp: A pointer to the UDSResponse that will be modified.
//   - payload: A byte slice containing the payload to be assigned to the response.
//
// Returns:
//   - A pointer to the modified UDSResponse with the status set to Success and the payload assigned.
func Ok(resp *UDSResponse, payload []byte) *UDSResponse {
	resp.Status = Success
	resp.Payload = payload
	return resp
}

// Error sets the status code and error message in the UDSResponse.
//
// Parameters:
//
//	resp - The UDSResponse to be modified.
//	error - The error message to be set in the response payload.
//	statusCode - The status code to be set in the response.
//
// Returns:
//
//	The modified UDSResponse with the specified status code and error message.
func Error(resp *UDSResponse, error string, statusCode status) *UDSResponse {
	resp.Status = statusCode
	resp.Payload = []byte(error)
	return resp
}
