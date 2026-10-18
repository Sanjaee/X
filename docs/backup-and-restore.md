# Backup & Disaster Recovery Guide

## 1. Automated PostgreSQL Daily Backup

The backup script is located at `scripts/backup-db.sh`. It dumps the database directly through the `x-postgres` container into a compressed `.sql.gz` file with a SHA256 checksum and a 14-day retention rotation policy.

### Manual Backup Command
```bash
bash scripts/backup-db.sh
```

### Automated Host Cron Configuration
```cron
# Run daily database backup at 02:00 AM UTC
0 2 * * * /opt/x/scripts/backup-db.sh >> /var/log/x-backup.log 2>&1
```

---

## 2. Disaster Recovery & Database Restoration

To restore a database snapshot into `x-postgres`:

```bash
# 1. Uncompress and restore database dump into container
gunzip -c /var/backups/x-db/x_prod_YYYYMMDD_HHMMSS.sql.gz | docker exec -i x-postgres psql -U postgres -d x_prod

# 2. Reconcile counters and indexes
docker exec -it x-postgres psql -U postgres -d x_prod -c "VACUUM ANALYZE;"

# 3. Verify user and post integrity
docker exec -it x-postgres psql -U postgres -d x_prod -c "SELECT COUNT(*) FROM users; SELECT COUNT(*) FROM posts;"
```

---

## 3. RPO and RTO Targets

- **RPO (Recovery Point Objective)**: 24 hours with daily snapshots (or <15 minutes if PostgreSQL WAL archiving is enabled in Phase B).
- **RTO (Recovery Time Objective)**: <15 minutes to provision a new VPS, execute Docker Compose up, and restore the `.sql.gz` database dump.
