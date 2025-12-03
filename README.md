# Reverse Proxy

A simple reverse proxy server with caching capabilities.

## Features

- HTTP reverse proxy with configurable upstream
- LRU TTL cache for GET requests
- Connection pooling with configurable limits
- Debug mode for detailed logging
- Configurable cache TTL and capacity

## Steps to Run

1. Copy the sample configuration:
   ```bash
   cp sample.config.toml config.toml
   ```

2. (Optional) Provide a custom config path using `--config` flag if needed

3. Run the proxy:
   - **Debug mode:**
     ```bash
     go run -tags debug cmd/proxy/main.go --config config.toml
     ```
   - **Normal mode:**
     ```bash
     go run cmd/proxy/main.go --config config.toml
     ```

4. Run the example server:
   ```bash
   go run examples/main.go
   ```

5. Test the proxy:
   ```bash
   curl localhost:8000
   ```

6. Test stream responses:
   ```bash
   curl localhost:8000/stream
   ```

