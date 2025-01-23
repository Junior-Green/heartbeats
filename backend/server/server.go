package server

import (
	"time"

	"github.com/Junior-Green/heartbeats/logger"
	"github.com/google/uuid"
	"github.com/guregu/null/v5"
)

type Metrics struct {
	Latency     []LatencyMarker     `json:"latency"`
	PacketLoss  []PacketLossMarker  `json:"packet_loss"`
	Throughput  []ThroughputMarker  `json:"throughput"`
	DnsResolved []DnsResolvedMarker `json:"dns_resolved"`
	StatusCode  []StatusCodeMarker  `json:"status_code"`
	Rtt         []RttMarker         `json:"rtt"`
}

type Server struct {
	Id       string `json:"id"`
	Host     string `json:"hostname"`
	Online   bool   `json:"online"`
	Favorite bool   `json:"favorite"`
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

type RttMarker struct {
	Date time.Time `json:"date"`
	Rtt  null.Int  `json:"rtt"`
}

type StatusCodeMarker struct {
	Date       time.Time `json:"date"`
	StatusCode null.Int  `json:"status_code"`
}

// NewServer creates a new Server instance with a unique ID and the specified host.
// It returns the created Server and an error if the ID generation fails.
//
// Parameters:
//   - host: The hostname or IP address for the server.
//
// Returns:
//   - Server: The newly created Server instance.
//   - error: An error if the UUID generation fails.
func NewServer(host string) (Server, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		logger.Debugf("Error generating server ID: %v", err)
		return Server{}, err
	}

	return Server{Id: id.String(), Host: host}, nil
}
