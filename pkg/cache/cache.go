package cache

type Cache[K comparable, V any] interface {
	// Get returns the value for key and true if present (and not expired).
	Get(key K) (V, bool)

	// Set stores the value for key using the cache's default TTL (if any).
	Set(key K, value V)

	// SetWithTTL stores the value for key with a custom ttl (ttl > 0).
	SetWithTTL(key K, value V, ttlSeconds int)

	// Delete removes the key from the cache.
	Delete(key K)

	// Len returns the number of items currently stored (non-expired items). May trigger lazy eviction.
	Len() int

	// GetAll returns a copy of all the cache contents (non-expired items).
	GetAll() map[K]V

	//// TTL Specific ////

	// StartCleanupDaemon starts a background cleanup cronjob that periodically removes expired entries.
	StartCleanupDaemon()

	// StopCleanupDaemon stops the background cleanup cronjob if running.
	StopCleanupDaemon()

	// Close stops cleanup cronjob and releases resources. After Close the cache can still be used, but TTL cronjob won't run.
	Close()
}
