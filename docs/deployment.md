# Production Deployment Guide (Single 4GB RAM VPS)

## 1. Quick Start / Production Deploy

To deploy the production-hardened stack on a clean or updated VPS:

```bash
# 1. Clone repository into workspace
cd /opt/x

# 2. Configure environment from clean template
cp .env.example .env
# Set your secure DB_PASSWORD, REDIS_PASSWORD, NEXTAUTH_SECRET, JWT_SECRET in .env

# 3. Build and launch all services with Docker Compose
docker compose -f deploy/docker/docker-compose.prod.yml build
docker compose -f deploy/docker/docker-compose.prod.yml up -d --remove-orphans

# 4. Restart Nginx to flush upstream resolver cache
docker restart x-proxy

# 5. Verify live readiness
curl -f http://localhost/health/ready
```

---

## 2. Container Ports & Security Model

All containers run inside the internal isolated bridge network (`x_backend`).

| Container | Image | Ports (Host Bound) | Security Scope |
| :--- | :--- | :--- | :--- |
| `x-proxy` | `nginx:alpine` | `80:80`, `443:443` | **Public** (Internet Facing) |
| `x-web` | `node:20-alpine` | None (Internal Port 3000) | Isolated (`x_backend`) |
| `x-api` | `golang:1.24-alpine` | None (Internal Port 8080) | Isolated (`x_backend`) |
| `x-postgres`| `postgres:15-alpine` | None (Internal Port 5432) | Isolated (`x_backend`) |
| `x-redis` | `redis:7-alpine` | None (Internal Port 6379) | Isolated (`x_backend`) |
| `x-rabbitmq`| `rabbitmq:3-alpine` | None (Internal Port 5672) | Isolated (`x_backend`) |

---

## 3. Database Migration Procedures

Migrations reside in `services/api/migrations/`.
On server boot, the Go backend automatically checks and bootstraps missing schema updates (`ensureSearchSchema`, `ensureRBACSchema`, `ensureOutboxSchema`).
To inspect database migration status manually:
```bash
docker exec -it x-postgres psql -U postgres -d x_prod -c "\dt"
```

---

## 4. Rollback Procedure

If a deployed version experiences unexpected behavior or errors:

```bash
# 1. Revert Git repository to the previous stable commit
git checkout <previous_stable_commit>

# 2. Rebuild the application containers
docker compose -f deploy/docker/docker-compose.prod.yml build api web

# 3. Recreate the containers with zero stale IP cache
docker compose -f deploy/docker/docker-compose.prod.yml up -d --force-recreate api web
docker restart x-proxy

# 4. Verify health status
curl -i http://localhost/health/ready
```
