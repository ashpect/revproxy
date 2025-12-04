package proxy

import (
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/ashpect/revproxy/pkg/cache"
	"github.com/ashpect/revproxy/pkg/utils"
)

type proxy struct {
	upstream             *url.URL
	client               *http.Client
	preserveOriginalHost bool
	cache                cache.Cache[string, *CachedResponse]
}

type ProxyOption func(*proxy)

func WithPreserveOriginalHost(preserve bool) ProxyOption {
	return func(p *proxy) {
		p.preserveOriginalHost = preserve
	}
}

func WithClient(client *http.Client) ProxyOption {
	return func(p *proxy) {
		p.client = client
	}
}

func WithCache(cache cache.Cache[string, *CachedResponse]) ProxyOption {
	return func(p *proxy) {
		p.cache = cache
	}
}

func NewProxy(upstream *url.URL, client *http.Client, opts ...ProxyOption) *proxy {
	p := &proxy{
		upstream:             upstream,
		client:               client,
		preserveOriginalHost: false,
		cache:                nil,
	}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

func (p *proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Only cache GET requests
	isCacheable := r.Method == http.MethodGet
	uniqueKey := p.getUniqueReqKey(r)

	if isCacheable && p.cache != nil {
		utils.Debug("Checking cache for key: %s", uniqueKey)
		cachedResp, ok := p.cache.Get(uniqueKey)
		if ok {
			utils.Debug("Cache hit for key: %s", uniqueKey)
			utils.Debug("Serving cached response for key: %s", uniqueKey)
			p.serveCachedResponse(w, cachedResp)
			utils.Debug("Cached response served for key: %s", uniqueKey)
			return
		}
	}
	utils.Debug("Cache miss for key: %s", uniqueKey)
	outReq, err := p.buildUpstreamRequest(r)
	if err != nil {
		http.Error(w, "bad upstream request", http.StatusInternalServerError)
		log.Printf("build upstream request error: %v", err)
		return
	}

	resp, err := p.client.Do(outReq)

	// TODO : Better error handling
	if err != nil {
		http.Error(w, "upstream error", http.StatusBadGateway)
		log.Printf("upstream request error: %v", err)
		return
	}
	defer resp.Body.Close()

	removeHopByHopHeaders(resp.Header)

	// Read the response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("error reading response body: %v", err)
		http.Error(w, "error reading response", http.StatusInternalServerError)
		return
	}

	// Copy headers to response writer
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value) // key is case insensitive
		}
	}

	// Create cached response and store in cache if cache is available and request is GET
	ttl := parseMaxAge(resp.Header.Get("Cache-Control"))
	if isCacheable && p.cache != nil {
		cachedResp := &CachedResponse{
			Status:   resp.StatusCode,
			Header:   resp.Header.Clone(),
			Body:     bodyBytes,
			CachedAt: time.Now(),
		}
		utils.Debug("Caching response for key: %s with ttl: %d", uniqueKey, ttl)
		if ttl > 0 {
			p.cache.SetWithTTL(uniqueKey, cachedResp, ttl)
		} else {
			p.cache.Set(uniqueKey, cachedResp) // use default TTL
		}
		utils.Debug("Cached response stored for key: %s", uniqueKey)
	}

	done := make(chan bool)
	go func() {
		select {
		case <-time.Tick(10 * time.Millisecond):
			w.(http.Flusher).Flush()
		case <-done:
			return
		}
	}()
	w.WriteHeader(resp.StatusCode)

	// Write the body to response writer
	if _, err := w.Write(bodyBytes); err != nil {
		log.Printf("error writing response body: %v", err)
	}
	close(done)
}

func (p *proxy) getUniqueReqKey(r *http.Request) string {
	return r.URL.String()
}

func (p *proxy) serveCachedResponse(w http.ResponseWriter, cachedResp *CachedResponse) {
	// Copy headers to response writer
	for key, values := range cachedResp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	// Set status code and write body
	w.WriteHeader(cachedResp.Status)
	if _, err := w.Write(cachedResp.Body); err != nil {
		log.Printf("error writing cached response body: %v", err)
	}
}

func (p *proxy) buildUpstreamRequest(req *http.Request) (*http.Request, error) {
	// TESTING
	utils.PrintRequest(req, "Initial request")

	// Clone keeps method, headers, body, context, etc.
	ctx := req.Context()
	outReq := req.Clone(ctx)

	// Rewrite URL to point to upstream
	outReq.URL.Scheme = p.upstream.Scheme
	outReq.URL.Host = p.upstream.Host
	outReq.URL.Path = singleJoiningSlash(p.upstream.Path, req.URL.Path)

	// Required for http.Client.Do
	outReq.RequestURI = ""

	// Set host header
	if p.preserveOriginalHost {
		outReq.Host = req.Host
	} else {
		outReq.Host = p.upstream.Host
	}

	removeHopByHopHeaders(outReq.Header)

	// Add new header
	outReq.Header.Set("X-Forwarded-Host", req.Host)
	proto := "http"
	if req.TLS != nil {
		proto = "https"
	}
	outReq.Header.Set("X-Forwarded-Proto", proto)

	s, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		log.Printf("error splitting host port: %v", err)
	} else {
		outReq.Header.Set("X-Forwarded-For", s)
	}

	// TESTING
	utils.PrintRequestWithMetadata(outReq, "Final request", p.upstream, p.preserveOriginalHost)

	return outReq, nil
}

// parseMaxAge extracts the max-age or s-maxage value from Cache-Control header.
// Returns 0 if not found or invalid.
func parseMaxAge(cacheControl string) int {
	cacheControl = strings.ToLower(cacheControl)
	directives := strings.Split(cacheControl, ",")

	for _, directive := range directives {
		directive = strings.TrimSpace(directive)

		// Check for max-age=value or s-maxage=value
		if strings.HasPrefix(directive, "max-age=") {
			if maxAge, err := strconv.Atoi(strings.TrimPrefix(directive, "max-age=")); err == nil && maxAge > 0 {
				return maxAge
			}
		}
		if strings.HasPrefix(directive, "s-maxage=") {
			if maxAge, err := strconv.Atoi(strings.TrimPrefix(directive, "s-maxage=")); err == nil && maxAge > 0 {
				return maxAge
			}
		}
	}

	return 0
}
