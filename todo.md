## Basic Plan

### Todo (Must to have):
- [x] Basic reverse proxy (no net/http/httputil)
- [x] Timeouts
- [x] Use a custom transport
- [] Error Handling
- [] Better logging
- [] Deployment configs
- [] Caching
    - [x]Implement a builder for cache
    - [x]LRU impl + thread safety
    - []Use LRU with TTL (Support TTL per entry)
    - []Server initiated cache-control 
        - []200 GET, cache it
        - []Cache-control specified
        - []Cache control setup from config itself

### Todo (Security)
- []Global Semaphore
- []Rate limiting
- []Max header/body size

### Todo (Good to have):
- []Tune connection pooling
- []Support other hop to hop headers (ws etc)
- []Health check based conn checker (to avoid direct hit + circuit breaker)
- []Compression if > x ?  
- []Tests
- []Load balancing

### HTTPS
- []TLS

### Benchmarking
- [] Benchmarking script that simulates clients for various cases.
- []The benchmarking script must measure:
    - []Total failures (e.g., no response received or connection dropped).
    - []Total throughput of the server.
    - []Latency percentiles
