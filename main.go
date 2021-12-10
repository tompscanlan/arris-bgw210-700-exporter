package main

import (
	"net/http"
	"os"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promlog"
	"github.com/prometheus/common/promlog/flag"
	"github.com/prometheus/exporter-toolkit/web"
	webflag "github.com/prometheus/exporter-toolkit/web/kingpinflag"
	"gopkg.in/alecthomas/kingpin.v2"
)

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

var (
	namespace   = "arris"
	deviceName  = "Arris BGW210-700"
	metricsPath = "/metrics"
	version     = "0.0.1"

	logger   log.Logger
	scrapeUp = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "up"),
		"Was the last scrape of the router successful.", nil, nil)
)

type InterfaceMetric int
type metricInfo struct {
	Desc *prometheus.Desc
	Type prometheus.ValueType
}

func main() {
	var (
		webConfig     = webflag.AddFlags(kingpin.CommandLine)
		scrapeURI     = kingpin.Flag("scrape-uri", "URI on router which to scrape.").Default("http://192.168.1.254/cgi-bin/broadbandstatistics.ha").String()
		scrapeTimeout = kingpin.Flag("scrape.timeout", "Timeout for trying to get stats from router.").Default("5s").Duration()
		addr          = kingpin.Flag("listen-address", "The address to listen on for HTTP requests.").Default(":9085").String()
	)

	promlogConfig := &promlog.Config{}
	logger = promlog.New(promlogConfig)

	flag.AddFlags(kingpin.CommandLine, promlogConfig)
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	level.Info(logger).Log("msg", "Starting "+deviceName+" exporter", "version", version)

	exporter, err := NewExporter(*scrapeURI, *scrapeTimeout, logger)
	if err != nil {
		level.Error(logger).Log("msg", "Error creating an exporter", "err", err)
		os.Exit(1)
	}
	prometheus.MustRegister(exporter)

	level.Info(logger).Log("msg", "Listening on address", "address", *addr)
	http.HandleFunc("/", defaultPageHandler)
	http.Handle("/metrics", promhttp.Handler())

	srv := &http.Server{Addr: *addr}
	if err := web.ListenAndServe(srv, *webConfig, logger); err != nil {
		level.Error(logger).Log("msg", "Error starting HTTP server", "err", err)
		os.Exit(1)
	}

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
