//
//  WindowVC.swift
//  HeartBeats
//
//  Created by Junior Green on 2025-01-08.
//

import AppKit
import SwiftUI

final class CustomWindowController: NSWindowController {
  init() {
    let root = NSHostingView(rootView: HomeView())
    // Create the window
    let window = NSWindow(
      contentRect: NSRect(x: 0, y: 0, width: 1250, height: 750),
      styleMask: [.titled, .closable, .fullSizeContentView, .miniaturizable, .fullScreen],
      backing: .buffered,
      defer: false
    )

    // Set up the window properties
    window.title = ""
    window.isReleasedWhenClosed = false
    window.center()
    window.contentView = root
    window.titlebarAppearsTransparent = true
    window.isMovableByWindowBackground = true

    super.init(window: window)
  }

  @available(*, unavailable)
  required init?(coder: NSCoder) {
    fatalError("init(coder:) has not been implemented")
  }
}

class AppDelegate: NSObject, NSApplicationDelegate {
  var customWindowController: CustomWindowController?

  func applicationDidFinishLaunching(_ notification: Notification) {
    // Instantiate and show the custom window
    customWindowController = CustomWindowController()
    customWindowController?.showWindow(nil)
  }
}
