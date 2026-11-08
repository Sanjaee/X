-- ==============================================================================
-- Migration 000010: TOTP Two-Factor Authentication (2FA) Support
-- ==============================================================================

ALTER TABLE users ADD COLUMN IF NOT EXISTS totp_secret VARCHAR(128) NOT NULL DEFAULT '';
ALTER TABLE users ADD COLUMN IF NOT EXISTS totp_enabled BOOLEAN NOT NULL DEFAULT FALSE;

