//
//  NetworkManager.swift
//  HeartBeats
//
//  Created by Junior Green on 2025-01-07.
//

final class NetworkManager {
  private let logger = Logger.shared
  private let client: UDSClient = .init(socketUrl: UDSClient.serviceUrl())
  private let delegate: UDSClientDelegate?
  
  static let shared = NetworkManager()
  
  private init() {
    self.delegate = UDSDelegate()
    client.delegate = delegate
    
//    do {
//      try client.start()
//    }
//    catch {
//      logger.log(error.localizedDescription)
//    }
  }

//
//  deinit {
//    self.client.stop()
//  }
  
  // MARK: - API Methods

  func getServer(hostname: String) throws -> Server {
    throw HBError.DefaultError
  }
  
  func favoriteServer(hostname: String) throws {
    throw HBError.DefaultError
  }
  
  func addServer(hostname: String) throws {
    throw HBError.DefaultError
  }
  
  func getAllServers() throws -> [Server] {
    throw HBError.DefaultError
  }
}
