package main

import (
	"log"
	"net/http"
	"net/url"

	"github.com/ashpect/revproxy/pkg/proxy"
)

func main() {
	// TODO : Read from config file
	upstreamURL := &url.URL{
		Scheme: "http",
		Host:   "localhost:9000",
		Path:   "/api",
	}
	listenAddr := ":8000"

	proxyHandler := proxy.New(upstreamURL)

	// Initialize the server
	server := &http.Server{
		Addr:    listenAddr,
		Handler: proxyHandler,
		// TODO : Add other settings to the server.
	}

	log.Println("reverse proxy listening on", listenAddr, " forwarding to", upstreamURL.String())

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}
