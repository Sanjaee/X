DROP TABLE IF EXISTS outbox_events CASCADE;
DROP TABLE IF EXISTS notification_preferences CASCADE;
DROP TABLE IF EXISTS message_reactions CASCADE;
DROP TABLE IF EXISTS conversation_read_states CASCADE;
DROP TABLE IF EXISTS follow_requests CASCADE;

ALTER TABLE messages DROP COLUMN IF EXISTS client_message_id;
ALTER TABLE messages DROP COLUMN IF EXISTS message_type;
ALTER TABLE messages DROP COLUMN IF EXISTS reply_to_id;
ALTER TABLE messages DROP COLUMN IF EXISTS edited_at;
ALTER TABLE messages DROP COLUMN IF EXISTS deleted_at;

ALTER TABLE conversation_members DROP COLUMN IF EXISTS last_read_message_id;

ALTER TABLE notifications DROP COLUMN IF EXISTS entity_type;
ALTER TABLE notifications DROP COLUMN IF EXISTS entity_id;
ALTER TABLE notifications DROP COLUMN IF EXISTS dedupe_key;
ALTER TABLE notifications DROP COLUMN IF EXISTS read_at;
