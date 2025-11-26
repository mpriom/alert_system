# Mock Alerts API

Simulated third-party alerts API with configurable failure rate for testing resilience.

## API
```
GET /alerts              # All alerts (limit 100)
GET /alerts?since=<ts>   # Alerts since ISO8601 timestamp
GET /health              # Health check
```

## Response Format
```json
{
  "alerts": [
    {
      "source": "siem-1",
      "severity": "high",
      "description": "Suspicious login",
      "created_at": "2025-01-10T12:34:56Z"
    }
  ]
}
```

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `DB_HOST` | `localhost` | Database host |
| `DB_PORT` | `5432` | Database port |
| `DB_USER` | `postgres` | Database user |
| `DB_PASSWORD` | `postgres` | Database password |
| `DB_NAME` | `alerts_db` | Database name |
| `PORT` | `8081` | Server port |
| `MOCK_FAILURE_RATE` | `0.25` | Failure rate (0-1) |

## Failure Simulation

Set `MOCK_FAILURE_RATE` to control random 500 errors:
- `0` = No failures
- `0.25` = 25% failure rate (default)
- `1` = Always fail

## Run Locally
```bash
go run main.go
```