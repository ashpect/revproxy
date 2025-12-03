package main

import (
	"log"
	"net/http"
	"net/url"
	"fmt"

	"github.com/ashpect/revproxy/pkg/client"
	"github.com/ashpect/revproxy/pkg/config"
	"github.com/ashpect/revproxy/pkg/proxy"

)

func main() {

	// Load configs
	config.LoadConfig()
	SystemCfg := config.SystemConfig
	ProxyCfg := SystemCfg.Proxy

	fmt.Println("SystemCfg", SystemCfg)

	// Transport and client builders
	transport := client.NewTransport(
		client.WithMaxIdleConns(ProxyCfg.MaxIdleConns),
		client.WithMaxIdleConnsPerHost(ProxyCfg.MaxIdleConnsPerHost),
		client.WithIdleConnTimeout(ProxyCfg.IdleConnTimeout),
	)
	client := client.NewClient(
		client.WithTransport(transport),
	)

	upstreamURL, err := url.Parse(ProxyCfg.UpstreamURL)
	if err != nil {
		log.Fatalf("invalid upstream URL: %v", err)
	}

	// Proxyhandler builder
	proxyHandler := proxy.NewProxy(upstreamURL, client)

	// Initialize the server
	server := &http.Server{
		Addr:    SystemCfg.ListenAddr,
		Handler: proxyHandler,
		// TODO : Add other settings to the server.
	}

	log.Println("reverse proxy listening on", SystemCfg.ListenAddr, " forwarding to", upstreamURL.String())

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}
