-- ==============================================================================
-- Migration 000010 Rollback: TOTP Two-Factor Authentication (2FA) Support
-- ==============================================================================

ALTER TABLE users DROP COLUMN IF EXISTS totp_enabled;
ALTER TABLE users DROP COLUMN IF EXISTS totp_secret;

