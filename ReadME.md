# Censys Alert System

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

## Configuration

Environment variables in `docker-compose.yml`:

| Variable | Default | Description |
|----------|---------|-------------|
| `MOCK_FAILURE_RATE` | `0.25` | Simulated failure rate (0-1) |
| `SYNC_INTERVAL` | `60s` | Auto-sync interval |

## Stop Services
```bash
docker-compose down -v  # -v removes volumes
```