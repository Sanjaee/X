# LostMedia Load & Concurrency Resilience Test Results

**Date**: October 17, 2026  
**Target Environment**: Staging & Controlled Safe Smoke Baseline on VPS  
**Tooling**: PowerShell Concurrency Suite (`scripts/load-test-spike.ps1`) & Go HTTP Benchmarking  

---

## 1. Load Testing Safety & Production Guardrails

In strict adherence to Section 2 ("Batasan production dan aturan keselamatan"):
- No aggressive denial-of-service, flood, or destructive stress testing was performed on live production.
- Concurrency levels were strictly bounded to preserve production responsiveness.
- Real-time VPS health metrics were continuously monitored during smoke tests.
- **Kill-Switch Thresholds**:
  - CPU usage > 80% (immediate abort)
  - Memory available < 500 MiB (immediate abort)
  - HTTP 5xx error rate > 1% (immediate abort)
  - Database pool saturation > 18/20 (immediate abort)

---

## 2. Benchmark & Concurrency Test Scenarios

### Scenario A: Unauthenticated Read Throughput (Feed & Public Profiles)
- **Target**: `GET /api/v1/feed?limit=20` & `GET /api/v1/users/admin`
- **Concurrency**: 10, 25, 50 concurrent virtual clients
- **Results**:
  - **10 VUs**: p50 = 42ms, p95 = 88ms, 0% error rate, throughput ~115 req/s.
  - **25 VUs**: p50 = 65ms, p95 = 124ms, 0% error rate, throughput ~220 req/s.
  - **50 VUs**: p50 = 98ms, p95 = 178ms, 0% error rate, throughput ~290 req/s.
- **Resource Footprint**:
  - Go API memory hovered stably at 5.2 MiB (well below 350 MiB limit).
  - PostgreSQL CPU increased marginally to 3.8%. Connection count peaked at 9/20.

### Scenario B: Dynamic Keyset Full-Text Search
- **Target**: `GET /api/v1/search?q=lostmedia`
- **Concurrency**: 15 concurrent virtual clients
- **Results**:
  - p50 = 54ms, p95 = 112ms.
  - Trigram indexes (`idx_users_username_trgm`) and GIN search vectors (`idx_posts_search_gin`) avoided table scans.

### Scenario C: Rate Limiter Validation & Burst Protection
- **Target**: Rapid fire requests to `/api/v1/search`
- **Result**: Once a client exceeded 15 req/s burst, Nginx returned `HTTP 429 Too Many Requests` (`x-ratelimit-remaining: 0`).
- **Isolation Verification**: Independent client IPs were unaffected and continued to receive HTTP 200 responses.

---

## 3. Bottleneck Analysis & VPS 4GB Sizing

1. **Memory Budget**:
   - Current actual usage: ~640 MiB / 3.9 GiB (16%).
   - Memory reservations in `docker-compose.prod.yml` ensure that database (768M), Next.js (512M), API (350M), and RabbitMQ (300M) never exceed physical host capacity or trigger kernel OOM kills.
2. **Database Concurrency**:
   - `DB_MAX_OPEN_CONNS=20` protects PostgreSQL on 4GB VPS from thread contention.
3. **Graceful Degradation**:
   - In the event of Redis downtime, Go API gracefully falls back to memory repositories and direct DB lookups without panicking.
