-- ==============================================================================
-- Migration 000008: Post Interactions, Reposts, Quote Posts & Profile Tab Indexes
-- ==============================================================================

-- 1. Ensure quotes_count column exists on posts table
ALTER TABLE posts ADD COLUMN IF NOT EXISTS quotes_count INTEGER NOT NULL DEFAULT 0;

-- 2. Performance indexes for post interactions
CREATE INDEX IF NOT EXISTS idx_post_likes_user_created ON post_likes(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_post_likes_post ON post_likes(post_id);

CREATE INDEX IF NOT EXISTS idx_post_reposts_user_created ON post_reposts(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_post_reposts_post ON post_reposts(post_id);

-- 3. Profile tab optimization indexes
-- Posts tab: original posts and quotes (where reply_to_id IS NULL)
CREATE INDEX IF NOT EXISTS idx_posts_author_original ON posts(author_id, created_at DESC) WHERE reply_to_id IS NULL AND deleted_at IS NULL;

-- Replies tab: posts that reply to other posts (where reply_to_id IS NOT NULL)
CREATE INDEX IF NOT EXISTS idx_posts_author_replies ON posts(author_id, created_at DESC) WHERE reply_to_id IS NOT NULL AND deleted_at IS NULL;

-- Conversation threads and Quote relationships
CREATE INDEX IF NOT EXISTS idx_posts_reply_to_tree ON posts(reply_to_id, created_at ASC) WHERE reply_to_id IS NOT NULL AND deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_posts_quote_rel ON posts(quote_post_id, created_at DESC) WHERE quote_post_id IS NOT NULL AND deleted_at IS NULL;
