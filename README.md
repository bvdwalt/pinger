# Pinger

A simple Go application that periodically pings HTTP endpoints on a configurable cron schedule.

## Features

- Configure multiple endpoints to ping via YAML
- Cron-based scheduling (e.g., every 5 minutes)
- Custom API key header support
- Custom User-Agent header support
- Full HTTP request/response logging for debugging and monitoring
- Parameterized endpoints with iterations (ping multiple organizations with one config)
- Logs response status codes and duration
- Graceful shutdown

## Usage

1. Copy the example configuration:
   ```bash
   cp config-example.yaml config.yaml
   ```

2. Edit `config.yaml` with your endpoints and API key

3. Run the pinger:
   ```bash
   go run ./cmd/pinger
   ```
   
   Or build and run:
   ```bash
   go build -o pinger ./cmd/pinger
   ./pinger
   ```

## Configuration

The `config.yaml` file supports:

- **schedule**: Cron expression (e.g., `*/5 * * * *` for every 5 minutes)
- **timeout-seconds**: HTTP request timeout in seconds
- **api-key-header-name**: Header name for API authentication
- **api-key-value**: Your API key
- **user-agent**: Custom User-Agent header (optional)
- **endpoints**: List of endpoints to ping
  - Simple endpoints with `name`, `url`, and `method`
  - Parameterized endpoints with `iterations` for dynamic URL/name substitution

See `config-example.yaml` for examples.

## Logging

Pinger logs all HTTP activity including:
- HTTP request details (method, URL, headers)
- HTTP response status and headers
- Endpoint response time and status code
- Any errors encountered during requests

This comprehensive logging is useful for:
- Debugging endpoint issues
- Monitoring API availability
- Tracking response times
- Identifying configuration problems

## Testing

```bash
go test -v --cover
```
## Docker

### Build and run with Docker

1. Build the image:
   ```bash
   docker build -t pinger .
   ```

2. Run the container:
   ```bash
   docker run -v $(pwd)/config.yaml:/app/config.yaml:ro pinger
   ```