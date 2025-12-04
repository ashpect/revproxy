package cache

import "time"

// Len returns number of non-expired items.
// Uses read lock since it only reads the map length
func (c *LRUWithTTL[K, V]) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}

// Get returns value if present and not expired
// Marks the element as most-recent
func (c *LRUWithTTL[K, V]) Get(key K) (value V, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var zero V
	element, ok := c.items[key]
	if !ok {
		return zero, false
	}
	entry := element.Value.(*ttlEntry[K, V])

	if isExpired(entry) {
		c.ll.Remove(element)
		delete(c.items, key)
		return zero, false
	}

	c.ll.MoveToFront(element)
	return entry.value, true
}

// GetAll returns a shallow copy of the current contents.
// Uses read lock since it only reads the map
func (c *LRUWithTTL[K, V]) GetAll() map[K]V {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make(map[K]V, len(c.items))
	for k, ele := range c.items {
		out[k] = ele.Value.(*ttlEntry[K, V]).value
	}
	return out
}

// Helper functions for testing, just returns the value based on key without moving them at front
func (c *LRUWithTTL[K, V]) Get_Exclusive(key K) (value V, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	element, ok := c.items[key]
	if !ok {
		var zero V
		return zero, false
	}
	return element.Value.(*ttlEntry[K, V]).value, true
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

// SetWithTTL using default TTL if expiresAt is not set
func (c *LRUWithTTL[K, V]) Set(key K, value V) {
	c.setWithTTLInternal(key, value, time.Now().Add(c.defaultTTL))
}

// SetWithTTL stores value with a specific ttlSeconds
// ttlSeconds = 0 explicitly means no expiry
func (c *LRUWithTTL[K, V]) SetWithTTL(key K, value V, ttlSeconds int) {
	if ttlSeconds > 0 {
		expiresAt := time.Now().Add(time.Duration(ttlSeconds) * time.Second)
		c.setWithTTLInternal(key, value, expiresAt)
	} else if ttlSeconds == 0 { // no expiry
		c.setWithTTLInternal(key, value, time.Time{})
	} else {
		panic("ttlSeconds must be >= 0")
	}
}

// Actual setting
func (c *LRUWithTTL[K, V]) setWithTTLInternal(key K, value V, expiresAt time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// update if it's existing
	if element, ok := c.items[key]; ok {
		entry := element.Value.(*ttlEntry[K, V])
		entry.value = value
		entry.expiresAt = expiresAt
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
		expiresAt: expiresAt,
	}
	element := c.ll.PushFront(entry)
	c.items[key] = element
}

// isExpired checks whether an entry is expired. (expirytime - currenttime)
func isExpired[K comparable, V any](entry *ttlEntry[K, V]) bool {
	if entry.expiresAt.IsZero() {
		return false // zero time means no expiry
	}
	return time.Since(entry.expiresAt) > 0
}

// CRONJOB

// Close stops cleanup cronjob if running.
func (c *LRUWithTTL[K, V]) Close() {
	c.StopCleanupDaemon()
}

// StartCleanupDaemon starts a background goroutine that periodically evicts expired items.
// interval must be > 0 seconds.
func (c *LRUWithTTL[K, V]) StartCleanupDaemon() {
	if c.cleanupInterval <= 0 {
		panic("cleanup interval must be > 0")
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	go func() {
		ticker := time.NewTicker(c.cleanupInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				c.cleanupExpired()
			case <-c.cleanupStop:
				return
			}
		}
	}()
}

// cleanupExpired iterates through the linked list and removes expired entries and also deletes the entry from the map.
func (c *LRUWithTTL[K, V]) cleanupExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()

	current := c.ll.Front()
	for current != nil {
		next := current.Next()
		entry := current.Value.(*ttlEntry[K, V])

		expired := isExpired(entry)
		if expired {
			c.ll.Remove(current)
			delete(c.items, entry.key)
		}

		current = next
	}
}

func (c *LRUWithTTL[K, V]) StopCleanupDaemon() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.cleanupRunning {
		close(c.cleanupStop)
		c.cleanupStop = make(chan struct{})
		c.cleanupRunning = false
	}
}
