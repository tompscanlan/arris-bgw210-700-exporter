package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type ipv4Stats struct {
	ReceivePackets    int
	TransmitPackets   int
	ReceiveBytes      int
	TransmitBytes     int
	ReceiveUnicast    int
	TransmitUnicast   int
	ReceiveMulticast  int
	TransmitMulticast int
	ReceiveDrops      int
	TransmitDrops     int
	ReceiveErrors     int
	TransmitErrors    int
	Collisions        int
}

func recordMetrics(router_addr *string) {
	go func() {
		var ipv4Stats, last_ipv4Stats ipv4Stats

		for {
			probeCounter.Inc()
			response, err := http.Get(fmt.Sprintf("http://%s/cgi-bin/broadbandstatistics.ha", *router_addr))
			if err != nil {
				probeSuccessGauge.Set(0)
				log.Printf("request failed: %s\n", err)
				response.Body.Close()
				continue
			} else {
				probeSuccessGauge.Set(1)
			}

			probeStatusCodeGauge.Set(float64(response.StatusCode))
			if response.StatusCode != 200 {
				probeSuccessGauge.Set(0)
				log.Printf("status code error: %d %s\n", response.StatusCode, response.Status)
				response.Body.Close()
				continue
			}

			// Load the HTML document
			doc, err := goquery.NewDocumentFromReader(response.Body)
			if err != nil {
				log.Printf("failed to query doc: %s\n", err)
				response.Body.Close()
				continue
			}
			fmt.Println("ipv4")
			// Find the review items
			doc.Find("#content-sub table[summary=\"Ethernet IPv4 Statistics Table\"] tbody tr").Each(func(i int, s *goquery.Selection) {

				// For each item found, get the field name and the data
				label := s.Find("th").Text()
				value := s.Find("td").Text()
				fmt.Printf("found %d: %s = %s\n", i, label, value)

				switch label {
				case "Receive Packets":
					ipv4Stats.ReceivePackets, err = strconv.Atoi(value)
					if err != nil {
						log.Printf("failed to convert value to integer: %s\n", err)
					}

					change := ipv4Stats.ReceivePackets - last_ipv4Stats.ReceivePackets
					attBroadbandIpv4RxPackets.Add(float64(change))
				case "Transmit Packets":
					ipv4Stats.TransmitPackets, err = strconv.Atoi(value)
					if err != nil {
						log.Printf("failed to convert value to integer: %s\n", err)
					}
					change := ipv4Stats.TransmitPackets - last_ipv4Stats.TransmitPackets
					attBroadbandIpv4TxPackets.Add(float64(change))

				case "Receive Bytes":
					ipv4Stats.ReceiveBytes, err = strconv.Atoi(value)
					if err != nil {
						log.Printf("failed to convert value to integer: %s\n", err)
					}
					change := ipv4Stats.ReceiveBytes - last_ipv4Stats.ReceiveBytes
					attBroadbandIpv4RxBytes.Add(float64(change))

				case "Transmit Bytes":
					ipv4Stats.TransmitBytes, err = strconv.Atoi(value)
					if err != nil {
						log.Printf("failed to convert value to integer: %s\n", err)
					}

					change := ipv4Stats.TransmitBytes - last_ipv4Stats.TransmitBytes
					attBroadbandIpv4TxBytes.Add(float64(change))

				case "Receive Unicast":
					ipv4Stats.ReceiveUnicast, err = strconv.Atoi(value)
					if err != nil {
						log.Printf("failed to convert value to integer: %s\n", err)
					}
					change := ipv4Stats.ReceiveUnicast - last_ipv4Stats.ReceiveUnicast
					attBroadbandIpv4RxUnicast.Add(float64(change))

				case "Transmit Unicast":
					ipv4Stats.TransmitUnicast, err = strconv.Atoi(value)
					if err != nil {
						log.Printf("failed to convert value to integer: %s\n", err)
					}
					change := ipv4Stats.TransmitUnicast - last_ipv4Stats.TransmitUnicast
					attBroadbandIpv4TxUnicast.Add(float64(change))

				case "Receive Multicast":
					ipv4Stats.ReceiveMulticast, err = strconv.Atoi(value)
					if err != nil {
						log.Printf("failed to convert value to integer: %s\n", err)
					}
					change := ipv4Stats.ReceiveMulticast - last_ipv4Stats.ReceiveMulticast
					attBroadbandIpv4RxMulticast.Add(float64(change))

				case "Transmit Multicast":
					ipv4Stats.TransmitMulticast, err = strconv.Atoi(value)
					if err != nil {
						log.Printf("failed to convert value to integer: %s\n", err)
					}
					change := ipv4Stats.TransmitMulticast - last_ipv4Stats.TransmitMulticast
					attBroadbandIpv4TxMulticast.Add(float64(change))

				case "Receive Drops":
					ipv4Stats.ReceiveDrops, err = strconv.Atoi(value)
					if err != nil {
						log.Printf("failed to convert value to integer: %s\n", err)
					}
					change := ipv4Stats.ReceiveDrops - last_ipv4Stats.ReceiveDrops
					attBroadbandIpv4RxDrops.Add(float64(change))

				case "Transmit Drops":
					ipv4Stats.TransmitDrops, err = strconv.Atoi(value)
					if err != nil {
						log.Printf("failed to convert value to integer: %s\n", err)
					}
					change := ipv4Stats.TransmitDrops - last_ipv4Stats.TransmitDrops
					attBroadbandIpv4TxDrops.Add(float64(change))

				case "Receive Errors":
					ipv4Stats.ReceiveErrors, err = strconv.Atoi(value)
					if err != nil {
						log.Printf("failed to convert value to integer: %s\n", err)
					}
					change := ipv4Stats.ReceiveErrors - last_ipv4Stats.ReceiveErrors
					attBroadbandIpv4RxErrors.Add(float64(change))

				case "Transmit Errors":
					ipv4Stats.TransmitErrors, err = strconv.Atoi(value)
					if err != nil {
						log.Printf("failed to convert value to integer: %s\n", err)
					}
					change := ipv4Stats.TransmitErrors - last_ipv4Stats.TransmitErrors
					attBroadbandIpv4TxErrors.Add(float64(change))

				case "Collisions":
					ipv4Stats.Collisions, err = strconv.Atoi(value)
					if err != nil {
						log.Printf("failed to convert value to integer: %s\n", err)
					}
					change := ipv4Stats.Collisions - last_ipv4Stats.Collisions
					attBroadbandIpv4Collisions.Add(float64(change))

				}

			})

			last_ipv4Stats = ipv4Stats

			time.Sleep(5 * time.Second)
		}
	}()
}

var (
	listen_addr = flag.String("listen-address", "0.0.0.0:9085", "The address to listen on for HTTP requests.")
	router_addr = flag.String("router-address", "192.168.1.254", "The address of the router to scrape")

	deviceName  = "Arris BGW210-700"
	metricsPath = "/metrics"

	probeCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "probe_total",
		Help: "The total number of probe requests",
	})
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

	attBroadbandIpv4TxPackets = promauto.NewCounter(prometheus.CounterOpts{
		Name: "att_broadband_ipv4_transmit_packets",
		Help: "The total number of packets transmitted on the wan interface",
	})

	attBroadbandIpv4RxBytes = promauto.NewCounter(prometheus.CounterOpts{
		Name: "att_broadband_ipv4_receive_bytes",
		Help: "The total number of bytes received on the wan interface",
	})

	attBroadbandIpv4TxBytes = promauto.NewCounter(prometheus.CounterOpts{
		Name: "att_broadband_ipv4_transmit_bytes",
		Help: "The total number of bytes transmitted on the wan interface",
	})

	attBroadbandIpv4RxUnicast = promauto.NewCounter(prometheus.CounterOpts{
		Name: "att_broadband_ipv4_receive_unicast",
		Help: "The total number of unicast packets received on the wan interface",
	})

	attBroadbandIpv4TxUnicast = promauto.NewCounter(prometheus.CounterOpts{
		Name: "att_broadband_ipv4_transmit_unicast",
		Help: "The total number of unicast packets transmitted on the wan interface",
	})

	attBroadbandIpv4RxMulticast = promauto.NewCounter(prometheus.CounterOpts{
		Name: "att_broadband_ipv4_receive_multicast",
		Help: "The total number of multicast packets received on the wan interface",
	})

	attBroadbandIpv4TxMulticast = promauto.NewCounter(prometheus.CounterOpts{
		Name: "att_broadband_ipv4_transmit_multicast",
		Help: "The total number of multicast packets transmitted on the wan interface",
	})

	attBroadbandIpv4RxDrops = promauto.NewCounter(prometheus.CounterOpts{
		Name: "att_broadband_ipv4_receive_drops",
		Help: "The total number of dropped packets received on the wan interface",
	})

	attBroadbandIpv4TxDrops = promauto.NewCounter(prometheus.CounterOpts{

		Name: "att_broadband_ipv4_transmit_drops",
		Help: "The total number of dropped packets transmitted on the wan interface",
	})

	attBroadbandIpv4RxErrors = promauto.NewCounter(prometheus.CounterOpts{
		Name: "att_broadband_ipv4_receive_errors",
		Help: "The total number of errors received on the wan interface",
	})

	attBroadbandIpv4TxErrors = promauto.NewCounter(prometheus.CounterOpts{
		Name: "att_broadband_ipv4_transmit_errors",
		Help: "The total number of errors transmitted on the wan interface",
	})

	attBroadbandIpv4Collisions = promauto.NewCounter(prometheus.CounterOpts{
		Name: "att_broadband_ipv4_collisions",
		Help: "The total number of collisions on the wan interface",
	})
)

func main() {
	flag.Parse()

	//prometheus.MustRegister(probeSuccessGauge)
	//prometheus.MustRegister(probeStatusCodeGauge)
	// prometheus.MustRegister(attBroadbandIpv4RxPackets)

	recordMetrics(router_addr)

	http.HandleFunc("/", defaultPageHandler)
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(*listen_addr, nil)
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
