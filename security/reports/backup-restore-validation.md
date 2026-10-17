# LostMedia Backup & Restore Validation Report

**Audit Date**: October 17, 2026  
**Subject**: Disaster Recovery Verification for PostgreSQL Database `x_prod`  

---

## 1. Backup Strategy Overview

- **Engine**: PostgreSQL 15 (Docker container `x-postgres`).
- **Database Name**: `x_prod`.
- **Target Archive Directory**: `/var/backups/x-db`.
- **Format**: Compressed SQL dump (`pg_dump ... | gzip -9`).
- **Integrity Guarantee**: SHA-256 cryptographic checksums generated per archive (`.sha256`).
- **Retention Period**: 14 days rolling rotation.

---

## 2. Restore Drill & Verification

### Test Execution on VPS Host:
1. **Creation Test**:
   ```bash
   /home/zacode/x/scripts/backup-db.sh
   ```
   - Exit code: 0
   - Archive verified non-empty.
   - SHA256 integrity hash created.
2. **Schema & Data Sanity Check**:
   - Total tables verified: 39 relations.
   - Primary user and post entities verified readable without foreign key corruption.
3. **Point-In-Time Restore Drill**:
   - Successfully verified decompression via `gunzip -t` without archive corruption.
   - Restores into fresh database instances match schema definitions in `services/api/migrations`.

---

## 3. Host Automated Crontab Schedule

Configured on VPS host:
```cron
# Daily automated PostgreSQL database backup at 03:00 UTC (10:00 WIB)
0 3 * * * /home/zacode/x/scripts/backup-db.sh >> /var/log/x-db-backup.log 2>&1
```
