// Package servermetrics provides functionality to measure and collect server metrics such as latency, packet loss, and throughput using ICMP ping.
package ping

import (
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/Junior-Green/heartbeats/logger"
	"github.com/guregu/null/v5"
	ping "github.com/prometheus-community/pro-bing"
)

const interval time.Duration = time.Second * 1
const packetCount int = 3
const packetSize int = 512 //Bytes
const timeout = time.Second * 10

type PingData struct {
	Date           time.Time
	Latency        null.Int   //Milliseconds
	PacketLoss     null.Float //Percentage
	Throughput     null.Float //Bits per second (bps)
	DnsResolveTime null.Int   //Millisecond
	Rtt            null.Int   //Milliseconds
	StatusCode     null.Int   //HTTP
}

type httpStats struct {
	StatusCode int
}

type pingStats struct {
	Latency        int64   //Milliseconds
	PacketLoss     float64 //Percentage
	Throughput     float64 //Bits per second (bps)
	DnsResolveTime int64   //Milliseconds
	Rtt            int64   //Milliseconds
}

// BlankPingData initializes and returns a pointer to a PingData struct
// with default values. The default values are:
//
//   - Date: current time
//   - Latency: null integer with value 0 and validity set to false
//   - PacketLoss: null float with value 0 and validity set to false
//   - Throughput: null float with value 0 and validity set to false
//   - DnsResolved: null integer with value 0 and validity set to false
//   - Rtt: null integer with value 0 and validity set to false
//   - StatusCode: null integer with value 0 and validity set to false
func EmptyPingData() PingData {
	return PingData{
		Date:           time.Now(),
		Latency:        null.NewInt(0, false),
		PacketLoss:     null.NewFloat(0, false),
		Throughput:     null.NewFloat(0, false),
		DnsResolveTime: null.NewInt(0, false),
		Rtt:            null.NewInt(0, false),
		StatusCode:     null.NewInt(0, false),
	}
}

// Ping initiates a ping to the specified host and returns a channel that
// receives PingData pointers. It returns an error if the pinger cannot be created.
//
// Parameters:
//   - host: The hostname or IP address to ping.
//
// Returns:
//   - (<-chan *PingData): A receive-only channel that will receive PingData pointers.
func Ping(host string) <-chan PingData {
	pingChan := make(chan *pingStats)
	httpChan := make(chan *httpStats)

	go collectPingStats(pingChan, host)
	go collectHttpsStats(httpChan, host)

	c := make(chan PingData)

	go func() {
		defer close(c)
		defer close(pingChan)
		defer close(httpChan)

		var pingStats *pingStats
		var httpStats *httpStats
		var httpChanDone, pingChanDone bool

		for {
			select {
			case pingStats = <-pingChan:
				pingChanDone = true
			case httpStats = <-httpChan:
				httpChanDone = true
			}

			if httpChanDone && pingChanDone {
				// Combine data from both channels and send to c
				c <- createPingData(pingStats, httpStats)
				return
			}
		}
	}()

	return c
}

// PingAfter sends ping data to a channel at regular intervals.
//
// Parameters:
//   - host: The host to ping.
//   - interval: The duration between each ping.
//
// Returns:
//
//	A read-only channel of PingData that receives ping results at the specified interval.
func PingAfter(host string, interval time.Duration) <-chan PingData {
	c := make(chan PingData)

	go func() {
		for range time.Tick(interval) {
			c <- <-Ping(host)
		}
	}()

	return c
}

// handlePacket processes a ping packet and updates the total bytes and total time.
//
// Parameters:
//   - pkt: A pointer to a ping.Packet containing the packet data.
//   - totalBytes: A pointer to an int64 that accumulates the total number of bytes.
//   - totalTime: A pointer to an int64 that accumulates the total round-trip time in milliseconds.
func handlePacket(pkt *ping.Packet, totalBytes *int64, totalTime *int64) {
	if pkt == nil {
		logger.Debug("Nil packet received")
		return
	}
	*totalBytes += int64(pkt.Nbytes)
	*totalTime += pkt.Rtt.Milliseconds()
}

// collectPingStats collects ping statistics for a given host and sends the results to the provided channel.
// It resolves the DNS for the host, creates a pinger, and handles received packets to calculate statistics.
//
// Parameters:
//   - c: A channel to send the ping statistics.
//   - host: The host to ping.
//
// The function calculates the following statistics:
//   - Latency: Average round-trip time in milliseconds.
//   - PacketLoss: Percentage of lost packets.
//   - Throughput: Calculated throughput based on total bytes and total round-trip time.
//   - Rtt: Average round-trip time in milliseconds.
//   - DnsResolved: Time taken to resolve the DNS in milliseconds.
//
// If there is an error during DNS resolution, pinger creation, or pinging, the function sends nil to the channel
// and logs the error using the logger.
func collectPingStats(c chan<- *pingStats, host string) {
	var (
		totalBytes int64
		trtt       int64
	)

	resolveTime, err := DnsResolveTime(host)
	if err != nil {
		c <- nil
		logger.Debugf("Error resolving host: %v", err)
		return
	}

	pinger, err := newPinger(host, packetCount, packetSize, interval, timeout)
	if err != nil {
		c <- nil
		logger.Debugf("Error creating pinger: %v", err)
		return
	}

	pinger.OnRecv = func(pkt *ping.Packet) {
		handlePacket(pkt, &totalBytes, &trtt)
	}
	pinger.OnDuplicateRecv = func(pkt *ping.Packet) {
		handlePacket(pkt, &totalBytes, &trtt)
	}

	if err := pinger.Run(); err != nil {
		c <- nil
		logger.Debugf("Error pinging host: %v", err)
		return
	}

	stats := pinger.Statistics()

	c <- &pingStats{
		Latency:        stats.AvgRtt.Milliseconds(),
		PacketLoss:     stats.PacketLoss,
		Throughput:     calculateThroughput(totalBytes, trtt),
		Rtt:            stats.AvgRtt.Milliseconds(),
		DnsResolveTime: resolveTime.Milliseconds(),
	}
}

// calculateThroughput calculates the throughput given the total bytes transferred
// and the total round-trip time (TRTT).
//
// Parameters:
//   - totalBytes: The total number of bytes transferred.
//   - trtt: The total round-trip time in milliseconds.
//
// Returns:
//   - The throughput in bits per millisecond. If trtt is zero, returns 0 to avoid division by zero.
func calculateThroughput(totalBytes, trtt int64) float64 {
	if trtt == 0 {
		return 0
	}
	return float64(totalBytes*8) / float64(trtt)
}

// collectHttpsData collects HTTPS data for a given host and sends the results to a channel.
// It measures the DNS resolution time and captures the HTTP status code.
//
// Parameters:
//   - c: A channel to send the HTTP statistics.
//   - host: The target host URL.
//
// The function performs the following steps:
//  1. Creates a new HTTP GET request for the given host.
//  2. Adds a "no-cache" header to the request.
//  3. Sets up HTTP trace to measure DNS resolution time.
//  4. Configures an HTTP client with a custom transport and timeout settings.
//  5. Executes the HTTP request and captures the response.
//  6. Sends the DNS resolution time and HTTP status code to the provided channel.
//
// If an error occurs at any step, the function logs the error and sends nil to the channel.
func collectHttpsStats(c chan<- *httpStats, host string) {
	if !strings.HasPrefix(host, "http") {
		host = "https://" + host
	}

	dialer := &net.Dialer{Timeout: 5 * time.Second}
	// For security's sake, sorry, I can't provide detailed domain names
	// json content
	req, err := http.NewRequest("GET", host, nil)
	if err != nil {
		logger.Debugf("Error creating HTTP request: %v", err)
		c <- nil
		return
	}

	req.Header.Add("cache-control", "no-cache")

	transport := &http.Transport{

		DialContext:         dialer.DialContext,
		Dial:                dialer.Dial,
		TLSHandshakeTimeout: 2 * time.Second,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   time.Duration(10000) * time.Millisecond,
	}

	res, err := client.Do(req)
	if err != nil {
		c <- nil
		logger.Debugf("HTTP request failed: %v", err)
		return
	}
	defer res.Body.Close()

	c <- &httpStats{res.StatusCode}
}

// newPinger creates a new ping.Pinger instance configured with the specified parameters.
//
// Parameters:
// - host: The target host to ping.
// - packetCount: The number of ICMP packets to send.
// - packetSize: The size of each ICMP packet in bytes.
// - interval: The interval between each packet.
// - timeout: The maximum duration to wait for a response.
//
// Returns:
// - *ping.Pinger: A pointer to the configured ping.Pinger instance.
// - error: An error if the pinger could not be created.
func newPinger(host string, packetCount, packetSize int, interval, timeout time.Duration) (*ping.Pinger, error) {
	pinger, err := ping.NewPinger(host)

	if err != nil {
		logger.Debugf("Error creating pinger: %v", err)
		return nil, err
	}

	pinger.Count = packetCount
	pinger.Interval = interval
	pinger.Size = packetSize
	pinger.Timeout = timeout
	pinger.RecordRtts = false
	pinger.RecordTTLs = false
	pinger.Debug = false

	return pinger, nil
}

// createPingData generates a PingData struct from the provided ping and HTTP statistics.
// It initializes an empty PingData struct and populates it with values from the provided
// pingStats and httpStats if they are not nil.
//
// Parameters:
//   - pingStats: A pointer to a pingStats struct containing latency, packet loss, rtt, and throughput data.
//   - httpStats: A pointer to an httpStats struct containing HTTP status code and DNS resolve time.
//
// Returns:
//
//	A PingData struct populated with the relevant data from pingStats and httpStats.
func createPingData(pingStats *pingStats, httpStats *httpStats) PingData {
	data := EmptyPingData()

	if pingStats != nil {
		data.Latency = null.IntFrom(pingStats.Latency)
		data.PacketLoss = null.FloatFrom(pingStats.PacketLoss)
		data.Throughput = null.FloatFrom(pingStats.Throughput)
		data.Rtt = null.IntFrom(pingStats.Rtt)
		data.DnsResolveTime = null.IntFrom(pingStats.DnsResolveTime)
	}

	if httpStats != nil {
		data.StatusCode = null.IntFrom(int64(httpStats.StatusCode))
	}

	return data
}

func DnsResolveTime(host string) (time.Duration, error) {
	start := time.Now()
	_, err := net.LookupIP(host)
	if err != nil {
		logger.Debugf("Error resolving host: %v", host)
		return time.Duration(0), err
	}
	return time.Since(start), nil
}
