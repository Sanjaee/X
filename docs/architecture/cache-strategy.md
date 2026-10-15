# Architecture: Multi-Tier Caching & Memory Management Strategy

## 1. Overview
The caching architecture balances ultra-low latency (<5ms on hot queries) with strict memory bounds to prevent unbounded memory growth on both the backend (Redis RAM) and the frontend (browser DOM & JS Heap).

---

## 2. Server-Side Caching (Redis)

### Key Namespaces and TTL Policies

| Cache Namespace | Key Format | TTL | Purpose |
| :--- | :--- | :--- | :--- |
| **Search Suggestions** | `search:sugg:<prefix>` | 300s (5m) | High-throughput autocomplete suggestions |
| **Search Top Tab** | `search:posts:top:<sha256(q+filters)>` | 60s | Caches first page of popular search queries |
| **Search Latest Tab** | `search:posts:latest:<sha256(q+filters)>` | 15s | Very short TTL to preserve near real-time updates |
| **Recent Searches** | `search:recent:<userId>` | 3600s | Caches active user's recent search pills |
| **Search Rate Limiting** | `ratelimit:search:<ip/userId>` | 60s | Sliding window (max 60 queries/min) |

### Cache Key Normalization
All search query strings are trimmed, lowercased, and query operators are sorted alphabetically before generating the SHA-256 cache key. This guarantees identical keys for `golang from:jack` and `from:jack golang`.

### Invalidation Strategies
- **Post Deletion**: Automatically clears cached search keys related to the author.
- **User Block / Mute**: Search results for authenticated users bypass public cache or append `viewer_id` to the cache key to prevent privacy leaks.

---

## 3. Client-Side Bounded Cache (TanStack Query + TanStack Virtual)

### Bounded Page Windowing (`maxPages`)
To prevent infinite scroll from consuming hundreds of megabytes of browser RAM, the feed and search queries configure `maxPages`:
```typescript
const { data, fetchNextPage, hasNextPage } = useInfiniteQuery({
  queryKey: searchKeys.list(query, tab, filters),
  queryFn: ({ pageParam }) => fetchSearchPosts({ query, tab, cursor: pageParam }),
  initialPageParam: '',
  getNextPageParam: (lastPage) => lastPage.page_info?.next_cursor ?? null,
  maxPages: 5, // Automatically evicts pages outside the active window!
  staleTime: 60_000,
  gcTime: 300_000,
});
```

### Virtual DOM Cleanup
- Combined with `@tanstack/react-virtual`, DOM nodes remain under 2,000 at all times.
- Elements unmounted from the DOM release their listeners, internal states, and image resources.
- Browser garbage collection collects discarded page objects smoothly without frame drops.
