package main

import (
	"log"
	"net/http"
	"net/url"

	"github.com/ashpect/revproxy/pkg/cache"
	"github.com/ashpect/revproxy/pkg/client"
	"github.com/ashpect/revproxy/pkg/config"
	"github.com/ashpect/revproxy/pkg/proxy"
	"github.com/ashpect/revproxy/pkg/utils"
)

func main() {

	// Load configs
	systemCfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load system config: %v", err)
	}
	proxyCfg := systemCfg.ProxyCfg
	cacheCfg := systemCfg.CacheCfg
	utils.Debug("config: %+v", systemCfg)

	// Transport and client builders
	transport := client.NewTransport(
		client.WithMaxIdleConns(proxyCfg.MaxIdleConns),
		client.WithMaxIdleConnsPerHost(proxyCfg.MaxIdleConnsPerHost),
		client.WithIdleConnTimeout(proxyCfg.IdleConnTimeout),
	)
	client := client.NewClient(
		client.WithTransport(transport),
	)

	upstreamURL, err := url.Parse(proxyCfg.UpstreamURL)
	if err != nil {
		log.Fatalf("invalid upstream URL: %v", err)
	}

	// Cache builder
	cache, err := cache.NewLRUTTL(
		cache.WithCapacity[string, *proxy.CachedResponse](cacheCfg.CacheCapacity),
		cache.WithDefaultTTL[string, *proxy.CachedResponse](cacheCfg.DefaultTTL),
		cache.WithCleanupStart[string, *proxy.CachedResponse](true),
	)
	if err != nil {
		log.Fatalf("failed to create cache: %v", err)
	}

	// Proxyhandler builder
	proxyHandler := proxy.NewProxy(upstreamURL, client, proxy.WithCache(cache))

	// Initialize the server
	server := &http.Server{
		Addr:    systemCfg.ListenAddr,
		Handler: proxyHandler,
	}
	utils.Log("reverse proxy listening on %s forwarding to %s", systemCfg.ListenAddr, upstreamURL.String())
	utils.Log("server starting...")

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}
