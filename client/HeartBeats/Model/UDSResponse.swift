//
//  UDSResponse.swift
//  HeartBeats
//
//  Created by Junior Green on 2025-02-01.
//

import Foundation

struct UDSResponse: Codable, Identifiable, Hashable {
  let id: UUID
  let status: StatusCode
  let payload: Data?

  enum CodingKeys: String, CodingKey {
    case id
    case status
    case payload
  }
}
