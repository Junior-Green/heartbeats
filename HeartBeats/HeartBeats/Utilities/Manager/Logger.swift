//
//  Logger.swift
//  HeartBeats
//
//  Created by Junior Green on 2025-01-12.
//

import Foundation

import Foundation

class Logger: ObservableObject {
  @Published var log: String = ""

  /* our singleton */
  static let shared = Logger()

  init() {
    self.log("Greetings from a new Logger")
  }

  func log(_ newEntry: String) {
    let dateFormatter = DateFormatter()
    dateFormatter.dateFormat = "HH:mm:ss.SSS "
    let msg = "\n" + dateFormatter.string(from: Date()) + newEntry
    self.log += msg
    debugPrint(msg)
  }

  func logError(_ error: Error) {
    if let error = error as? UDSocket.UDSErr {
      let msg = ("ERROR: " + String(describing: error.kind) + "\n"
        + "    " + error.localizedDescription)
      self.log(msg)
      debugPrint(msg)
    } else {
      let msg = "FOREIGN ERROR: \(String(describing: error))"
      self.log(msg)
      debugPrint(msg)
    }
  }
}
