# Operational Runbook — Low-Memory VPS (4GB RAM)

This runbook contains step-by-step procedures for operating, diagnosing, and maintaining the social media stack under production conditions on a single or clustered VPS.

---

## 1. Quick Health Check & Status Inspection

Run these diagnostic commands to get an immediate status snapshot:

```bash
# 1. Check all container statuses and uptimes
docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"

# 2. Check live resource consumption across all containers
docker stats --no-stream --format "table {{.Name}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.MemPerc}}\t{{.NetIO}}"

# 3. Check host memory and swap
free -h

# 4. Check API readiness and internal metrics
curl -i http://localhost/health/ready
curl -s http://localhost/health/metrics | jq .
```

---

## 2. Emergency Incident Response Procedures

### 2.1 Scenario A: Server Running Out of Memory (OOM Panic)
**Symptoms**: High swap usage (>500MB), SSH latency high, Docker killing containers with exit code 137.

**Step-by-step mitigation**:
1. Check which container is consuming the most RAM:
   ```bash
   docker stats --no-stream --format "table {{.Name}}\t{{.MemUsage}}\t{{.MemPerc}}" | sort -k3 -hr
   ```
2. If `x-api` has ballooned, trigger manual Go GC by curling pprof or restarting the container:
   ```bash
   docker restart x-api
   ```
3. If Redis has filled up, flush the ephemeral cache (all permanent data is safe in Postgres):
   ```bash
   docker exec -it x-redis redis-cli FLUSHDB
   ```
4. If Postgres is holding large buffer cache under heavy query load:
   ```bash
   # Terminate any queries running longer than 10 seconds:
   docker exec -it x-postgres psql -U postgres -d x_db -c \
     "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE (now() - pg_stat_activity.query_start) > interval '10 seconds' AND state != 'idle';"
   ```

### 2.2 Scenario B: Database Connection Pool Saturation
**Symptoms**: API logs show `pq: remaining connection slots are reserved for non-replication superuser connections` or `connection pool timeout`.

**Step-by-step mitigation**:
1. Inspect current active connections by state:
   ```bash
   docker exec -it x-postgres psql -U postgres -d x_db -c \
     "SELECT state, count(*) FROM pg_stat_activity GROUP BY state;"
   ```
2. Kill idle-in-transaction connections:
   ```bash
   docker exec -it x-postgres psql -U postgres -d x_db -c \
     "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE state = 'idle in transaction' AND (now() - state_change) > interval '5 seconds';"
   ```
3. Restart `x-api` to cleanly reset application connection pools:
   ```bash
   docker restart x-api
   ```

### 2.3 Scenario C: RabbitMQ Queue Backlog & Worker Stall
**Symptoms**: Background notifications or indexing delayed; RabbitMQ memory climbing.

**Step-by-step mitigation**:
1. Check queue depths:
   ```bash
   docker exec -it x-rabbitmq rabbitmqctl list_queues name messages consumers
   ```
2. Check if consumers are connected:
   - If consumers = 0, check `x-api` logs: `docker logs --tail 100 x-api`
   - Restart API to reconnect consumers: `docker restart x-api`
3. If poisoned messages are blocking queues, inspect the Dead-Letter Queue (`x.dlq`):
   ```bash
   docker exec -it x-rabbitmq rabbitmqctl list_queues name messages | grep dlq
   ```

---

## 3. Deployment & Update Runbook

When deploying updates without downtime:

1. Pull or build latest code:
   ```bash
   docker compose -f deploy/docker/docker-compose.yml build
   ```
2. Apply database migrations before starting new binaries:
   ```bash
   docker compose -f deploy/docker/docker-compose.yml run --rm x-api ./migrate up
   ```
3. Recreate containers with resource limits:
   ```bash
   docker compose -f deploy/docker/docker-compose.yml up -d
   ```
4. Refresh reverse proxy to ensure upstream resolution is updated:
   ```bash
   docker restart x-proxy
   ```
5. Verify health:
   ```bash
   curl -i http://localhost/health/ready
   ```

---

## 4. Periodic Maintenance Window (Weekly)

Run weekly off-peak (e.g. Sunday 03:00 UTC):
1. **Clean Docker dangling resources**:
   ```bash
   docker system prune -f
   ```
2. **Vacuum & Analyze Postgres**:
   ```bash
   docker exec -it x-postgres psql -U postgres -d x_db -c "VACUUM ANALYZE;"
   ```
3. **Prune processed outbox events older than 7 days**:
   ```bash
   docker exec -it x-postgres psql -U postgres -d x_db -c \
     "DELETE FROM outbox_events WHERE status = 'published' AND created_at < NOW() - INTERVAL '7 days';"
   ```
