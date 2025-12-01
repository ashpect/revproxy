package main

import (
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/ashpect/revproxy/pkg/proxy"
	"github.com/ashpect/revproxy/pkg/client"
)

func main() {
	// TODO : Read from config file
	upstreamURL := &url.URL{
		Scheme: "http",
		Host:   "localhost:9000",
		Path:   "/api",
	}
	listenAddr := ":8000"
	maxIdleConns := 100
	maxIdleConnsPerHost := 100
	idleConnTimeout := 10 * time.Second

	transport := client.NewTransport(
		client.WithMaxIdleConns(maxIdleConns),
		client.WithMaxIdleConnsPerHost(maxIdleConnsPerHost),
		client.WithIdleConnTimeout(idleConnTimeout),
	)
	client := client.NewClient(
		client.WithTransport(transport),
	)
	proxyHandler := proxy.New(upstreamURL, client)

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
