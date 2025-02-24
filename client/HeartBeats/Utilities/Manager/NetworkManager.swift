import Combine
import Foundation
import Network

final class NetworkManager {
  private let logger: Logger = .shared
  private let connectionManager: ConnectionManager
  private var responseQueue: [UUID: CheckedContinuation<UDSResponse, any Error>] = [:]
  
  static let shared = NetworkManager(socketPath: URL.socketFile.path())
  
  private static let timeout: TimeInterval = 2
  
  private init(socketPath: String) {
    let conn = NWConnection(to: .unix(path: socketPath), using: .tcp)
    self.connectionManager = ConnectionManager(connection: conn)
    connectionManager.onRecieve = onRecieve(_:)
  }
  
  private func sendRequest(_ request: UDSRequest, timeout: TimeInterval = NetworkManager.timeout) async throws -> UDSResponse {
    let data = try request.toJSON()
    let resp = try await withCheckedThrowingContinuation { (continuation: CheckedContinuation<UDSResponse, Error>) in
      do {
        try connectionManager.send(data: data)
      } catch {
        continuation.resume(throwing: error)
        return
      }
      self.responseQueue[request.id] = continuation
      self.startTimeoutTimer(for: request.id, timeout: timeout)
    }
    
    return resp
  }

  private func startTimeoutTimer(for requestID: UUID, timeout: TimeInterval) {
    Task.detached {
      _ = Timer(timeInterval: timeout, repeats: false) { [responseQueue = self.responseQueue] _ in
        guard let continuation = responseQueue[requestID] else {
          return
        }
        continuation.resume(throwing: NetworkError.timeout)
      }
    }
  }
  
  private func onRecieve(_ response: UDSResponse) {
    if let cont = self.responseQueue[response.id] {
      cont.resume(returning: response)
    } else {
      logger.logError(NetworkError.unhandledResponse)
      logger.log("Response id: \(response.id) not handled.")
    }
  }
    
  private func handleResponseStatus(_ status: StatusCode) throws {
    switch status {
    case .BadRequest:
      throw NetworkError.badRequest
    case .Duplicate:
      throw NetworkError.duplicate
    case .Internal:
      throw NetworkError.internalError
    case .NotFound:
      throw NetworkError.notFound
    default:
      break
    }
  }
    
  // MARK: - API Methods
    
  func ping() async throws -> Bool {
    let req = UDSRequest(action: .GET, resource: "/", payload: nil)
    let resp = try await sendRequest(req)
    try handleResponseStatus(resp.status)
    return true
  }
    
  func getServer(host: String) async throws -> Server {
    let data = try JSONEncoder().encode(host)
    
    let req = UDSRequest(action: .GET, resource: "/server/host", payload: Payload(data))
    let resp = try await sendRequest(req)
    try handleResponseStatus(resp.status)
    guard let payload = resp.payload else { throw NetworkError.badRequest }
    return try JSONDecoder().decode(Server.self, from: payload)
  }
    
  func updateServerFavorite(host: String, favorite: Bool) async throws -> Bool {
    let data = try JSONEncoder().encode(ServerFavorite(host: host, favorite: favorite))
    
    let req = UDSRequest(action: .PUT, resource: "/server/favorite", payload: Payload(data))
    let resp = try await sendRequest(req)
    try handleResponseStatus(resp.status)
    return true
  }
    
  func addServer(host: String) async throws -> Bool {
    let data = try JSONEncoder().encode(Server(host: host))
    
    let req = UDSRequest(action: .POST, resource: "/server", payload: Payload(data))
    let resp = try await sendRequest(req)
    try handleResponseStatus(resp.status)
    return true
  }
    
  func getAllServers() throws -> [Server] {
    throw HBError.defaultError
  }
}

// MARK: - Helper Structs

struct ServerFavorite: Codable {
  let host: String
  let favorite: Bool
  
  enum CodingKeys: String, CodingKey {
    case host
    case favorite
  }
}
