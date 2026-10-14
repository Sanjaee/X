# Performance Engineering & Optimization Strategy

**Status:** Phase 3 Approved  
**Targets:** API p95 < 30ms, Timeline render < 100ms, Zero N+1 Queries

---

## 1. Timeline Fan-Out & Query Optimization

Timeline generation is the most query-intensive operation in a microblogging platform.

### Two-Tier Timeline Architecture
1. **Pull Model (Fan-out on Read)**:
   For standard users with moderate followings (< 5,000 followings), timeline is assembled on-demand via indexed SQL:
   ```sql
   SELECT p.*, u.username, u.display_name, u.avatar_url,
          EXISTS(SELECT 1 FROM post_likes pl WHERE pl.post_id = p.id AND pl.user_id = $1) AS is_liked,
          EXISTS(SELECT 1 FROM post_reposts pr WHERE pr.post_id = p.id AND pr.user_id = $1) AS is_reposted,
          EXISTS(SELECT 1 FROM post_bookmarks pb WHERE pb.post_id = p.id AND pb.user_id = $1) AS is_bookmarked
   FROM posts p
   JOIN follows f ON f.following_id = p.author_id AND f.follower_id = $1
   JOIN users u ON u.id = p.author_id
   WHERE p.deleted_at IS NULL
     AND (p.created_at, p.id) < ($2, $3) -- Keyset pagination
   ORDER BY p.created_at DESC, p.id DESC
   LIMIT 20;
   ```
2. **Push Model (Fan-out on Write - Future Scaling)**:
   For accounts with millions of followers ("celebrity accounts"), posts are published to a Redis timeline list or message broker topic, rather than forcing millions of individual feed joins.

---

## 2. Keyset (Cursor-Based) Pagination

Avoid `OFFSET / LIMIT` pagination. As offset grows, PostgreSQL must scan and discard thousands of unneeded rows ($O(N)$ penalty).

### Keyset Pattern:
- **Cursor**: Base64 encoded compound tuple `{"t": 1728900000, "id": "uuid"}`.
- **SQL Operator**:
  ```sql
  WHERE (created_at, id) < (to_timestamp($cursor_time), $cursor_id)
  ```
- Constant time execution ($O(1)$) utilizing the composite B-Tree index `(created_at DESC, id DESC)`.

---

## 3. Database Connection Pooling & Concurrency

Configured in Go `database/sql` driver:
```go
db.SetMaxOpenConns(50)                  // Bound maximum active connections
db.SetMaxIdleConns(25)                  // Maintain warm idle connections
db.SetConnMaxLifetime(15 * time.Minute) // Periodically refresh connections
db.SetConnMaxIdleTime(5 * time.Minute)  // Reclaim idle connections
```

---

## 4. Frontend Asset Optimization

- Dynamic import of heavy modules (emoji picker, video player, image crop tool).
- Next.js Image component (`next/image`) for automatic WebP/AVIF transcoding and responsive `srcset` generation.
- TanStack Query cache deduplication: window focus re-fetching disabled for timeline to prevent jitter; stale time set to 30 seconds.
