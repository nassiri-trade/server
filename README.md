## Forex Factory Calendar Service

Go Fiber service that ingests the ForexFactory weekly calendar feed, deduplicates entries using a deterministic hash, persists them to PostgreSQL, and exposes HTTP endpoints for manual sync and querying.

### Features

- Hourly sync from [`ff_calendar_thisweek.json`](https://nfs.faireconomy.media/ff_calendar_thisweek.json)
- Resty-based HTTP client with graceful error handling
- PostgreSQL persistence with deterministic hashes for idempotent upserts
- Structured logging with zerolog (configurable log levels)
- Auto-provisions user profiles on first position/trade ingestion
- Fiber API:
  - `GET /health`
  - `GET /api/v1/events?limit=100`
  - `POST /api/v1/events/sync`
  - `POST /api/v1/users/:user_id/positions`
  - `POST /api/v1/users/:user_id/trades`
  - `GET /api/v1/users/:user_id/report?limit=1000`
  - `GET /api/v1/users?limit=100`
  - `POST /api/v1/trading-data` (unified endpoint for positions and trades)
- Configurable via environment variables (see below)
- Docker Compose stack with PostgreSQL database
- Built-in Swagger UI served at `/swagger/index.html`

### Running Locally

```bash
docker compose up --build
```

The API will be available at `http://localhost:3000`.

### Environment Variables

| Variable | Default | Description |
| --- | --- | --- |
| `SERVER_PORT` | `3000` | Fiber listen port |
| `DATABASE_DSN` | `postgres://user:password@localhost:5432/trading_db?sslmode=disable` | PostgreSQL connection string |
| `CALENDAR_FEED_URL` | `https://nfs.faireconomy.media/ff_calendar_thisweek.json` | ForexFactory JSON feed |
| `SCHEDULER_INTERVAL` | `1h` | Sync cadence (Go duration string) |
| `LOG_LEVEL` | `info` | Logging level (debug, info, warn, error, fatal) |

You can override these in a `.env` file or via Compose.

> The default DSN connects to a local PostgreSQL instance. Make sure PostgreSQL is running and the database exists before starting the server.

### Project Structure

- `cmd/server`: Application entrypoint
- `internal/domain`: Entities and interfaces
- `internal/usecase`: Business logic (sync + list)
- `internal/infra`: External adapters (PostgreSQL via GORM, Resty feed client, structured logging)
- `internal/transport/http`: Fiber routing and handlers
- `scripts/migrations`: SQL migrations applied at startup

### Manual Operations

Trigger a manual sync:

```bash
curl -X POST http://localhost:3000/api/v1/events/sync
```

Fetch recent events (custom limit):

```bash
curl "http://localhost:3000/api/v1/events?limit=50"
```

Upsert a live position snapshot:

```bash
curl -X POST http://localhost:3000/api/v1/users/mohammad/positions \
  -H "Content-Type: application/json" \
  -d '{
        "platform": "MT5",
        "name": "Mohammad Baba",
        "login": "90354117",
        "server": "LiteFinance-MT5-Demo",
        "broker": "LiteFinance Global LLC",
        "time": "2025.11.08 12:03:32",
        "balance": 59236.97,
        "position": {
          "ticket": 189676791,
          "symbol": "BTCUSD",
          "side": "DEAL_TYPE_BUY",
          "volume": 0.26,
          "entryTime": "2025.11.08 12:03",
          "entry": 102310.05000,
          "sl": 100072.78000,
          "tp": 103574.23000,
          "profit": 0,
          "riskPercent": 0.98,
          "tickValue": 0.01,
          "tickSize": 0.01,
          "magic": 0,
          "reason": "POSITION_REASON_EXPERT",
          "comment": "BuyMarket"
        }
      }'
```

Record a trade history entry:

```bash
curl -X POST http://localhost:3000/api/v1/users/mohammad/trades \
  -H "Content-Type: application/json" \
  -d '{
        "ticket": 173877656,
        "pid": 187771524,
        "symbol": "XAUEUR_o",
        "side": "DEAL_TYPE_SELL",
        "entryType": "DEAL_ENTRY_OUT",
        "volume": 0.04,
        "entryTime": "2025.11.05 22:56",
        "entry": 3466.54,
        "sl": 3372.18,
        "tp": 3600.32,
        "profit": -63.28,
        "riskPercent": 0.74,
        "tickValue": 1.16,
        "tickSize": 0.01,
        "magic": 0,
        "reason": "DEAL_REASON_CLIENT"
      }'
```

Generate a performance report:

```bash
curl "http://localhost:3000/api/v1/users/mohammad/report?limit=1000"
```

Swagger UI (after build/start):

```bash
open http://localhost:3000/swagger/index.html
```

### References

- Resty HTTP client: [go-resty/resty](https://github.com/go-resty/resty)
- ForexFactory JSON feed: [ff_calendar_thisweek.json](https://nfs.faireconomy.media/ff_calendar_thisweek.json)
- Financial metrics reference: [Investopedia - Key Trading Performance Metrics](https://www.investopedia.com/articles/trading/09/key-trading-metrics.asp)

