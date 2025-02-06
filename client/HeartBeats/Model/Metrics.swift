//
//  ServerMetrics.swift
//  HeartBeats
//
//  Created by Junior Green on 2025-01-11.
//

import Foundation

protocol Marker {
  var date: Date { get }
}

struct Metrics: Codable {
  let latency: [LatencyMarker]
  let packetLoss: [PacketLossMarker]
  let throughput: [ThroughputMarker]
  let dnsResolveTime: [DnsResolveTimeMarker]
  let statusCode: [StatusCodeMarker]
  let rtt: [RttMarker]

  enum CodingKeys: String, CodingKey {
    case latency
    case packetLoss = "packet_loss"
    case throughput
    case dnsResolveTime = "dns_resolved"
    case statusCode = "status_code"
    case rtt
  }
}

struct LatencyMarker: Codable, Marker {
  let latency: Int
  let date: Date

  enum CodingKeys: String, CodingKey {
    case latency
    case date
  }
}

struct PacketLossMarker: Codable, Marker {
  let packetLoss: Float
  let date: Date

  enum CodingKeys: String, CodingKey {
    case packetLoss = "packet_loss"
    case date
  }
}

struct ThroughputMarker: Codable, Marker {
  let throughput: Float
  let date: Date

  enum CodingKeys: String, CodingKey {
    case throughput
    case date
  }
}

struct DnsResolveTimeMarker: Codable, Marker {
  let dnsResolved: Int
  let date: Date

  enum CodingKeys: String, CodingKey {
    case dnsResolved = "dns_resolved"
    case date
  }
}

struct StatusCodeMarker: Codable, Marker {
  let statusCode: Int
  let date: Date

  enum CodingKeys: String, CodingKey {
    case statusCode = "status_code"
    case date
  }
}

struct RttMarker: Codable, Marker {
  let rtt: Int
  let date: Date

  enum CodingKeys: String, CodingKey {
    case rtt
    case date
  }
}

// TODO: Convert these fields to Models
// Host List (Overview):
//
//    Domain/Host name
//    Ping Status: "Up" (Green) or "Down" (Red)
//    Latency (Average, Min, Max)
//    Packet Loss (%) or Success Rate
//    Uptime (%) or Downtime (Last 24 hours)
//
// Detailed Metrics for Selected Host:
//
//    Ping History (graph of latency over time)
//    Real-time Latency (Current)
//    Uptime/Downtime
