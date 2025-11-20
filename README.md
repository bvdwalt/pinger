# Pinger

A simple Go application that periodically pings HTTP endpoints on a configurable cron schedule.

## Features

- Configure multiple endpoints to ping via YAML
- Cron-based scheduling (e.g., every 5 minutes)
- Custom API key header support
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
   go run .
   ```

## Configuration

The `config.yaml` file supports:

- **schedule**: Cron expression (e.g., `*/5 * * * *` for every 5 minutes)
- **timeout-seconds**: HTTP request timeout
- **api-key-header-name**: Header name for API authentication
- **api-key-value**: Your API key
- **endpoints**: List of endpoints to ping
  - Simple endpoints with `name`, `url`, and `method`
  - Parameterized endpoints with `iterations` for dynamic URL/name substitution

See `config-example.yaml` for examples.

## Testing

```bash
go test -v --cover
```
