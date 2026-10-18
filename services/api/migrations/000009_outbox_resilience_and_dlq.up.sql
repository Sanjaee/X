-- ==============================================================================
-- Migration 000009: Outbox Resilience, Retry Schedule, DLQ & Retention Cleanup
-- ==============================================================================

ALTER TABLE outbox_events ADD COLUMN IF NOT EXISTS last_error TEXT NULL;
ALTER TABLE outbox_events ADD COLUMN IF NOT EXISTS next_retry_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

-- Index for polling pending or retrying outbox events ready for execution
CREATE INDEX IF NOT EXISTS idx_outbox_poll_ready ON outbox_events(next_retry_at ASC) WHERE status IN ('pending', 'retrying');

-- Index for retention cleanup of old published events
CREATE INDEX IF NOT EXISTS idx_outbox_retention_cleanup ON outbox_events(published_at) WHERE status = 'published';
