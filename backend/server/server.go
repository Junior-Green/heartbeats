package server

import (
	"time"

	"github.com/guregu/null/v5"
)

type Metrics struct {
	Latency     []LatencyMarker     `json:"latency"`
	PacketLoss  []PacketLossMarker  `json:"packet_loss"`
	Throughput  []ThroughputMarker  `json:"throughput"`
	DnsResolved []DnsResolvedMarker `json:"dns_resolved"`
	StatusCode  []StatusCodeMarker  `json:"status_code"`
}

type Server struct {
	Hostname string  `json:"hostname"`
	Online   bool    `json:"online"`
	Favorite bool    `json:"favorite"`
	Metrics  Metrics `json:"metrics"`
}

type LatencyMarker struct {
	Date    time.Time `json:"date"`
	Latency null.Int  `json:"latency"`
}

type PacketLossMarker struct {
	Date       time.Time  `json:"date"`
	PacketLoss null.Float `json:"packet_loss"`
}

type ThroughputMarker struct {
	Date       time.Time  `json:"date"`
	Throughput null.Float `json:"throughput"`
}

type DnsResolvedMarker struct {
	Date        time.Time `json:"date"`
	DnsResolved null.Int  `json:"dns_resolved"`
}

type StatusCodeMarker struct {
	Date       time.Time `json:"date"`
	StatusCode null.Int  `json:"status_code"`
}
