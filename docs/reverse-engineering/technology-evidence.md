# Technology Evidence Matrix

**Status:** Phase 1 Complete  
**Classification Rules:**
- `VERIFIED`: Directly supported by verifiable public artifact or network evidence.
- `INFERRED`: Plausible deduction based on observable behaviors and industry standards.
- `UNKNOWN`: Cannot be determined from observable external evidence.
- `PROPOSED`: Chosen for our independent microblogging platform implementation.

---

## 1. Comprehensive Technology Matrix

| Technology Domain | Target Observation | Classification | Evidence / Basis | Proposed Stack for Our Reconstruction |
| :--- | :--- | :--- | :--- | :--- |
| **Frontend Framework** | TanStack Router + React SSR Streaming | **VERIFIED** | `self.$_TSR`, `$_TSR.router`, hydration stream scripts in HTML | Next.js 15+ App Router (React 19, Server Components) |
| **Serialization** | Seroval Stream | **VERIFIED** | `__SEROVAL_STREAM__: true` in explore HTML payload | Standard JSON + React RSC Streaming Protocol |
| **CSS & Styling** | Tailwind CSS v4.3.3 | **VERIFIED** | `/*! tailwindcss v4.3.3 */` header in `styles-BPdMqb99.css` | Tailwind CSS with custom theme matching X design tokens |
| **Typography** | TwitterChirp Custom Font | **VERIFIED** | Preloaded `Chirp-Regular.woff2`, `Chirp-Bold.woff2` | Inter / Outfit from Google Fonts |
| **Iconography** | Embedded Inline SVG Symbols | **VERIFIED** | `icon-reply-stroke`, `icon-retweet-stroke`, `icon-heart-stroke` | Lucide React icons matching standard stroke dimensions |
| **Error Monitoring** | Sentry | **VERIFIED** | Meta tag `sentry-trace`, `sentry-public_key` | Sentry SDK (optional) / structured slog logging |
| **PWA & Manifest** | Web App Manifest v2 | **VERIFIED** | `manifest.json` with shortcuts (`compose/post`, `explore`, `notifications`) | Web App Manifest + service worker offline capability |
| **Push Notifications** | Firebase Cloud Messaging (GCM) | **VERIFIED** | `gcm_sender_id: "49625052041"` in `manifest.json` | Web Push API (VAPID keys) + WebSocket |
| **API Architecture** | GraphQL over HTTP + ReadableStream | **VERIFIED** | `graphqlRequestStream`, streaming ReadableStream in DOM | REST API (`/api/v1`) + WebSocket for real-time |
| **Backend Language** | Unobservable from public web | **UNKNOWN** | Server hides runtime headers; historical Scala/Java unproven | Go (Golang 1.23+) |
| **HTTP Server / Router** | Unobservable from public web | **UNKNOWN** | Reverse proxy headers only (`x-frame-options`, etc.) | Go `net/http` + Chi router |
| **Database (Primary)** | Relational / Distributed Graph | **UNKNOWN** | Backend internals hidden behind API gateway | PostgreSQL 16+ (relational, ACID, normalized) |
| **Database Querying** | Unobservable | **UNKNOWN** | No raw SQL or DB traces exposed in public output | sqlc / database/sql with migration scripts |
| **Caching Layer** | Redis / Memcached | **INFERRED** | Sub-millisecond timeline delivery & rate-limiting behavior | Redis 7+ for cache, sessions, and rate limiting |
| **Media Storage** | Distributed Blob Storage | **INFERRED** | Dedicated CDNs (`pbs.twimg.com`, `video.twimg.com`) | S3-Compatible Storage (MinIO dev, AWS S3 / Cloudflare R2 prod) |
| **Reverse Proxy** | Edge Proxy / CDN | **INFERRED** | Cloudflare / Fastly CDN headers and TLS termination | Nginx / Caddy reverse proxy in Docker container |
| **Deployment / Containers**| Kubernetes / Mesos (historical) | **UNKNOWN** | Private cloud/data center topology unobservable | Docker & Docker Compose modular deployment |

---

## 2. Key Takeaways for Reconstruction

1. **Styling Parity**: Because X officially employs Tailwind CSS, our implementation can reproduce identical visual spacing, flexbox alignment, and responsive behaviors natively using standard Tailwind utility classes.
2. **Layout Exactness**: We have exact pixel-level layout specifications extracted directly from X's stylesheet:
   - Left Navigation Rail: `88px` (compact) / `275px` (full)
   - Primary Feed: `600px`
   - Right Rail: `420px` (min `258px`)
   - Sticky Header: `53px`
3. **Decoupled Architecture**: While X uses GraphQL streaming, a well-structured REST API with OpenAPI contracts and WebSocket provides superior DX, faster execution in Go, and straightforward testability for our independent platform.
