package ping

import (
	"net/http"
	"testing"
	"time"

	"github.com/guregu/null/v5"
	ping "github.com/prometheus-community/pro-bing"
	"github.com/stretchr/testify/assert"
)

func TestNewPinger(t *testing.T) {
	tests := []struct {
		name        string
		host        string
		packetSize  int
		packetCount int
		interval    time.Duration
		timeout     time.Duration
		wantErr     bool
	}{
		{
			name:        "Valid host",
			host:        "google.com",
			packetSize:  8,
			packetCount: 3,
			interval:    time.Second,
			timeout:     time.Second * 10,
			wantErr:     false,
		},
		{
			name:        "Invalid host",
			host:        "invalid.host",
			packetSize:  8,
			packetCount: 3,
			interval:    time.Second,
			timeout:     time.Second * 10,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pinger, err := newPinger(tt.host, tt.packetCount, tt.packetSize, tt.interval, tt.timeout)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, pinger)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, pinger)
				assert.Equal(t, tt.packetCount, pinger.Count)
				assert.Equal(t, tt.interval, pinger.Interval)
				assert.Equal(t, tt.packetSize, pinger.Size)
				assert.Equal(t, tt.timeout, pinger.Timeout)
				assert.False(t, pinger.RecordRtts)
				assert.False(t, pinger.RecordTTLs)
				assert.False(t, pinger.Debug)
			}
		})
	}
}

func TestCalculateThroughput(t *testing.T) {
	tests := []struct {
		name       string
		totalBytes int64
		trtt       int64
		want       float64
	}{
		{
			name:       "Zero TRTT",
			totalBytes: 1024,
			trtt:       0,
			want:       0,
		},
		{
			name:       "Non-zero TRTT",
			totalBytes: 1024,
			trtt:       10,
			want:       819.2,
		},
		{
			name:       "Large totalBytes and TRTT",
			totalBytes: 1048576,
			trtt:       1000,
			want:       8388.608,
		},
		{
			name:       "Small totalBytes and TRTT",
			totalBytes: 1,
			trtt:       1,
			want:       8,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateThroughput(tt.totalBytes, tt.trtt)
			assert.Equal(t, tt.want, got)
		})
	}
}
func TestHandlePacket(t *testing.T) {
	tests := []struct {
		name       string
		packet     *ping.Packet
		totalBytes int64
		totalTime  int64
		wantBytes  int64
		wantTime   int64
	}{
		{
			name: "Valid packet",
			packet: &ping.Packet{
				Nbytes: 64,
				Rtt:    10 * time.Millisecond,
			},
			totalBytes: 0,
			totalTime:  0,
			wantBytes:  64,
			wantTime:   10,
		},
		{
			name: "Accumulate packet data",
			packet: &ping.Packet{
				Nbytes: 128,
				Rtt:    20 * time.Millisecond,
			},
			totalBytes: 64,
			totalTime:  10,
			wantBytes:  192,
			wantTime:   30,
		},
		{
			name: "Zero packet data",
			packet: &ping.Packet{
				Nbytes: 0,
				Rtt:    0,
			},
			totalBytes: 100,
			totalTime:  50,
			wantBytes:  100,
			wantTime:   50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handlePacket(tt.packet, &tt.totalBytes, &tt.totalTime)
			assert.Equal(t, tt.wantBytes, tt.totalBytes)
			assert.Equal(t, tt.wantTime, tt.totalTime)
		})
	}
}

func TestEmptyPingData(t *testing.T) {
	data := EmptyPingData()

	assert.False(t, data.Latency.Valid)
	assert.Equal(t, int64(0), data.Latency.Int64)

	assert.False(t, data.PacketLoss.Valid)
	assert.Equal(t, 0.0, data.PacketLoss.Float64)

	assert.False(t, data.Throughput.Valid)
	assert.Equal(t, 0.0, data.Throughput.Float64)

	assert.False(t, data.DnsResolveTime.Valid)
	assert.Equal(t, int64(0), data.DnsResolveTime.Int64)

	assert.False(t, data.StatusCode.Valid)
	assert.Equal(t, int64(0), data.StatusCode.Int64)

	assert.False(t, data.Rtt.Valid)
	assert.Equal(t, int64(0), data.Rtt.Int64)

	assert.WithinDuration(t, time.Now(), data.Date, time.Second)
}
func TestCreatePingData(t *testing.T) {
	tests := []struct {
		name      string
		pingStats *pingStats
		httpStats *httpStats
		want      PingData
	}{
		{
			name:      "Both stats nil",
			pingStats: nil,
			httpStats: nil,
			want: PingData{
				Latency:        null.NewInt(0, false),
				PacketLoss:     null.NewFloat(0, false),
				Throughput:     null.NewFloat(0, false),
				DnsResolveTime: null.NewInt(0, false),
				Rtt:            null.NewInt(0, false),
				StatusCode:     null.NewInt(0, false),
			},
		},
		{
			name: "Only pingStats provided",
			pingStats: &pingStats{
				Latency:        100,
				PacketLoss:     0.5,
				Throughput:     1000,
				Rtt:            6,
				DnsResolveTime: 5,
			},
			httpStats: nil,
			want: PingData{
				Latency:        null.IntFrom(100),
				PacketLoss:     null.FloatFrom(0.5),
				Throughput:     null.FloatFrom(1000),
				DnsResolveTime: null.IntFrom(5),
				Rtt:            null.IntFrom(6),
				StatusCode:     null.NewInt(0, false),
			},
		},
		{
			name:      "Only httpStats provided",
			pingStats: nil,
			httpStats: &httpStats{
				StatusCode: http.StatusOK,
			},
			want: PingData{
				Latency:        null.NewInt(0, false),
				PacketLoss:     null.NewFloat(0, false),
				Throughput:     null.NewFloat(0, false),
				DnsResolveTime: null.NewInt(0, false),
				Rtt:            null.NewInt(0, false),
				StatusCode:     null.IntFrom(http.StatusOK),
			},
		},
		{
			name: "Both stats provided",
			pingStats: &pingStats{
				Latency:        100,
				PacketLoss:     0.5,
				Throughput:     1000,
				Rtt:            4,
				DnsResolveTime: 50,
			},
			httpStats: &httpStats{
				StatusCode: http.StatusOK,
			},
			want: PingData{
				Latency:        null.IntFrom(100),
				PacketLoss:     null.FloatFrom(0.5),
				Throughput:     null.FloatFrom(1000),
				DnsResolveTime: null.IntFrom(50),
				Rtt:            null.IntFrom(4),
				StatusCode:     null.IntFrom(http.StatusOK),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := createPingData(tt.pingStats, tt.httpStats)
			assert.Equal(t, tt.want.Latency, got.Latency)
			assert.Equal(t, tt.want.PacketLoss, got.PacketLoss)
			assert.Equal(t, tt.want.Throughput, got.Throughput)
			assert.Equal(t, tt.want.DnsResolveTime, got.DnsResolveTime)
			assert.Equal(t, tt.want.Rtt, got.Rtt)
			assert.Equal(t, tt.want.StatusCode, got.StatusCode)
		})
	}
}

func TestPing(t *testing.T) {
	tests := []struct {
		name      string
		host      string
		reachable bool
		wantErr   bool
	}{
		{
			name:      "Valid host",
			host:      "google.com",
			reachable: true,
			wantErr:   false,
		},
		{
			name:      "Invalid host",
			host:      "invalid.host",
			reachable: false,
			wantErr:   true,
		},
		{
			name:      "Empty host",
			host:      "",
			reachable: false,
			wantErr:   true,
		},
		{
			name:      "Localhost",
			host:      "localhost",
			reachable: false,
			wantErr:   true,
		},
		{
			name:      "IP address",
			host:      "8.8.8.8",
			reachable: true,
			wantErr:   true,
		},
		{
			name:      "Non-existent domain",
			host:      "nonexistent.domain",
			reachable: false,
			wantErr:   true,
		},
		{
			name:      "Timeout host",
			host:      "10.255.255.1", // Assuming this IP will timeout
			reachable: false,
			wantErr:   false,
		},
		{
			name:      "HTTPS URL",
			host:      "https://example.com",
			reachable: false,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Ping(tt.host)
			select {
			case result := <-c:
				printPingData(t, result)

				assert.WithinDuration(t, time.Now(), result.Date, time.Second*3)
				if tt.wantErr {
					assert.False(t, result.DnsResolveTime.Valid && result.Latency.Valid && result.PacketLoss.Valid && result.Throughput.Valid && result.StatusCode.Valid)
				} else if !tt.reachable {
					assert.Zero(t, result.DnsResolveTime.Int64)
					assert.Zero(t, result.Latency.Int64)
					assert.Zero(t, result.PacketLoss.Float64)
					assert.Zero(t, result.Throughput.Float64)
					assert.Zero(t, result.Rtt.Int64)
					assert.Zero(t, result.StatusCode.Int64)
				} else {
					assert.GreaterOrEqual(t, result.Latency.Int64, int64(0))
					assert.Less(t, result.PacketLoss.Float64, 100.0)
					assert.Greater(t, result.Throughput.Float64, 0.0)
					assert.Greater(t, result.Rtt.Int64, int64(0))
					assert.Greater(t, result.DnsResolveTime.Int64, int64(0))
					assert.Equal(t, int64(200), result.StatusCode.Int64)
				}

			case <-time.After(10 * time.Second):
				if tt.reachable {
					t.Fatal("Test timed out")
				}
			}
		})
	}
}

func TestCollectHttpsStats(t *testing.T) {
	tests := []struct {
		name       string
		host       string
		statusCode int
		wantErr    bool
		timeout    bool
	}{
		{
			name:       "Valid host",
			host:       "http://example.com",
			statusCode: http.StatusOK,
			wantErr:    false,
			timeout:    false,
		},
		{
			name:       "No protocol scheme",
			host:       "example.com",
			statusCode: http.StatusOK,
			wantErr:    false,
			timeout:    false,
		},
		{
			name:       "Invalid host",
			host:       "http://invalid.host",
			statusCode: 0,
			wantErr:    true,
			timeout:    false,
		},
		{
			name:       "Valid HTTPS host",
			host:       "https://example.com",
			statusCode: http.StatusOK,
			wantErr:    false,
			timeout:    false,
		},
		{
			name:       "Non-existent domain",
			host:       "http://nonexistent.domain",
			statusCode: 0,
			wantErr:    true,
			timeout:    false,
		},
		{
			name:       "IP address",
			host:       "8.8.8.8",
			statusCode: 200,
			wantErr:    false,
			timeout:    false,
		},
		{
			name:       "Timeout host",
			host:       "https://10.255.255.1", // Assuming this IP will timeout
			statusCode: 0,
			wantErr:    true,
			timeout:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := make(chan *httpStats)
			go collectHttpsStats(c, tt.host)

			select {
			case result := <-c:
				if tt.wantErr {
					assert.Nil(t, result)
				} else {
					assert.NotNil(t, result)
					assert.Equal(t, tt.statusCode, result.StatusCode)
					t.Logf("StatusCode: %d", result.StatusCode)
				}
			case <-time.After(5 * time.Second):
				if !tt.timeout {
					t.Fatal("Test timed out")
				}
			}
		})
	}
}

func TestCollectPingStats(t *testing.T) {
	tests := []struct {
		name       string
		host       string
		wantErr    bool
		reachable  bool
		packetLoss float64
	}{
		{
			name:       "Valid host",
			host:       "google.com",
			wantErr:    false,
			reachable:  true,
			packetLoss: 0.0,
		},
		{
			name:       "Invalid host",
			host:       "invalid.host",
			wantErr:    true,
			reachable:  false,
			packetLoss: 100.0,
		},
		{
			name:       "Localhost",
			host:       "localhost",
			wantErr:    false,
			reachable:  false,
			packetLoss: 0.0,
		},
		{
			name:       "IP address",
			host:       "8.8.8.8",
			wantErr:    false,
			reachable:  true,
			packetLoss: 0.0,
		},
		{
			name:       "Valid HTTPS host",
			host:       "https://example.com",
			wantErr:    true,
			reachable:  false,
			packetLoss: 0.0,
		},
		{
			name:       "Non-existent domain",
			host:       "nonexistent.domain",
			wantErr:    true,
			reachable:  false,
			packetLoss: 100.0,
		},
		{
			name:       "Timeout host",
			host:       "10.255.255.1", // Assuming this IP will timeout
			wantErr:    true,
			reachable:  false,
			packetLoss: 100.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := make(chan *pingStats)
			go collectPingStats(c, tt.host)

			select {
			case result := <-c:
				printPingStats(t, result)
				if tt.wantErr {
					assert.Nil(t, result)
				} else if !tt.reachable {
					assert.NotNil(t, result)
					assert.Zero(t, result.Latency)
					assert.Zero(t, result.PacketLoss)
					assert.Zero(t, result.Throughput)
					assert.Zero(t, result.Rtt)
					assert.GreaterOrEqual(t, result.DnsResolveTime, int64(0))
				} else {
					assert.NotNil(t, result)
					assert.GreaterOrEqual(t, result.Latency, int64(0))
					assert.Less(t, result.PacketLoss, 100.0)
					assert.Greater(t, result.Throughput, 0.0)
					assert.Greater(t, result.Rtt, int64(0))
					assert.GreaterOrEqual(t, result.DnsResolveTime, int64(0))
				}
			case <-time.After(10 * time.Second):
				if !tt.wantErr {
					t.Fatal("Test timed out")
				}
			}
		})
	}
}

func TestDnsResolveTime(t *testing.T) {
	tests := []struct {
		name    string
		host    string
		wantErr bool
	}{
		{
			name:    "Valid host",
			host:    "google.com",
			wantErr: false,
		},
		{
			name:    "Invalid host",
			host:    "invalid.host",
			wantErr: true,
		},
		{
			name:    "Empty host",
			host:    "",
			wantErr: true,
		},
		{
			name:    "Valid HTTPS host",
			host:    "https://example.com",
			wantErr: true,
		},
		{
			name:    "www subdomain",
			host:    "www.example.com",
			wantErr: false,
		},
		{
			name:    "Localhost",
			host:    "localhost",
			wantErr: false,
		},
		{
			name:    "IP address",
			host:    "8.8.8.8",
			wantErr: false,
		},
		{
			name:    "Non-existent domain",
			host:    "nonexistent.domain",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			duration, err := DnsResolveTime(tt.host)
			t.Logf("Resolved time: %v", duration)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Zero(t, duration)
			} else {
				assert.NoError(t, err)
				assert.Greater(t, duration, time.Duration(0))
			}
		})
	}
}

func TestPingAfter(t *testing.T) {
	tests := []struct {
		name     string
		host     string
		interval time.Duration
		wantErr  bool
	}{
		{
			name:     "Valid host with interval",
			host:     "google.com",
			interval: time.Second,
			wantErr:  false,
		},
		{
			name:     "Invalid host with interval",
			host:     "invalid.host",
			interval: time.Second,
			wantErr:  true,
		},
		{
			name:     "Empty host with interval",
			host:     "",
			interval: time.Second,
			wantErr:  true,
		},
		{
			name:     "Localhost with interval",
			host:     "localhost",
			interval: time.Second,
			wantErr:  true,
		},
		{
			name:     "IP address with interval",
			host:     "8.8.8.8",
			interval: time.Second,
			wantErr:  false,
		},
		{
			name:     "Non-existent domain with interval",
			host:     "nonexistent.domain",
			interval: time.Second,
			wantErr:  true,
		},
		{
			name:     "Timeout host with interval",
			host:     "10.255.255.1", // Assuming this IP will timeout
			interval: time.Second,
			wantErr:  true,
		},
		{
			name:     "HTTPS URL with interval",
			host:     "https://example.com",
			interval: time.Second,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := PingAfter(tt.host, tt.interval)

			for i := 0; i < 3; i++ {
				select {
				case result := <-c:
					printPingData(t, result)

					assert.WithinDuration(t, time.Now(), result.Date, time.Second*3)
					if tt.wantErr {
						assert.False(t, result.DnsResolveTime.Valid && result.Latency.Valid && result.PacketLoss.Valid && result.Throughput.Valid && result.StatusCode.Valid)
					} else {
						assert.GreaterOrEqual(t, result.Latency.Int64, int64(0))
						assert.Less(t, result.PacketLoss.Float64, 100.0)
						assert.Greater(t, result.Throughput.Float64, 0.0)
						assert.GreaterOrEqual(t, result.DnsResolveTime.Int64, int64(0))
						assert.Greater(t, result.Rtt.Int64, int64(0))
						assert.Equal(t, int64(200), result.StatusCode.Int64)
					}

				case <-time.After(10 * time.Second):
					if !tt.wantErr {
						t.Fatal("Test timed out")
					}
				}
			}

		})
	}
}

func printPingData(t *testing.T, p PingData) {
	t.Logf("Latency: %d ms", p.Latency.Int64)
	t.Logf("PacketLoss: %.2f %%", p.PacketLoss.Float64)
	t.Logf("Throughput: %.2f bps", p.Throughput.Float64)
	t.Logf("DnsResolveTime: %d ms", p.DnsResolveTime.Int64)
	t.Logf("Rtt: %d ms", p.Rtt.Int64)
	t.Logf("StatusCode: %d", p.StatusCode.Int64)
}

func printPingStats(t *testing.T, p *pingStats) {
	if p == nil {
		t.Log("<nil>")
		return
	}

	t.Logf("Latency: %d ms", p.Latency)
	t.Logf("PacketLoss: %.2f %%", p.PacketLoss)
	t.Logf("Throughput: %.2f bps", p.Throughput)
	t.Logf("Rtt: %d ms", p.Rtt)
	t.Logf("DNS Resolve time: %d ms", p.DnsResolveTime)
}
