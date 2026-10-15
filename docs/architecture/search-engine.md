# Architecture: High-Performance Search Engine

## 1. System Architecture & End-to-End Flow
The search engine provides sub-50ms search query execution across millions of posts, accounts, and media items.

```
[Browser / Client]
       |
       |  GET /api/v1/search/posts?q=golang+from:jack&tab=top&cursor=...
       v
   [NGINX Reverse Proxy (x-proxy)]
       |
       v
   [Go API Service (x-api)]
       |-- 1. Authentication & Viewer Context Extraction
       |-- 2. Query AST Parser (internal/search/parser.go)
       |       * Extracts text terms: "golang"
       |       * Extracts filters: Author="jack"
       |-- 3. Redis Cache Check (Key: search:v1:posts:<tab>:<hash>)
       |       * Hit: Return cached payload immediately (<5ms)
       |       * Miss: Proceed to PostgreSQL
       v
   [PostgreSQL Engine (x-postgres)]
       |-- 4. Full-Text Execution (to_tsvector & plainto_tsquery)
       |-- 5. GIN Index Scan & Trigram Similarity
       |-- 6. Privacy & Block Filters (excludes blocked & private accounts)
       |-- 7. Multi-signal Ranking Score computation
       v
   [Response Serializer]
       |-- 8. Enrich with Author profile, Media attachments, Like/Repost state
       |-- 9. Generate opaque Keyset cursor
       +-- 10. Write-through to Redis (TTL: 60s)
```

---

## 2. Query AST Parser & Supported Operators
The query parser safely extracts operators from the raw search string without raw SQL injection vulnerabilities:

| Operator | Example | Extracted Constraint |
| :--- | :--- | :--- |
| `from:username` | `from:elonmusk` | Filter by `author.username = 'elonmusk'` |
| `to:username` | `to:jack` | Filter by `reply_to.author.username = 'jack'` |
| `@username` | `@openai` | Mentions username in content or tag |
| `has:media` or `filter:media` | `launch has:media` | Enforces `has_media = TRUE` |
| `filter:images` | `space filter:images` | Filter by image attachments |
| `filter:videos` | `game filter:videos` | Filter by video attachments |
| `min_faves:N` | `min_faves:100` | Enforces `p.likes_count >= 100` |
| `since:YYYY-MM-DD` | `since:2026-01-01` | Enforces `p.created_at >= '2026-01-01'` |
| `until:YYYY-MM-DD` | `until:2026-06-01` | Enforces `p.created_at <= '2026-06-01'` |

All remaining tokens are combined into a sanitized full-text `tsquery` using `plainto_tsquery('simple', sanitizedText)`.

---

## 3. Search Tabs Architecture

### 1. Top Tab (`/api/v1/search/posts?tab=top`)
- Blends text relevance (`ts_rank_cd`), engagement log scale (`LOG(likes + 1)`), freshness exponential decay, and author authority.
- Paginated using composite cursor `(score, id)`.

### 2. Latest Tab (`/api/v1/search/posts?tab=latest`)
- Pure reverse-chronological search results.
- Strictly bounded by `created_at DESC, id DESC`.
- High efficiency using index `(created_at DESC, id DESC)`.

### 3. People Tab (`/api/v1/search/users?q=...`)
- Searches usernames, display names, and bios using PostgreSQL `pg_trgm` similarity.
- Exact match bonus (`username = q` receives highest boost).
- Follower count logarithmically weighted.

### 4. Media Tab (`/api/v1/search/posts?tab=media`)
- Returns only posts with valid image or video attachments (`has_media = TRUE`).
- Lazy-loads media thumbnails to optimize viewport consumption.

---

## 4. Search Autocomplete & Suggestions
- Endpoint: `GET /api/v1/search/suggestions?q=...`
- Returns:
  - Top matching query completions.
  - Top 3 matching accounts with avatar, handle, verified badge, and follower counts.
- Debounced at 200ms on frontend with `AbortController` cancellation for pending requests.
- Backed by Redis cache (`search:sugg:<prefix>`) with 5-minute TTL.

---

## 5. Recent Search History
- Endpoints:
  - `GET /api/v1/search/recent`: Lists user's recent search queries (up to 10).
  - `DELETE /api/v1/search/recent/{id}`: Removes an individual search pill.
  - `DELETE /api/v1/search/recent`: Clears entire search history for the user.
- Stored in PostgreSQL `recent_searches` table with user ownership validation.
