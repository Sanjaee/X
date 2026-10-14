-- Migration 000002: Subscriptions, Verification Badges, and Entitlements

-- 1. Alter posts table to allow long posts up to 10,000 characters (enforced by entitlement engine)
ALTER TABLE posts ALTER COLUMN content TYPE VARCHAR(10000);

-- 2. Subscription Plans Catalog
CREATE TABLE IF NOT EXISTS subscription_plans (
    id VARCHAR(50) PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    description TEXT NOT NULL,
    monthly_price_cents INTEGER NOT NULL DEFAULT 0,
    annual_price_cents INTEGER NOT NULL DEFAULT 0,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    max_post_length INTEGER NOT NULL DEFAULT 280,
    can_get_verified_badge BOOLEAN NOT NULL DEFAULT FALSE,
    can_edit_posts BOOLEAN NOT NULL DEFAULT FALSE,
    can_undo_posts BOOLEAN NOT NULL DEFAULT FALSE,
    max_video_minutes INTEGER NOT NULL DEFAULT 2,
    max_media_count INTEGER NOT NULL DEFAULT 4,
    reduced_ads BOOLEAN NOT NULL DEFAULT FALSE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Seed Initial Tiered Plans (Free, Basic, Premium, Premium+)
INSERT INTO subscription_plans (id, name, description, monthly_price_cents, annual_price_cents, currency, max_post_length, can_get_verified_badge, can_edit_posts, can_undo_posts, max_video_minutes, max_media_count, reduced_ads)
VALUES
('free', 'Free', 'Standard microblogging experience for everyone.', 0, 0, 'USD', 280, FALSE, FALSE, FALSE, 2, 4, FALSE),
('basic', 'Basic', 'Essential features: edit posts, bookmark folders, and up to 1,000 characters.', 300, 3000, 'USD', 1000, FALSE, TRUE, TRUE, 10, 4, FALSE),
('premium', 'Premium', 'The signature X experience: Blue verification badge, 4,000 characters, 60min video, 50% fewer ads.', 800, 8400, 'USD', 4000, TRUE, TRUE, TRUE, 60, 4, TRUE),
('premium_plus', 'Premium+', 'Everything in Premium with no ads in For You, largest reply boost, 10,000 characters, and exclusive creator tools.', 1600, 16800, 'USD', 10000, TRUE, TRUE, TRUE, 180, 4, TRUE)
ON CONFLICT (id) DO NOTHING;

-- 3. Subscriptions Table
CREATE TABLE IF NOT EXISTS subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    plan_id VARCHAR(50) NOT NULL REFERENCES subscription_plans(id),
    billing_cycle VARCHAR(20) NOT NULL DEFAULT 'monthly', -- 'monthly' or 'annual'
    status VARCHAR(30) NOT NULL DEFAULT 'pending', -- 'pending', 'active', 'past_due', 'canceled', 'expired', 'suspended'
    current_period_start TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    current_period_end TIMESTAMPTZ NOT NULL,
    cancel_at_period_end BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_subscriptions_user_status ON subscriptions(user_id, status);

-- 4. Payment Orders Table
CREATE TABLE IF NOT EXISTS payment_orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    plan_id VARCHAR(50) NOT NULL REFERENCES subscription_plans(id),
    billing_cycle VARCHAR(20) NOT NULL DEFAULT 'monthly',
    amount_cents INTEGER NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    provider VARCHAR(30) NOT NULL DEFAULT 'plisio',
    provider_order_id VARCHAR(255) NULL,
    provider_invoice_url TEXT NULL,
    status VARCHAR(30) NOT NULL DEFAULT 'pending', -- 'pending', 'completed', 'expired', 'failed', 'refunded'
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_payment_orders_user ON payment_orders(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_payment_orders_provider_id ON payment_orders(provider_order_id);

-- 5. Payment Transactions / Webhook History
CREATE TABLE IF NOT EXISTS payment_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL REFERENCES payment_orders(id) ON DELETE CASCADE,
    txn_id TEXT NULL,
    amount_cents INTEGER NOT NULL,
    currency VARCHAR(10) NOT NULL,
    status VARCHAR(30) NOT NULL,
    raw_payload JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 6. Invoices Table
CREATE TABLE IF NOT EXISTS invoices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL REFERENCES payment_orders(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    amount_cents INTEGER NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    status VARCHAR(20) NOT NULL DEFAULT 'paid',
    pdf_url TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_invoices_user ON invoices(user_id, created_at DESC);

-- 7. Verification Badges
CREATE TABLE IF NOT EXISTS verification_badges (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    badge_type VARCHAR(20) NOT NULL DEFAULT 'blue', -- 'blue', 'gold', 'grey', 'staff'
    display_label VARCHAR(50) NOT NULL DEFAULT 'Verified',
    description TEXT NOT NULL DEFAULT 'Verified Account',
    color_token VARCHAR(20) NOT NULL DEFAULT '#1d9bf0',
    status VARCHAR(20) NOT NULL DEFAULT 'approved', -- 'pending', 'approved', 'rejected', 'suspended', 'revoked'
    expires_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_user_badge UNIQUE (user_id, badge_type)
);

CREATE INDEX IF NOT EXISTS idx_verification_badges_user ON verification_badges(user_id, status);

-- 8. Verification Applications
CREATE TABLE IF NOT EXISTS verification_applications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    badge_type VARCHAR(20) NOT NULL DEFAULT 'blue',
    category VARCHAR(50) NOT NULL DEFAULT 'individual',
    document_url TEXT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- 'pending', 'approved', 'rejected', 'suspended'
    review_notes TEXT NULL,
    reviewed_by UUID NULL REFERENCES users(id),
    reviewed_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_verification_app_status ON verification_applications(status, created_at DESC);
