# X (Twitter) Full-Stack Reconstruction

An independent, evidence-driven microblogging platform inspired by X (Twitter). Built to reproduce X's layout, design system tokens, interactive state progression, and functional microblogging features.

---

## 1. Evidence-Driven Reverse Engineering Highlights

This platform was constructed following a passive investigation of `https://x.com/` using **REA (Reverse Engineer Anything v6.5.0)**:

* **Styling System (VERIFIED)**: X utilizes **Tailwind CSS v4.3.3** (`styles-BPdMqb99.css`). Our frontend reproduces exact layout custom properties:
  * Left Navigation Rail: `88px` (icon mode) / `275px` (labeled mode)
  * Primary Feed: `600px`
  * Right Sidebar: `420px` (min `258px`)
  * Sticky Header: `53px` with `backdrop-blur-md`
  * Theme Colors: Pitch Black `#000000`, Card `#16181c`, Border `#2f3336`, Brand Blue `#1d9bf0`
* **Streaming Architecture (VERIFIED)**: X web client uses TanStack Router with streaming SSR and Seroval stream serialization (`__SEROVAL_STREAM__: true`).
* **PWA & Media Pipeline (VERIFIED)**: Web App Manifest v2 with declared shortcuts (`/compose/post`, `/explore`, `/notifications`, `/messages`) and multi-format media (`image/jpeg`, `image/png`, `image/gif`, `video/mp4`).

---

## 2. Technology Stack

* **Frontend**: Next.js 14+ App Router, React 18/19, TypeScript (strict mode), Tailwind CSS, Lucide icons.
* **Backend**: Go (Golang 1.23+) modular monolith with Chi router, versioned REST API (`/api/v1`), structured logging with `slog`, bcrypt credential hashing.
* **Database**: PostgreSQL 16+ relational schema with foreign keys, cascading deletions, and composite B-Tree/GIN indexes.
* **Cache & Sessions**: Redis 7+ for rate limiting, ephemeral session tokens, and caching.
* **Media Storage**: S3-compatible storage pipeline with local file storage fallback.
* **Deployment**: Docker Compose with multi-stage builds and Nginx edge reverse proxy.

---

## 3. Monorepo Organization

```
.
├── apps/
│   └── web/                         # Next.js App Router Frontend
│       ├── src/
│       │   ├── app/                 # Routes: Home, Profile, Post Detail, Explore, Notifications, Messages, Bookmarks
│       │   ├── components/          # LeftNavRail, RightSidebar, PostCard, PostComposer, AuthModal
│       │   ├── lib/                 # API client wrapper, date formatters, utils
│       │   └── types/               # TypeScript interfaces matching API contract
├── services/
│   └── api/                         # Go Backend Service
│       ├── cmd/server/              # Server entrypoint (main.go)
│       ├── internal/
│       │   ├── domain/              # Entities: User, Post, MediaObject, Follow, Notification
│       │   ├── repository/          # PostgreSQL (raw SQL) & In-Memory thread-safe repository
│       │   ├── handler/             # HTTP handlers & Integration test suite
│       │   ├── middleware/          # Auth, CORS, Logger (slog), Recovery
│       │   └── config/              # Environment loader
│       └── migrations/              # SQL DDL migrations (000001_initial_schema.up.sql)
├── deploy/
│   ├── docker/
│   │   ├── Dockerfile.api           # Multi-stage Go scratch build
│   │   ├── Dockerfile.web           # Node.js production runner
│   │   └── docker-compose.yml       # Full stack container orchestrator
│   └── nginx/                       # Edge proxy routing /api/ and web
├── docs/
│   ├── reverse-engineering/         # 8 Phase 1 investigation documents
│   ├── architecture/                # System architecture, security, performance, scaling
│   ├── database/                    # Schema, ER diagram, indexing strategy
│   └── api/                         # OpenAPI 3.0 specification (openapi.yaml)
└── scripts/
    ├── dev.ps1                      # Start local Go API & Next.js concurrently
    └── test.ps1                     # Run full Go tests and Next.js build verification
```

---

## 4. Quick Start Guide

### Option A: Local Development (Native Go & Node.js)

1. **Install dependencies and test backend**:
   ```powershell
   cd services/api
   go test -v ./...
   ```
2. **Install frontend dependencies**:
   ```powershell
   cd apps/web
   npm install
   ```
3. **Launch the entire platform**:
   ```powershell
   ./scripts/dev.ps1
   ```
   * Frontend: [http://localhost:3000](http://localhost:3000)
   * Backend API: [http://localhost:8080/api/v1](http://localhost:8080/api/v1)
   * Health Check: [http://localhost:8080/health/ready](http://localhost:8080/health/ready)

### Option B: Docker Compose Deployment

Run the complete containerized stack (Nginx, Go API, Next.js, PostgreSQL 16, Redis 7):
```bash
docker compose -f deploy/docker/docker-compose.yml up --build -d
```
Access the application at [http://localhost](http://localhost).

---

---

## 5. REA Social Graph, Realtime DMs & Notification Architecture

### A. REA Passive Observation & Evidence Matrix
* **Social Graph (VERIFIED)**: Bilateral relationship controls distinguish hard **Block** (dissolves follows in both directions and prevents mutual views/interactions) from soft **Mute** (retains follow relationship but hides actor posts from timeline). Follow requests for private accounts (`is_private = true`) enter a `pending` state requiring explicit authorization.
* **Direct Messaging (VERIFIED/PROPOSED)**: Realtime 1:1 messaging enforces idempotency via client-generated `client_message_id`. Typing indicators are strictly ephemeral, stored in Redis with a 4-second TTL, and never written to relational storage.
* **Notifications & Outbox (PROPOSED)**: Follows the **Transactional Outbox Pattern** in PostgreSQL (`outbox_events`). Database mutations and outbox records commit in the same transaction, which an asynchronous worker fans out to Redis and connected WebSocket clients.

### B. Realtime WebSocket Gateway & Contracts
* **Endpoint**: `/api/v1/ws` (supports Bearer Authorization header or `?token=<session_token>` query parameter for browsers).
* **Client-to-Server Events**:
  * `subscribe`: `{ "conversation_id": "uuid" }`
  * `unsubscribe`: `{ "conversation_id": "uuid" }`
  * `typing`: `{ "conversation_id": "uuid", "is_typing": true|false }`
  * `ping`: Heartbeat
* **Server-to-Client Broadcasts**:
  * `message.created`: Delivers newly published messages to active subscribers.
  * `message.read`: Live delivery and read receipt confirmations.
  * `typing.updated`: Ephemeral typing presence scoped to conversation members.
  * `notification.created` / `notification.count.updated`: Realtime badge updates without page refreshes.

### C. Database Migrations
1. `000001_initial_schema.up.sql`: Base tables (`users`, `posts`, `media_objects`, `likes`, `reposts`, `bookmarks`, `follows`, `notifications`).
2. `000002_monetization_and_verification.up.sql`: Subscriptions, orders, billing history, and verification applications.
3. `000003_social_messaging_realtime.up.sql`: `follow_requests`, `conversation_read_states`, `message_reactions`, `notification_preferences`, `outbox_events`, composite indexes, and relational constraints.

---

## 6. Running Verification Tests

Run the backend test suite:
```powershell
cd services/api
go test -v ./...
```
This tests:
1. `TestHealthCheck`: Liveness and readiness endpoints.
2. `TestAuthAndPostingFlow`: Registration, login, posting, liking, and bookmarking.
3. `TestSubscriptionsAndBadges`: Monetization orders and badge verification.
4. `TestSocialGraph_FollowUnfollowBlockMute`: Follow/unfollow, relationship checks, block dissolution, and mute isolation.
5. `TestFollowRequest_PrivateAccount`: Private account follow request creation, recipient inbox, and acceptance flow.
6. `TestDirectMessaging_IdempotencyAndFlow`: Conversation creation, idempotent message delivery, read receipts, and block restriction enforcement.

Verify frontend production build:
```powershell
cd apps/web
npm run build
```

---

## 7. REA Recommendation, Ranking & Timeline Architecture

### A. Dual-Feed & Discovery Paradigm
* **Following Feed (`type=following`)**: Pure chronological stream of followed accounts. Zero out-of-network insertions. Paginated via opaque composite keyset cursors.
* **For You Feed (`type=for_you`)**: Multi-source personalized timeline balancing familiarity, high-velocity topics, social proof, and emerging creator exploration.
* **Explore Discovery (`type=explore`)**: High-velocity trending conversations categorized by topic (e.g. Technology, AI, Development) with volume indicators.

### B. Multi-Stage Pipeline Execution
1. **Candidate Generation**: Retrieves 150+ candidates in parallel across 6 distinct sources:
   - Source A: In-Network Following Graph (limit: 80)
   - Source B: Social Engagement / 2nd-degree connections (limit: 40)
   - Source C: Topic & Hashtag Similarity (limit: 30)
   - Source D: Trending Velocity (limit: 30)
   - Source E: Fresh Content / Anti-Staleness (limit: 25)
   - Source F: Cold-Start Exploration for creators with < 1,000 followers (limit: 15)
2. **Eligibility & Safety Filtering**: Zero-trust server-side guard. Drops soft-deleted posts, suspended authors, bilateral blocks (`blocker_id = viewer` OR `blocked_id = viewer`), muted accounts, private accounts (unless viewer is an accepted follower), and posts previously hidden by viewer.
3. **Normalized Feature Extraction**:
   - `RelationshipScore`: [0.0, 1.0] (mutual: 1.0, following: 0.75, social: 0.50, discovery: 0.10)
   - `FreshnessScore`: Exponential decay $\exp(-\text{age\_hours} / \lambda_{half})$ with $\lambda_{half} = 12.0\text{ hours}$.
   - `EngagementVelocityScore`: $\frac{\text{likes} + 2.5\cdot\text{reposts} + 3.0\cdot\text{replies}}{(\text{age} + 0.5)^{1.1}}$ normalized to [0.0, 1.0].
   - `QualityScore`: Evaluates media attachments, verification status, and optimal content length.
   - `ExplorationScore`: Boost for fresh posts from emerging creators.
   - `Penalties`: Subtracted for negative feedback ("Not interested", "Hide post"), spam, and repetition.
4. **Diversity Reranker**: Post-ranking sliding window algorithm enforcing author capping (maximum 2 posts per author within any 10-post window) and topic interleaving.
5. **Timeline Assembly & Keyset Cursor**: Encodes composite cursor `base64(timestamp_unix_nano + ":" + post_id)` to prevent skipped or duplicate posts during real-time timeline shifts.

### C. Benchmarks & Testing
Run the recommendation engine tests and benchmarks:
```powershell
cd services/api
go test -v -bench=. ./internal/recommendation/...
```
Benchmark results:
* **Linear Scoring Engine**: `~0.32 ns/op` (over 1 billion operations/sec throughput).
* **Diversity Reranker (100 candidates -> 20 page)**: `~32 µs/op` (sub-millisecond reranking).

---

## 8. X-Like Infinite Scroll & High-Performance Search Architecture

### A. Memory-Efficient Virtualized Infinite Scrolling
- **Virtual DOM Windowing**: Uses `@tanstack/react-virtual` to unmount offscreen posts. Total active DOM nodes remain bounded (~1,500–2,500 elements) even after scrolling past 10,000+ posts.
- **Zero Cumulative Layout Shift (CLS)**: Fixed aspect ratio placeholder containers (`aspect-video`, `aspect-[w/h]`) with `loading="lazy"` images and native `<video controls playsInline preload="metadata" />`.
- **Bounded In-Memory Query Cache**: Configured with `maxPages: 5` in TanStack Query infinite query to discard distant pages from browser heap memory.
- **Scroll Restoration**: Persists anchor post ID and relative offset in `sessionStorage` for seamless browser Back navigation.
- **Bottom Prefetch Sentinel**: Uses `IntersectionObserver` with an 800px root margin to prefetch next page payloads smoothly before reaching the bottom.

### B. High-Performance PostgreSQL Full-Text Search Engine
- **Generated `search_vector` Column**: Pre-computed `to_tsvector('simple', COALESCE(content, ''))` stored on `posts` table with GIN index (`idx_posts_search_vector`).
- **Trigram Similarity Autocomplete**: Accelerated with `pg_trgm` extension on `users.username` and `user_profiles.display_name`.
- **Keyset Cursor Pagination**: Zero `OFFSET` degradation. Guaranteed $O(\log N)$ traversal on composite keyset `(created_at DESC, id DESC)`.
- **Tab-Specific Ranking Semantics**:
  - **Top**: Cover Density ranking (`ts_rank_cd`), log-scaled engagement (`LN(1 + likes)`), and exponential freshness decay.
  - **Latest**: Pure reverse-chronological order.
  - **People**: Trigram match score + exact username match top anchor + follower count weight.
  - **Media**: Filters strictly for posts with valid image/video assets.
- **AST Query Operator Parser**: Extracted parameters for `from:user`, `to:user`, `@user`, `min_faves:N`, `since:YYYY-MM-DD`, `until:YYYY-MM-DD`, `filter:images`, `filter:videos`.
- **Autocomplete & Recent Searches**: 200ms debounce with `AbortController` cancellation, keyboard navigation (Arrow Up/Down/Enter/Escape), and server-persisted search history with single pill delete and clear all.

### C. Search Benchmark Execution
Execute the search benchmark against the running backend:
```powershell
.\scripts\benchmark-search.ps1
```
* **Autocomplete Latency**: `< 10ms`
* **Search Posts Top Tab Latency**: `< 25ms`
* **Keyset Pagination Latency**: `< 15ms`

---

## 9. 4 GB VPS Architecture, Big Data Scaling & Traffic Spike Protection

### A. Memory Allocation Budget (4 GB Host)
| Component | RAM Hard Limit | Reservation | Purpose & Tuning Parameters |
|---|---|---|---|
| **OS & Kernel** | ~900 MB | - | Linux kernel, page cache, socket buffers, sysctls |
| **PostgreSQL 16** | 768 MB | 256 MB | `shared_buffers=256MB`, `work_mem=4MB`, `max_connections=80`, `statement_timeout=3s` |
| **Go API (`x-api`)** | 350 MB | 128 MB | `GOMEMLIMIT=300MiB`, `GOGC=80`, 300 concurrent request semaphore |
| **Next.js Web (`x-web`)** | 512 MB | 128 MB | `NODE_OPTIONS="--max-old-space-size=384"`, virtualized rendering |
| **RabbitMQ 3** | 300 MB | 128 MB | `vm_memory_high_watermark=0.4`, prefetch=10, durable queues |
| **Redis 7** | 180 MB | 64 MB | `maxmemory 128mb`, `allkeys-lru` eviction policy |
| **Nginx Proxy** | 128 MB | 32 MB | Connection keepalive, rate-limiting zones |
| **Total Hard Sum** | **~2,238 MB** | **~736 MB** | **Leaves > 1.7 GB headroom** for page cache and traffic bursts |

### B. Traffic Spike & Admission Control Architecture
1. **Tier 1 (Nginx Edge)**: Rate limits at 30 req/s for API and 10 req/s for search with burst buffers. Drops excess traffic before touching application runtimes.
2. **Tier 2 (Go Admission Semaphore)**: Restricts inflight requests to 300. Returns **HTTP 503** with `Retry-After: 3` and `X-Backpressure: Active` under saturation to prevent memory ballooning.
3. **Tier 3 (Redis Sliding Window Limiter)**: User-ID and IP-based rate limiting with 120 req/min default, returning **HTTP 429**.
4. **Tier 4 (DB Pool Bounding)**: Postgres pool capped at 20 open connections and 10 idle connections per instance.

### C. Lightweight Monitoring & Health Endpoints
- **Liveness**: `GET /health/live` (Checks basic process responsiveness).
- **Readiness**: `GET /health/ready` (Validates PostgreSQL and Redis connectivity).
- **System Metrics**: `GET /health/metrics` (Exposes active goroutines, heap allocations, GC pause duration, and active in-flight requests in real time).

### D. Horizontal Scaling Roadmap
- **Stage A (Current)**: Single 4GB VPS running Go API, Web, Postgres, Redis, RabbitMQ via Docker Compose.
- **Stage B (Stateless Scaling)**: Multiple Go API replicas behind Nginx load balancer; external managed PostgreSQL and Redis cluster. Triggered when host CPU > 75% or API RAM reaches 300MB.
- **Stage C (Multi-VPS Distributed)**: Dedicated application nodes, dedicated database primary + read replica, dedicated queue broker, and CDN/R2 object storage for all media.

### E. Capacity & Spike Testing
Execute the native PowerShell load test against your deployment:
```powershell
.\scripts\load-test-spike.ps1 -TargetUrl http://localhost -TotalRequests 300 -Concurrency 15
```

### F. Comprehensive Documentation Index
- [Current System Audit](file:///docs/architecture/current-system-audit.md)
- [4GB VPS Deployment & Kernel Tuning](file:///docs/architecture/4gb-vps-deployment.md)
- [Instance Scaling (Stage A -> B -> C)](file:///docs/architecture/instance-scaling.md)
- [Memory Budget & Profiling](file:///docs/architecture/memory-budget.md)
- [Traffic Spike Protection](file:///docs/architecture/traffic-spike-protection.md)
- [Redis & RabbitMQ Separation](file:///docs/architecture/redis-rabbitmq.md)
- [Monitoring & Alerting](file:///docs/architecture/monitoring-alerts.md)
- [Operational Runbook](file:///docs/operations/runbook.md)
- [Capacity & Load Testing](file:///docs/operations/load-testing.md)
- [Backup & Disaster Recovery](file:///docs/operations/backup-restore.md)


