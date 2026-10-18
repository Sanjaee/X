# System Design & Architecture Specification: Production X Platform (4 GB RAM VPS)

## 1. System Overview & Core Philosophy

This document defines the production system design, reliability safeguards, and scaling strategy for our social media platform (similar to X/Twitter), specifically engineered to deliver low latency, high resilience, and zero OOM risks on a **single VPS with 4 GB RAM and 2 vCPUs**.

### Architectural Tenets
1. **Modular Monolith**: Single clean Go binary backend (`x-api`) handling REST and WebSockets; Next.js 14 frontend (`x-web`) handling UI and server-rendering; Nginx (`x-proxy`) managing TLS termination and rate limiting.
2. **PostgreSQL as Single Source of Truth**: All transactional records (users, posts, follows, interactions, subscriptions, notifications) persist in PostgreSQL 15/16.
3. **Controlled Ephemeral Redis**: Redis 7 is strictly restricted to ephemeral workloads: sliding window rate limiting, active session token caching, public user profile caching, and WebSocket presence.
4. **Resilient Transactional Outbox Worker**: High-reliability event fanout with atomic PostgreSQL writes, exponential backoff with random jitter, max-retry Dead-Letter handling (`status='failed'`), and automated retention cleanup.
5. **Zero Public Port Leakage**: No databases or caches (PostgreSQL 5432, Redis 6379, RabbitMQ 5672) are exposed to the public internet. All container traffic traverses an internal Docker bridge network (`x_backend`), with only Nginx ports 80/443 exposed publicly.

---

## 2. Before & After Architecture

### Prior Architecture (Issues Identified)
- `docker-compose.prod.yml` relied on a non-existent external Docker network (`media_backend`), causing build/deploy failures.
- Postgres (5432) and Redis (6379) ports were bound to `0.0.0.0` in base configurations, posing significant security exposure on public servers.
- Hardcoded fallback secrets (R2 Cloudflare credentials, Plisio keys, default passwords) existed in `config.go` and `.env.example`.
- `RequestTimeoutMiddleware` only injected an HTTP header without invoking `context.WithTimeout`, meaning slow database operations or stalled requests never timed out.
- `ProcessAndUpload` buffered unlimited file sizes into RAM via `io.ReadAll` and only validated the file extension string (`.jpg`), creating both memory exhaustion and MIME spoofing vulnerabilities.
- Outbox worker processed events with basic queries without concurrency locking, lacked retry backoff, lacked a Dead-Letter Queue (DLQ), and had no retention cleanup mechanism, causing unbounded database table growth.
- Every authenticated request triggered two PostgreSQL queries (`GetSessionByTokenHash` + `GetUserByID`) without cache-aside assistance.

### Upgraded Architecture (Production-Hardened)
```text
Users / Mobile / Browsers
          │
      Internet
          │ (Port 80 / 443)
┌─────────▼──────────────────────────────────────────────┐
│ Nginx Reverse Proxy (x-proxy)                           │
│ - Rate Limiting: API 30 req/s, Search 10 req/s          │
│ - Buffer tuning & Real IP restoration                   │
│ - Static uploads routing & WebSocket upgrade           │
└─────────┬──────────────────────────────────┬───────────┘
          │ (Internal Network: x_backend)    │
┌─────────▼──────────────┐         ┌─────────▼───────────┐
│ Next.js 14 Web (x-web) │         │ Go API Monolith     │
│ - Max Old Space: 384MB │         │ (x-api)             │
│ - SSR & React Virtual  │         │ - GOMEMLIMIT=350MiB │
└────────────────────────┘         │ - Active In-Flight: │
                                   │   Max 300 Requests  │
                                   └─────────┬───────────┘
                                             │
      ┌──────────────────────────────────────┼─────────────────────────┐
      │                                      │                         │
┌─────▼───────────────────┐    ┌─────────────▼──────────┐   ┌──────────▼──────────────┐
│ PostgreSQL 15/16        │    │ Redis 7.0 (x-redis)    │   │ Transactional Outbox    │
│ (x-postgres)            │    │ - Max Memory: 96MB     │   │ Worker (In-Process)     │
│ - Max Conns: 80         │    │ - Policy: allkeys-lru  │   │ - Exponential Backoff   │
│ - Shared Buffers: 192MB │    │ - Cache: Public Profile│   │ - Jitter (0-1000ms)     │
│ - Work Mem: 4MB         │    │   & Sessions (15m TTL) │   │ - Max Retries: 5 (DLQ)  │
│ - Statement Timeout: 8s │    │ - Rate Limit Sliding   │   │ - Auto Purge: 7 Days    │
│ - Dedicated Migrations  │    │   Window (ZSet)        │   └─────────────────────────┘
└─────────────────────────┘    └────────────────────────┘
```

---

## 3. VPS 4 GB RAM — Memory Budgeting & Tuning

| Component | Indicative Allocation | Hard Container Limit | Memory Safeguard Mechanism |
| :--- | :---: | :---: | :--- |
| **Linux OS, Kernel, Docker Engine** | 700 MB | Host | `vm.swappiness=10`, `vm.min_free_kbytes=65536` |
| **PostgreSQL 15/16** | 600–700 MB | 800 MB | `shared_buffers=192MB`, `work_mem=4MB`, `max_connections=80` |
| **Go Backend API (`x-api`)** | 250–350 MB | 450 MB | `GOMEMLIMIT=350MiB`, `GOGC=80`, In-flight limit 300 |
| **Next.js Frontend (`x-web`)** | 250–350 MB | 512 MB | `NODE_OPTIONS="--max-old-space-size=384"` |
| **Redis 7 Cache (`x-redis`)** | 60–100 MB | 150 MB | `maxmemory 96mb`, `maxmemory-policy allkeys-lru` |
| **RabbitMQ 3 (`x-rabbitmq`)** | 120–200 MB | 250 MB | `RABBITMQ_VM_MEMORY_HIGH_WATERMARK=0.4` |
| **Nginx Proxy (`x-proxy`)** | 20–40 MB | 128 MB | `worker_connections 1024`, `client_body_buffer_size 128k` |
| **Safety Burst Headroom** | **~1,200 MB** | — | Preserved for file system cache, builds & traffic surges |

---

## 4. Database Strategy & Migrations

### Source of Truth Guarantees
- **Cursor-Based Keyset Pagination**: Feeds, user posts, replies, and notifications query using `(created_at DESC, id DESC)` over B-Tree indexes, completely avoiding expensive offset scans (`OFFSET N`).
- **N+1 Prevention**: Media objects, parent posts, and quote post embeddings are batched using PostgreSQL `ANY($1)` arrays rather than sequential per-row queries.
- **Outbox Resilience (Migration 000009)**:
  - Adds `last_error TEXT NULL` and `next_retry_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`.
  - Filtered indexes: `idx_outbox_poll_ready` on `(next_retry_at ASC) WHERE status IN ('pending', 'retrying')`.
  - Retention index: `idx_outbox_retention_cleanup` on `(published_at) WHERE status = 'published'`.

---

## 5. Background Queue, Outbox & Worker Design

1. **Transactional Outbox Pattern**: When a post interaction, direct message, or notification occurs, the event is written within the same database transaction to `outbox_events`.
2. **Worker Polling**:
   - Polling interval: Every 2 seconds.
   - Batch size: 20 events.
   - Condition: `status IN ('pending', 'retrying') AND next_retry_at <= NOW()`.
3. **Exponential Backoff & Jitter**:
   $$\text{Backoff} = 2^{\min(\text{retry\_count}, 6)} \times 2\text{s} + \text{rand}(0, 1000)\text{ms}$$
4. **Dead-Letter Queue (DLQ)**: When `retry_count >= 5`, the status changes to `failed` and logs an alert.
5. **Disk Retention Purge**: Every 1 hour, events in `published` status older than 7 days are pruned automatically.

---

## 6. Security & Upload Pipeline

1. **Magic-Byte Sniffing**: File uploads are inspected using `http.DetectContentType` on the initial 512 bytes. Scripts, executables, or HTML disguised as images are rejected immediately with HTTP 400.
2. **Resource Throttling**: Strict 50 MB hard cap on video uploads and 15 MB on images to prevent heap inflation.
3. **Context Timeout Cancellation**: All HTTP requests run under `context.WithTimeout` (15s standard, 5s for search endpoints), with persistent WebSockets bypassed.
4. **Admission Control**: Rate limiting at Nginx (30 req/s general, 10 req/s search) backed by Go atomic concurrency backpressure (HTTP 503 after 300 active requests).

---

## 7. Scaling Roadmap (From 1 VPS to Scale)

- **Stage 1 (Current — Single 4GB VPS)**:
  - Modular Monolith + Postgres + Ephemeral Redis + Local/R2 Object Storage.
  - Handles up to ~150-250 concurrent active users and ~500k daily requests.
- **Stage 2 (Read Load Growth — 8GB-16GB VPS or Two Nodes)**:
  - Separate Next.js frontend to Vercel/Node VPS or CDN edge.
  - Dedicated PostgreSQL instance with automated daily WAL archiving.
- **Stage 3 (High Traffic — Read Replica & Sharding Evaluation)**:
  - PostgreSQL Read Replica for feed ranking and search read replicas.
  - Dedicated Redis cluster for timeline caching.
