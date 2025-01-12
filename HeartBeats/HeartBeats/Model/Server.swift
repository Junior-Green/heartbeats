//
//  ServerGroup.swift
//  HeartBeats
//
//  Created by Junior Green on 2025-01-11.
//

struct Server: Codable {
  let hostname: String = ""
  let favorite: Bool = false
  let metrics: ServerMetrics? = nil

  init(from decoder: any Decoder) throws {}

  func encode(to encoder: any Encoder) throws {}
}
