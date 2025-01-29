//
//  ProcessManager.swift
//  HeartBeats
//
//  Created by Junior Green on 2025-01-29.
//

import AppKit
import Foundation

final class ProcessManager {
  static let udsserver = ProcessManager()

  private let process: Process

  private init() {
    process = Process()
    process.executableURL = Bundle.main.url(forResource: Files.goExecutable as String, withExtension: nil)
    process.environment = ["SOCKET_PATH": URL.socketFile.path()]

    do {
      try process.run()
    } catch {
      Logger.shared.logError(error)
      showNSAlert(err: error)
      NSApplication.shared.terminate(nil)
    }
  }

  deinit {
    process.terminate()
  }
}
