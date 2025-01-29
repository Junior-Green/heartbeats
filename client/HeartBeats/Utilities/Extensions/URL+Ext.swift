//
//  URL+Ext.swift
//  HeartBeats
//
//  Created by Junior Green on 2025-01-29.
//

import AppKit

extension URL {
  static let databaseFile = URL.applicationSupportDirectory.appending(path: Files.database as String, directoryHint: .notDirectory)
}

extension URL {
  static let socketFile = URL.temporaryDirectory.appending(path: Files.uds as String, directoryHint: .notDirectory)
}
