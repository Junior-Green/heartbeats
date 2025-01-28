//
//  Constants.swift
//  HeartBeats
//
//  Created by Junior Green on 2025-01-28.
//

import AppKit

enum Env: String {
  case debug
  case appStore
}

enum Files {
  static let socketFile: NSString = NSString(string: "heartbeats.socket")
  static let databaseFile: NSString = NSString(string: "heartbeats.db")
}

enum AppInfo {
  static let version: String = Bundle.main.object(forInfoDictionaryKey: "CFBundleShortVersionString") as! String
  static let build: String = Bundle.main.object(forInfoDictionaryKey: "CFBundleVersion") as! String
  static let name = Bundle.main.object(forInfoDictionaryKey: "CFBundleDisplayName") as! String
}

enum BuildMode {
  static var isDebug: Bool {
    #if DEBUG
    return true
    #else
    return false
    #endif
  }

  static var env: Env {
    if isDebug {
      return .debug
    } else {
      return .appStore
    }
  }
}
