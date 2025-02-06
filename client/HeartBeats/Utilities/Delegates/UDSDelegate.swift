import AppKit
import Combine
import Foundation
import Network

protocol UDSClientDelegate: AnyObject {
  func didAcceptSocketConnection()
  func didRecieveData(data: Data?, err: NetworkError?)
  func didFailToAcceptSocketConnection()
  func didSocketConnetionClose()
}

class UDSDelegate: UDSClientDelegate {
  private let logger: Logger
  private var responseQueue: [UUID: Future<UDSResponse, Error>.Promise] = [:]
    
  init(logger: Logger = Logger.shared) {
    self.logger = logger
  }
    
  func didSocketConnetionClose() {
    logger.log("Socket connection closed")
  }
    
  func didFailToAcceptSocketConnection() {
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
