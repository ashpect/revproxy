## Basic Plan

### Todo :
Reverse proxy (no net/http/httputil)
Timeouts
Logging
Error Handling and Graceful Shutdown
Caching 
Config

### Todo2 : 
Support Keep-Alive
Health check based conn checker (to avoid direct hit + circuit breaker)
Header sanitization (hop-by-hop and if anything else user defined)
TLS termination ?
Compression if > x ?  
tests

### If time allows :
Connection pooling
Load balancing
WebSocket upgrade support
Global Semaphore ?

### Security : 
Basic rate limiting
Prevention of ddos ? 
Stress testing

### Basic deployment :  
Dockerfile + README
k8s yaml and etc
