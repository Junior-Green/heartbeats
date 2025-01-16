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

	assert.False(t, data.DnsResolved.Valid)
	assert.Equal(t, int64(0), data.DnsResolved.Int64)

	assert.False(t, data.StatusCode.Valid)
	assert.Equal(t, int64(0), data.StatusCode.Int64)

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
				Latency:     null.NewInt(0, false),
				PacketLoss:  null.NewFloat(0, false),
				Throughput:  null.NewFloat(0, false),
				DnsResolved: null.NewInt(0, false),
				StatusCode:  null.NewInt(0, false),
			},
		},
		{
			name: "Only pingStats provided",
			pingStats: &pingStats{
				Latency:    100,
				PacketLoss: 0.5,
				Throughput: 1000,
			},
			httpStats: nil,
			want: PingData{
				Latency:     null.IntFrom(100),
				PacketLoss:  null.FloatFrom(0.5),
				Throughput:  null.FloatFrom(1000),
				DnsResolved: null.NewInt(0, false),
				StatusCode:  null.NewInt(0, false),
			},
		},
		{
			name:      "Only httpStats provided",
			pingStats: nil,
			httpStats: &httpStats{
				DnsResolveTime: 50 * time.Millisecond,
				StatusCode:     http.StatusOK,
			},
			want: PingData{
				Latency:     null.NewInt(0, false),
				PacketLoss:  null.NewFloat(0, false),
				Throughput:  null.NewFloat(0, false),
				DnsResolved: null.IntFrom(50),
				StatusCode:  null.IntFrom(http.StatusOK),
			},
		},
		{
			name: "Both stats provided",
			pingStats: &pingStats{
				Latency:    100,
				PacketLoss: 0.5,
				Throughput: 1000,
			},
			httpStats: &httpStats{
				DnsResolveTime: 50 * time.Millisecond,
				StatusCode:     http.StatusOK,
			},
			want: PingData{
				Latency:     null.IntFrom(100),
				PacketLoss:  null.FloatFrom(0.5),
				Throughput:  null.FloatFrom(1000),
				DnsResolved: null.IntFrom(50),
				StatusCode:  null.IntFrom(http.StatusOK),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := createPingData(tt.pingStats, tt.httpStats)
			assert.Equal(t, tt.want.Latency, got.Latency)
			assert.Equal(t, tt.want.PacketLoss, got.PacketLoss)
			assert.Equal(t, tt.want.Throughput, got.Throughput)
			assert.Equal(t, tt.want.DnsResolved, got.DnsResolved)
			assert.Equal(t, tt.want.StatusCode, got.StatusCode)
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
			name:       "Timeout host",
			host:       "http://10.255.255.1", // Assuming this IP will timeout
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
					assert.Greater(t, result.DnsResolveTime, time.Duration(0))
					t.Logf("StatusCode: %d", result.StatusCode)
					t.Logf("DnsResolveTime: %dms", result.DnsResolveTime.Milliseconds())
				}
			case <-time.After(5 * time.Second):
				if !tt.timeout {
					t.Fatal("Test timed out")
				}
			}
		})
	}
}
func TestCreatePingStats(t *testing.T) {
	tests := []struct {
		name       string
		pStats     *ping.Statistics
		totalBytes int64
		trtt       int64
		want       *pingStats
	}{
		{
			name: "Valid statistics",
			pStats: &ping.Statistics{
				AvgRtt:     100 * time.Millisecond,
				PacketLoss: 0.5,
			},
			totalBytes: 1024,
			trtt:       100,
			want: &pingStats{
				Latency:    100,
				PacketLoss: 0.5,
				Throughput: 81.92,
			},
		},
		{
			name: "Zero TRTT",
			pStats: &ping.Statistics{
				AvgRtt:     50 * time.Millisecond,
				PacketLoss: 0.0,
			},
			totalBytes: 1024,
			trtt:       0,
			want: &pingStats{
				Latency:    50,
				PacketLoss: 0.0,
				Throughput: 0,
			},
		},
		{
			name: "Zero totalBytes",
			pStats: &ping.Statistics{
				AvgRtt:     200 * time.Millisecond,
				PacketLoss: 1.0,
			},
			totalBytes: 0,
			trtt:       100,
			want: &pingStats{
				Latency:    200,
				PacketLoss: 1.0,
				Throughput: 0,
			},
		},
		{
			name: "Zero totalBytes and TRTT",
			pStats: &ping.Statistics{
				AvgRtt:     300 * time.Millisecond,
				PacketLoss: 0.2,
			},
			totalBytes: 0,
			trtt:       0,
			want: &pingStats{
				Latency:    300,
				PacketLoss: 0.2,
				Throughput: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := createPingStats(tt.pStats, tt.totalBytes, tt.trtt)
			assert.Equal(t, tt.want.Latency, got.Latency)
			assert.Equal(t, tt.want.PacketLoss, got.PacketLoss)
			assert.Equal(t, tt.want.Throughput, got.Throughput)
		})
	}
}

func TestPing(t *testing.T) {
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
		{
			name:    "Timeout host",
			host:    "10.255.255.1", // Assuming this IP will timeout
			wantErr: true,
		},
		{
			name:    "HTTPS host",
			host:    "https://example.com",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := Ping(tt.host)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, c)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, c)

				select {
				case result := <-c:
					assert.NotNil(t, result)
					assert.WithinDuration(t, time.Now(), result.Date, time.Second)
					t.Logf("Latency: %d ms", result.Latency.Int64)
					t.Logf("PacketLoss: %.2f %%", result.PacketLoss.Float64)
					t.Logf("Throughput: %.2f bps", result.Throughput.Float64)
					t.Logf("DnsResolved: %d ms", result.DnsResolved.Int64)
					t.Logf("StatusCode: %d", result.StatusCode.Int64)
				case <-time.After(10 * time.Second):
					t.Fatal("Test timed out")
				}
			}
		})
	}
}
