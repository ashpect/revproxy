package proxy

import (
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/ashpect/revproxy/pkg/utils"
)

type proxy struct {
	upstream             *url.URL
	client               *http.Client
	preserveOriginalHost bool
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

func NewProxy(upstream *url.URL, client *http.Client, opts ...ProxyOption) *proxy {
	p := &proxy{
		upstream:             upstream,
		client:               client,
		preserveOriginalHost: false,
	}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

func (p *proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	outReq, err := p.buildUpstreamRequest(r)
	if err != nil {
		http.Error(w, "bad upstream request", http.StatusInternalServerError)
		log.Printf("build upstream request error: %v", err)
		return
	}

	resp, err := p.client.Do(outReq)

	// TODO : Better error handling, differenttiate between error types (eg: timeout, conn ref, etc)
	if err != nil {
		http.Error(w, "upstream error", http.StatusBadGateway)
		log.Printf("upstream request error: %v", err)
		return
	}
	defer resp.Body.Close()

	removeHopByHopHeaders(resp.Header)

	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value) // key is case insensitive
		}
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

	if _, err := io.Copy(w, resp.Body); err != nil {
		log.Printf("error copying response body: %v", err)
	}
	close(done)
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
