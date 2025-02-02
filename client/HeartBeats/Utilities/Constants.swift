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
  static let uds: String = "heartbeats.socket"
  static let database: String = "heartbeats.db"
  static let goExecutable: String = "com.heartbeats.universal"
  static let goLog: String = "heartbeats.log"
  static let tempDir: String = FileManager.default.temporaryDirectory.appending(path: AppInfo.name).path()

  static var appSuppDir: String? {
    do {
      return try URL(for: .applicationSupportDirectory, in: .userDomainMask, appropriateFor: nil, create: true)
        .appending(path: AppInfo.name)
        .path()
    } catch {
      showNSAlert(err: error)
      return nil
    }
  }
}

enum Action: String, Codable {
  case GET
  case PUT
  case POST
  case DELETE
}

enum StatusCode: Int, Codable {
  case Success = 0
  case BadRequest = 1
  case NotFound = 2
  case Internal = 3
  case Duplicate = 4
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
