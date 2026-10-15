-- ==============================================================================
-- Migration 000005: Full-Text Search Vector, Trigram Fuzzy Indexes, and Recent Searches
-- ==============================================================================

-- 1. Enable Trigram Extension for Substring & Fuzzy Matching
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- 2. Add Generated tsvector Column to Posts for Full-Text Search
DO $$ 
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'posts' AND column_name = 'search_vector'
    ) THEN
        ALTER TABLE posts
        ADD COLUMN search_vector tsvector
        GENERATED ALWAYS AS (to_tsvector('simple', COALESCE(content, ''))) STORED;
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_posts_search_vector 
ON posts USING GIN (search_vector);

-- 3. Trigram Indexes for User Handle & Display Name Autocomplete
CREATE INDEX IF NOT EXISTS idx_users_username_trgm 
ON users USING GIN (username gin_trgm_ops);

CREATE INDEX IF NOT EXISTS idx_user_profiles_display_name_trgm 
ON user_profiles USING GIN (display_name gin_trgm_ops);

-- 4. Keyset Pagination & Filtering Indexes for Search
CREATE INDEX IF NOT EXISTS idx_posts_search_keyset 
ON posts (created_at DESC, id DESC) 
WHERE deleted_at IS NULL;

-- 5. User Search History / Recent Searches Table
CREATE TABLE IF NOT EXISTS recent_searches (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    query VARCHAR(255) NOT NULL,
    searched_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, query)
);

CREATE INDEX IF NOT EXISTS idx_recent_searches_user 
ON recent_searches (user_id, searched_at DESC);
