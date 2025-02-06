//
//  Logger.swift
//  HeartBeats
//
//  Created by Junior Green on 2025-01-12.
//

import Foundation

class Logger: ObservableObject {
  @Published var log: String = ""

  /* our singleton */
  static let shared = Logger()

  func log(_ newEntry: String) {
    let dateFormatter = DateFormatter()
    dateFormatter.dateFormat = "HH:mm:ss.SSS "
    let msg = "\n" + dateFormatter.string(from: Date()) + newEntry
    self.log += msg
    debugPrint(msg)
  }

  func logError(_ error: Error) {
    let msg = ("ERROR: " + String(describing: error) + "\n"
      + "    " + error.localizedDescription)
    self.log(msg)
    debugPrint(msg)
  }
}
