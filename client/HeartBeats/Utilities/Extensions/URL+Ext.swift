//
//  URL+Ext.swift
//  HeartBeats
//
//  Created by Junior Green on 2025-01-29.
//

import AppKit

extension URL {
  static let databaseFile = URL(fileURLWithPath: "\(Files.appSuppDir!)/\(Files.database)", isDirectory: false)
}

extension URL {
  static let socketFile = URL(fileURLWithPath: "\(Files.tempDir)/\(Files.uds)", isDirectory: false)
}

extension URL {
  static let goLogFile = URL(fileURLWithPath: "\(Files.tempDir)/\(Files.goLog)", isDirectory: false)
}
