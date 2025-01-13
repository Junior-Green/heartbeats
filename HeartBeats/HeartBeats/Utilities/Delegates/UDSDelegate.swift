//
//  UDSDelegate.swift
//  HeartBeats
//
//  Created by Junior Green on 2025-01-13.
//

import Foundation

class UDSDelegate: UDSClientDelegate {
  func handleSocketClientMsgDict(_ aDict: [AnyHashable: AnyHashable]?, client: UDSClient?, error: (any Error)?) {
    if error != nil {
      Logger.shared.logError(error!)
    }
  }

  func handleSocketClientDisconnect(_ client: UDSClient?) {
    Logger.shared.log("Client has disconnected")
  }

  func handleSocketServerDisconnect(_ client: UDSClient?) {
    Logger.shared.log("Client has disconnected")
  }
}
