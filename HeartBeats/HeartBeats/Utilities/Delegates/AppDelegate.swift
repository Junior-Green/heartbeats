//
//  AppDelegate.swift
//  HeartBeats
//
//  Created by Junior Green on 2025-01-09.
//

import AppKit
import SwiftUI

class AppDelegate: NSObject, NSApplicationDelegate {
  var customWindowController: RootWindowController?

  func applicationDidFinishLaunching(_ notification: Notification) {
    // Instantiate and show the custom window
    customWindowController = RootWindowController()
    
    customWindowController?.showWindow(nil)
  }
}
