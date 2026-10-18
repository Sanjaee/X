# Production Monitoring & Troubleshooting Guide (4 GB RAM VPS)

## 1. Monitoring & Observability Endpoints

The Go backend exposes safe observability endpoints:
- `GET /health/live`: Lightweight liveness check (200 OK if process is running).
- `GET /health/ready`: Readiness probe verifying PostgreSQL connection and Redis connection.
- `GET /health/metrics`: Protected runtime metrics (Goroutines, memory allocations, GC pauses, in-flight requests, and outbox event lag). Restricted to localhost / internal networks.

### Quick Live Dashboard
Run the interactive host monitor:
```bash
bash scripts/monitoring.sh
```

Or view Docker container memory and CPU stats:
```bash
docker stats --format "table {{.Name}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.MemPerc}}\t{{.NetIO}}"
```

---

## 2. Common Incident Runbooks

### Incident 1: High Memory Pressure / Out Of Memory (OOM)
- **Check RAM**: `free -h` and `docker stats --no-stream`
- **Mitigate**:
  - If Go memory is high: `GOMEMLIMIT=350MiB` and `GOGC=80` automatically force garbage collection before container OOM. If necessary: `docker restart x-api`.
  - If Redis memory is high: `docker exec -it x-redis redis-cli info memory`. Redis automatically evicts oldest cache keys via `allkeys-lru` at 96 MB.

### Incident 2: High Number of HTTP 429 Responses
- **Cause**: Client IP exceeded Nginx rate limits (30 req/s general, 10 req/s search) or backend sliding-window limit (120 req/min).
- **Verify**: Check `nginx` access log:
  ```bash
  docker logs --tail 50 x-proxy | grep "429"
  ```
- **Action**: Legitimate client can be whitelisted or granted higher quota via admin dashboard.

### Incident 3: Outbox Queue Backlog or Delayed Notifications
- **Check pending events**:
  ```bash
  docker exec -it x-postgres psql -U postgres -d x_prod -c \
    "SELECT status, count(*), min(next_retry_at), max(retry_count) FROM outbox_events GROUP BY status;"
  ```
- **Inspect Dead-Letter Queue**:
  ```bash
  docker exec -it x-postgres psql -U postgres -d x_prod -c \
    "SELECT id, event_type, retry_count, last_error FROM outbox_events WHERE status = 'failed' LIMIT 10;"
  ```
- **Retry failed events after fixing root cause**:
  ```bash
  docker exec -it x-postgres psql -U postgres -d x_prod -c \
    "UPDATE outbox_events SET status = 'pending', next_retry_at = NOW(), retry_count = 0 WHERE status = 'failed';"
  ```
