-- Rollback Migration 000004
DROP INDEX IF EXISTS idx_posts_trending_velocity;
DROP TABLE IF EXISTS ranking_experiments;
DROP TABLE IF EXISTS post_features;
DROP TABLE IF EXISTS user_topic_preferences;
DROP TABLE IF EXISTS feed_feedback;
DROP TABLE IF EXISTS feed_impressions;
