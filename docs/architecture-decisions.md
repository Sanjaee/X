# Architecture Decision Records (ADR)

## ADR-001: Modular Monolith vs Microservices for 4 GB VPS Target
- **Status**: Accepted
- **Context**: The platform is targeting deployment on a single VPS with 4 GB RAM and 2 vCPUs.
- **Decision**: Retain and optimize the Go + Next.js modular monolith architecture. Reject microservices.
- **Rationale**: A microservices approach would require multiple runtime containers, separate service discovery, higher base memory footprint (>2.5 GB idle), and inter-service network overhead. A modular monolith uses ~600 MB total idle RAM, fits comfortably within the 4 GB budget, and simplifies transactional consistency.

## ADR-002: In-Process Transactional Outbox Worker with Dead-Letter Handling
- **Status**: Accepted
- **Context**: Event broadcasting (notifications, real-time message alerts) must not block HTTP requests and must survive process restarts without data loss.
- **Decision**: Use a PostgreSQL-backed Transactional Outbox pattern processed by an in-process background worker with exponential backoff, random jitter, and a max-retry Dead-Letter cutoff (5 retries).
- **Rationale**: RabbitMQ consumes ~150-250MB idle RAM and introduces distributed commit complexity. PostgreSQL transactions guarantee that an event is only persisted if the underlying post/message is saved. In-process workers eliminate extra process overhead.

## ADR-003: Redis Ephemeral Caching & Cache-Aside Strategy
- **Status**: Accepted
- **Context**: Frequent database queries for active sessions and public user profiles during viral spikes.
- **Decision**: Use Redis 7 with an LRU eviction policy (`allkeys-lru`) capped at 96 MB for:
  1. Active user sessions (15-minute TTL, invalidated on logout).
  2. Public user profiles (5-minute TTL, invalidated on profile update).
  3. Sliding window rate limiting.
- **Rationale**: Reading public profiles and verifying authentication tokens on every request created avoidable DB load. Cache-aside with automatic invalidation provides sub-millisecond responses without risking stale data.

## ADR-004: Strict Magic-Byte Content Sniffing & Upload Throttling
- **Status**: Accepted
- **Context**: Media uploads directly buffered in RAM created Out-Of-Memory risks on 4 GB VPS and security risks from file extension spoofing.
- **Decision**: Enforce:
  1. Magic-byte verification using `http.DetectContentType` on the initial 512 bytes.
  2. Hard file size ceiling: 15 MB for images, 50 MB for videos.
  3. Context cancellation timeout on all media routes.
- **Rationale**: Prevents malicious executable/HTML uploads disguised as `.jpg` and prevents heap ballooning from concurrent large file uploads.

## ADR-005: Network Isolation and Port Hardening
- **Status**: Accepted
- **Context**: Container databases and caches should not be exposed to the public internet.
- **Decision**: In `docker-compose.prod.yml`, only Nginx (`x-proxy`) exposes public ports (`80:80`, `443:443`). PostgreSQL, Redis, and internal services communicate exclusively over an isolated internal Docker bridge network (`x_backend`).
- **Rationale**: Eliminates brute-force port scanning and credential attacks against database and cache ports directly from the internet.
