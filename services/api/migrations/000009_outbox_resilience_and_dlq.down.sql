DROP INDEX IF EXISTS idx_outbox_retention_cleanup;
DROP INDEX IF EXISTS idx_outbox_poll_ready;
ALTER TABLE outbox_events DROP COLUMN IF EXISTS next_retry_at;
ALTER TABLE outbox_events DROP COLUMN IF EXISTS last_error;
