package client

import (
	"net/http"
	"time"
)

const defaultClientTimeout = 30 * time.Second

type ClientOption func(*http.Client)

func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *http.Client) {
		c.Timeout = timeout
	}
}

func WithTransport(transport *http.Transport) ClientOption {
	return func(c *http.Client) {
		c.Transport = transport
	}
}

func NewClient(opts ...ClientOption) *http.Client {
	client := &http.Client{
		Timeout:   defaultClientTimeout,
		Transport: http.DefaultTransport,
	}
	
	for _, opt := range opts {
		opt(client)
	}
	return client
}