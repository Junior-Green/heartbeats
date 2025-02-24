//
//  ConnectionManager.swift
//  HeartBeats
//
//  Created by Junior Green on 2025-02-23.
//

import Foundation
import Network

final class ConnectionManager {
  private static let bufferSize: Int = 50000
  private static let prefixLength: Int = 4
  
  private let connection: NWConnection
  private let logger: Logger
  
  private var buffer = Data()
  
  var onRecieve: ((UDSResponse) -> Void)?
  
  init(
    connection: NWConnection,
    logger: Logger = Logger.shared
  ) {
    self.connection = connection
    self.logger = logger
    connection.start(queue: .global())
    connection.stateUpdateHandler = handleState()
    read()
  }
  
  deinit {
    self.connection.cancel()
  }
  
  private func handleState() -> ((NWConnection.State) -> Void) {
    return { state in
      switch state {
      case .setup:
        self.logger.log("The connection has been initialized but not started.")
      case .waiting:
        self.logger.log("The connection is waiting for a network.")
      case .preparing:
        self.logger.log("The connection in the process of being established.")
      case .ready:
        self.logger.log("The connection is established, and ready to send and receive data.")
      case .failed(let error):
        self.logger.logError(error)
      case .cancelled:
        self.logger.log("The connection has been canceled.")
      @unknown default:
        fatalError("ERROR: Unknown NWConnection.State not handled")
      }
    }
  }
  
  private func processBuffer() {
    while buffer.count >= 4 {
      let length = Int(buffer.prefix(4).withUnsafeBytes { $0.load(as: UInt32.self).bigEndian })
      if buffer.count >= length + 4 {
        let jsonData = buffer.subdata(in: 4 ..< (length + 4))
        buffer.removeFirst(length + 4)
        do {
          let resp = try JSONDecoder().decode(UDSResponse.self, from: jsonData)
          onRecieve?(resp)
        } catch {
          logger.logError(error)
        }
      } else {
        break // Wait for more
      }
    }
  }
  
  private func read() {
    connection.receive(minimumIncompleteLength: ConnectionManager.prefixLength, maximumLength: ConnectionManager.bufferSize) { data, _, isComplete, error in
      if let data = data, !data.isEmpty {
        self.buffer.append(data)
        self.processBuffer()
      }
      if isComplete { print("Connection closed"); return }
      if let error = error { self.logger.logError(error); return }
      self.read() // Keep going
    }
  }
  
  func send(data: Data) throws {
    var error: NWError?
    connection.send(content: data, completion: .contentProcessed { err in
      if let err = err {
        error = err
      }
    })
    
    if error != nil {
      throw error!
    }
  }
}
