# Architectural Audit: Current System, Bottlenecks, and Resource Utilization

## 1. Executive Summary & Objective
This audit analyzes the full-stack social media monorepo to identify performance bottlenecks, memory leak risks, unbounded concurrency hazards, and architectural constraints when deploying to a low-resource **4 GB RAM VPS**.

### Core Methodology
- Analysis grounded in **REA (Reverse Engineer Anything)** telemetry from large-scale social networks (e.g., X/Twitter).
- Identification of structural differences between high-memory distributed infrastructure and resource-constrained single/multi-VPS environments.
- Classification of all findings as `[VERIFIED]`, `[INFERRED]`, or `[PROPOSED]`.

---

## 2. Component-by-Component Resource Audit

### A. Go API Service (`x-api`)
| Attribute | Current State | Risk / Bottleneck | Target Budget (4 GB VPS) |
| :--- | :--- | :--- | :--- |
| **Connection Pooling** | `SetMaxOpenConns(50)`, `SetMaxIdleConns(25)` | Too high for small Postgres instance. 2 replicas would consume 100 connections alone. | `SetMaxOpenConns(20)`, `SetMaxIdleConns(10)` |
| **Goroutine Management** | Unbounded `go func()` in handlers (e.g. async recent searches, impression tracking) | Traffic spike of 5,000 req/s creates 5,000 goroutines, inflating RSS memory to OOM. | Bounded worker pools with channel queues or semaphores |
| **Request Concurrency** | No global inflight request limit | Unbounded incoming HTTP traffic causes memory starvation. | Semaphore middleware capping concurrent requests at 300 |
| **Memory Ceiling** | No `GOMEMLIMIT` set | Go runtime GC only triggers at default `GOGC=100`, risking Linux kernel OOM kills. | `GOMEMLIMIT=300MiB`, `GOGC=80` |
| **Body Size Limits** | Handled at Nginx (50M), but API handlers read multipart directly into memory | Uploading 40MB file directly strains API heap. | Streamed disk spooling with 10MB chunk threshold |

### B. PostgreSQL Database (`x-postgres`)
| Attribute | Current State | Risk / Bottleneck | Target Budget (4 GB VPS) |
| :--- | :--- | :--- | :--- |
| **Memory Allocation** | Alpine default configuration (`shared_buffers=128MB`, unbounded `max_connections=100`) | Unoptimized cache hit ratio and risk of swap trashing. | `shared_buffers=256MB`, `work_mem=4MB`, `max_connections=80` |
| **Query Plans & Indexes** | High-quality partial GIN indexes (`to_tsvector`), trigram (`pg_trgm`), and keyset indexes | Excellent coverage, but lacks statement timeout protection. | `statement_timeout=8000ms`, `idle_in_transaction_session_timeout=15000ms` |
| **Pagination Strategy** | Pure Keyset cursor pagination (`created_at DESC, id DESC`) | Eliminates `OFFSET` degradation. Deterministic $O(\log N)$ traversal. | Maintained |
| **Resource Isolation** | No Docker container memory limits | Postgres could expand past 2 GB during analytical or heavy batch operations. | Capped via Docker Compose `mem_limit: 768m` |

### C. Redis Cache & Ephemeral State (`x-redis`)
| Attribute | Current State | Risk / Bottleneck | Target Budget (4 GB VPS) |
| :--- | :--- | :--- | :--- |
| **Memory Management** | `redis:7-alpine` without `--maxmemory` | Redis will continuously consume RAM until host runs out of memory. | `--maxmemory 128mb --maxmemory-policy allkeys-lru` |
| **Persistence** | In-memory with default RDB snapshotting | Adequate for caching and rate limiting; not used as durable source of truth. | Retain RDB; ensure keys have mandatory TTLs |

### D. Reverse Proxy & Frontend (`x-proxy` & `x-web`)
| Attribute | Current State | Risk / Bottleneck | Target Budget (4 GB VPS) |
| :--- | :--- | :--- | :--- |
| **Nginx Buffers** | Standard proxy buffers without rate limiting zones | Susceptible to HTTP flood attacks and request amplification. | Add `limit_req_zone` (30r/s burst=50) for `/api/` |
| **Next.js SSR Memory** | Standalone Node.js server inside container | Node.js V8 heap defaults to ~1.4 GB if unconstrained. | Set `NODE_OPTIONS="--max-old-space-size=384"` with Docker `mem_limit: 512m` |

---

## 3. High-Traffic Risk Matrix & Mitigations

```
+-----------------------------------------------------------------------------------------+
| Threat / Failure Mode      | Impact on 4GB VPS  | Remediation Strategy                  |
+----------------------------+--------------------+---------------------------------------+
| Sudden 10x traffic spike   | Host RAM exhaustion| Bounded inflight concurrency semaphore|
| Uncached search text flood | CPU 100% on GIN    | Redis query cache + AST operator limits|
| Viral post like/follow storm| Lock contention   | Batched counter updates + Redis buffer|
| Large video upload flood   | API OOM kills      | Direct streaming to Cloudflare R2 / S3 |
| Unbounded WebSocket clients| File descriptor OOM| Max 2,000 active WS + idle timeout    |
+-----------------------------------------------------------------------------------------+
```

---

## 4. Architectural Readiness Score
- **Stateless API Tier**: Ready (Session tokens stored in Postgres/Redis, no local session affinity required).
- **Database Indexing**: Ready (Composite B-Trees and GIN indexes in place).
- **Asynchronous Decoupling**: Ready for RabbitMQ durable worker queue integration.
- **Container Hardening**: Requires explicit Docker resource limits and tuned runtime limits.
