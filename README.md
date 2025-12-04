# Reverse Proxy
A minimal HTTP reverse proxy embeding cache feature. Features a generic, thread-safe LRU cache with TTL support, configurable connection pooling, and HTTP standards compliance.

## Steps to Run

1. Copy the sample configuration:
   ```bash
   cp sample.config.toml config.toml
   ```

2. (Optional) Provide a custom config path using `--config` flag if needed

3. Run all tests in the repository:
   ```bash
   go test ./... -v
   ```

4. Run the proxy:
   - **Debug mode:**
     ```bash
     go run -tags debug cmd/proxy/main.go --config config.toml
     ```
   - **Normal mode:**
     ```bash
     go run cmd/proxy/main.go --config config.toml
     ```

5. Run the example server:
   ```bash
   go run examples/main.go
   ```

6. Test the proxy:
   ```bash
   curl localhost:8000
   curl localhost:8000/stream
   ````

#### Ways to Build Cache :
1. `cache.NewLRUTTL()` - cache with standard settings from server (10 capacity, 10 min TTS, 10ms ceanup interval)
2. Recommended options to use with `cache.NewLRUTTL()`:
   - `WithCapacity`: Sets the maximum number of items the cache can store.
   - `WithDefaultTTL`: Sets the default Time-To-Live for items without an explicit TTL, specified in seconds.
   - `WithCleanupInterval`: Sets the frequency (in seconds) at which expired items are cleaned up by daemon
   - `WithCleanupStart`: Determines whether you want to use TTL cleanup service or not.
   - `WithItemsMap` : Import an items map as an initial cache and build your cache upon it using options.
You can also insert an element outliner to not follow TTS by seeting it's expiresAt as `0 time.time`
The code is hightly composable for both proxy and cache, builders are specified for each structs for each extensibility in the future.

### Done 
- [x] Basic reverse proxy (no net/http/httputil)
- [x] Stream
- [x] Use a custom transport
- [x] debug utils
- [x] Caching
    - [x] Implement a builder for cache
    - [x] LRU impl + thread safety
    - [x] Use LRU with TTL (Support TTL per entry)
    - [x] Generic cache
    - [x] Cache unit-tests
    - [x] Server initiated cache-control 
        - [x] 200 GET, cache it
        - [x] Cache-control specified
        - [x] Cache control setup from config itself

General code improvements/optimizations are marked in code as TODO
#### Planned Features
- [] Deployment configs
- [] Support other hop to hop headers (ws etc)
- [] Health check based conn checker (to avoid direct hit + circuit breaker)
- [] Compression if > x ?  
- [] Load balancing
#### Benchmarking
- [] Benchmarking script that simulates clients for various cases measuring failures, throughput and latency
#### Security : 
- [] Rate limiting
- [] Max header/body size
#### Better Observability : 
- [] Better logging and writing to 2 files, .info and .err for preserving server logs
- [] Fully functional test suite, currently exists for pkg cache