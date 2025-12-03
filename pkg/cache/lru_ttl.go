package cache

import (
	"container/list"
	"errors"
	"fmt"
	"sync"
	"time"
)

const defaultTTL = 10 * time.Minute
const defaultCleanupInterval = 1 * time.Hour

// LRUOption is a functional option for building LRUTTL cache
type LRUOption[K comparable, V any] func(*LRUWithTTL[K, V])

// ttlEntry stored in list.Element
type ttlEntry[K comparable, V any] struct {
	key       K
	value     V
	createdAt time.Time
	ttl       time.Duration
}

// LRU cache with TTL based cleanup
type LRUWithTTL[K comparable, V any] struct {
	mu              sync.Mutex
	capacity        int
	ll              *list.List
	items           map[K]*list.Element
	defaultTTL      time.Duration
	cleanupInterval time.Duration
	cleanupStop     chan struct{}
	cleanupRunning  bool
}

// WithDefaultTTL sets a default TTL (in seconds) used by Set(). if ttlSeconds <= 0 not allowed as it can result in a cache which keeps growing
func WithDefaultTTL[K comparable, V any](ttlSeconds int) LRUOption[K, V] {
	return func(c *LRUWithTTL[K, V]) {
		if ttlSeconds > 0 {
			c.defaultTTL = time.Duration(ttlSeconds) * time.Second
		} else {
			panic("default TTL must be > 0")
		}
	}
}

// WithCleanupInterval configures automatic cleanup interval (seconds). intervalSeconds > 0 for TTL based cleanup
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

// NewLRUTTL creates an LRU cache with TTL based cleanup.
// Capacity must be > 0. Provide options to configure TTL and cleanup interval.
func NewLRUTTL[K comparable, V any](capacity int, opts ...LRUOption[K, V]) (*LRUWithTTL[K, V], error) {
	if capacity <= 0 {
		return nil, errors.New("capacity must be > 0")
	}

	c := &LRUWithTTL[K, V]{
		capacity:        capacity,
		ll:              list.New(),
		items:           make(map[K]*list.Element, capacity),
		defaultTTL:      defaultTTL,
		cleanupInterval: defaultCleanupInterval,
		cleanupStop:     make(chan struct{}),
		cleanupRunning:  true,
	}

	for _, o := range opts {
		o(c)
	}

	if c.cleanupRunning {
		c.StartCleanupDaemon(c.cleanupInterval)
	}
	return c, nil
}

// isExpired checks whether an entry is expired (considers per-entry ttl).
func (c *LRUWithTTL[K, V]) isExpired(entry *ttlEntry[K, V]) (bool, error) {
	var ttl time.Duration
	if entry.ttl > 0 {
		ttl = entry.ttl
	} else if entry.ttl == -1 { // No expiry
		return false, nil
	} else {
		return false, errors.New("ttl must be > 0 or -1 for no expiry")
	}
	return time.Since(entry.createdAt) > ttl, nil
}

// Get returns value if present and not expired; marks as most-recent.
func (c *LRUWithTTL[K, V]) Get(key K) (value V, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var zero V
	element, ok := c.items[key]
	if !ok {
		return zero, false
	}
	entry := element.Value.(*ttlEntry[K, V])
	expired, err := c.isExpired(entry)
	if err != nil {
		// Invalid TTL configuration - treat as expired and remove
		c.ll.Remove(element)
		delete(c.items, key)
		return zero, false
	}
	if expired {
		c.ll.Remove(element)
		delete(c.items, key)
		return zero, false
	}
	c.ll.MoveToFront(element)
	return entry.value, true
}

// SetWithTTL using default TTL
func (c *LRUWithTTL[K, V]) Set(key K, value V) {
	c.setWithTTLInternal(key, value, c.defaultTTL)
}

// Delete removes the key from the cache (both the linked list node and the items map).
func (c *LRUWithTTL[K, V]) Delete(key K) {
	c.mu.Lock()
	defer c.mu.Unlock()
	element, ok := c.items[key]
	if !ok {
		return
	}
	c.ll.Remove(element)
	delete(c.items, key)
}

// SetWithTTL stores value with a specific ttlSeconds
// ttlSeconds = -1 explicitly means no expiry
func (c *LRUWithTTL[K, V]) SetWithTTL(key K, value V, ttlSeconds int) {
	var ttl time.Duration
	if ttlSeconds > 0 {
		ttl = time.Duration(ttlSeconds) * time.Second
	} else if ttlSeconds < 0 {
		ttl = time.Duration(-1) * time.Second // -1 to represent no expiry
	} else {
		panic("ttlSeconds must be > 0 or -1 for no expiry")
	}
	c.setWithTTLInternal(key, value, ttl)
}

// Actual setting
func (c *LRUWithTTL[K, V]) setWithTTLInternal(key K, value V, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// update existing
	if element, ok := c.items[key]; ok {
		entry := element.Value.(*ttlEntry[K, V])
		entry.value = value
		entry.createdAt = time.Now()
		entry.ttl = ttl
		c.ll.MoveToFront(element)
		return
	}

	// if its full, evict to create space
	if len(c.items) >= c.capacity {
		tail := c.ll.Back()
		if tail == nil {
			panic("capacity is 0")
		}
		entry := tail.Value.(*ttlEntry[K, V])
		c.ll.Remove(tail)
		delete(c.items, entry.key)
	}

	// insert new
	entry := &ttlEntry[K, V]{
		key:       key,
		value:     value,
		createdAt: time.Now(),
		ttl:       ttl,
	}
	element := c.ll.PushFront(entry)
	c.items[key] = element

}

// Len returns number of non-expired items.
func (c *LRUWithTTL[K, V]) Len() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.items)
}

// GetAll returns a shallow copy of the current contents.
func (c *LRUWithTTL[K, V]) GetAll() map[K]V {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make(map[K]V, len(c.items))
	for k, ele := range c.items {
		out[k] = ele.Value.(*ttlEntry[K, V]).value
	}
	return out
}

// CRONJOB - todo

// Close stops cleanup cronjob if running.
func (c *LRUWithTTL[K, V]) Close() {
	c.StopCleanupDaemon()
}

// StartCleanupDaemon starts a background goroutine that periodically evicts expired items.
// interval must be > 0 seconds.
func (c *LRUWithTTL[K, V]) StartCleanupDaemon(interval time.Duration) {
	fmt.Println("Starting cleanup daemon")
}

// StopCleanupDaemon stops the janitor if running.
func (c *LRUWithTTL[K, V]) StopCleanupDaemon() {
	fmt.Println("Stopping cleanup daemon")
}
