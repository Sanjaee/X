# Backup, Disaster Recovery & High Availability Guide

This document defines the backup, point-in-time recovery, and disaster resilience procedures for the X-like social media platform.

---

## 1. High Availability vs. Backup vs. Disaster Recovery

It is vital to understand the architectural distinctions:

1. **High Availability (HA)**: Real-time redundancy (multi-node cluster with automated failover). If one node crashes, traffic instantly routes to another with zero perceptible downtime.
   > **Note**: A single 4 GB VPS with Docker Compose is **NOT** High Availability. If the physical host experiences a hardware failure or kernel panic, downtime will occur until the server reboots or is reprovisioned.
2. **Backup**: Point-in-time snapshots of raw data (PostgreSQL database dump, uploaded media files, configuration files) stored securely outside the primary VPS.
3. **Disaster Recovery (DR)**: The documented, reproducible procedure to rebuild the entire application from scratch on a brand-new VPS using only the off-site backups and git repository.

---

## 2. Automated PostgreSQL Daily Backup Script

Save this script as `scripts/backup-db.sh` on the VPS host:

```bash
#!/usr/bin/env bash
set -euo pipefail

BACKUP_DIR="/var/backups/x-db"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
BACKUP_FILE="${BACKUP_DIR}/x_db_${TIMESTAMP}.sql.gz"
RETENTION_DAYS=14

mkdir -p "${BACKUP_DIR}"

echo "[$(date)] Starting PostgreSQL backup..."

# Dump database through Docker container directly into compressed gzip
docker exec -t x-postgres pg_dump -U postgres -d x_db --clean --if-exists | gzip -9 > "${BACKUP_FILE}"

# Calculate SHA256 checksum
sha256sum "${BACKUP_FILE}" > "${BACKUP_FILE}.sha256"

# Verify dump is not empty
if [ ! -s "${BACKUP_FILE}" ]; then
    echo "[ERROR] Backup file is empty! Alerting ops..."
    exit 1
fi

echo "[$(date)] Backup completed successfully: ${BACKUP_FILE} ($(du -h "${BACKUP_FILE}" | cut -f1))"

# Prune backups older than retention window
find "${BACKUP_DIR}" -name "x_db_*.sql.gz*" -type f -mtime +${RETENTION_DAYS} -delete

# (Optional) Sync to offsite S3 or backup storage:
# aws s3 cp "${BACKUP_FILE}" s3://my-x-backups/daily/
```

Configure via host cron (`crontab -e`):
```cron
# Run daily database backup at 02:00 AM UTC
0 2 * * * /path/to/scripts/backup-db.sh >> /var/log/x-backup.log 2>&1
```

---

## 3. Disaster Recovery: Step-by-Step Restore Procedure

In the event of total server corruption or migrating to a new VPS:

### Step 1: Provision Clean Server & Install Docker
```bash
sudo apt-get update && sudo apt-get install -y docker.io docker-compose-plugin git gzip
```

### Step 2: Clone Codebase & Configure Environment
```bash
git clone https://github.com/my-org/x.git /opt/x
cd /opt/x
cp .env.example .env
# Edit .env with production database passwords and secrets
```

### Step 3: Start Supporting Infrastructure
```bash
docker compose -f deploy/docker/docker-compose.yml up -d x-postgres x-redis x-rabbitmq
```

### Step 4: Restore PostgreSQL Database from Backup
```bash
# 1. Copy backup archive to host
scp user@backup-server:/var/backups/x-db/x_db_20261015_020000.sql.gz /tmp/

# 2. Decompress and pipe directly into postgres container
gunzip -c /tmp/x_db_20261015_020000.sql.gz | docker exec -i x-postgres psql -U postgres -d x_db

# 3. Verify record integrity
docker exec -it x-postgres psql -U postgres -d x_db -c "SELECT count(*) FROM users; SELECT count(*) FROM posts;"
```

### Step 5: Start Application Services & Proxy
```bash
docker compose -f deploy/docker/docker-compose.yml up -d
docker restart x-proxy
```

### Step 6: Verify Service Health
```bash
curl -i http://localhost/health/ready
```

---

## 4. Cache & Queue Disaster Behavior

- **Redis**: Redis holds only ephemeral data (rate limits, presence, cache). When recovering from a disaster, Redis starts completely empty. The Go API automatically hydrates user profiles and timelines on demand from PostgreSQL.
- **RabbitMQ**: RabbitMQ queues are marked `durable: true`. Messages on disk survive container restarts. In the event of catastrophic message loss, the Transactional Outbox pattern in PostgreSQL allows re-publishing all pending or recent events:
  ```sql
  UPDATE outbox_events SET status = 'pending' WHERE created_at > NOW() - INTERVAL '24 hours';
  ```
