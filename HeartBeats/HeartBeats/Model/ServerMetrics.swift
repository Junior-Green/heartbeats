//
//  ServerMetrics.swift
//  HeartBeats
//
//  Created by Junior Green on 2025-01-11.
//

import Foundation

struct ServerMetrics {
  let online: Bool
  let latency: [LatencyPoint]
  let packetLoss: [PacketLossPoint]
  let throughput: [ThroughputPoint]
  let dnsResolveTime: [DNSResolveTimePoint]
  let statusCode: [StatusCodePoint]
}

struct LatencyPoint {
  let latency: Int
  let at: Date
}

struct PacketLossPoint {
  let percentage: Float
  let at: Date
}

struct ThroughputPoint {
  let kilobytes: Int
  let at: Date
}

struct DNSResolveTimePoint {
  let milliseconds: Int
  let at: Date
}

struct StatusCodePoint {
  let code: Int
  let at: Date
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
