package main

import (
	"flag"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func recordMetrics() {

	go func() {
		ipv4Stats := make(map[string]int)
		ipv6Stats := make(map[string]int)

		for {
			ipv4StatsCurrent := make(map[string]int)
			// ipv6StatsCurrent := make(map[string]int)

			doc, err := grabDocument("http://192.168.1.254/cgi-bin/broadbandstatistics.ha")
			if err != nil {
				log.Printf("failed to grab doc: %s\n", err)
				continue
			}

			// fmt.Println("ipv4")
			// Grab the ipv4 stats section
			doc.Find("#content-sub table[summary=\"Ethernet IPv4 Statistics Table\"]").Each(func(i int, s *goquery.Selection) {
				key := s.Find("th").Text()
				log.Printf("found ipv4 key %s\n", key)
				counter, err := strconv.Atoi(s.Find("tr").Text())

				if err != nil {
					log.Printf("err for %s value: %s", key, err)
				} else {
					ipv4Stats[key] = counter
				}
			})

			// fmt.Println("ipv6")
			// Grab the ipv6 stats section
			doc.Find("#content-sub table[summary=\"IPv6 Statistics Table\"]").Each(func(i int, s *goquery.Selection) {
				key := s.Find("th").Text()
				log.Printf("found ipv6 key %s\n", key)
				counter, err := strconv.Atoi(s.Find("tr").Text())

				if err != nil {
					log.Printf("err for %s value: %s", key, err)
				} else {
					ipv6Stats[key] = counter
				}
			})

			for k, v := range ipv4Stats {
				if stat, ok := ipv4StatsCurrent[k]; ok {

					if stat > v {
						//  if the stat counter has wrapped around
						attBroadbandIpv4RxPackets.Add(float64((math.MaxInt64 - v) + stat))
					} else {
						// otherwise, difference between last and current values
						attBroadbandIpv4RxPackets.Add(float64(v - stat))
					}
				}
			}

			time.Sleep(2 * time.Second)

		}
	}()
}

var (
	addr              = flag.String("listen-address", ":9085", "The address to listen on for HTTP requests.")
	deviceName        = "Arris BGW210-700"
	metricsPath       = "/metrics"
	probeSuccessGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "probe_success",
		Help: "Displays whether or not the probe was a success",
	})
	probeStatusCodeGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "probe_status_code",
		Help: "HTTP response status code, where 200 is normal success",
	})

	attBroadbandIpv4RxPackets = promauto.NewCounter(prometheus.CounterOpts{
		Name: "att_broadband_ipv4_receive_packets",
		Help: "The total number of packets received on the wan interface",
	})
)

// Receive Packets 	91399113
// Transmit Packets 	51871120
// Receive Bytes 	118049999769
// Transmit Bytes 	41305517047
// Receive Unicast 	91399112
// Transmit Unicast 	51869821
// Receive Multicast 	-2
// Transmit Multicast 	485
// Receive Drops 	-3
// Transmit Drops 	6015
// Receive Errors 	-3
// Transmit Errors 	-3
// Collisions 	-3

// IPv3 Statistics
// Transmit Packets 	806025
// Transmit Errors 	-3
// Transmit Discards 	8930

type InterfaceMetric int

const (
	RxPackets InterfaceMetric = iota
	TxPackets
	RxBytes
	TxBytes
	RxUnicast
	TxUnicast
	RxMulticast
	TxMulticast
	RxDrops
	TxDrops
	RxErrors
	TxErrors
)

func main() {
	flag.Parse()

	//prometheus.MustRegister(probeSuccessGauge)
	//prometheus.MustRegister(probeStatusCodeGauge)
	//prometheus.MustRegister(attBroadbandIpv4RxPackets)

	recordMetrics()

	http.HandleFunc("/", defaultPageHandler)
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(*addr, nil)
}

func defaultPageHandler(w http.ResponseWriter, _ *http.Request) {
	w.Write([]byte(`<html>
			<head><title>` + deviceName + ` Exporter</title></head>
			<body>
			<h1>` + deviceName + ` Exporter</h1>
			<p><a href="` + metricsPath + `">Metrics</a></p>
			</body>
			</html>`))
}
