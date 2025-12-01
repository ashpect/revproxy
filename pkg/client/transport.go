package client

import (
	"net/http"
	"time"
)

type TransportOption func(*http.Transport)

func WithMaxIdleConns(maxIdleConns int) TransportOption {
	return func(t *http.Transport) {
		t.MaxIdleConns = maxIdleConns
	}
}

func WithMaxIdleConnsPerHost(maxIdleConnsPerHost int) TransportOption {
	return func(t *http.Transport) {
		t.MaxIdleConnsPerHost = maxIdleConnsPerHost
	}
}

func WithIdleConnTimeout(timeout time.Duration) TransportOption {
	return func(t *http.Transport) {
		t.IdleConnTimeout = timeout
	}
}

func NewTransport(opts ...TransportOption) *http.Transport {
	transport := &http.Transport{}
	for _, opt := range opts {
		opt(transport)
	}
	return transport
}