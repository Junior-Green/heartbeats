//
//  ProcessManager.swift
//  HeartBeats
//
//  Created by Junior Green on 2025-01-29.
//

import AppKit
import Foundation

final class ProcessManager {
  static let shared = ProcessManager()

  private let process: Process

  private init() {
    process = Process()
    process.executableURL = Bundle.main.url(forResource: Files.goExecutable as String, withExtension: nil)
    process.environment = [
      "SOCKET_PATH": URL.socketFile.path(),
      "DB_PATH": URL.databaseFile.path(),
      "MODE": BuildMode.env.rawValue,
      "PID": String(describing: ProcessInfo().processIdentifier)
    ]

    do {
      debugPrint("starting backend agent process")
      try process.run()
    } catch {
      Logger.shared.logError(error)
      showNSAlert(err: error)

      DispatchQueue.main.async {
        NSApplication.shared.terminate(nil)
      }
    }
  }

  func terminate() {
    debugPrint("terminated backend agent process")
    process.terminate()
  }

  deinit {
    terminate()
  }
}
