import Foundation

//
//  UDSRequest.swift
//  HeartBeats
//
//  Created by Junior Green on 2025-02-01.
//

struct UDSRequest: Codable, Identifiable {
  let id: UUID = .init()
  let action: Action
  let resource: String
  let payload: Payload?

  enum CodingKeys: String, CodingKey {
    case id
    case action
    case resource
    case payload
  }
}

struct Payload: Codable {
  let data: Data

  init(_ data: Data) {
    self.data = data
  }

  enum CodingKeys: String, CodingKey {
    case data
  }
}
