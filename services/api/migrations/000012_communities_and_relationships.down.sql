-- Migration 000012: Rollback Communities
DROP TABLE IF EXISTS community_posts CASCADE;
DROP TABLE IF EXISTS community_invitations CASCADE;
DROP TABLE IF EXISTS community_join_requests CASCADE;
DROP TABLE IF EXISTS community_rules CASCADE;
DROP TABLE IF EXISTS community_members CASCADE;
DROP TABLE IF EXISTS communities CASCADE;
