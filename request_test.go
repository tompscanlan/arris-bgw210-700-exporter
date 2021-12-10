package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"testing"
	"time"

	"github.com/go-kit/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/prometheus/common/promlog"
)

func init() {

	promlogConfig := &promlog.Config{}
	logger = promlog.New(promlogConfig)

}

func expectMetrics(t *testing.T, c prometheus.Collector, fixture string) {
	exp, err := os.Open(path.Join("test_data", fixture))
	if err != nil {
		t.Fatalf("Error opening fixture file %q: %v", fixture, err)
	}
	if err := testutil.CollectAndCompare(c, exp); err != nil {
		t.Fatal("Unexpected metrics returned:", err)
	}
}

func TestErrorResponse(t *testing.T) {
	server := newServerWithResponseFromString(t, "Not the right response, or response with no parsable data", 500)
	defer server.Close()

	e, err := NewExporter(server.URL, 5*time.Second, log.NewNopLogger())
	if err != nil {
		t.Errorf("Expected nil error")
		return
	}

	expectMetrics(t, e, "server_error_response.metrics")
}

func TestUnparsableResponse(t *testing.T) {
	server := newServerWithResponseFromString(t, "Not the right response, or response with no parsable data", 200)
	defer server.Close()

	e, _ := NewExporter(server.URL, 5*time.Second, log.NewNopLogger())

	expectMetrics(t, e, "invalid_response.metrics")
}

func TestGrabDocGoodStatusTestHttp(t *testing.T) {
	server := newServerWithResponseFromFile(t, "broadbandstatistics.ha", 200)
	defer server.Close()
	d, err := grabDocument(server.URL)

	if err != nil {
		t.Errorf("Expected nil error")
		return
	}

	if len(d.Find("body").Nodes) == 0 {
		t.Errorf("Expected body to be present" + d.Find("body").Text())
	}

}

func TestGrabDocBadStatus(t *testing.T) {
	server := newServerWithResponseFromFile(t, "broadbandstatistics.ha", 500)
	defer server.Close()
	d, err := grabDocument(server.URL)

	if err == nil {
		t.Errorf("Expected non-nil error")
		return
	}

	if d != nil {
		t.Errorf("Expected document to be empty")
	}
}

func responseFromLocalFile(t *testing.T, file string) []byte {
	page_reader, err := os.Open(path.Join("test_data", file))
	if err != nil {
		t.Fatalf("Error opening fixture file %q: %v", file, err)
	}
	buf := new(bytes.Buffer)
	buf.ReadFrom(page_reader)
	return buf.Bytes()
}

func getHandler(response []byte, statusCode int) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCode)
		w.Write(response)
	}
}

func newServerWithResponseFromString(t *testing.T, response string, statusCode int) *httptest.Server {
	responseBytes := []byte(response)

	server := httptest.NewServer(
		getHandler(responseBytes, statusCode),
	)
	return server

}

func newServerWithResponseFromFile(t *testing.T, file string, statusCode int) *httptest.Server {
	responseBytes := responseFromLocalFile(t, file)

	server := httptest.NewServer(
		getHandler(responseBytes, statusCode),
	)
	return server

}
