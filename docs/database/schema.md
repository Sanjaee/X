# Database Schema Specification

**Status:** Phase 3 Approved  
**Engine:** PostgreSQL 16+  
**Dialect:** Standard Relational SQL with Foreign Keys, Constraints & Indexes

---

## 1. Table Definitions

### 1.1 `users`
Core identity table storing credentials and account status.

| Column | Type | Constraints | Description |
| :--- | :--- | :--- | :--- |
| `id` | `UUID` | `PRIMARY KEY DEFAULT gen_random_uuid()` | Unique user identifier |
| `username` | `VARCHAR(30)` | `NOT NULL UNIQUE` | Unique handle (alphanumeric, lowercase, underscores) |
| `email` | `VARCHAR(255)` | `NOT NULL UNIQUE` | User email address |
| `password_hash` | `VARCHAR(255)` | `NOT NULL` | bcrypt or Argon2id hashed password |
| `role` | `VARCHAR(20)` | `NOT NULL DEFAULT 'user'` | 'user', 'moderator', 'admin' |
| `is_verified` | `BOOLEAN` | `NOT NULL DEFAULT FALSE` | Blue verification badge status |
| `verified_type` | `VARCHAR(20)` | `DEFAULT 'none'` | 'blue', 'business', 'government' |
| `is_private` | `BOOLEAN` | `NOT NULL DEFAULT FALSE` | Protected account toggle |
| `created_at` | `TIMESTAMPTZ`| `NOT NULL DEFAULT NOW()` | Account creation timestamp |
| `updated_at` | `TIMESTAMPTZ`| `NOT NULL DEFAULT NOW()` | Last update timestamp |
| `deleted_at` | `TIMESTAMPTZ`| `NULL` | Soft delete timestamp (deactivation) |

---

### 1.2 `user_profiles`
Public profile attributes.

| Column | Type | Constraints | Description |
| :--- | :--- | :--- | :--- |
| `user_id` | `UUID` | `PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE` | 1-to-1 link to user |
| `display_name` | `VARCHAR(50)` | `NOT NULL` | Visible user name |
| `bio` | `VARCHAR(160)` | `NOT NULL DEFAULT ''` | Short biography |
| `avatar_url` | `TEXT` | `NOT NULL DEFAULT ''` | Avatar image URL |
| `banner_url` | `TEXT` | `NOT NULL DEFAULT ''` | Banner image URL |
| `location` | `VARCHAR(100)` | `NOT NULL DEFAULT ''` | Location string |
| `website` | `VARCHAR(255)` | `NOT NULL DEFAULT ''` | External link |
| `birth_date` | `DATE` | `NULL` | Birth date |
| `followers_count` | `INTEGER` | `NOT NULL DEFAULT 0` | Denormalized follower counter |
| `following_count` | `INTEGER` | `NOT NULL DEFAULT 0` | Denormalized following counter |
| `posts_count` | `INTEGER` | `NOT NULL DEFAULT 0` | Denormalized posts counter |
| `updated_at` | `TIMESTAMPTZ`| `NOT NULL DEFAULT NOW()` | Last updated timestamp |

---

### 1.3 `sessions`
Active user session tokens.

| Column | Type | Constraints | Description |
| :--- | :--- | :--- | :--- |
| `id` | `UUID` | `PRIMARY KEY DEFAULT gen_random_uuid()` | Session ID |
| `user_id` | `UUID` | `NOT NULL REFERENCES users(id) ON DELETE CASCADE` | Owner user |
| `token_hash` | `VARCHAR(64)` | `NOT NULL UNIQUE` | SHA-256 hash of the session token |
| `ip_address` | `INET` | `NULL` | Client IP address |
| `user_agent` | `TEXT` | `NOT NULL DEFAULT ''` | Client device user agent |
| `expires_at` | `TIMESTAMPTZ`| `NOT NULL` | Session expiry timestamp |
| `created_at` | `TIMESTAMPTZ`| `NOT NULL DEFAULT NOW()` | Creation timestamp |

---

### 1.4 `posts`
Core publishing table representing top-level posts, replies, and quote posts.

| Column | Type | Constraints | Description |
| :--- | :--- | :--- | :--- |
| `id` | `UUID` | `PRIMARY KEY DEFAULT gen_random_uuid()` | Post identifier |
| `author_id` | `UUID` | `NOT NULL REFERENCES users(id) ON DELETE CASCADE` | Author user |
| `content` | `VARCHAR(280)` | `NOT NULL DEFAULT ''` | Post text body |
| `reply_to_id` | `UUID` | `NULL REFERENCES posts(id) ON DELETE SET NULL` | Parent post if this is a reply |
| `quote_post_id`| `UUID` | `NULL REFERENCES posts(id) ON DELETE SET NULL` | Quoted post reference |
| `replies_count`| `INTEGER` | `NOT NULL DEFAULT 0` | Denormalized replies counter |
| `reposts_count`| `INTEGER` | `NOT NULL DEFAULT 0` | Denormalized reposts counter |
| `likes_count` | `INTEGER` | `NOT NULL DEFAULT 0` | Denormalized likes counter |
| `views_count` | `BIGINT` | `NOT NULL DEFAULT 0` | Denormalized view counter |
| `created_at` | `TIMESTAMPTZ`| `NOT NULL DEFAULT NOW()` | Published timestamp |
| `updated_at` | `TIMESTAMPTZ`| `NOT NULL DEFAULT NOW()` | Edit timestamp |
| `deleted_at` | `TIMESTAMPTZ`| `NULL` | Soft delete timestamp |

---

### 1.5 `media_objects` & `post_media`
Media storage and attachments.

```sql
CREATE TABLE media_objects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    uploader_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    storage_key TEXT NOT NULL UNIQUE,
    url TEXT NOT NULL,
    media_type VARCHAR(20) NOT NULL, -- 'image', 'video', 'gif'
    mime_type VARCHAR(100) NOT NULL,
    file_size_bytes BIGINT NOT NULL,
    width INTEGER NULL,
    height INTEGER NULL,
    duration_seconds NUMERIC(6,2) NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE post_media (
    post_id UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    media_id UUID NOT NULL REFERENCES media_objects(id) ON DELETE CASCADE,
    position SMALLINT NOT NULL DEFAULT 0, -- order of display in 1-4 grid
    PRIMARY KEY (post_id, media_id)
);
```

---

### 1.6 Engagement Tables (`post_likes`, `post_reposts`, `post_bookmarks`)

```sql
CREATE TABLE post_likes (
    post_id UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (post_id, user_id)
);

CREATE TABLE post_reposts (
    post_id UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (post_id, user_id)
);

CREATE TABLE post_bookmarks (
    post_id UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (post_id, user_id)
);
```

---

### 1.7 Social Graph (`follows`, `blocks`, `mutes`)

```sql
CREATE TABLE follows (
    follower_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    following_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (follower_id, following_id),
    CHECK (follower_id != following_id)
);

CREATE TABLE blocks (
    blocker_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    blocked_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (blocker_id, blocked_id),
    CHECK (blocker_id != blocked_id)
);

CREATE TABLE mutes (
    muter_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    muted_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (muter_id, muted_id),
    CHECK (muter_id != muted_id)
);
```

---

### 1.8 Hashtags & Categorization

```sql
CREATE TABLE hashtags (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tag VARCHAR(100) NOT NULL UNIQUE, -- stored in lowercase without '#'
    posts_count INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE post_hashtags (
    post_id UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    hashtag_id UUID NOT NULL REFERENCES hashtags(id) ON DELETE CASCADE,
    PRIMARY KEY (post_id, hashtag_id)
);
```

---

### 1.9 Notifications

```sql
CREATE TABLE notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    recipient_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    actor_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type VARCHAR(30) NOT NULL, -- 'like', 'reply', 'repost', 'mention', 'follow'
    post_id UUID NULL REFERENCES posts(id) ON DELETE CASCADE,
    is_read BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

---

### 1.10 Direct Messaging (`conversations`, `conversation_members`, `messages`)

```sql
CREATE TABLE conversations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    is_group BOOLEAN NOT NULL DEFAULT FALSE,
    name VARCHAR(100) NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE conversation_members (
    conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_read_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (conversation_id, user_id)
);

CREATE TABLE messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    sender_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content TEXT NOT NULL DEFAULT '',
    media_id UUID NULL REFERENCES media_objects(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

---

### 1.11 Moderation & Administration (`reports`, `audit_events`, `user_settings`)

```sql
CREATE TABLE reports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    reporter_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    reported_user_id UUID NULL REFERENCES users(id) ON DELETE CASCADE,
    reported_post_id UUID NULL REFERENCES posts(id) ON DELETE CASCADE,
    reason VARCHAR(50) NOT NULL,
    details TEXT NOT NULL DEFAULT '',
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- 'pending', 'resolved', 'dismissed'
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE audit_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    actor_id UUID NULL REFERENCES users(id) ON DELETE SET NULL,
    action VARCHAR(50) NOT NULL,
    resource_type VARCHAR(50) NOT NULL,
    resource_id UUID NULL,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    ip_address INET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE user_settings (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    theme VARCHAR(20) NOT NULL DEFAULT 'dark',
    accent_color VARCHAR(20) NOT NULL DEFAULT 'blue',
    allow_dms_from VARCHAR(20) NOT NULL DEFAULT 'everyone',
    email_notifications BOOLEAN NOT NULL DEFAULT TRUE,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```
