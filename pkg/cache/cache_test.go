package cache

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// CachedResponse represents a cached HTTP response

type CachedResponse struct {
	Status    int
	Header    http.Header
	Body      []byte
	CachedAt  time.Time
	ExpiresAt time.Time
}

func TestLRUTTL_GET_SET_DELETE(t *testing.T) {
	cache, err := NewLRUTTL[string, CachedResponse](100)
	if err != nil {
		t.Fatalf("NewLRUTTL error: %v", err)
	}

	cache.Set("key1", CachedResponse{
		Status: 200,
		Header: http.Header{
			"Content-Type":   []string{"application/json"},
			"Content-Length": []string{"4"},
			"Cache-Control":  []string{"public, max-age=3600"},
		},
		Body:      []byte("body"),
		CachedAt:  time.Now(),
		ExpiresAt: time.Now().Add(10 * time.Second),
	})

	value, ok := cache.Get("key1")

	assert.True(t, ok)
	assert.Equal(t, value.Status, 200)
	assert.Equal(t, value.Header.Get("Content-Type"), "application/json")
	assert.Equal(t, value.Header.Get("Content-Length"), "4")
	assert.Equal(t, value.Header.Get("Cache-Control"), "public, max-age=3600")
	assert.Equal(t, string(value.Body), "body")

	cache.Delete("key1")
	value, ok = cache.Get("key1")
	assert.False(t, ok)
	assert.Equal(t, CachedResponse{}, value)
}

func TestLRUTTL_eviction(t *testing.T) {
	// single capacity, so eviction should happen
	cache, err := NewLRUTTL[string, CachedResponse](1,
		WithDefaultTTL[string, CachedResponse](3),
		WithCleanupStart[string, CachedResponse](true),
	)
	if err != nil {
		t.Fatalf("NewLRUTTL error: %v", err)
	}

	cache.Set("key1", CachedResponse{
		Status:    200,
		Header:    http.Header{},
		Body:      []byte("body"),
		CachedAt:  time.Now(),
		ExpiresAt: time.Now().Add(10 * time.Second),
	})

	// expect second key which evicts the first one
	cache.Set("key2", CachedResponse{
		Status:    200,
		Header:    http.Header{},
		Body:      []byte("body"),
		CachedAt:  time.Now(),
		ExpiresAt: time.Now().Add(10 * time.Second),
	})

	_, ok := cache.Get("key1")
	assert.False(t, ok)

	_, ok = cache.Get("key2")
	assert.True(t, ok)
}
