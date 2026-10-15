# Architecture: Search Indexing Strategy

## 1. Storage & Indexing Strategy
PostgreSQL serves as the primary source of truth for all post and user search queries.

```
 [Write Path]                                    [Read Path]
Post Creation / Update                           Search Request
       |                                               |
       v                                               v
[PostgreSQL Table: posts]                     [PostgreSQL GIN / Trigram Index]
       |                                               |
       +--> GENERATED tsvector column                  +--> ts_rank / similarity query
       |    (search_vector)                            |
       v                                               v
[GIN Index: idx_posts_search_vector]            [Fast Sub-50ms Candidate Set]
```

---

## 2. PostgreSQL Full-Text Indexes & Schema

### Generated `tsvector` Column
To prevent re-computing `to_tsvector` on every search query, a generated column is added to the `posts` table:
```sql
ALTER TABLE posts
ADD COLUMN IF NOT EXISTS search_vector tsvector
GENERATED ALWAYS AS (
    to_tsvector('simple', COALESCE(content, ''))
) STORED;

CREATE INDEX IF NOT EXISTS idx_posts_search_vector
ON posts USING GIN (search_vector);
```

### Trigram Index for Fast User & Handle Autocomplete
```sql
CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE INDEX IF NOT EXISTS idx_users_username_trgm
ON users USING gin (username gin_trgm_ops);

CREATE INDEX IF NOT EXISTS idx_user_profiles_display_name_trgm
ON user_profiles USING gin (display_name gin_trgm_ops);
```

### Composite Filtering Indexes
```sql
-- For Latest Tab queries
CREATE INDEX IF NOT EXISTS idx_posts_created_at_id_active
ON posts (created_at DESC, id DESC)
WHERE deleted_at IS NULL;

-- For Media Tab queries
CREATE INDEX IF NOT EXISTS idx_posts_media_created_at
ON posts (created_at DESC, id DESC)
WHERE deleted_at IS NULL AND has_media = TRUE;
```

---

## 3. Data Integrity & Content Lifecycle

### Soft Deletion Handling
When a post is soft-deleted (`deleted_at IS NOT NULL`), it is automatically excluded by the partial index `WHERE deleted_at IS NULL`. It will never appear in any search result tab.

### Privacy Enforcement
- Private account posts (`users.is_private = TRUE`) are excluded from public search results unless the search viewer is an approved follower of the author.
- Blocked user relationships (`blocks` table) strictly omit posts and profiles from appearing in the viewer's search results.

---

## 4. Scalability Path: OpenSearch Integration
When search volume scales beyond 10,000 queries per second or dataset exceeds 100M posts:
1. **Transactional Outbox**: Post mutations write an event to an `outbox_events` table.
2. **RabbitMQ Stream / Debezium CDC**: Publishes post indexing events to `search.index.v1` queue.
3. **Indexer Worker**: Consumes events idempotently and pushes bulk documents to OpenSearch / Elasticsearch with KNN dense embeddings for semantic search.
