import Foundation

class UDSClient {
  private var socket: Int32?
  private let logger = Logger.shared
  private let socketPath: String
  private weak var delegate: UDSClientDelegate?
  
  private static let retries: Int = 5
  
  /// Initializes the Client with an app group identifier and a socket name.
  /// - Parameters:
  ///   - appGroup: The application group identifier.
  ///   - socketName: The name of the socket.
  init(socketPath: String, delegate: UDSClientDelegate) {
    self.socketPath = socketPath
    self.delegate = delegate
  }
  
  /// Attempts to connect to the Unix socket.
  func start() {
    logger.log("Attempting to initialize connection to socket path: \(socketPath)")
    
    guard let socket = createSocket() else {
      delegate?.didFailToAcceptSocketConnection(err: NetworkError.socketSetup("Could not create socket"))
      return
    }
    self.socket = socket
    
    do {
      try acceptConnection(descriptor: self.socket!, path: socketPath)
    } catch {
      delegate?.didFailToAcceptSocketConnection(err: NetworkError.socketSetup(error.localizedDescription))
      return
    }
    
    Task { [self] in
      try await Task.sleep(for: .seconds(3))
      self.readData()
    }
  }
  
  func sendData(_ data: Data) {
    guard let socket = socket else {
      logger.log("No connected client.")
      return
    }

    if data.isEmpty {
      logger.log("No data to send!")
      return
    }

    data.withUnsafeBytes { (bytes: UnsafeRawBufferPointer) in
      let pointer = bytes.bindMemory(to: UInt8.self)
      let bytesWritten = Darwin.send(socket, pointer.baseAddress!, data.count, 0)

      if bytesWritten == -1 {
        self.logger.log("Error sending data")
        return
      }
      self.logger.log("\(bytesWritten) bytes written")
    }
  }
  
  /// Reads data from the connected socket.
  func readData() {
    Task.detached { [weak self] in
      while true {
        var buffer = [UInt8](repeating: 0, count: 1024)
        guard let socketDescriptor = self?.socket else {
          self?.logger.log("Socket descriptor is nil")
          return
        }
        let bytesRead = read(socketDescriptor, &buffer, buffer.count)
        if bytesRead == -1, errno == EWOULDBLOCK {
          self?.logger.log("socket is set to not block")
          continue // No data yet, but keep looping
        } else if bytesRead == 0 {
          self?.logger.log("socket is non blocking")
          continue
        }
        
        // Print the data for debugging purposes
        let data = Data(buffer[..<bytesRead])
        self?.delegate?.didRecieveData(data: data, err: nil)
      }
      
      if let socket = self?.socket {
        close(socket)
        self?.delegate?.didSocketConnetionClose()
        self?.socket = nil
      }
    }
  }
  
  private func createSocket() -> Int32? {
    let socket = Darwin.socket(AF_UNIX, SOCK_STREAM, 0)
    guard socket != -1 else {
      return nil
    }
    
    var flags = fcntl(socket, F_GETFL, 0)
    flags &= ~O_NONBLOCK
    if fcntl(socket, F_SETFL, flags) != 0 {
      return nil
    }
    
    return socket
  }
  
  private func acceptConnection(descriptor: Int32, path: String) throws {
    var address = sockaddr_un()
    address.sun_family = sa_family_t(AF_UNIX)
    path.withCString { ptr in
      withUnsafeMutablePointer(to: &address.sun_path.0) { dest in
        _ = strcpy(dest, ptr)
      }
    }
    
    var attempts = 0
    
    while attempts < UDSClient.retries {
      if Darwin.connect(descriptor, withUnsafePointer(to: &address) { $0.withMemoryRebound(to: sockaddr.self, capacity: 1) { $0 } }, socklen_t(MemoryLayout<sockaddr_un>.size)) == -1 {
        sleep(UInt32(pow(2.0, Double(attempts))))
        attempts += 1
        continue
      }
      return
    }
    throw NetworkError.socketSetup("Error connecting to socket - \(String(cString: strerror(errno)))")
  }
}
