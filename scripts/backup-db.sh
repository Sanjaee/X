#!/usr/bin/env bash
# ==============================================================================
# LOSTMEDIA / X — AUTOMATED PRODUCTION DATABASE BACKUP & ROTATION
# ==============================================================================
set -euo pipefail

BACKUP_DIR="/var/backups/x-db"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
DB_NAME="${DB_NAME:-media_prod}"
DB_USER="${DB_USER:-postgres}"
BACKUP_FILE="${BACKUP_DIR}/${DB_NAME}_${TIMESTAMP}.sql.gz"
RETENTION_DAYS=14

mkdir -p "${BACKUP_DIR}"
chmod 700 "${BACKUP_DIR}"

echo "[$(date)] Starting PostgreSQL backup for database ${DB_NAME}..."

# Dump database through Docker container directly into compressed gzip
docker exec x-postgres pg_dump -U "${DB_USER}" -d "${DB_NAME}" --clean --if-exists | gzip -9 > "${BACKUP_FILE}"

# Calculate SHA256 checksum
sha256sum "${BACKUP_FILE}" > "${BACKUP_FILE}.sha256"

# Verify dump is not empty
if [ ! -s "${BACKUP_FILE}" ]; then
    echo "[$(date)] ERROR: Backup file is empty! Alerting ops..."
    exit 1
fi

FILESIZE=$(du -h "${BACKUP_FILE}" | cut -f1)
echo "[$(date)] Backup completed successfully: ${BACKUP_FILE} (${FILESIZE})"

# Prune backups older than retention window
find "${BACKUP_DIR}" -name "${DB_NAME}_*.sql.gz*" -type f -mtime +${RETENTION_DAYS} -delete
echo "[$(date)] Retention policy applied: removed archives older than ${RETENTION_DAYS} days."
