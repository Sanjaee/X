# Architecture Decision Records (ADR)

**Status:** Phase 3 Approved  
**Project:** X / Twitter Independent Full-Stack Reconstruction

---

## ADR 1: Backend Language — Go (Golang)

### Context
A microblogging platform requires high concurrency (thousands of concurrent timeline reads and WebSocket connections), low latency (< 20ms p95 for API responses), and minimal memory overhead per connection.

### Alternatives Considered
1. **Node.js / Express or Fastify**: Easy TypeScript code sharing with frontend, but single-threaded event loop can suffer during heavy CPU tasks (e.g., password hashing, media hashing, serialization) and uses significantly more memory per WebSocket connection.
2. **Python / FastAPI**: Rich ecosystem, but higher latency, GIL limitations, and higher memory footprint under high concurrency.
3. **Go (Golang)**: Native goroutines (cheap concurrency, ~2KB stack), statically typed, compiled binary, rapid startup, deterministic garbage collection, excellent standard library.

### Decision
**Adopt Go (1.23+)** with a modular monolith architecture. Business domain logic is kept strictly isolated from transport protocols (HTTP/WebSocket) and persistence engines.

---

## ADR 2: Primary Database — PostgreSQL (Relational)

### Context
Social data exhibits strict relational integrity: posts belong to users, likes reference posts and users, reply threads form trees, and follow relationships form directional graphs.

### Alternatives Considered
1. **MongoDB / Document Store**: Allows flexible schema, but lacks ACID multi-document transactions across complex operations (e.g. creating a post, updating hashtag counts, generating notifications, and debiting rate limits simultaneously). Joins and referential integrity must be manually maintained in application code.
2. **Neo4j / Dedicated Graph DB**: Excellent for graph traversal, but high operational complexity and poor fit for large text content and time-series feeds.
3. **PostgreSQL 16+**: Industry-standard relational integrity, powerful JSONB when needed, built-in full-text search (`tsvector`), excellent indexing (B-Tree, GIN, BRIN), and battle-tested ACID transactions.

### Decision
**Adopt PostgreSQL 16+** with fully normalized tables, explicit foreign keys, `ON DELETE CASCADE / SET NULL` constraints, and optimized composite indexes for timeline queries.

---

## ADR 3: API Architecture — RESTful JSON + WebSocket (vs GraphQL)

### Context
The investigated target (`x.com`) uses GraphQL streaming. However, implementing an enterprise-grade GraphQL gateway with DataLoader, streaming fragments, and persisted queries introduces excessive maintenance overhead for an independent modular project.

### Alternatives Considered
1. **GraphQL Gateway**: Flexible client field selection, but complex caching, complex query cost analysis, and difficult N+1 prevention without extensive tooling.
2. **REST API with OpenAPI Contract + WebSocket**:
   - Standard HTTP semantics (`GET`, `POST`, `PATCH`, `DELETE`).
   - Clean URL routing (`/api/v1/posts`, `/api/v1/users/{username}`).
   - Deterministic caching in Redis or HTTP edge proxies.
   - Type-safe client generation via OpenAPI 3.0 specification.
   - Dedicated WebSocket channel for real-time notifications and direct messages.

### Decision
**Adopt versioned REST API (`/api/v1`)** with OpenAPI 3.0 specification, paired with a lightweight Go WebSocket hub for bidirectional events.

---

## ADR 4: Frontend Framework — Next.js 15 App Router

### Context
Public social media platforms require immediate first-contentful-paint (FCP), dynamic metadata for SEO and link unfurling (Open Graph cards on social previews), and rich interactive client states.

### Decision
**Adopt Next.js 15 App Router** with React Server Components (RSC) by default for public profiles and post detail pages, paired with Client Components for dynamic timeline interaction and TanStack Query for cache invalidation.
