DROP TABLE IF EXISTS recent_searches;
DROP INDEX IF EXISTS idx_posts_search_keyset;
DROP INDEX IF EXISTS idx_user_profiles_display_name_trgm;
DROP INDEX IF EXISTS idx_users_username_trgm;
DROP INDEX IF EXISTS idx_posts_search_vector;
ALTER TABLE posts DROP COLUMN IF EXISTS search_vector;
