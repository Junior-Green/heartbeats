//
//  Alert.swift
//  HeartBeats
//
//  Created by Junior Green on 2025-01-07.
//

import SwiftUI

struct AlertItem: Identifiable {
  let id = UUID()
  let title: Text
  let message: Text
  let dismissButton: Alert.Button
}


//  static let invalidData = AlertItem(title: Text("Server Error"),
//                                     message: Text("The data received from the server was invalid. Please contact support."),
//                                     dismissButton: .default(Text("OK")))
enum AlertContext {
  // MARK: - Network Alerts
}
