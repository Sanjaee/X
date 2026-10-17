-- Migration 000006: Dynamic Roles, Memberships, RBAC, Post Limits & Owner Control

-- 1. Roles Table
CREATE TABLE IF NOT EXISTS roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    is_system BOOLEAN NOT NULL DEFAULT FALSE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_roles_key ON roles(key);

-- 2. Permissions Table
CREATE TABLE IF NOT EXISTS permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key VARCHAR(100) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_permissions_key ON permissions(key);

-- 3. Role Permissions Join Table
CREATE TABLE IF NOT EXISTS role_permissions (
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (role_id, permission_id)
);

-- 4. User Roles Table
CREATE TABLE IF NOT EXISTS user_roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    assigned_by UUID NULL REFERENCES users(id) ON DELETE SET NULL,
    assigned_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NULL,
    reason TEXT NULL,
    CONSTRAINT uq_user_role UNIQUE (user_id, role_id)
);

CREATE INDEX IF NOT EXISTS idx_user_roles_user ON user_roles(user_id);
CREATE INDEX IF NOT EXISTS idx_user_roles_role ON user_roles(role_id);

-- 5. Membership Plans Table
CREATE TABLE IF NOT EXISTS membership_plans (
    id VARCHAR(50) PRIMARY KEY,
    key VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    price_amount_cents INTEGER NOT NULL DEFAULT 0,
    currency VARCHAR(10) NOT NULL DEFAULT 'USD',
    billing_interval VARCHAR(20) NOT NULL DEFAULT 'monthly',
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 6. Membership Entitlements Table
CREATE TABLE IF NOT EXISTS membership_entitlements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    plan_id VARCHAR(50) NOT NULL REFERENCES membership_plans(id) ON DELETE CASCADE,
    entitlement_key VARCHAR(50) NOT NULL,
    limit_value BIGINT NOT NULL DEFAULT 0,
    unit VARCHAR(50) NOT NULL DEFAULT 'count',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_plan_entitlement UNIQUE (plan_id, entitlement_key)
);

CREATE INDEX IF NOT EXISTS idx_entitlements_plan ON membership_entitlements(plan_id);

-- 7. User Memberships Table
CREATE TABLE IF NOT EXISTS user_memberships (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    plan_id VARCHAR(50) NOT NULL REFERENCES membership_plans(id) ON DELETE CASCADE,
    status VARCHAR(30) NOT NULL DEFAULT 'active',
    starts_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_user_membership UNIQUE (user_id, plan_id)
);

CREATE INDEX IF NOT EXISTS idx_user_memberships_user_status ON user_memberships(user_id, status);

-- 8. User Usage Counters Table (Atomic Quota Engine)
CREATE TABLE IF NOT EXISTS user_usage_counters (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    period_key VARCHAR(50) NOT NULL,
    period_start TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    posts_created INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_user_period UNIQUE (user_id, period_key)
);

CREATE INDEX IF NOT EXISTS idx_usage_counters_user_period ON user_usage_counters(user_id, period_key);

-- 9. Audit Logs Table
CREATE TABLE IF NOT EXISTS audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    actor_user_id UUID NULL REFERENCES users(id) ON DELETE SET NULL,
    action VARCHAR(100) NOT NULL,
    target_type VARCHAR(50) NOT NULL,
    target_id VARCHAR(100) NOT NULL,
    reason TEXT NOT NULL DEFAULT '',
    metadata_redacted JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_audit_logs_actor ON audit_logs(actor_user_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_target ON audit_logs(target_type, target_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_created ON audit_logs(created_at DESC);

-- 10. Augment posts table if missing visibility or moderation_status
ALTER TABLE posts ADD COLUMN IF NOT EXISTS visibility VARCHAR(20) NOT NULL DEFAULT 'public';
ALTER TABLE posts ADD COLUMN IF NOT EXISTS moderation_status VARCHAR(20) NOT NULL DEFAULT 'active';

-- ==========================================
-- SEED INITIAL SYSTEM DATA (Idempotent)
-- ==========================================

-- Seed Roles
INSERT INTO roles (key, name, description, is_system, is_active)
VALUES
('member', 'Member', 'Default standard role for all registered users', TRUE, TRUE),
('vip', 'VIP', 'Special membership role with enhanced limits and benefits', TRUE, TRUE),
('mvp', 'MVP', 'High tier membership role with top limits and creator tools', TRUE, TRUE),
('mod', 'Moderator', 'Content moderation and report management', TRUE, TRUE),
('god', 'God Admin', 'High-level platform administrator with explicit administrative controls', TRUE, TRUE),
('owner', 'Owner', 'Platform Owner with full governance and security oversight', TRUE, TRUE)
ON CONFLICT (key) DO UPDATE SET 
    name = EXCLUDED.name, 
    description = EXCLUDED.description, 
    is_system = EXCLUDED.is_system;

-- Seed Permissions
INSERT INTO permissions (key, name, description)
VALUES
-- Post
('post.create', 'Create Post', 'Ability to create new posts subject to tier quota'),
('post.read', 'Read Post', 'Ability to read public posts and feeds'),
('post.update.own', 'Update Own Post', 'Ability to edit content of own posts'),
('post.delete.own', 'Delete Own Post', 'Ability to delete own posts'),
('post.update.any', 'Update Any Post', 'Administrative ability to edit any post'),
('post.delete.any', 'Delete Any Post', 'Moderation ability to delete or purge any post'),
('post.pin.own', 'Pin Own Post', 'Ability to pin a post to own profile'),
('post.pin.any', 'Pin Any Post', 'Ability to pin global announcement posts'),
-- Media
('media.upload.image', 'Upload Image', 'Ability to upload image assets'),
('media.upload.video', 'Upload Video', 'Ability to upload video assets'),
('media.delete.own', 'Delete Own Media', 'Ability to delete uploaded media'),
('media.delete.any', 'Delete Any Media', 'Ability to remove any media asset'),
-- Profile
('profile.update.own', 'Update Own Profile', 'Ability to edit own profile bio, avatar, and banner'),
('profile.update.any', 'Update Any Profile', 'Administrative ability to edit any user profile'),
('profile.manage.colors', 'Manage Profile Colors', 'Ability to customize profile theme and badge colors'),
('profile.view.analytics', 'View Analytics', 'Access to advanced post and impression analytics'),
-- Moderation
('moderation.post.hide', 'Hide Post', 'Hide offensive posts from public feeds'),
('moderation.post.restore', 'Restore Post', 'Restore hidden or reported posts'),
('moderation.user.warn', 'Warn User', 'Issue formal moderation warning to user'),
('moderation.user.suspend', 'Suspend User', 'Temporarily or permanently suspend user account'),
('moderation.report.manage', 'Manage Reports', 'Review and resolve reported posts and profiles'),
-- Membership
('membership.view', 'View Memberships', 'View available membership plans'),
('membership.manage.own', 'Manage Own Membership', 'Subscribe, renew, or cancel own membership'),
('membership.assign', 'Assign Membership', 'Administrative assignment of membership to users'),
('membership.revoke', 'Revoke Membership', 'Administrative revocation of membership'),
('membership.plan.manage', 'Manage Membership Plans', 'Create, update, or alter pricing and entitlements of plans'),
-- Administration
('role.read', 'Read Roles', 'View system roles and permission matrices'),
('role.create', 'Create Role', 'Create custom roles'),
('role.update', 'Update Role', 'Edit role metadata and assign permissions'),
('role.delete', 'Delete Role', 'Delete custom non-system roles'),
('permission.manage', 'Manage Permissions', 'Configure permission mappings across roles'),
('user.read', 'Read User List', 'Search and view detailed user accounts in admin dashboard'),
('user.update', 'Update User', 'Administrative edit of user properties'),
('user.suspend', 'Suspend User Account', 'Administrative freeze/suspension of user'),
('audit.read', 'Read Audit Logs', 'View administrative and security audit trail'),
('system.settings.manage', 'Manage System Settings', 'Configure core platform parameters')
ON CONFLICT (key) DO UPDATE SET 
    name = EXCLUDED.name, 
    description = EXCLUDED.description;

-- Helper function to map permissions to roles
DO $$
DECLARE
    r_member UUID;
    r_vip UUID;
    r_mvp UUID;
    r_mod UUID;
    r_god UUID;
    r_owner UUID;
BEGIN
    SELECT id INTO r_member FROM roles WHERE key = 'member';
    SELECT id INTO r_vip FROM roles WHERE key = 'vip';
    SELECT id INTO r_mvp FROM roles WHERE key = 'mvp';
    SELECT id INTO r_mod FROM roles WHERE key = 'mod';
    SELECT id INTO r_god FROM roles WHERE key = 'god';
    SELECT id INTO r_owner FROM roles WHERE key = 'owner';

    -- Member permissions
    INSERT INTO role_permissions (role_id, permission_id)
    SELECT r_member, id FROM permissions 
    WHERE key IN (
        'post.create', 'post.read', 'post.update.own', 'post.delete.own', 'post.pin.own',
        'media.upload.image', 'media.delete.own',
        'profile.update.own',
        'membership.view', 'membership.manage.own'
    )
    ON CONFLICT DO NOTHING;

    -- VIP permissions
    INSERT INTO role_permissions (role_id, permission_id)
    SELECT r_vip, id FROM permissions 
    WHERE key IN (
        'post.create', 'post.read', 'post.update.own', 'post.delete.own', 'post.pin.own',
        'media.upload.image', 'media.upload.video', 'media.delete.own',
        'profile.update.own', 'profile.manage.colors', 'profile.view.analytics',
        'membership.view', 'membership.manage.own'
    )
    ON CONFLICT DO NOTHING;

    -- MVP permissions
    INSERT INTO role_permissions (role_id, permission_id)
    SELECT r_mvp, id FROM permissions 
    WHERE key IN (
        'post.create', 'post.read', 'post.update.own', 'post.delete.own', 'post.pin.own',
        'media.upload.image', 'media.upload.video', 'media.delete.own',
        'profile.update.own', 'profile.manage.colors', 'profile.view.analytics',
        'membership.view', 'membership.manage.own'
    )
    ON CONFLICT DO NOTHING;

    -- Moderator permissions
    INSERT INTO role_permissions (role_id, permission_id)
    SELECT r_mod, id FROM permissions 
    WHERE key IN (
        'post.create', 'post.read', 'post.update.own', 'post.delete.own', 'post.pin.own', 'post.delete.any',
        'media.upload.image', 'media.upload.video', 'media.delete.own',
        'profile.update.own',
        'moderation.post.hide', 'moderation.post.restore', 'moderation.user.warn', 'moderation.user.suspend', 'moderation.report.manage',
        'user.read', 'audit.read',
        'membership.view', 'membership.manage.own'
    )
    ON CONFLICT DO NOTHING;

    -- God Admin permissions
    INSERT INTO role_permissions (role_id, permission_id)
    SELECT r_god, id FROM permissions 
    WHERE key IN (
        'post.create', 'post.read', 'post.update.own', 'post.delete.own', 'post.update.any', 'post.delete.any', 'post.pin.own', 'post.pin.any',
        'media.upload.image', 'media.upload.video', 'media.delete.own', 'media.delete.any',
        'profile.update.own', 'profile.update.any', 'profile.manage.colors', 'profile.view.analytics',
        'moderation.post.hide', 'moderation.post.restore', 'moderation.user.warn', 'moderation.user.suspend', 'moderation.report.manage',
        'role.read', 'user.read', 'user.update', 'user.suspend', 'audit.read', 'system.settings.manage',
        'membership.view', 'membership.manage.own', 'membership.assign', 'membership.revoke'
    )
    ON CONFLICT DO NOTHING;

    -- Owner permissions (ALL permissions)
    INSERT INTO role_permissions (role_id, permission_id)
    SELECT r_owner, id FROM permissions
    ON CONFLICT DO NOTHING;
END $$;

-- Seed Membership Plans
INSERT INTO membership_plans (id, key, name, description, price_amount_cents, currency, billing_interval, is_active)
VALUES
('free', 'free', 'Member', 'Free tier for standard community use', 0, 'USD', 'monthly', TRUE),
('vip', 'vip', 'VIP Member', 'Enhanced posting limit, video uploads, and verified VIP badge', 100, 'USD', 'monthly', TRUE),
('mvp', 'mvp', 'MVP', 'Maximum limits, highest upload quota, and top tier benefits', 100, 'USD', 'monthly', TRUE)
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    price_amount_cents = EXCLUDED.price_amount_cents,
    is_active = EXCLUDED.is_active;

-- Seed Membership Entitlements
-- Free Plan Entitlements: 2 posts/day, 2 images/post, 0 videos/post, 280 chars, 5MB file
INSERT INTO membership_entitlements (plan_id, entitlement_key, limit_value, unit)
VALUES
('free', 'posts_per_day', 2, 'count'),
('free', 'max_images_per_post', 2, 'count'),
('free', 'max_videos_per_post', 0, 'count'),
('free', 'max_file_size_mb', 5, 'megabytes'),
('free', 'max_post_length', 280, 'characters'),
('free', 'max_video_minutes', 0, 'minutes'),
('free', 'can_view_analytics', 0, 'boolean'),
('free', 'can_manage_colors', 0, 'boolean')
ON CONFLICT (plan_id, entitlement_key) DO UPDATE SET limit_value = EXCLUDED.limit_value;

-- VIP Plan Entitlements: 10 posts/day, 4 images/post, 1 video/post, 4000 chars, 50MB file, 10 min video
INSERT INTO membership_entitlements (plan_id, entitlement_key, limit_value, unit)
VALUES
('vip', 'posts_per_day', 10, 'count'),
('vip', 'max_images_per_post', 4, 'count'),
('vip', 'max_videos_per_post', 1, 'count'),
('vip', 'max_file_size_mb', 50, 'megabytes'),
('vip', 'max_post_length', 4000, 'characters'),
('vip', 'max_video_minutes', 10, 'minutes'),
('vip', 'can_view_analytics', 1, 'boolean'),
('vip', 'can_manage_colors', 1, 'boolean')
ON CONFLICT (plan_id, entitlement_key) DO UPDATE SET limit_value = EXCLUDED.limit_value;

-- MVP Plan Entitlements: 30 posts/day, 4 images/post, 1 video/post, 10000 chars, 200MB file, 60 min video
INSERT INTO membership_entitlements (plan_id, entitlement_key, limit_value, unit)
VALUES
('mvp', 'posts_per_day', 30, 'count'),
('mvp', 'max_images_per_post', 4, 'count'),
('mvp', 'max_videos_per_post', 1, 'count'),
('mvp', 'max_file_size_mb', 200, 'megabytes'),
('mvp', 'max_post_length', 10000, 'characters'),
('mvp', 'max_video_minutes', 60, 'minutes'),
('mvp', 'can_view_analytics', 1, 'boolean'),
('mvp', 'can_manage_colors', 1, 'boolean')
ON CONFLICT (plan_id, entitlement_key) DO UPDATE SET limit_value = EXCLUDED.limit_value;

-- Migrate existing users to have member role and free membership by default
DO $$
DECLARE
    r_member UUID;
    r_owner UUID;
    owner_user_id UUID;
BEGIN
    SELECT id INTO r_member FROM roles WHERE key = 'member';
    SELECT id INTO r_owner FROM roles WHERE key = 'owner';

    -- Ensure all users have member role
    INSERT INTO user_roles (user_id, role_id, reason)
    SELECT u.id, r_member, 'Default migration assignment'
    FROM users u
    ON CONFLICT (user_id, role_id) DO NOTHING;

    -- Ensure all users have free membership if they have no active subscription/membership
    INSERT INTO user_memberships (user_id, plan_id, status)
    SELECT u.id, 'free', 'active'
    FROM users u
    ON CONFLICT (user_id, plan_id) DO NOTHING;

    -- Bootstrap Owner account if afrizaahmad18@gmail.com exists
    SELECT id INTO owner_user_id FROM users WHERE LOWER(TRIM(email)) = 'afrizaahmad18@gmail.com' LIMIT 1;
    IF owner_user_id IS NOT NULL THEN
        -- Assign owner role
        INSERT INTO user_roles (user_id, role_id, reason)
        VALUES (owner_user_id, r_owner, 'Initial Bootstrap Owner Assignment')
        ON CONFLICT (user_id, role_id) DO NOTHING;

        -- Also update users table legacy column for backward compatibility
        UPDATE users SET role = 'owner' WHERE id = owner_user_id;

        -- Assign MVP membership
        INSERT INTO user_memberships (user_id, plan_id, status)
        VALUES (owner_user_id, 'mvp', 'active')
        ON CONFLICT (user_id, plan_id) DO UPDATE SET status = 'active';

        -- Log in audit_logs
        INSERT INTO audit_logs (actor_user_id, action, target_type, target_id, reason, metadata_redacted)
        VALUES (
            owner_user_id, 
            'owner.bootstrap', 
            'user', 
            owner_user_id::text, 
            'Initial platform owner bootstrapped from verified account afrizaahmad18@gmail.com',
            '{"email": "afrizaahmad18@gmail.com", "role": "owner", "membership": "mvp"}'::jsonb
        );
    END IF;
END $$;
