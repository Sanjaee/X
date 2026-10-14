# Database Indexing & Query Acceleration Strategy

**Status:** Phase 3 Approved  
**Engine:** PostgreSQL 16+

---

## 1. Indexing Strategy Summary

| Target Table | Columns Indexed | Index Type | Accelerated Query Pattern |
| :--- | :--- | :--- | :--- |
| `posts` | `(author_id, created_at DESC)` | B-Tree | User profile timeline (`GET /users/{u}/posts`) |
| `posts` | `(created_at DESC, id DESC) WHERE deleted_at IS NULL` | Partial B-Tree | Global & Following feed keyset cursor pagination |
| `posts` | `(reply_to_id, created_at ASC) WHERE reply_to_id IS NOT NULL` | Partial B-Tree | Descendant reply tree traversal (`GET /posts/{id}/replies`) |
| `posts` | `USING GIN (to_tsvector('english', content))` | GIN | Full-text search queries (`GET /search?q=keyword`) |
| `follows` | `(follower_id, following_id)` | B-Tree (PK) | "Is user A following user B?" and timeline fan-out join |
| `follows` | `(following_id, follower_id)` | B-Tree | "Who are the followers of user B?" (`/users/{u}/followers`) |
| `post_likes` | `(user_id, created_at DESC)` | B-Tree | "Posts liked by user" tab on profile (`/users/{u}/likes`) |
| `post_bookmarks` | `(user_id, created_at DESC)` | B-Tree | User saved bookmarks tab (`GET /bookmarks`) |
| `notifications` | `(recipient_id, created_at DESC)` | B-Tree | Notification list (`GET /notifications`) |
| `notifications` | `(recipient_id) WHERE is_read = FALSE` | Partial B-Tree | Fast unread badge counter query (`SELECT COUNT(*) ...`) |
| `messages` | `(conversation_id, created_at DESC)` | B-Tree | Chat thread message pagination |
| `hashtags` | `(posts_count DESC)` | B-Tree | Trending hashtags list |

---

## 2. Complete SQL Migration Script (`V1__initial_schema.sql`)

The migration file below creates all tables, constraints, foreign keys, triggers for counter synchronization, and indexes.
