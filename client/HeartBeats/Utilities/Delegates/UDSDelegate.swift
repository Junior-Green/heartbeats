import AppKit
import Combine
import Foundation
import Network

protocol UDSClientDelegate: AnyObject {
  func didAcceptSocketConnection()
  func didRecieveData(data: Data?, err: NetworkError?)
  func didFailToAcceptSocketConnection(err: NetworkError)
  func didSocketConnetionClose()
}

class UDSDelegate: UDSClientDelegate {
  private let logger: Logger
  private static let timeout: TimeInterval = 3
  private var responseQueue: [UUID: Future<UDSResponse, Error>.Promise] = [:]
    
  init(logger: Logger = Logger.shared) {
    self.logger = logger
  }
    
  func didSocketConnetionClose() {
    logger.log("Socket connection closed")
  }
    
  func didFailToAcceptSocketConnection(err: NetworkError) {
    DispatchQueue.main.async {
      showNSAlert(item: NSAlertContext.clientSocket)
      NSApplication.shared.terminate(nil)
    }
  }
    
  func didAcceptSocketConnection() {
    logger.log("Socket connection accepted")
  }
  
  func queueRequest(requestId: UUID, promise: @escaping Future<UDSResponse, Error>.Promise) {
    responseQueue.updateValue(promise, forKey: requestId)
    
    Task {
      let _ = Timer(timeInterval: UDSDelegate.timeout, repeats: false, block: { [responseQueue] _ in
        guard let promise = responseQueue[requestId] else {
          return
        }
        
        promise(.failure(NetworkError.timeout))
      })
    }
  }
    
  func didRecieveData(data: Data?, err: NetworkError?) {
    guard let data = data, err == nil else {
      logger.logError(err!)
      return
    }

    do {
      let resp = try JSONDecoder().decode(UDSResponse.self, from: data)
      guard let promise = responseQueue[resp.id] else { return }
      promise(.success(resp))
      responseQueue.removeValue(forKey: resp.id)
    } catch {
      logger.logError(error)
    }
  }
}
