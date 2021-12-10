package main

import (
	"io"
	"sync"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

// Exporter collects HAProxy stats from the given URI and exports them using
// the prometheus metrics package.
type Exporter struct {
	URI       string
	mutex     sync.RWMutex
	fetchStat func() (io.ReadCloser, error)

	up           prometheus.Gauge
	totalScrapes prometheus.Counter
	logger       log.Logger
}

// NewExporter returns an initialized Exporter.
func NewExporter(uri string, timeout time.Duration, logger log.Logger) (*Exporter, error) {

	level.Info(logger).Log("msg", "Initializing exporter", "uri", uri, "timeout", timeout)

	return &Exporter{
		fetchStat: fetchHTTP(uri, false, timeout),
		URI:       uri,
		logger:    logger,
		up: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "up",
			Help:      "Up/Down status of the last scrape.",
		}),
		totalScrapes: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "exporter_total_scrapes",
			Help:      "Current total scrapes.",
		}),
	}, nil
}

// Describe describes all the metrics ever exported. It
// implements prometheus.Collector.
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	// for _, m := range frontendMetrics {
	// 	ch <- m.Desc
	// }
	// for _, m := range backendMetrics {
	// 	ch <- m.Desc
	// }
	// for _, m := range e.serverMetrics {
	// 	ch <- m.Desc
	// }

	ch <- e.up.Desc()
	ch <- e.totalScrapes.Desc()
}

// Collect fetches the stats from the routerand delivers them
// as Prometheus metrics. It implements prometheus.Collector.
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	e.mutex.Lock() // To protect metrics from concurrent collects.
	defer e.mutex.Unlock()

	up := e.scrape(ch)

	ch <- prometheus.MustNewConstMetric(scrapeUp, prometheus.GaugeValue, up)
	ch <- e.totalScrapes

}
