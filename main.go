package main

import (
	"flag"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"time"
)

func recordMetrics() {
	go func() {
		for {
			response, err := http.Get("http://192.168.1.254/cgi-bin/broadbandstatistics.haX")
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
				log.Printf("status code error: %d %s\n", response.StatusCode, response.Status)
				response.Body.Close()
				continue
			}

			// Load the HTML document
			doc, err := goquery.NewDocumentFromReader(response.Body)
			if err != nil {
				log.Printf("failed to query doc: %s\n",err)
				response.Body.Close()
				continue
			}
			fmt.Println("ipv4")
			// Find the review items
			doc.Find("#content-sub table[summary=\"Ethernet IPv4 Statistics Table\"]").Each(func(i int, s *goquery.Selection) {

				// For each item found, get the title
				label := s.Find("th").Text()
				value := s.Find("tr").Text()
				fmt.Printf("found %d: %s = %s\n", i, label, value)
			})

			fmt.Println("ipv6")
			doc.Find("#content-sub table[summary=\"IPv6 Statistics Table\"]").Each(func(i int, s *goquery.Selection) {

				// For each item found, get the title
				label := s.Find("th").Text()
				value := s.Find("tr").Text()
				fmt.Printf("found %d: %s = %s\n", i, label, value)
			})

			time.Sleep(2 * time.Second)
		}
	}()
}

var (
	addr                      = flag.String("listen-address", ":9085", "The address to listen on for HTTP requests.")
	deviceName                = "Arris BGW210-700"
	metricsPath               = "/metrics"
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
