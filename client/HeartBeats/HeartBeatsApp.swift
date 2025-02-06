//
//  HeartBeatsApp.swift
//  HeartBeats
//
//  Created by Junior Green on 2025-01-06.
//

import Cocoa

@main
enum HeartBeatsApp {
  static let appDelegate = AppDelegate()

  static func main() {
    // Custom Initialization
    setupEnvironment()

    // Start AppKit's main loop
    let app = NSApplication.shared
    app.delegate = appDelegate
    app.run()
  }

  private static func setupEnvironment() {
    guard let appSuppDir = Files.appSuppDir else {
      NSApplication.shared.terminate(nil)
      return
    }

    createDirectory(url: URL(fileURLWithPath: appSuppDir))
    createDirectory(url: URL(fileURLWithPath: Files.tempDir))

    if !FileManager.default.fileExists(atPath: URL.databaseFile.path()) {
      createFile(url: URL.databaseFile)
    }
    createFile(url: URL.goLogFile)

    // Trigger .init
    _ = ProcessManager.shared
    _ = NetworkManager.shared
  }

  private static func createDirectory(url: URL) {
    debugPrint("Attempting to create directory at: \(url.path())")
    if !FileManager.default.fileExists(atPath: url.path) {
      do {
        try FileManager.default.createDirectory(atPath: url.path(), withIntermediateDirectories: true, attributes: nil)
        debugPrint("Created directory at: \(url.path())")
      } catch {
        showNSAlert(err: error)
        NSApplication.shared.terminate(nil)
      }
    }
  }

  private static func createFile(url: URL) {
    debugPrint("Attempting to create file: \(url.path())")
    if !FileManager.default.createFile(atPath: url.path(), contents: nil) {
      showNSAlert(item: NSAlertContext.createFile(filePath: url).alertItem)
      NSApplication.shared.terminate(nil)
    }
    debugPrint("Created file at: \(url.path())")
  }
}
