# Monitoring & Alerting Architecture for Low-Memory VPS

This document outlines the lightweight observability strategy designed to monitor platform health, diagnose traffic spikes, and trigger actionable alerts without consuming more than 150–250 MB RAM on a 4 GB VPS.

---

## 1. Observability Philosophy for Resource-Constrained Hosts

Traditional enterprise monitoring stacks (Datadog agents, heavy ELK clusters, Grafana Loki with Tempo) easily consume 1.5–3 GB of RAM—exceeding 50% of the entire VPS capacity.

Our lightweight approach relies on:
1. **Low-Overhead Metrics Endpoint (`/health/metrics`)**: Direct JSON/Prometheus metrics exposed by Go API without scraping overhead.
2. **Standard Docker Stats & OS Counters**: Fast, zero-memory introspection.
3. **Structured Logging via Go `log/slog`**: Structured JSON logs written to stdout, aggregated without bulky daemon processes.
4. **Targeted Alert Thresholds**: Alerting on sustained saturation rather than transient jitter.

---

## 2. Key Metrics Catalog

### 2.1 Host & Container Vital Signs
| Metric | Source | Normal Range | Warning Threshold | Critical Alert |
|---|---|---|---|---|
| **Available Host RAM** | `/proc/meminfo` | > 1.2 GB | < 600 MB (sustained 3m) | < 300 MB (Immediate OOM Risk) |
| **Swap Usage** | `free -m` | < 100 MB | > 300 MB | > 800 MB (Thrashing warning) |
| **Host CPU Utilization** | OS / `top` | 10% – 50% | > 75% for 5 min | > 90% for 2 min |
| **Container OOM Kills** | `docker events` | 0 | >= 1 in 1h | Any container killed |
| **Disk Space Available** | `df -h /` | > 20% | < 15% | < 5% |

### 2.2 Database (PostgreSQL)
| Metric | Source | Normal Range | Alert Threshold | Action |
|---|---|---|---|---|
| **Active Connections** | `pg_stat_activity` | 5 – 25 | > 60 (limit is 80) | Check for leaked connections in API pool |
| **Slow Queries (>1s)** | `pg_stat_statements` | 0 – 2/min | > 10/min | Review EXPLAIN plans; missing index |
| **Transaction Deadlocks**| `pg_stat_database` | 0 | > 0 | Review lock order in updates |

### 2.3 Cache (Redis)
| Metric | Source | Normal Range | Alert Threshold | Action |
|---|---|---|---|---|
| **Used Memory** | `INFO memory` | 30 – 80 MB | > 115 MB (max: 128MB) | Check TTL coverage; prune large keys |
| **Eviction Rate** | `evicted_keys` | < 10/sec | > 100/sec sustained | Hot cache thrashing; scale Redis or expand RAM |

### 2.4 Message Broker (RabbitMQ)
| Metric | Source | Normal Range | Alert Threshold | Action |
|---|---|---|---|---|
| **Queue Depth (Backlog)**| `rabbitmqctl list_queues` | < 100 | > 1,000 for 5 min | Spin up worker concurrency or throttle outbox |
| **Oldest Message Age** | RabbitMQ metrics | < 5s | > 60s | Worker stalled; restart worker container |
| **Dead-Letter Messages** | `x.dlq` count | 0 | > 5 | Inspect poisoned payload in DLQ |

### 2.5 API & Application Performance
| Metric | Target SLO | Alert Condition |
|---|---|---|
| **HTTP p95 Latency** | < 150 ms | > 500 ms for 3 minutes |
| **HTTP p99 Latency** | < 350 ms | > 1,500 ms for 2 minutes |
| **HTTP 5xx Error Rate** | < 0.1% | > 1% over 3 minutes |
| **HTTP 429 Rate Limit Rejections** | Baseline info | > 50/sec (indicates DDoS or bot sweep) |
| **Goroutine Count** | 50 – 300 | > 1,000 (indicates goroutine leak) |
| **WebSocket Active Connections** | 100 – 1,000 | > 1,500 (approaching socket memory limit) |

---

## 3. Lightweight Tooling Configuration

### 3.1 Option A: Ultra-Lightweight (Built-in Script / Cron)
Zero added RAM overhead. A lightweight shell script run via cron every minute checks memory, container health, and queue depth, notifying via webhook (Discord/Slack/Telegram) on violation.

### 3.2 Option B: Embedded Prometheus + Node Exporter
- **Node Exporter**: ~15 MB RAM.
- **Prometheus**: Configured with `--storage.tsdb.retention.time=3d` and `--storage.tsdb.max-block-duration=2h` to cap RAM at ~80–120 MB.
- **Dashboard**: Run Grafana externally or query Prometheus directly via CLI / lightweight UI.

---

## 4. Alert Response Matrix

```
+---------------------+---------------------------------------------------------------+
| Alert               | Immediate Mitigation Step                                     |
+---------------------+---------------------------------------------------------------+
| RAM Available < 400M| 1. Trigger Go GC: /debug/pprof/heap or restart API container |
|                     | 2. Clear Redis ephemeral cache: FLUSHDB (safe for ephemeral)  |
|                     | 3. Check docker stats to identify rogue container             |
+---------------------+---------------------------------------------------------------+
| DB Conn > 65        | 1. Query: SELECT pid, state, query FROM pg_stat_activity      |
|                     | 2. Terminate idle in transaction: pg_terminate_backend(pid)  |
|                     | 3. Reduce API pool size via env: DB_MAX_OPEN_CONNS=15         |
+---------------------+---------------------------------------------------------------+
| RabbitMQ Backlog    | 1. Inspect worker logs: docker logs --tail 100 x-api          |
| > 2,000 messages    | 2. Verify DB isn't locking on event processing                |
|                     | 3. Temporarily increase worker concurrency from 3 to 6        |
+---------------------+---------------------------------------------------------------+
| API p95 > 1s        | 1. Check if Search or Recommendation endpoint is saturating  |
|                     | 2. Backpressure semaphore activates HTTP 503 automatically   |
|                     | 3. Enable aggressive caching or degrade to chronological feed |
+---------------------+---------------------------------------------------------------+
```
