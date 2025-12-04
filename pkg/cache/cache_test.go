package cache

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// CachedResponse represents a cached HTTP response

type CachedResponse struct {
	Status   int
	Header   http.Header
	Body     []byte
	CachedAt time.Time
}

// Map construction and basic GET/SET/DELETE tests
func TestLRUTTL_GET_SET_DELETE(t *testing.T) {
	cache, err := NewLRUTTL(WithCapacity[string, *CachedResponse](100))
	if err != nil {
		t.Fatalf("NewLRUTTL error: %v", err)
	}

	cache.SetWithTTL("key1", &CachedResponse{
		Status: 200,
		Header: http.Header{
			"Content-Type":   []string{"application/json"},
			"Content-Length": []string{"4"},
			"Cache-Control":  []string{"public, max-age=3600"},
		},
		Body: []byte("body"),
	}, 10)

	value, ok := cache.Get_Exclusive("key1")

	assert.True(t, ok)
	assert.Equal(t, value.Status, 200)
	assert.Equal(t, value.Header.Get("Content-Type"), "application/json")
	assert.Equal(t, value.Header.Get("Content-Length"), "4")
	assert.Equal(t, value.Header.Get("Cache-Control"), "public, max-age=3600")
	assert.Equal(t, string(value.Body), "body")

	cache.Delete("key1")
	value, ok = cache.Get_Exclusive("key1")
	assert.False(t, ok)
	assert.Equal(t, (*CachedResponse)(nil), value)
}

// TODO : Add the builder tests with options to cover their unit tests

// Test TTL functionality
func TestLRUTTL_ttl(t *testing.T) {
	cache, err := NewLRUTTL(WithCapacity[string, *CachedResponse](100),
		WithDefaultTTL[string, *CachedResponse](3),
		WithCleanupInterval[string, *CachedResponse](2),
		WithCleanupStart[string, *CachedResponse](true))
	if err != nil {
		t.Fatalf("NewLRUTTL error: %v", err)
	}

	cache.SetWithTTL("key1", &CachedResponse{
		Status: 200,
		Header: http.Header{},
		Body:   []byte("body"),
	}, 3)

	time.Sleep(4 * time.Second)

	_, ok := cache.Get_Exclusive("key1")
	assert.False(t, ok)
}

// Tests LRU eviction logic
func TestLRUTTL_eviction(t *testing.T) {
	// single capacity, so eviction should happen
	cache, err := NewLRUTTL(WithCapacity[string, *CachedResponse](2),
		WithDefaultTTL[string, *CachedResponse](2),
		WithCleanupStart[string, *CachedResponse](true),
	)
	if err != nil {
		t.Fatalf("NewLRUTTL error: %v", err)
	}

	cache.SetWithTTL("key1", &CachedResponse{
		Status:   200,
		Header:   http.Header{},
		Body:     []byte("body"),
		CachedAt: time.Now(),
	}, 100)

	// expect second key which evicts the first one
	cache.SetWithTTL("key2", &CachedResponse{
		Status:   200,
		Header:   http.Header{},
		Body:     []byte("body"),
		CachedAt: time.Now(),
	}, 100)

	cache.Get("key1") // Cache hit on key1 so now at front of link list

	cache.SetWithTTL("key3", &CachedResponse{
		Status:   200,
		Header:   http.Header{},
		Body:     []byte("body"),
		CachedAt: time.Now(),
	}, 100)

	// Key 3 should evict key2 as key1 is not least recently used
	_, ok := cache.Get_Exclusive("key1")
	assert.True(t, ok)

	_, ok = cache.Get_Exclusive("key2")
	assert.False(t, ok)

	_, ok = cache.Get_Exclusive("key3")
	assert.True(t, ok)
}

// Test LRU + TTL
func TestLRU_TTL(t *testing.T) {
	cache, err := NewLRUTTL(WithCapacity[string, *CachedResponse](3),
		WithDefaultTTL[string, *CachedResponse](3),
		WithCleanupInterval[string, *CachedResponse](1),
		WithCleanupStart[string, *CachedResponse](true),
	)
	if err != nil {
		t.Fatalf("NewLRUTTL error: %v", err)
	}

	// Add and get key1 and key2
	cache.SetWithTTL("key1", &CachedResponse{
		Status:   200,
		Header:   http.Header{},
		Body:     []byte("body"),
		CachedAt: time.Now(),
	}, 3)

	cache.SetWithTTL("key2", &CachedResponse{
		Status:   200,
		Header:   http.Header{},
		Body:     []byte("body"),
		CachedAt: time.Now(),
	}, 10)

	cache.SetWithTTL("key3", &CachedResponse{
		Status:   200,
		Header:   http.Header{},
		Body:     []byte("body"),
		CachedAt: time.Now(),
	}, 100)

	// Hit key1 and key2
	_, ok := cache.Get("key1")
	assert.True(t, ok)

	_, ok = cache.Get("key2")
	assert.True(t, ok)
	// Now key3 is least recently used

	// Emulating a case when a newer entry comes and size is full, so eviction should happen for key3
	cache.SetWithTTL("key4", &CachedResponse{
		Status:   200,
		Header:   http.Header{},
		Body:     []byte("body4"),
		CachedAt: time.Now(),
	}, 100)

	time.Sleep(4 * time.Second)
	// Sleeps 5 seconds so key1 expires

	// Hence final remaining should be key2 and key4
	_, ok = cache.Get_Exclusive("key1")
	assert.False(t, ok)
	_, ok = cache.Get_Exclusive("key2")
	assert.True(t, ok)
	_, ok = cache.Get_Exclusive("key3")
	assert.False(t, ok)
	_, ok = cache.Get_Exclusive("key4")
	assert.True(t, ok)
}
