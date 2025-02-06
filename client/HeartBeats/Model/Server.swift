//
//  ServerGroup.swift
//  HeartBeats
//
//  Created by Junior Green on 2025-01-11.
//

import Foundation

struct Server: Codable, Identifiable {
  let id: UUID
  let host: String
  var favorite: Bool
  var online: Bool

  init(id: UUID, host: String, favorite: Bool, online: Bool) {
    self.id = id
    self.host = host
    self.favorite = favorite
    self.online = online
  }

  init(host: String) {
    self.id = UUID()
    self.host = host
    self.favorite = false
    self.online = false
  }

  enum CodingKeys: String, CodingKey {
    case id
    case host
    case favorite
    case online
  }
}
