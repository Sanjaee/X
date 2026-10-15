# Architecture: High-Performance Virtualized Infinite Scrolling

## 1. Overview & Core Philosophy
The infinite scrolling subsystem is engineered to support browsing tens of thousands of posts without degrading browser performance, exhausting mobile RAM, or causing Cumulative Layout Shift (CLS).

### Key Pillars
1. **Bounded DOM Footprint**: Keep active DOM nodes under 2,500 elements regardless of how deep the user scrolls.
2. **Deterministic Keyset Cursors**: Zero SQL `OFFSET`. Every page is fetched using `(created_at, id)` keyset tuples.
3. **Bounded In-Memory Query Cache**: Evict distant pages from the TanStack Query store (`maxPages: 5`) while caching initial pages for instant tab switching.
4. **Zero Layout Shift Media Anchors**: Media attachments declare their intrinsic aspect ratio before binary assets download.
5. **Exact Scroll Restoration**: Store viewport anchor post ID in `sessionStorage` to restore exact positioning upon browser Back navigation.

---

## 2. Virtualization Architecture with TanStack Virtual

### Virtual Window Model
```
[Viewport Container: overflow-y-auto]
  |
  +-- [Total Height Spacer: height = totalVirtualSize px]
        |
        +-- [Rendered Virtual Slice: overscan = 5 items]
              |-- Post Card #42 (top: 8400px)
              |-- Post Card #43 (top: 8620px)
              |-- Post Card #44 [IN VIEWPORT] (top: 8810px)
              |-- Post Card #45 [IN VIEWPORT] (top: 9050px)
              |-- Post Card #46 [IN VIEWPORT] (top: 9310px)
              |-- Post Card #47 (top: 9550px)
              |-- Post Card #48 (top: 9790px)
```

Items outside `[startIndex - overscan, endIndex + overscan]` are unmounted. Their heights are tracked dynamically in the virtualizer cache via `measureElement`.

### CLS Prevention Strategy
- Text content is rendered with standard font metric baselines.
- Single image posts render in a container with a calculated aspect ratio: `style={{ aspectRatio: `${width}/${height}` }}` or fallback `aspect-video` (`16/9`).
- Image elements use CSS `loading="lazy"` and placeholder background colors (`bg-muted`) to ensure the layout engine allocates physical space before image bytes arrive.

---

## 3. Cursor Pagination Contract

### Keyset Cursor Format
All cursor values are opaque URL-safe base64 strings containing:
```json
{
  "t": "2026-10-15T14:32:01.123456Z",
  "id": "c1f78e4a-912b-4e63-8ef2-7e4a1a3e811b"
}
```

### Response Schema
```json
{
  "data": [
    {
      "id": "post_uuid",
      "author_id": "user_uuid",
      "content": "Post text content...",
      "created_at": "2026-10-15T14:32:01Z",
      "author": { ... }
    }
  ],
  "page_info": {
    "next_cursor": "eyJ0IjoiMjAyNi0xMC0xNVQxNDozMjowMS4xMjM0NTZaIiwiaWQiOiJjMWY3OGU0YS05MTJiLTRlNjMtOGVmMi03ZTRhMWEzZTgxMWIifQ==",
    "has_next_page": true
  }
}
```

### Database Query Execution
```sql
SELECT ...
FROM posts p
WHERE (p.created_at, p.id) < ($1, $2)
  AND p.deleted_at IS NULL
ORDER BY p.created_at DESC, p.id DESC
LIMIT $3;
```
This guarantees logarithmic index traversal `O(log N)` across indexes `(created_at DESC, id DESC)`.

---

## 4. Scroll Restoration Mechanism

When navigating away from a feed (e.g. clicking into a post detail or user profile):
1. The feed captures:
   - Route path + query parameters (`window.location.pathname + window.location.search`).
   - The top-most fully visible post ID (`anchorPostId`).
   - Relative offset from the viewport top (`anchorOffset`).
2. Stored in `sessionStorage` under `scroll_anchor_${routeKey}`.
3. Upon returning via browser Back:
   - The query cache serves cached pages instantly (`staleTime: 60_000`).
   - `virtualizer.scrollToIndex(targetIndex, { align: 'start' })` restores the precise frame before the user left.

---

## 5. Lifecycle State Machine
```
               +-----------------+
               | initial-loading |
               +--------+--------+
                        |
                        v
                 +--------------+
                 |    ready     |<-------------------+
                 +------+-------+                    |
                        | (Sentinel triggers fetch)  |
                        v                            |
             +---------------------+                 |
             | fetching-next-page  |                 |
             +---+--------------+--+                 |
                 |              |                    |
        (Success)|              |(Network Failure)   |
                 v              v                    |
          +--------------+  +-------------------+    |
          | hasNextPage? |  | retryable-error   +----+
          +---+------+---+  +-------------------+ (Retry clicked)
              |      |
        (Yes) |      | (No)
              |      v
              |  +-------------+
              |  | end-of-list |
              |  +-------------+
              v
           [ready]
```
