# Feed Performance, Indexing & Caching Architecture

## 1. High-Performance Timeline Strategy

Scaling social feeds to millions of concurrent reads demands avoiding full table scans, preventing memory spikes, and eliminating N+1 query patterns.

---

## 2. Cursor Pagination Mechanics

Using traditional SQL `OFFSET` pagination is disastrous on high-volume feeds because `OFFSET 10000` forces the database to read and discard 10,000 rows. Furthermore, when new posts arrive in real time, offset pagination causes posts to shift down, creating duplicate reads or skipped posts on the client.

### Opaque Composite Cursor
Our pagination uses an opaque base64-encoded cursor:
```
cursor = base64(timestamp_unix_nano + ":" + post_id)
```

The database query leverages keyset pagination:
```sql
WHERE (p.created_at, p.id) < ($1, $2)
ORDER BY p.created_at DESC, p.id DESC
LIMIT $3;
```
This guarantees $O(\log N)$ index traversal regardless of page depth.

---

## 3. Database Indexing Matrix

The following specialized indexes are deployed in PostgreSQL 16:

| Index Name | Target Table | Columns & Condition | Purpose |
|---|---|---|---|
| `idx_posts_feed_cursor` | `posts` | `(created_at DESC, id DESC) WHERE deleted_at IS NULL` | Accelerated keyset timeline pagination |
| `idx_posts_author_created` | `posts` | `(author_id, created_at DESC) WHERE deleted_at IS NULL` | Rapid in-network Following candidate extraction |
| `idx_post_likes_recent` | `post_likes` | `(user_id, created_at DESC)` | Fast social engagement extraction for viewer |
| `idx_follows_lookup` | `follows` | `(follower_id, following_id)` | Fast relational checks ($O(1)$) |
| `idx_post_hashtags_tag` | `post_hashtags`| `(hashtag_id, post_id)` | Instant topic candidate retrieval |
| `idx_posts_velocity` | `posts` | `(created_at DESC, likes_count DESC)` | Trending candidate candidate extraction |

---

## 4. Elimination of N+1 Queries (Batch Loading)

When loading 20 posts for a timeline page, querying author profiles, media items, and like states individually would generate $20 \times 4 = 80$ database roundtrips.

Our Go repository executes structured **Batch Loading**:
1. Single query extracts the 20 post IDs and author IDs.
2. `SELECT ... WHERE user_id = ANY($1)` loads all author profiles in 1 query.
3. `SELECT ... WHERE post_id = ANY($1)` loads all media attachments in 1 query.
4. If viewer is authenticated, a single batch query checks relational states (`is_liked`, `is_reposted`, `is_bookmarked`) using PostgreSQL arrays.
5. In-memory assembly joins the structs in $O(M)$ time.

---

## 5. Multi-Tier Caching Architecture

```mermaid
graph TD
    Client[Client Feed Request] --> L1[L1: Process In-Memory Cache\nTTL: 5s, For Hot Public Trends]
    L1 -- Miss --> L2[L2: Redis Feed Cache\nKey: feed:for_you:{userId}:{cursor}\nTTL: 60s]
    L2 -- Miss --> Engine[Ranking Engine Pipeline]
    Engine --> PG[(PostgreSQL Keyset Query)]
    Engine --> CacheFill[Write-Through to Redis]
```

### Cache Invalidation Strategy
- **User Actions**: When viewer likes, hides, or follows someone, the user's specific feed cache is tagged with a generation timestamp or evicted to reflect the new state immediately.
- **New Posts by Followees**: We utilize an **on-read ranking model with cached candidate pools**, avoiding expensive fan-out-on-write across millions of followers for inactive users.
