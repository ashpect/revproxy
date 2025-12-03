package proxy

import (
	"net/http"
	"time"
)

type CachedResponse struct {
	Status    int
	Header    http.Header
	Body      []byte
	CachedAt  time.Time
	ExpiresAt time.Time
}
