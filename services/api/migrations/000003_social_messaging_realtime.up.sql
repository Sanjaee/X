-- ==============================================================================
-- Migration 000003: Social Graph, Advanced Direct Messaging, Realtime Outbox
-- ==============================================================================

-- 1. Follow Requests (for Private Accounts)
CREATE TABLE IF NOT EXISTS follow_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    requester_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    target_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- 'pending', 'accepted', 'rejected', 'cancelled'
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(requester_id, target_id),
    CHECK (requester_id != target_id)
);

CREATE INDEX IF NOT EXISTS idx_follow_requests_target ON follow_requests(target_id, status, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_follow_requests_requester ON follow_requests(requester_id, status);

-- 2. Enhance Messages for Idempotency, Attachments, and Threading
ALTER TABLE messages ADD COLUMN IF NOT EXISTS client_message_id VARCHAR(100) NULL;
ALTER TABLE messages ADD COLUMN IF NOT EXISTS message_type VARCHAR(20) NOT NULL DEFAULT 'text';
ALTER TABLE messages ADD COLUMN IF NOT EXISTS reply_to_id UUID NULL REFERENCES messages(id) ON DELETE SET NULL;
ALTER TABLE messages ADD COLUMN IF NOT EXISTS edited_at TIMESTAMPTZ NULL;
ALTER TABLE messages ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ NULL;

CREATE INDEX IF NOT EXISTS idx_messages_client_id ON messages(conversation_id, client_message_id) WHERE client_message_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_messages_cursor ON messages(conversation_id, created_at DESC, id DESC);

-- 3. Conversation Read States
ALTER TABLE conversation_members ADD COLUMN IF NOT EXISTS last_read_message_id UUID NULL REFERENCES messages(id) ON DELETE SET NULL;

CREATE TABLE IF NOT EXISTS conversation_read_states (
    conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    last_read_message_id UUID NULL REFERENCES messages(id) ON DELETE SET NULL,
    last_read_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (conversation_id, user_id)
);

-- 4. Message Reactions
CREATE TABLE IF NOT EXISTS message_reactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    message_id UUID NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    emoji VARCHAR(16) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(message_id, user_id, emoji)
);

CREATE INDEX IF NOT EXISTS idx_message_reactions_msg ON message_reactions(message_id);

-- 5. Notification Preferences & Outbox
CREATE TABLE IF NOT EXISTS notification_preferences (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    email_dms BOOLEAN NOT NULL DEFAULT TRUE,
    email_follows BOOLEAN NOT NULL DEFAULT TRUE,
    push_dms BOOLEAN NOT NULL DEFAULT TRUE,
    push_mentions BOOLEAN NOT NULL DEFAULT TRUE,
    push_likes BOOLEAN NOT NULL DEFAULT TRUE,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 6. Enhance Notifications for Deduplication and Granular Entities
ALTER TABLE notifications ADD COLUMN IF NOT EXISTS entity_type VARCHAR(30) NULL;
ALTER TABLE notifications ADD COLUMN IF NOT EXISTS entity_id UUID NULL;
ALTER TABLE notifications ADD COLUMN IF NOT EXISTS dedupe_key VARCHAR(100) NULL;
ALTER TABLE notifications ADD COLUMN IF NOT EXISTS read_at TIMESTAMPTZ NULL;

CREATE INDEX IF NOT EXISTS idx_notifications_dedupe ON notifications(recipient_id, dedupe_key) WHERE is_read = FALSE AND dedupe_key IS NOT NULL;

-- 7. Transactional Outbox Events
CREATE TABLE IF NOT EXISTS outbox_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_type VARCHAR(50) NOT NULL,
    aggregate_type VARCHAR(50) NOT NULL,
    aggregate_id UUID NOT NULL,
    payload JSONB NOT NULL DEFAULT '{}'::jsonb,
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- 'pending', 'published', 'failed'
    retry_count INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    published_at TIMESTAMPTZ NULL
);

CREATE INDEX IF NOT EXISTS idx_outbox_pending ON outbox_events(created_at ASC) WHERE status = 'pending';
