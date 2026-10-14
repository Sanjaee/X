# Scaling Roadmap: From Single Node to High-Traffic Cluster

**Status:** Phase 3 Approved  
**Evolution Stages:** Stage 1 (MVP) -> Stage 2 (Clustered) -> Stage 3 (Large Scale)

---

## Stage 1: Single Node VPS (0 - 50,000 MAU)
- **Deployment**: Single 8-Core / 16GB RAM VPS (Hetzner / DigitalOcean / AWS EC2).
- **Architecture**: Docker Compose with Nginx, Go API (1 instance), PostgreSQL (local SSD with WAL), Redis (single instance), MinIO or Cloudflare R2 for media.
- **Timeline Engine**: Pull-based fan-out on read with PostgreSQL composite indexes.
- **Estimated Cost**: $20 - $50 / month.

---

## Stage 2: Horizontally Scaled Stateless Services (50,000 - 1,000,000 MAU)
- **Deployment**: Managed Kubernetes (EKS / GKE) or multi-container service (ECS).
- **Database**: Managed PostgreSQL (AWS RDS / Supabase) with 1 Primary (writes) and 2 Read Replicas (read-heavy feeds).
- **API**: Multiple Go API pods auto-scaling behind an AWS ALB or Cloudflare Load Balancer.
- **WebSocket Synchronization**: Redis Pub/Sub backplane connecting multiple Go WebSocket nodes.
- **Background Jobs**: Dedicated RabbitMQ or Redis Streams worker pool for transcoding video, sending push notifications, and recalculating follower counts.

---

## Stage 3: Large-Scale Distributed Architecture (> 1,000,000 MAU)
- **Hybrid Fan-Out Engine**:
  - Regular users: Fan-out on read from partitioned read replicas.
  - Celebrity / High-follower accounts: Fan-out on write into pre-computed Redis Sorted Sets (`ZADD timeline:{user_id} {timestamp} {post_id}`).
- **Search**: Migration from PostgreSQL `tsvector` to dedicated OpenSearch / MeiliSearch cluster.
- **Global Edge Caching**: Cloudflare Enterprise with edge worker token verification for static user profile timelines.
