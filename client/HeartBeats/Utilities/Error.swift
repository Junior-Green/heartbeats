//
//  Error.swift
//  HeartBeats
//
//  Created by Junior Green on 2025-01-07.
//

import Network
import SwiftUI

enum HBError: Error {
  case defaultError
  case serverResponseError(_ nwError: NWError)
}

enum NetworkError: Error {
  case timeout
  case badRequest
  case duplicate
  case internalError
  case notFound
}
