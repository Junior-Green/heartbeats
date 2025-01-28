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
  static let appSupport = AlertItem(title: Text("File Path Error"),
                                     message: Text("Unable to open necessary files in /Library/Application Support. Invalid permissions or does not exist."),
                                     dismissButton: .default(Text("OK")))

  // MARK: - Network Alerts
}
