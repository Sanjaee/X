DROP INDEX IF EXISTS idx_posts_quote_rel;
DROP INDEX IF EXISTS idx_posts_reply_to_tree;
DROP INDEX IF EXISTS idx_posts_author_replies;
DROP INDEX IF EXISTS idx_posts_author_original;
DROP INDEX IF EXISTS idx_post_reposts_post;
DROP INDEX IF EXISTS idx_post_reposts_user_created;
DROP INDEX IF EXISTS idx_post_likes_post;
DROP INDEX IF EXISTS idx_post_likes_user_created;
ALTER TABLE posts DROP COLUMN IF EXISTS quotes_count;
