//
//  RootWindowController.swift
//  HeartBeats
//
//  Created by Junior Green on 2025-01-09.
//

import AppKit
import SwiftUI

final class RootWindowController: NSWindowController {
  private let windowWidth: CGFloat = 1100
  private let windowHeight: CGFloat = 670

  init() {
    let rootView = RootView()
      .frame(minWidth: windowWidth, maxWidth: .infinity, minHeight: windowHeight, maxHeight: .infinity)
      .background(LinearGradient(colors: [Color.brandPrimary, Color.brandSecondary], startPoint: .top, endPoint: .bottom))

    // Create the window
    let window = NSWindow(
      contentRect: NSRect(x: 0, y: 0, width: windowWidth, height: windowHeight),
      styleMask: [.titled, .closable, .miniaturizable, .fullSizeContentView, .resizable],
      backing: .buffered,
      defer: false
    )

    // Set up the window properties
    window.title = ""
    window.backgroundColor = NSColor.clear
    window.isReleasedWhenClosed = false
    window.center()
    window.isOpaque = false
    window.isMovableByWindowBackground = true
    window.titlebarAppearsTransparent = true

    window.contentView = NSHostingView(rootView: rootView)

    window.contentView?.wantsLayer = true
    window.contentView?.layer?.cornerRadius = 40
    window.contentView?.layer?.masksToBounds = true
    window.contentView?.layer?.maskedCorners = [.layerMaxXMinYCorner, .layerMinXMaxYCorner]

    super.init(window: window)
  }

  @available(*, unavailable)
  required init?(coder: NSCoder) {
    fatalError("init(coder:) has not been implemented")
  }
}
