-- Migration 000006 Down: Revert dynamic roles, memberships, RBAC & quotas

DROP TABLE IF EXISTS audit_logs CASCADE;
DROP TABLE IF EXISTS user_usage_counters CASCADE;
DROP TABLE IF EXISTS user_memberships CASCADE;
DROP TABLE IF EXISTS membership_entitlements CASCADE;
DROP TABLE IF EXISTS membership_plans CASCADE;
DROP TABLE IF EXISTS user_roles CASCADE;
DROP TABLE IF EXISTS role_permissions CASCADE;
DROP TABLE IF EXISTS permissions CASCADE;
DROP TABLE IF EXISTS roles CASCADE;
