import Foundation

//
//  UDSRequest.swift
//  HeartBeats
//
//  Created by Junior Green on 2025-02-01.
//

struct UDSRequest: Codable, Identifiable {
  let id: UUID = UUID()
  let action: Action
  let resource: String
  let payload: Data?

  enum CodingKeys: String, CodingKey {
    case id
    case action
    case resource
    case payload
  }
}
