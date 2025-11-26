# Alert Service

Fetches alerts from external API, enriches them, and exposes via REST API.

## Features

- Periodic sync with configurable interval
- Initial sync on startup (fetches since last known alert)
- Retry logic for failed API calls
- Alert enrichment (type + random IP)
- Context-aware with graceful shutdown

## API
```
GET  /alerts         # All alerts
GET  /alerts?id=xyz  # Single alert
GET  /alerts?days=7  # Last 7 days
POST /sync           # Trigger manual sync
GET  /health         # Health check
```

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `DB_HOST` | `localhost` | Database host |
| `DB_PORT` | `5432` | Database port |
| `DB_USER` | `postgres` | Database user |
| `DB_PASSWORD` | `postgres` | Database password |
| `DB_NAME` | `alerts_db` | Database name |
| `MOCK_API_URL` | `http://localhost:8081` | External API URL |
| `SYNC_INTERVAL` | `60s` | Periodic sync interval |

## Sync Behavior

| Event | Behavior |
|-------|----------|
| Startup (empty DB) | Fetches all alerts from external API |
| Startup (existing data) | Fetches alerts since last stored alert |
| Periodic | Runs every `SYNC_INTERVAL`, fetches new alerts |
| Manual (`POST /sync`) | Triggers immediate sync |

## Run Locally
```bash
go run main.go
```

## Tests
```bash
go test -v ./...
```

## Project Structure
```
├── internal/
│   ├── handlers/    # HTTP handlers
│   ├── service/     # Business logic
│   ├── storage/     # Database layer
│   └── models/      # Data models
├── external/        # External API client
└── config/          # Configuration
```