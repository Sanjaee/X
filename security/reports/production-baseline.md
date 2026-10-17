# LostMedia Production Baseline Measurements

**Recorded**: October 17, 2026, 07:55:00 UTC (14:55:00 WIB)  
**Host**: Linux 5.15.0-191-generic x86_64 Ubuntu (103.89.2.68)  
**Target Platform**: [https://lostmedia.pro/](https://lostmedia.pro/)  

---

## 1. System Infrastructure Baseline

| Metric | Measured Baseline Value | Threshold Safe Limit | Health Status |
| :--- | :--- | :--- | :--- |
| **System Uptime** | 17 days, 17 hours | N/A | Operational |
| **CPU Load Average** | 0.10, 0.14, 0.15 (1-core / 2-vCPU) | < 2.0 | Nominal (Normal) |
| **Total Memory** | 3911 MiB (4.0 GB) | N/A | Normal |
| **Used Memory** | 640 MiB (16.3%) | < 3200 MiB (80%) | Optimal |
| **Available Memory** | 2924 MiB (74.7%) | > 600 MiB | High headroom |
| **Swap Usage** | 0 MiB used / 4095 MiB total | < 500 MiB | Zero thrashing |
| **Root Disk `/` Usage** | 16 GB used / 58 GB total (28%) | < 80% (46 GB) | Healthy |

---

## 2. Docker Container Baseline Resource Consumption

| Container Name | CPU % | Memory Usage / Limit | Memory % | Network I/O (Rx/Tx) | PIDs |
| :--- | :--- | :--- | :--- | :--- | :--- |
| `x-web` | 0.00% | 79.3 MiB / 512 MiB | 15.49% | 6.93 MB / 15.1 MB | 22 |
| `x-api` | 0.05% | 4.99 MiB / 350 MiB | 1.43% | 29.2 MB / 36.1 MB | 9 |
| `x-proxy` | 0.00% | 4.43 MiB / 128 MiB | 3.46% | 23.0 MB / 22.7 MB | 2 |
| `x-postgres` | 0.01% | 36.39 MiB / 768 MiB | 4.74% | 38.7 MB / 31.3 MB | 7 |
| `x-rabbitmq` | 0.22% | 115.5 MiB / 300 MiB | 38.49% | 47.0 kB / 126 B | 28 |
| `x-redis` | 0.61% | 3.99 MiB / 180 MiB | 2.22% | 531 kB / 199 kB | 6 |
| `global-nginx` | 0.00% | 7.78 MiB / 3.82 GiB | 0.20% | 23.4 MB / 25.2 MB | 3 |

---

## 3. Database & Broker Runtime Baseline

- **Active PostgreSQL Connections**: 7 connections (pool configured with max 20).
- **Redis Operations**: 59,244 commands processed, 0 ops/sec idle, memory consumption 3.99 MiB.
- **RabbitMQ Queue Depth**: 0 pending messages, zero message backlog across all virtual hosts.
- **Go Runtime Goroutines**: 9 active goroutines at idle.
- **Go Heap Allocation**: 0.85 MB allocated memory.
- **OOM Events**: 0 OOM kills recorded since container inception.
- **Container Restarts**: 0 unexpected restarts.

---

## 4. Network & HTTP Baseline

- **HTTPS Response Time (p50)**: ~78 ms (via Cloudflare Singapore edge).
- **HTTPS Response Time (p95)**: ~142 ms.
- **HTTP Status Baseline**:
  - `GET https://lostmedia.pro/` -> HTTP/2 200 (HTML document ~33 kB).
  - `GET https://lostmedia.pro/health/ready` -> HTTP/2 200 (`{"status":"ready"}`).
- **Open Host Ports**:
  - Port 22/tcp (OpenSSH)
  - Port 80/tcp (Nginx Edge Proxy)
  - Port 443/tcp (Nginx Edge Proxy TLS)
- **All other services (5432, 6379, 5672, 8080, 3000)**: Bound strictly to internal Docker bridge `media_backend`.
