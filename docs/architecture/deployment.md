# Deployment & Infrastructure Specification

**Status:** Phase 3 Approved  
**Platform Target:** Single VPS (MVP) scaling to Distributed Cluster

---

## 1. Container Topology (Docker Compose)

The entire platform runs predictably in local development and single-node production via Docker Compose:

| Container Name | Service | Base Image | Port Mapping | Purpose |
| :--- | :--- | :--- | :--- | :--- |
| `x-proxy` | Nginx Edge Proxy | `nginx:alpine` | `80:80`, `443:443` | SSL termination, reverse proxy, gzip/brotli compression |
| `x-web` | Next.js Frontend | `node:20-alpine` | `3000:3000` (internal) | Server-Side Rendering (SSR) & Static HTML/CSS/JS delivery |
| `x-api` | Go API Backend | `golang:1.23-alpine` (multi-stage) | `8080:8080` (internal) | Core REST API, WebSocket Hub, business logic |
| `x-postgres` | PostgreSQL DB | `postgres:16-alpine` | `5432:5432` | Primary ACID relational database |
| `x-redis` | Redis Cache | `redis:7-alpine` | `6379:6379` | Sessions, ephemeral rate limits, timeline cache |
| `x-minio` | MinIO Object Store | `minio/minio` | `9000:9000`, `9001:9001` | S3-compatible local development object storage |

---

## 2. Health & Readiness Probes

Both the Go API and Next.js services expose standardized health check endpoints:
- `GET /health/live`: Returns `200 OK` if the process is running.
- `GET /health/ready`: Returns `200 OK` only if database ping and Redis ping succeed.
  ```json
  {
    "status": "healthy",
    "database": "connected",
    "redis": "connected",
    "version": "1.0.0",
    "uptime_seconds": 3600
  }
  ```

---

## 3. Database Migration Strategy

- Migrations run automatically before API server startup using a dedicated migration runner (`golang-migrate/migrate` or embedded SQL runner in Go).
- Migrations are versioned with timestamps:
  - `000001_initial_schema.up.sql` / `.down.sql`
  - `000002_add_indexes.up.sql` / `.down.sql`
- Zero-downtime guidelines: Add columns with `NULL` or default values; avoid dropping active columns in a single deployment step.
