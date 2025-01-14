package servermetrics

import (
	"time"

	"github.com/guregu/null/v5"
	ping "github.com/prometheus-community/pro-bing"
)

const interval time.Duration = 3 * time.Second
const packetCount int = 20

type PingData struct {
	Date        time.Time
	Latency     null.Int   //Milliseconds
	PacketLoss  null.Float //Percentage
	Throughput  null.Float //Bytes per second (Bps)
	DnsResolved null.Int   //Milliseconds
	StatusCode  null.Int   //HTTP
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

func Ping(host string) (<-chan *PingData, error) {
	pinger, err := newPinger(host)
	if err != nil {
		return nil, err
	}

	c := make(chan *PingData)

	go func(c chan<- *PingData, pinger *ping.Pinger) {
		defer close(c)

		err := pinger.Run() // Blocks until finished.
		if err != nil {
			c <- nil
			return
		}

		stats := pinger.Statistics()

		c <- &PingData{
			Date:        time.Now(),
			Latency:     null.IntFrom(stats.AvgRtt.Milliseconds()),
			PacketLoss:  null.FloatFrom(stats.PacketLoss),
			// Throughput:  null.FloatFrom(stats.AvgRtt.Milliseconds()),
			// DnsResolved: null.IntFrom(int(stats.AvgRtt.Milliseconds())),
			// StatusCode:  null.IntFrom(int(stats.AvgRtt.Milliseconds())),
		}
	}(c, pinger)

	return c, nil
}

func newPinger(host string) (*ping.Pinger, error) {
	pinger, err := ping.NewPinger(host)
	if err != nil {
		return nil, err
	}

	pinger.Count = packetCount
	pinger.Interval = interval

	return pinger, nil
}
