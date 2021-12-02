package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/PuerkitoBio/goquery"
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
		probeSuccessGauge.Set(0)
		log.Printf("request failed: %s\n", err)
		if response != nil {
			response.Body.Close()
		}
		return nil, err
	} else {
		probeSuccessGauge.Set(1)
	}

	probeStatusCodeGauge.Set(float64(response.StatusCode))
	if response.StatusCode != 200 {
		log.Printf("status code error: %d %s\n", response.StatusCode, response.Status)
		response.Body.Close()
		return nil, fmt.Errorf("http code error %d", response.StatusCode)
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		log.Printf("failed to query doc: %s\n", err)
		response.Body.Close()
		return nil, err
	}
	return doc, nil
}
