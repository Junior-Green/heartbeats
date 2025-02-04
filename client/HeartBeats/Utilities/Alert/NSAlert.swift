//
//  NSAlert.swift
//  HeartBeats
//
//  Created by Junior Green on 2025-01-27.
//

import AppKit

struct NSAlertItem: Identifiable {
  let id = UUID()
  let alert: NSAlert.Style
  let title: String
  let message: String
  let buttonLabel: String
}

enum NSAlertContext {
  static let agentProcess = NSAlertItem(
    alert: .critical,
    title: "Error starting internal process",
    message: "Something went wrong when trying to start an agent process.",
    buttonLabel: "Exit")

  static let clientSocket = NSAlertItem(
    alert: .critical,
    title: "Cannot instantiate socket connection",
    message: "Something went wrong when initialising socket connection.",
    buttonLabel: "Exit")

  case createFile(filePath: URL)

  var alertItem: NSAlertItem {
    switch self {
    case .createFile(let filePath):
      return NSAlertItem(
        alert: .critical,
        title: "Cannot create file",
        message: "Cannot create file at \(filePath.path()). Please check write permissions or try running as adminstrator.",
        buttonLabel: "Exit")
    }
  }
}

func showNSAlert(item: NSAlertItem) {
  DispatchQueue.main.async {
    let alert = NSAlert()
    alert.messageText = item.title
    alert.addButton(withTitle: item.buttonLabel)
    alert.informativeText = item.message
    alert.alertStyle = item.alert

    alert.runModal()
  }
}

func showNSAlert(err: Error) {
  DispatchQueue.main.async {
    let alert = NSAlert(error: err)
    alert.runModal()
  }
}
