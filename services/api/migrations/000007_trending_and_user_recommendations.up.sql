-- ==============================================================================
-- Migration 000007: Trending Analytics, Acceleration Indexes & Recommendation Dismissals
-- ==============================================================================

-- 1. Table for tracking dismissed account recommendations with cooldown period
CREATE TABLE IF NOT EXISTS dismissed_user_recommendations (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    target_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    dismissed_until TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, target_user_id)
);

CREATE INDEX IF NOT EXISTS idx_dismissed_rec_user ON dismissed_user_recommendations(user_id, dismissed_until);

-- 2. Acceleration indexes for rolling window trending queries
CREATE INDEX IF NOT EXISTS idx_posts_trending_window ON posts(created_at DESC, author_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_post_hashtags_composite ON post_hashtags(hashtag_id, post_id);
