import Combine
import Foundation
import Network

final class NetworkManager {
  private let logger: Logger = .shared
  private let udsDelegate = UDSDelegate()
  private let client: UDSClient
    
  static let shared = NetworkManager()
    
  private static let timeout: Int = 1500 // ms
    
  private init() {
    self.client = UDSClient(socketPath: URL.socketFile.path(), delegate: udsDelegate)
    client.start()
  }
    
  private func sendRequest(_ request: UDSRequest) -> Future<UDSResponse, Error> {
    return Future<UDSResponse, Error> { promise in
      do {
        let requestData = try JSONEncoder().encode(request)
        self.udsDelegate.queueRequest(requestId: request.id, promise: promise)
        self.client.sendData(requestData)
      } catch {
        promise(.failure(error))
      }
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
    let resp = try await sendRequest(req).value
    try handleResponseStatus(resp.status)
    return true
  }
    
  func getServer(host: String) async throws -> Server {
    let data = try JSONEncoder().encode(host)
    
    let req = UDSRequest(action: .GET, resource: "/server/host", payload: Payload(data))
    let resp = try await sendRequest(req).value
    try handleResponseStatus(resp.status)
    guard let payload = resp.payload else { throw NetworkError.badRequest }
    return try JSONDecoder().decode(Server.self, from: payload)
  }
    
  func updateServerFavorite(host: String, favorite: Bool) async throws -> Bool {
    let data = try JSONEncoder().encode(ServerFavorite(host: host, favorite: favorite))
    
    let req = UDSRequest(action: .PUT, resource: "/server/favorite", payload: Payload(data))
    let resp = try await sendRequest(req).value
    try handleResponseStatus(resp.status)
    return true
  }
    
  func addServer(host: String) async throws -> Bool {
    let data = try JSONEncoder().encode(Server(host: host))
    
    let req = UDSRequest(action: .POST, resource: "/server", payload: Payload(data))
    let resp = try await sendRequest(req).value
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
