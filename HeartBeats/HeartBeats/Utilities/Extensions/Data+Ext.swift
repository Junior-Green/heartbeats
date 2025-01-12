//
//  Data+Ext.swift
//  HeartBeats
//
//  Created by Junior Green on 2025-01-12.
//

import Foundation

extension Data {
  mutating func append(data: Data, offset: Int, size: Int) {
    let safeSize = Swift.min(data.count - offset, size)
    let start = Int(data.startIndex) + Int(offset)
    let end = Int(start) + Int(safeSize)
    self.append(data[start ..< end])
  }
}
