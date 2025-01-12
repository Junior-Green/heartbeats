//
//  NetworkManager.swift
//  HeartBeats
//
//  Created by Junior Green on 2025-01-07.
//

import SwiftUI

final class NetworkManager: UDSClientDelegate {
  private let logger = Logger.shared
  private let client: UDSClient = .init(socketUrl: UDSClient.serviceUrl())
  
  static let shared = NetworkManager()
  
  private init() {
    client.delegate = self
    
    do {
      try client.start()
    }
    catch {
      logger.log(error.localizedDescription)
    }
  }
  
  deinit {
    self.client.stop()
  }
  
  
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
  
  // MARK: - Delegate Methods
  func handleSocketClientMsgDict(_ aDict: [AnyHashable: AnyHashable]?, client: UDSClient?, error: (any Error)?) {
    if error != nil {
      logger.logError(error!)
    }
  }
  
  func handleSocketClientDisconnect(_ client: UDSClient?) {
    logger.log("Client has disconnected")
  }
  
  func handleSocketServerDisconnect(_ client: UDSClient?) {
    logger.log("Client has disconnected")
  }
}
