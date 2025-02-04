import Foundation

class UDSClient {
  private var socket: Int32?
  private var clientSocket: Int32?
  private let logger = Logger.shared
  private let socketPath: String
  private let delegate: UDSClientDelegate

  /// Initializes the Server with an app group identifier and a socket name.
  /// - Parameters:
  ///   - socketName: The name of the socket.
  init(socketPath: String, delegate: UDSClientDelegate) {
    self.socketPath = socketPath
    self.delegate = delegate
  }

  /// Starts the server and begins listening for connections.
  func startBroadcasting() {
    createSocket()
    bindSocket()
    listenOnSocket()
    waitForConnection()
  }

  /// Creates a socket for communication.
  private func createSocket() {
    socket = Darwin.socket(AF_UNIX, SOCK_STREAM, 0)
    guard socket != nil, socket != -1 else {
      logger.log("Error creating socket")
      return
    }
    logger.log("Socket created successfully")
  }

  /// Binds the created socket to a specific address.
  private func bindSocket() {
    guard let socket = socket else { return }

    var address = sockaddr_un()
    address.sun_family = sa_family_t(AF_UNIX)
    socketPath.withCString { ptr in
      withUnsafeMutablePointer(to: &address.sun_path.0) { dest in
        _ = strcpy(dest, ptr)
      }
    }

    unlink(socketPath) // Remove any existing socket file

    if Darwin.bind(socket, withUnsafePointer(to: &address) { $0.withMemoryRebound(to: sockaddr.self, capacity: 1) { $0 } }, socklen_t(MemoryLayout<sockaddr_un>.size)) == -1 {
      logger.log("Error binding socket - \(String(cString: strerror(errno)))")
      return
    }
    logger.log("Binding to socket path: \(socketPath)")
  }

  /// Listens for connections on the bound socket.
  private func listenOnSocket() {
    guard let socket = socket else { return }

    if Darwin.listen(socket, 1) == -1 {
      logger.log("Error listening on socket - \(String(cString: strerror(errno)))")
      return
    }
    logger.log("Listening for connections...")
  }

  /// Waits for a connection and accepts it when available.
  private func waitForConnection() {
    DispatchQueue.global().async { [weak self] in
      self?.acceptConnection()
    }
  }

  /// Accepts a connection request from a client.
  private func acceptConnection() {
    guard let socket = socket else { return }

    var clientAddress = sockaddr_un()
    var clientAddressLen = socklen_t(MemoryLayout<sockaddr_un>.size)
    clientSocket = Darwin.accept(socket, withUnsafeMutablePointer(to: &clientAddress) { $0.withMemoryRebound(to: sockaddr.self, capacity: 1) { $0 } }, &clientAddressLen)

    if clientSocket == -1 {
      logger.log("Error accepting connection - \(String(cString: strerror(errno)))")
      delegate.didFailToAcceptSocketConnection()
      return
    }
    logger.log("Connection accepted!")
    delegate.didAcceptSocketConnection()
    readData()
  }

  /// Sends the provided data to the connected client.
  /// - Parameter data: The data to send.
  func sendData(_ data: Data) {
    guard let clientSocket = clientSocket else {
      logger.log("No connected client.")
      return
    }

    if data.isEmpty {
      logger.log("No data to send!")
      return
    }

    data.withUnsafeBytes { (bytes: UnsafeRawBufferPointer) in
      let pointer = bytes.bindMemory(to: UInt8.self)
      let bytesWritten = Darwin.send(clientSocket, pointer.baseAddress!, data.count, 0)

      if bytesWritten == -1 {
        logger.log("Error sending data")
        return
      }
      logger.log("\(bytesWritten) bytes written")
    }
  }

  func readData() {
    DispatchQueue.global().async(execute: .init(block: {
      while true {
        var buffer = [UInt8](repeating: 0, count: 1024)
        guard let socketDescriptor = self.socket else {
          self.logger.log("Socket descriptor is nil")
          return
        }
        let bytesRead = read(socketDescriptor, &buffer, buffer.count)
        if bytesRead <= 0 {
          self.logger.log("Error reading from socket or connection closed")
          self.delegate.didRecieveData(data: nil, err: NetworkError.badRequest)
          continue // exit loop on error or closure of connection
        }

        // Print the data for debugging purposes
        let data = Data(buffer[..<bytesRead])
        self.logger.log("Received data: \(data)")
        self.delegate.didRecieveData(data: data, err: nil)
      }

      if let socket = self.socket {
        close(socket)
      }
    }
    ))
  }

  /// Stops the server and closes any open connections.
  func stopBroadcasting() {
    if let clientSocket = clientSocket {
      logger.log("Closing client socket...")
      close(clientSocket)
    }
    if let socket = socket {
      logger.log("Closing server socket...")
      close(socket)
      self.socket = nil
    }
    unlink(socketPath)
    logger.log("Broadcasting stopped.")
  }
}
