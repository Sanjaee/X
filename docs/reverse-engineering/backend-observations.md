# Backend Observations & Public API Evidence

**Status:** Phase 1 Complete  
**Evidence Source:** Live network payloads, embedded bootstrap streams, HTTP headers, and manifest evidence

---

## 1. Observable API Interface Patterns

### Protocol and Architecture
- **VERIFIED**: X Web uses **GraphQL over HTTP POST / GET** with streaming responses.
  - Evidence from `explore` payload:
    ```javascript
    dehydratedData: {
      messagesBeforeDehydration: [],
      graphqlRequestStream: (stream) => new ReadableStream({ start(controller) { ... } })
    }
    ```
  - Responses are delivered as incremental chunks via `ReadableStream` and hydrated directly into the client routing tree.
- **VERIFIED**: Static resources and media are strictly distributed across partitioned subdomains:
  - `api.x.com` (Main API Gateway)
  - `pbs.twimg.com` (User avatars, banners, post image thumbnails & full resolutions)
  - `video.twimg.com` (HLS playlists, video segments)
  - `abs.twimg.com` (Static client scripts, fonts, CSS)
  - `t.co` (URL shortener, click tracking, malware scanning)

### Data Serialization & Hydration
- **VERIFIED**: Timestamps are transmitted as numeric millisecond epochs (e.g., `1791986821000`), formatted client-side using Intl API:
  - Relative: `1m`, `5m`, `18h`, `Oct 13`
  - Absolute: `2:07 PM · Oct 14, 2026`
- **VERIFIED**: Data hydration uses **Seroval Stream** protocol (`__SEROVAL_STREAM__: true`) to stream serialized state without blocking initial HTML rendering.

---

## 2. Authentication and Session Management

### Observable Behavior
- **VERIFIED**: Unauthenticated access is permitted for public user profiles (`/{username}`) and public post statuses (`/{username}/status/{id}`), but interactions (Like, Repost, Bookmark, Reply) trigger redirect to onboarding:
  `/i/jf/onboarding/web?mode=login&redirect_after_login={url}`
- **VERIFIED**: Supported public onboarding mechanisms:
  - Phone / SMS verification
  - Google Identity Services (One Tap & Redirect)
  - Sign in with Apple
  - Email / Username + Password form
- **INFERRED**: Session tracking relies on standard `HttpOnly; Secure; SameSite=None/Lax` cookies, with an anti-CSRF token verified in headers for mutative operations.

---

## 3. Media Upload Flow

### Observable Behavior
- **VERIFIED**: Supported MIME types declared in Web App Manifest:
  - Images: `image/jpeg`, `image/png`, `image/gif`
  - Video: `video/mp4`, `video/quicktime`
- **INFERRED**: Multi-step chunked upload protocol for media:
  1. `INIT`: Client sends total byte size, media category, and MIME type.
  2. `APPEND`: Client streams binary chunks (up to 5MB each).
  3. `FINALIZE`: Server assembles chunks, triggers async processing (transcoding, virus check).
  4. Post creation references the resulting `media_id`.

---

## 4. Unknowns (Internal Infrastructure)

In compliance with empirical reverse engineering discipline:
- **UNKNOWN**: Internal database technology (whether PostgreSQL, Cassandra, MySQL, CockroachDB, or custom key-value store).
- **UNKNOWN**: Internal RPC protocols between API gateway and downstream services (e.g., gRPC, Finagle, Thrift).
- **UNKNOWN**: Proprietary recommendation algorithm weights and model architecture.

---

## 5. Our Proposed Backend Architecture

To achieve equivalent functionality with maintainable, robust engineering:
- **Language**: **Go (Golang)** for high concurrency, low memory footprint, and rapid execution.
- **Router**: Clean HTTP routing using standard `net/http` or `go-chi/chi/v5`.
- **API Paradigm**: **RESTful JSON API under `/api/v1`**, augmented with **WebSocket** for real-time notifications and timeline events.
- **Documentation**: OpenAPI 3.0 specification (`docs/api/openapi.yaml`).
- **Database**: **PostgreSQL 16+** with relational schema, foreign keys, and indexes.
- **Storage**: S3-compatible API (MinIO in local dev, AWS S3 / Cloudflare R2 in prod).
