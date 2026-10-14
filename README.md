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

## 5. Running Verification Tests

Run the unified test suite:
```powershell
./scripts/test.ps1
```
This runs:
1. Go unit & integration tests (`TestHealthCheck`, `TestAuthAndPostingFlow`).
2. Next.js strict TypeScript check and production page generation (`npm run build`).
