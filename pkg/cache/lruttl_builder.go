package cache

import (
	"container/list"
	"sync"
	"time"
)

const defaultCapacity = 10
const defaultTTL = 10 * time.Minute
const defaultCleanupInterval = 10 * time.Millisecond

// LRUOption is a functional option for building LRUTTL cache
type LRUOption[K comparable, V any] func(*LRUWithTTL[K, V])

// ttlEntry stored in list.Element
type ttlEntry[K comparable, V any] struct {
	key       K
	value     V
	expiresAt time.Time
}

// LRU cache with TTL based cleanup
type LRUWithTTL[K comparable, V any] struct {
	capacity int
	mu       sync.RWMutex
	ll       *list.List
	items    map[K]*list.Element

	defaultTTL      time.Duration
	cleanupInterval time.Duration

	cleanupStop    chan struct{}
	cleanupRunning bool
}

// WithCapacity sets the capacity of the cache.
func WithCapacity[K comparable, V any](capacity int) LRUOption[K, V] {
	return func(c *LRUWithTTL[K, V]) {
		if capacity > 0 {
			c.capacity = capacity
		} else {
			panic("capacity must be > 0")
		}
	}
}

// WithDefaultTTL sets a default TTL (SECONDS) used by Set().
// DefaultTTL is used to evict in case upstream server response does not have a cache control header
func WithDefaultTTL[K comparable, V any](ttlSeconds int) LRUOption[K, V] {
	return func(c *LRUWithTTL[K, V]) {
		if ttlSeconds >= 0 {
			c.defaultTTL = time.Duration(ttlSeconds) * time.Second
		} else {
			panic("default TTL must be >= 0")
		}
	}
}

// WithCleanupInterval configures automatic cleanup interval (SECONDS). intervalSeconds > 0 for TTL based cleanup
func WithCleanupInterval[K comparable, V any](intervalSeconds int) LRUOption[K, V] {
	return func(c *LRUWithTTL[K, V]) {
		if intervalSeconds > 0 {
			c.cleanupInterval = time.Duration(intervalSeconds) * time.Second
		} else {
			panic("cleanup interval must be > 0")
		}
	}
}

// WithCleanupStart configures whether to start the cleanup cronjob on cache creation.
func WithCleanupStart[K comparable, V any](cleanupRunning bool) LRUOption[K, V] {
	return func(c *LRUWithTTL[K, V]) {
		c.cleanupRunning = cleanupRunning
	}
}

// WithItemsMap configures to use shallow copy of cache from a given items map (use at your own caution)
// TODO : Improve to handle edge cases like calling with capacity post this and also creating a linked list ?
func WithItemsMap[K comparable, V any](itemsMap map[K]*list.Element) LRUOption[K, V] {
	return func(c *LRUWithTTL[K, V]) {
		for k, v := range itemsMap {
			c.items[k] = v
		}
	}
}

// NewLRUTTL creates an LRU cache with TTL based cleanup.
// Capacity must be > 0. Provide options to configure TTL and cleanup interval.
func NewLRUTTL[K comparable, V any](opts ...LRUOption[K, V]) (*LRUWithTTL[K, V], error) {

	c := &LRUWithTTL[K, V]{
		capacity:        defaultCapacity,
		ll:              list.New(),
		items:           make(map[K]*list.Element, defaultCapacity),
		defaultTTL:      defaultTTL,
		cleanupInterval: defaultCleanupInterval,
		cleanupStop:     make(chan struct{}),
		cleanupRunning:  true,
	}

	for _, o := range opts {
		o(c)
	}

	if c.cleanupRunning {
		c.StartCleanupDaemon()
	}
	return c, nil
}
