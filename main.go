package main

import (
	"flag"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func getIntMetricsMapFromDocumentSection(doc *goquery.Document, selector string) map[string]int64 {

	metrics := make(map[string]int64)

	doc.Find(selector).Each(func(_ int, s *goquery.Selection) {
		key := s.Find("th").Text()
		key = strings.ToLower(key)
		key = strings.ReplaceAll(key, " ", "_")
		counter, err := strconv.Atoi(s.Find("tr").Text())

		if key != "" && err == nil {
			metrics[key] = int64(counter)
		} else {
			log.Printf("%s = %d or %s", key, counter, err)
		}
	})

	return metrics
}

func recordMetrics() {

	go func() {
		ipv4Stats := make(map[string]int64)
		ipv6Stats := make(map[string]int64)

		for {

			doc, err := grabDocument("http://192.168.1.254/cgi-bin/broadbandstatistics.ha")
			if err != nil {
				log.Printf("failed to grab doc: %s\n", err)
				continue
			}

			// Grab the ipv4 and ipv6 stats section
			ipv4StatsCurrent := getIntMetricsMapFromDocumentSection(doc, "#content-sub table[summary=\"Ethernet IPv4 Statistics Table\"]")
			ipv6StatsCurrent := getIntMetricsMapFromDocumentSection(doc, "#content-sub table[summary=\"IPv6 Statistics Table\"]")

			for k, v := range ipv4StatsCurrent {
				if stat, ok := ipv4Stats[k]; ok {
					_ = samplesToIncrement(stat, v)
				}
			}
			for k, v := range ipv6StatsCurrent {
				if stat, ok := ipv6Stats[k]; ok {
					_ = samplesToIncrement(stat, v)
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
