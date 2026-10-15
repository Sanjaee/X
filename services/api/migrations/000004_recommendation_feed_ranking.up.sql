-- ==============================================================================
-- Migration 000004: Recommendation Feed Ranking, Feedback, and Personalization
-- ==============================================================================

-- 1. Feed Impressions Tracking (with Exposure Dwell Time)
CREATE TABLE IF NOT EXISTS feed_impressions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NULL REFERENCES users(id) ON DELETE CASCADE,
    post_id UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    feed_type VARCHAR(30) NOT NULL DEFAULT 'for_you',
    dwell_time_ms INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_feed_impressions_user ON feed_impressions(user_id, post_id, created_at DESC) WHERE user_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_feed_impressions_post ON feed_impressions(post_id, created_at DESC);

-- 2. Feed Feedback (Hide, Not Interested, Report)
CREATE TABLE IF NOT EXISTS feed_feedback (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    post_id UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    feedback_type VARCHAR(30) NOT NULL, -- 'hide', 'not_interested', 'show_less', 'report'
    topic VARCHAR(100) NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, post_id, feedback_type)
);

CREATE INDEX IF NOT EXISTS idx_feed_feedback_user ON feed_feedback(user_id, post_id);
CREATE INDEX IF NOT EXISTS idx_feed_feedback_type ON feed_feedback(user_id, feedback_type);

-- 3. User Topic & Interest Preferences (Explicit & Inferred)
CREATE TABLE IF NOT EXISTS user_topic_preferences (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    topic VARCHAR(100) NOT NULL,
    affinity_score DOUBLE PRECISION NOT NULL DEFAULT 1.0,
    is_explicit BOOLEAN NOT NULL DEFAULT FALSE,
    is_muted BOOLEAN NOT NULL DEFAULT FALSE,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, topic)
);

CREATE INDEX IF NOT EXISTS idx_user_topic_pref ON user_topic_preferences(user_id, affinity_score DESC);
CREATE INDEX IF NOT EXISTS idx_user_topic_muted ON user_topic_preferences(user_id) WHERE is_muted = TRUE;

-- 4. Precomputed Post Features & Content Metadata
CREATE TABLE IF NOT EXISTS post_features (
    post_id UUID PRIMARY KEY REFERENCES posts(id) ON DELETE CASCADE,
    quality_score DOUBLE PRECISION NOT NULL DEFAULT 0.5,
    primary_topic VARCHAR(100) NOT NULL DEFAULT 'general',
    velocity_score DOUBLE PRECISION NOT NULL DEFAULT 0.0,
    has_media BOOLEAN NOT NULL DEFAULT FALSE,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_post_features_topic ON post_features(primary_topic, velocity_score DESC);
CREATE INDEX IF NOT EXISTS idx_post_features_quality ON post_features(quality_score DESC);

-- 5. Versioned Ranking Configurations & Experiments
CREATE TABLE IF NOT EXISTS ranking_experiments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(50) NOT NULL UNIQUE,
    version VARCHAR(20) NOT NULL,
    weights JSONB NOT NULL DEFAULT '{}'::jsonb,
    is_active BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Seed default baseline production ranking experiment
INSERT INTO ranking_experiments (name, version, weights, is_active)
VALUES (
    'production_baseline_v1',
    '1.0.0',
    '{
        "relationship_weight": 2.5,
        "freshness_weight": 3.0,
        "engagement_weight": 2.0,
        "topic_weight": 1.8,
        "quality_weight": 1.2,
        "exploration_weight": 1.0,
        "negative_penalty_weight": 5.0,
        "repetition_weight": 3.0,
        "freshness_half_life_hours": 12.0
    }'::jsonb,
    TRUE
)
ON CONFLICT (name) DO NOTHING;

-- 6. Additional Acceleration Indexes
CREATE INDEX IF NOT EXISTS idx_posts_trending_velocity ON posts(created_at DESC, likes_count DESC, reposts_count DESC) WHERE deleted_at IS NULL;
