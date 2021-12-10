package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

type HttpGetter interface {
	Get(url string) (resp *http.Response, err error)
}

var Client HttpGetter

func init() {
	Client = &http.Client{}
}

func grabDocument(url string) (*goquery.Document, error) {

	response, err := Client.Get(url)
	if err != nil {
		level.Error(logger).Log("error", err)
		if response != nil {
			response.Body.Close()
		}
		return nil, err
	}

	if response.StatusCode != 200 {
		level.Error(logger).Log("statusCode", response.StatusCode, "status", response.Status)
		response.Body.Close()
		return nil, fmt.Errorf("http code error %d", response.StatusCode)
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		level.Error(logger).Log("error", err)
		response.Body.Close()
		return nil, err
	}
	return doc, nil
}

func (e *Exporter) scrape(ch chan<- prometheus.Metric) (up float64) {
	e.totalScrapes.Inc()
	var err error

	body, err := e.fetchStat()
	if err != nil {
		level.Error(e.logger).Log("msg", "Can't scrape router", "err", err)
		return 0
	}
	defer body.Close()

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(body)
	if err != nil {
		level.Error(e.logger).Log("msg", "failed to parse page", "err", err)
		return 0
	}

	ipv4StatsCurrent := getIntMetricsMapFromDocumentSection(doc, "#content-sub table[summary=\"Ethernet IPv4 Statistics Table\"]", (e.logger))
	for k, v := range ipv4StatsCurrent {
		desc := prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "ipv4", k),
			"help",
			[]string{"protocol"},
			nil)

		ch <- prometheus.MustNewConstMetric(
			desc,
			prometheus.CounterValue,
			float64(v),
			"ipv4")
	}

	return 1
}

func fetchHTTP(uri string, sslVerify bool, timeout time.Duration) func() (io.ReadCloser, error) {
	tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: !sslVerify}}
	client := http.Client{
		Timeout:   timeout,
		Transport: tr,
	}

	return func() (io.ReadCloser, error) {
		resp, err := client.Get(uri)
		if err != nil {
			return nil, err
		}
		if !(resp.StatusCode >= 200 && resp.StatusCode < 300) {
			resp.Body.Close()
			return nil, fmt.Errorf("HTTP status %d", resp.StatusCode)
		}
		return resp.Body, nil
	}
}

// func recordMetrics() {

// 	go func() {
// 		ipv4Stats := make(map[string]int64)
// 		ipv6Stats := make(map[string]int64)

// 		for {

// 			doc, err := grabDocument("http://192.168.1.254/cgi-bin/broadbandstatistics.ha")
// 			if err != nil {
// 				level.Error(logger).Log("error", err, "msg", "failed to grab doc")
// 				continue
// 			}

// 			// Grab the ipv4 and ipv6 stats section
// 			ipv4StatsCurrent := getIntMetricsMapFromDocumentSection(doc, "#content-sub table[summary=\"Ethernet IPv4 Statistics Table\"]")
// 			ipv6StatsCurrent := getIntMetricsMapFromDocumentSection(doc, "#content-sub table[summary=\"IPv6 Statistics Table\"]")

// 			for k, v := range ipv4StatsCurrent {
// 				if stat, ok := ipv4Stats[k]; ok {
// 					_ = samplesToIncrement(stat, v)
// 				}
// 			}
// 			for k, v := range ipv6StatsCurrent {
// 				if stat, ok := ipv6Stats[k]; ok {
// 					_ = samplesToIncrement(stat, v)
// 				}
// 			}

// 			time.Sleep(2 * time.Second)

// 		}
// 	}()
// }

func getIntMetricsMapFromDocumentSection(doc *goquery.Document, selector string, logger log.Logger) map[string]int64 {

	metrics := make(map[string]int64)

	doc.Find(selector).Each(func(_ int, s *goquery.Selection) {

		s.Find("tr").Each(func(i int, si *goquery.Selection) {
			label := si.Find("th").Text()
			label = strings.ToLower(label)
			label = strings.ReplaceAll(label, " ", "_")
			value, err := strconv.Atoi(si.Find("td").Text())

			if err != nil {
				level.Debug(logger).Log("msg", fmt.Sprintf("failed to parse value of %s", label), "error", err)
			} else {

				metrics[label] = int64(value)
			}

		})
	})

	return metrics
}
