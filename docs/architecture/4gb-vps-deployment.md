# Architecture: 4 GB RAM VPS Production Deployment

## 1. Executive Summary & Design Target
This specification details how to deploy and operate the complete X social media platform on a single **4 GB RAM / 2 vCPU VPS** running Ubuntu 22.04/24.04 LTS or Debian 12 with NVMe SSD.

The core design principle is **Memory-First Isolation**: every service is allocated a deterministic RAM budget with hard limits in Docker Compose, ensuring that sudden traffic spikes or rogue requests never cause host-wide OOM (Out Of Memory) crashes.

---

## 2. Global Memory Budget Allocation (4,096 MB Total)

```
+-----------------------------------------------------------------------------------+
| Host Memory (4,096 MB)                                                            |
|                                                                                   |
| [OS Kernel, Buffers, Cache, Docker Daemon] : ~900 MB reserved                     |
|                                                                                   |
| Container Workloads (Hard Capped Total: ~2,350 MB / Peak Safety Margin: ~846 MB) |
|   |-- PostgreSQL 16 (x-postgres)      : Limit 768 MB (Reserve 256 MB)             |
|   |-- Go API Monolith (x-api)         : Limit 350 MB (Reserve 128 MB)             |
|   |-- Next.js Web Runner (x-web)      : Limit 512 MB (Reserve 128 MB)             |
|   |-- RabbitMQ Message Broker (x-mq)  : Limit 300 MB (Reserve 128 MB)             |
|   |-- Redis 7 Ephemeral Store (x-redis): Limit 180 MB (Reserve  64 MB)             |
|   +-- NGINX Reverse Proxy (x-proxy)   : Limit 128 MB (Reserve  32 MB)             |
|                                                                                   |
| Swap Space (ZRAM or NVMe Swapfile)     : 2,048 MB (Safety burst buffer only)      |
+-----------------------------------------------------------------------------------+
```

---

## 3. Host OS & Linux Kernel Tuning (`/etc/sysctl.conf`)

Apply these parameters on the host VPS to optimize memory recycling and network throughput:

```ini
# Reduce swappiness to prioritize RAM and avoid unnecessary disk thrashing
vm.swappiness = 10

# Memory overcommit: heuristic overcommit handling (prevents rogue allocations)
vm.overcommit_memory = 1

# Prevent memory fragmentation
vm.min_free_kbytes = 65536

# File descriptor limits for high concurrency WebSocket and HTTP connections
fs.file-max = 2097152

# Network socket backlog and buffer tuning
net.core.somaxconn = 4096
net.ipv4.tcp_max_syn_backlog = 4096
net.ipv4.tcp_fin_timeout = 15
net.ipv4.tcp_tw_reuse = 1
```

Apply with:
```bash
sudo sysctl -p
```

---

## 4. Container-by-Container Configuration

### 1. PostgreSQL 16
- **Config Flags**:
  - `shared_buffers = 256MB`
  - `work_mem = 4MB`
  - `maintenance_work_mem = 64MB`
  - `effective_cache_size = 512MB`
  - `max_connections = 80`
  - `statement_timeout = 8000` (8s timeout kills runaway full table scans)
  - `idle_in_transaction_session_timeout = 15000`
- **Docker Limits**: `mem_limit: 768m`, `cpus: 1.5`

### 2. Go API (`x-api`)
- **Environment Variables**:
  - `GOMEMLIMIT=300MiB` (Enforces Go 1.19+ garbage collection before Docker OOM)
  - `GOGC=80` (More aggressive GC cycle to keep RSS compact)
  - `MAX_INFLIGHT_REQUESTS=300` (Backpressure drops excess requests with HTTP 503)
  - `DB_MAX_OPEN_CONNS=20`
  - `DB_MAX_IDLE_CONNS=10`
- **Docker Limits**: `mem_limit: 350m`, `cpus: 1.0`

### 3. Redis 7
- **Command Line**:
  `redis-server --requirepass ${REDIS_PASSWORD} --maxmemory 128mb --maxmemory-policy allkeys-lru`
- **Docker Limits**: `mem_limit: 180m`, `cpus: 0.5`

### 4. RabbitMQ 3
- **Configuration**:
  - `RABBITMQ_VM_MEMORY_HIGH_WATERMARK=0.4` (Throttles publishers when broker hits 120MB)
  - Disk free limit: 50MB
- **Docker Limits**: `mem_limit: 300m`, `cpus: 0.5`

### 5. Next.js Web Runner (`x-web`)
- **Node Configuration**:
  - `NODE_OPTIONS="--max-old-space-size=384"`
- **Docker Limits**: `mem_limit: 512m`, `cpus: 1.0`

### 6. Nginx Edge Proxy (`x-proxy`)
- Rate limiting zones configured in `nginx.conf`:
  - 30 requests/second per IP with burst=50.
  - Client max body size: 50MB.
- **Docker Limits**: `mem_limit: 128m`, `cpus: 0.5`

---

## 5. Emergency Runbook During Traffic Spikes
1. **Check Live Memory Distribution**:
   ```bash
   docker stats --no-stream
   ```
2. **If Swap Usage > 50%**:
   Trigger immediate Redis cache eviction:
   ```bash
   docker exec x-redis redis-cli -a supersecretredispass MEMORY PURGE
   ```
3. **If Postgres Connection Saturation**:
   Inspect active queries:
   ```bash
   docker exec -it x-postgres psql -U postgres -d media_prod -c "SELECT pid, now() - query_start AS duration, query FROM pg_stat_activity WHERE state != 'idle' ORDER BY duration DESC LIMIT 5;"
   ```
