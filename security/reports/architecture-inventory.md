# LostMedia Security Architecture Inventory

**Date of Audit**: October 17, 2026  
**Auditor**: REA Security & Resiliency Team  
**System Target**: [lostmedia.pro](https://lostmedia.pro/)  
**Environment**: Production (Single-node 4GB VPS, Ubuntu 22.04 LTS / Linux 5.15)  

---

## 1. System Overview & Technology Stack

| Layer | Component | Version / Engine | Runtime Exposure | Dependencies |
| :--- | :--- | :--- | :--- | :--- |
| **Edge / CDN** | Cloudflare CDN & WAF | Cloudflare Managed | Public (172.67.x.x, 104.21.x.x) | DNS, SSL Termination |
| **Edge Reverse Proxy** | Global Nginx | `nginx:alpine` (1.31.5) | Host Ports `80/tcp`, `443/tcp` | Docker `media_backend` |
| **Internal Reverse Proxy** | Application Nginx (`x-proxy`) | `nginx:alpine` | Internal Docker port `80` | `x-web`, `x-api` |
| **Frontend Application** | Next.js App Router | Next.js 14, React 18, TailwindCSS | Internal Docker port `3000` | Node.js 20, NextAuth v4/v5 |
| **Backend API Gateway** | Go Core API (`x-api`) | Go 1.24 (`cmd/server`) | Internal Docker port `8080` | PostgreSQL, Redis, RabbitMQ, R2 |
| **Primary Relational DB** | PostgreSQL (`x-postgres`) | PostgreSQL 15-alpine | Internal Docker port `5432` | Volume `x_postgres_data` |
| **In-Memory Cache & State**| Redis (`x-redis`) | Redis 7-alpine (LRU 180MB) | Internal Docker port `6379` | Memory limit 180M |
| **Message Broker** | RabbitMQ (`x-rabbitmq`) | RabbitMQ 3-alpine | Internal Docker port `5672` | Memory limit 300M |
| **Media Asset Storage** | Cloudflare R2 | S3-compatible API | Public CDN Domain | AWS SDK v2 (Go) |
| **Payment Gateway** | Plisio Cryptocurrency | REST API + IPN Webhook | Webhook `/webhooks/payments/plisio` | HMAC-SHA1 signature |

---

## 2. Network Topology & Container Isolation

```
[ Internet Client ]
        │
        ▼ (HTTPS / TLS 1.3)
[ Cloudflare Edge WAF / CDN ]
        │
        ▼ (Origin SSL / Port 443)
[ VPS Host Firewall (UFW) ] ── (Port 22 SSH, 80 HTTP, 443 HTTPS only)
        │
        ▼
[ Docker Container: global-nginx ] (0.0.0.0:80, 0.0.0.0:443)
        │ (Docker Bridge Network: media_backend)
        ▼
[ Docker Container: x-proxy ] (port 80 internal)
   ├───► /api/* ────────► [ Docker Container: x-api ] (port 8080 internal)
   ├───► /uploads/* ────► [ Docker Container: x-api ] (port 8080 internal)
   └───► /* ────────────► [ Docker Container: x-web ] (port 3000 internal)
                                │
                        [ Supporting Services ]
                        ├── x-postgres:5432 (isolated)
                        ├── x-redis:6379 (requirepass, isolated)
                        └── x-rabbitmq:5672 (isolated)
```

### Network Exposure Verification
- Direct port scanning and `ss -tulpn` verify that **PostgreSQL (5432), Redis (6379), RabbitMQ (5672/15672), Go API (8080), and Next.js (3000) are NOT bound to host 0.0.0.0**.
- Only `global-nginx` (`80`, `443`) and OpenSSH (`22`) are exposed on host network interfaces.

---

## 3. Data Flow & Security Perimeters

### A. Authentication Flow
1. **Google OAuth**: User initiates Google OAuth via NextAuth at `/api/auth/signin/google`. NextAuth handles callback and exchanges profile credentials with Go backend endpoint `/api/v1/auth/google`.
2. **Username / Password**: User posts credentials to `/api/v1/auth/login`. Password verified via `bcrypt.CompareHashAndPassword`.
3. **Session Token**: On successful authentication, Go backend generates a high-entropy cryptographically random session token (32 bytes raw), hashes it with SHA-256 for persistent database storage (`sessions` table), and issues:
   - `Set-Cookie: session_token=<raw_token>; HttpOnly; SameSite=Lax; Path=/; Max-Age=2592000`
   - JSON payload `session_token` (stored in frontend client localStorage).

### B. Media Ingestion Flow
1. Authenticated user uploads file to `POST /api/v1/media/upload`.
2. Go backend `MediaService` evaluates extension and MIME type.
3. JPEGs and PNGs are decoded, stripped of EXIF metadata (privacy preservation), scaled if width/height exceeds 2048px, compressed at 85% quality, and uploaded to Cloudflare R2 bucket `lostmedia`.
4. R2 public CDN URL is recorded in `media_objects` table and associated with posts.

### C. Realtime Messaging & Notification Flow
1. Client connects via WebSocket to `GET /api/v1/ws?token=<session_token>`.
2. Go API validates session token, resolves `viewerID`, and registers client in memory connection hub.
3. Realtime broadcast events (new direct messages, mentions, notifications) are published to Redis pub/sub channels (`realtime:messages`, `realtime:notifications`) and multiplexed to connected client sockets.

---

## 4. Dependencies & Third-Party Integrations

1. **Cloudflare**: DNS management, DDoS protection, edge caching, SSL origin certificate, R2 object storage.
2. **Google Identity**: OAuth 2.0 client for single sign-on.
3. **Plisio**: Crypto payment processing (BTC, USDT, ETH, LTC) via IPN webhooks.
