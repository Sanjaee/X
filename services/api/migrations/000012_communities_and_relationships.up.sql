-- Migration 000012: Communities and Community Relationships

CREATE TABLE IF NOT EXISTS communities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug VARCHAR(64) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    avatar_url TEXT NOT NULL DEFAULT '',
    banner_url TEXT NOT NULL DEFAULT '',
    access_model VARCHAR(32) NOT NULL DEFAULT 'open', -- 'open', 'restricted'
    visibility VARCHAR(32) NOT NULL DEFAULT 'public',  -- 'public', 'unlisted'
    owner_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    members_count INT NOT NULL DEFAULT 1,
    posts_count INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_communities_slug ON communities(slug);
CREATE INDEX IF NOT EXISTS idx_communities_owner_id ON communities(owner_id);
CREATE INDEX IF NOT EXISTS idx_communities_discovery ON communities(access_model, visibility) WHERE deleted_at IS NULL;

-- Community Members
CREATE TABLE IF NOT EXISTS community_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    community_id UUID NOT NULL REFERENCES communities(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(32) NOT NULL DEFAULT 'member', -- 'owner', 'admin', 'moderator', 'member'
    status VARCHAR(32) NOT NULL DEFAULT 'active', -- 'active', 'suspended', 'banned'
    suspended_until TIMESTAMPTZ,
    joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_community_member UNIQUE (community_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_community_members_user ON community_members(user_id, status);
CREATE INDEX IF NOT EXISTS idx_community_members_comm_role ON community_members(community_id, role, status);

-- Community Rules
CREATE TABLE IF NOT EXISTS community_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    community_id UUID NOT NULL REFERENCES communities(id) ON DELETE CASCADE,
    rule_number INT NOT NULL,
    title VARCHAR(200) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_community_rule UNIQUE (community_id, rule_number)
);

-- Community Join Requests (For restricted communities)
CREATE TABLE IF NOT EXISTS community_join_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    community_id UUID NOT NULL REFERENCES communities(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status VARCHAR(32) NOT NULL DEFAULT 'pending', -- 'pending', 'approved', 'rejected', 'cancelled'
    response_note TEXT NOT NULL DEFAULT '',
    reviewed_at TIMESTAMPTZ,
    reviewed_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_comm_join_req_pending ON community_join_requests(community_id, status) WHERE status = 'pending';
CREATE INDEX IF NOT EXISTS idx_comm_join_req_user ON community_join_requests(user_id, community_id);

-- Community Invitations
CREATE TABLE IF NOT EXISTS community_invitations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    community_id UUID NOT NULL REFERENCES communities(id) ON DELETE CASCADE,
    inviter_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    invitee_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status VARCHAR(32) NOT NULL DEFAULT 'pending', -- 'pending', 'accepted', 'declined', 'expired'
    expires_at TIMESTAMPTZ NOT NULL DEFAULT (NOW() + INTERVAL '14 days'),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_comm_invites_invitee ON community_invitations(invitee_id, status) WHERE status = 'pending';

-- Community Posts Junction
CREATE TABLE IF NOT EXISTS community_posts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    post_id UUID UNIQUE NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    community_id UUID NOT NULL REFERENCES communities(id) ON DELETE CASCADE,
    is_pinned BOOLEAN NOT NULL DEFAULT FALSE,
    pinned_at TIMESTAMPTZ,
    pinned_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_comm_posts_feed ON community_posts(community_id, is_pinned DESC, created_at DESC);
