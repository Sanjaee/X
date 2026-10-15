# Architecture: Low-Memory Budgeting & Heap Management

## 1. Overview
Operating a complex social media platform on a **4 GB RAM VPS** requires strict budgeting of Resident Set Size (RSS), heap allocations, kernel buffers, and database caches.

---

## 2. Quantitative Memory Budget Breakdown

| Service / Subsystem | Hard RAM Limit | Target Steady RSS | Idle Memory | Primary Memory Consumers |
| :--- | :---: | :---: | :---: | :--- |
| **Linux OS & Kernel** | 900 MB (Reserved)| ~650 MB | ~450 MB | Page tables, slab allocators, systemd, Docker daemon |
| **PostgreSQL 16** | 768 MB | ~450 MB | ~180 MB | `shared_buffers` (256MB), connection buffers, GIN index pages |
| **Go API (`x-api`)** | 350 MB | ~180 MB | ~60 MB | Goroutine stacks, JSON serialization buffers, in-flight HTTP bodies |
| **Next.js Web (`x-web`)**| 512 MB | ~280 MB | ~120 MB | Node.js V8 heap, SSR page render trees, module cache |
| **RabbitMQ 3.13** | 300 MB | ~160 MB | ~95 MB | Erlang VM heap, message queue index, channel buffers |
| **Redis 7** | 180 MB | ~90 MB | ~25 MB | Key-value dictionary, LRU cache pages, client buffers |
| **NGINX Proxy** | 128 MB | ~45 MB | ~15 MB | Worker process memory, request buffer pools, SSL state |
| **Host Safety Buffer** | **958 MB** | — | — | **Unallocated headroom for kernel burst and page cache** |
| **TOTAL** | **4,096 MB** | **~1,855 MB** | **~945 MB** | **Guaranteed zero host-level OOM condition** |

---

## 3. Go Runtime Memory Controls (`GOMEMLIMIT`)

By default, Go's runtime only initiates garbage collection when the heap doubles (`GOGC=100`). In a container with a 350MB limit, this default can allow the heap to spike to 400MB before GC triggers, causing the Linux kernel to terminate the process with `Exit Code 137 (OOM Killed)`.

### Configuration
```bash
GOMEMLIMIT=300MiB
GOGC=80
```
- **`GOMEMLIMIT=300MiB`**: Signals the Go runtime to trigger aggressive GC cycles as allocations approach 300MB, leaving 50MB for OS thread stacks and CGO allocations.
- **`GOGC=80`**: Reduces heap trigger ratio from 100% to 80%, smoothing memory consumption.

---

## 4. Node.js Next.js Web Runner Memory Controls

Node.js calculates default heap limits based on total host physical RAM (which would be ~1.4GB on a 4GB machine), causing containers to exceed Docker constraints.
### Configuration
```bash
NODE_OPTIONS="--max-old-space-size=384"
```
Forces V8 garbage collection to aggressively compact old-space heap objects when usage reaches 384MB.

---

## 5. PostgreSQL Memory Sizing Math

$$\text{Total Postgres Memory} \approx \text{shared\_buffers} + (\text{max\_connections} \times \text{work\_mem}) + \text{maintenance\_work\_mem} + \text{base\_overhead}$$

Substituting our production parameters:
$$\text{Total} \approx 256\text{MB} + (80 \times 4\text{MB}) + 64\text{MB} + 50\text{MB} \approx 690\text{MB}$$

This stays strictly within the **768 MB Docker limit**.

---

## 6. What to Trim When Under Memory Pressure

If monitoring indicates available host memory is `< 500 MB`:
1. **Reduce Next.js Memory**: Set `--max-old-space-size=256`.
2. **Lower Postgres Connections**: Reduce `max_connections` to `50` and Go API `SetMaxOpenConns(15)`.
3. **Evict Redis Non-Essential Keys**:
   ```bash
   redis-cli -a <password> MEMORY PURGE
   ```
4. **Lower Worker Concurrency**: Reduce RabbitMQ worker concurrency from 5 to 2.
5. **Decouple Frontend**: Move Next.js to Cloudflare Pages / Vercel, freeing **512 MB** instantly on the backend VPS.
