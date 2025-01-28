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
  static let appSupport = NSAlertItem(alert: .critical,
                                      title: "File Path Error",
                                      message: "Do not have read/write access to /Library/Application Support. Invalid permissions or does not exists.",
                                      buttonLabel: "Exit")

  static let databaseFile = NSAlertItem(alert: .critical,
                                        title: "File Path Error",
                                        message: "Error occured while creating database resources. Invalid file permissions.",
                                        buttonLabel: "Exit")
}

func showNSAlert(item: NSAlertItem) {
  let alert = NSAlert()
  alert.messageText = item.title
  alert.addButton(withTitle: item.buttonLabel)
  alert.informativeText = item.message
  alert.alertStyle = item.alert

  alert.runModal()
}

func showNSAlert(err: Error) {
  let alert = NSAlert(error: err)
  alert.runModal()
}
