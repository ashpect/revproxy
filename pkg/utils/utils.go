package utils

import (
	"log"
	"net/http"
	"net/url"
)

// PrintRequest logs a request with a title/label in a nicely formatted way
func PrintRequest(req *http.Request, title string) {
	log.Printf("%s", title)
	log.Printf("Method: %s", req.Method)
	log.Printf("URL: %s", req.URL.String())
	log.Printf("Host: %s", req.Host)
	log.Println("Headers:")
	for key, values := range req.Header {
		for _, value := range values {
			log.Printf("    %s: %s", key, value)
		}
	}
}

// PrintRequestWithMetadata logs a request with additional metadata
func PrintRequestWithMetadata(req *http.Request, title string, upstream *url.URL, preserveOriginalHost bool) {
	log.Printf("  PreserveOriginalHost: %v", preserveOriginalHost)
	if upstream != nil {
		log.Printf("  Upstream: %s", upstream.String())
	}
	log.Printf("%s", title)
	log.Printf("Method: %s", req.Method)
	log.Printf("URL: %s", req.URL.String())
	log.Printf("Host: %s", req.Host)
	log.Println("Headers:")
	for key, values := range req.Header {
		for _, value := range values {
			log.Printf("    %s: %s", key, value)
		}
	}
}
