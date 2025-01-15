// Package servermetrics provides functionality to measure and collect server metrics such as latency, packet loss, and throughput using ICMP ping.
package servermetrics

import (
	"net"
	"net/http"
	"net/http/httptrace"
	"time"

	"github.com/Junior-Green/heartbeats/logger"
	"github.com/guregu/null/v5"
	ping "github.com/prometheus-community/pro-bing"
)

const interval time.Duration = 3 * time.Second
const packetCount int = 20
const packetSize int = 512 //Byte
const timeout = time.Second * 10

type PingData struct {
	Date        time.Time
	Latency     null.Int   //Milliseconds
	PacketLoss  null.Float //Percentage
	Throughput  null.Float //Bits per second (bps)
	DnsResolved null.Int   //Milliseconds
	StatusCode  null.Int   //HTTP
}

type httpStats struct {
	DnsResolveTime time.Duration
	StatusCode     int
}

// BlankPingData initializes and returns a pointer to a PingData struct
// with default values. The default values are:
//
//   - Date: current time
//   - Latency: null integer with value 0 and validity set to false
//   - PacketLoss: null float with value 0 and validity set to false
//   - Throughput: null float with value 0 and validity set to false
//   - DnsResolved: null integer with value 0 and validity set to false
//   - StatusCode: null integer with value 0 and validity set to false
func BlankPingData() *PingData {
	return &PingData{
		Date:        time.Now(),
		Latency:     null.NewInt(0, false),
		PacketLoss:  null.NewFloat(0, false),
		Throughput:  null.NewFloat(0, false),
		DnsResolved: null.NewInt(0, false),
		StatusCode:  null.NewInt(0, false),
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
//   - (error): An error if the pinger cannot be created.
func Ping(host string) (<-chan *PingData, error) {
	c := make(chan *PingData)
	defer close(c)

	go collectPingData(c, host)

	return c, nil
}

func PingAfter(host string, durations time.Duration) <-chan PingData {
	c := make(chan PingData)
	// go func() {
	// 	for {
	// 		time.Sleep(durations)
	// 		c <- PingData{
	// 			Date:        time.Now(),
	// 			Latency:     null.IntFrom(1),
	// 			PacketLoss:  null.FloatFrom(1.0),
	// 			Throughput:  null.IntFrom(1),
	// 			DnsResolved: null.IntFrom(1),
	// 			StatusCode:  null.IntFrom(1),
	// 		}
	// 	}
	// }()
	return c
}

// handlePacket processes a ping packet and updates the total bytes and total time.
//
// Parameters:
//   - pkt: A pointer to a ping.Packet containing the packet data.
//   - totalBytes: A pointer to an int64 that accumulates the total number of bytes.
//   - totalTime: A pointer to an int64 that accumulates the total round-trip time in milliseconds.
func handlePacket(pkt *ping.Packet, totalBytes *int64, totalTime *int64) {
	*totalBytes += int64(pkt.Nbytes)
	*totalTime += pkt.Rtt.Milliseconds()
}

// collectPingData collects ping data by running the provided pinger and sends the results to the given channel.
// It handles received and duplicate packets to accumulate total bytes and total round-trip time (trtt).
// The function blocks until the pinger finishes running, then sends the collected ping data to the channel.
//
// Parameters:
//   - c: A send-only channel to send the collected PingData.
//   - pinger: A pointer to a ping.Pinger instance used to perform the ping operations.
func collectPingData(c chan<- *PingData, host string) {
	pinger, err := newPinger(host)
	if err != nil {
		c <- nil
		return
	}

	var (
		totalBytes int64
		trtt       int64
		pStats     *ping.Statistics
	)

	httpChan := make(chan *httpStats)
	defer close(httpChan)

	go collectHttpsData(httpChan, host)

	pinger.OnRecv = func(pkt *ping.Packet) {
		handlePacket(pkt, &totalBytes, &trtt)
	}
	pinger.OnDuplicateRecv = func(pkt *ping.Packet) {
		handlePacket(pkt, &totalBytes, &trtt)
	}

	if err = pinger.Run(); err != nil {
		pStats = nil
		logger.Debugf("Error running pinger: %v", err)
	}
	pStats = pinger.Statistics()

	c <- createPingData(pStats, totalBytes, trtt, <-httpChan)
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
func calculateThroughput(totalBytes int64, trtt int64) float64 {
	if trtt == 0 {
		return 0
	}
	return float64(totalBytes*8) / float64(trtt)
}

// createPingData creates a new PingData instance with the provided statistics.
// It takes the following parameters:
// - stats: a pointer to ping.Statistics containing the ping statistics.
// - totalBytes: the total number of bytes transmitted.
// - trtt: the total round-trip time in milliseconds.
//
// The function returns a pointer to a PingData struct populated with the current date,
// average round-trip time (latency), packet loss, and throughput.
func createPingData(pStats *ping.Statistics, totalBytes int64, trtt int64, httpStats *httpStats) *PingData {
	if pStats == nil && httpStats == nil {
		return BlankPingData()
	}

	data := PingData{Date: time.Now()}

	if httpStats != nil {
		data.DnsResolved = null.IntFrom(httpStats.DnsResolveTime.Milliseconds())
		data.StatusCode = null.IntFrom(int64(httpStats.StatusCode))
	}

	if pStats != nil {
		data.Latency = null.IntFrom(pStats.AvgRtt.Milliseconds())
		data.PacketLoss = null.FloatFrom(pStats.PacketLoss)
		data.Throughput = null.FloatFrom(calculateThroughput(totalBytes, trtt))
	}

	return &data
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
func collectHttpsData(c chan<- *httpStats, host string) {
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

	var (
		dnsStart time.Time
		dnsEnd   time.Time
	)

	trace := &httptrace.ClientTrace{
		DNSStart: func(dnsInfo httptrace.DNSStartInfo) {
			dnsStart = time.Now()
		},
		DNSDone: func(dnsDoneInfo httptrace.DNSDoneInfo) {
			dnsEnd = time.Now()
		},
	}
	req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))

	transport := &http.Transport{

		DialContext:         (dialer).DialContext,
		Dial:                (dialer).Dial,
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

	c <- &httpStats{dnsEnd.Sub(dnsStart), res.StatusCode}
}

// newPinger creates a new ping.Pinger for the specified host with predefined settings.
// It returns a pointer to the ping.Pinger and an error if the pinger could not be created.
//
// Parameters:
//   - host: The target host to ping.
//
// Returns:
//   - *ping.Pinger: A pointer to the created ping.Pinger.
//   - error: An error if the pinger could not be created.
func newPinger(host string) (*ping.Pinger, error) {
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
