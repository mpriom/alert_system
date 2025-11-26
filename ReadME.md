# Alert System

A resilient alert ingestion system that fetches, enriches, and stores alerts from a third-party API.

## Quick Start
```bash
docker-compose up --build
```

Services will be available at:
- **Alert Service**: http://localhost:8080
- **Mock Alerts API**: http://localhost:8081
- **PostgreSQL**: localhost:5432

## Services

| Service | Description |
|---------|-------------|
| `alert-service` | Main service - fetches, enriches, and stores alerts |
| `mock-alerts-api` | Simulated third-party API with configurable failure rate |
| `postgres` | Shared PostgreSQL database |

## Endpoints

### Alert Service (port 8080)
- `GET /alerts` - List alerts (optional: `?id=<uuid>` or `?days=<int>`)
- `POST /sync` - Trigger manual sync
- `GET /health` - Health check

### Mock API (port 8081)
- `GET /alerts` - Fetch alerts (optional: `?since=<ISO8601>`)
- `GET /health` - Health check

## Testing with cURL

### Health Checks
```bash
# Alert service health
curl http://localhost:8080/health

# Mock API health
curl http://localhost:8081/health
```

### Get Alerts
```bash
# Get all alerts
curl http://localhost:8080/alerts

# Get alert by ID
curl http://localhost:8080/alerts?id=<uuid>

# Get alerts from last 7 days
curl http://localhost:8080/alerts?days=7

# Pretty print with jq
curl -s http://localhost:8080/alerts | jq
```

### Trigger Manual Sync
```bash
curl -X POST http://localhost:8080/sync
```

### Mock API (Direct)
```bash
# Get all alerts from mock API
curl http://localhost:8081/alerts

# Get alerts since timestamp
curl "http://localhost:8081/alerts?since=2025-01-01T00:00:00Z"
```

## Configuration

Environment variables in `docker-compose.yml`:

| Variable | Default | Description |
|----------|---------|-------------|
| `MOCK_FAILURE_RATE` | `0.25` | Simulated failure rate (0-1) |
| `SYNC_INTERVAL` | `60s` | Auto-sync interval |

## Stop Services
```bash
# Stop services
docker-compose down

# Stop and remove volumes
docker-compose down -v
```

## Run Tests
```bash
# Alert service tests
cd alert-service && go test -v ./...

# Mock API tests
cd mock-alerts-api && go test -v ./...
```