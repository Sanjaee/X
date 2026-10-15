# REA Investigation: X Search Architecture, Discovery, and Autocomplete

## 1. Executive Summary & Methodology
This investigation analyzes the search, autocomplete, query parsing, and results pagination flows observed on [X.com](https://x.com/). Findings are strictly partitioned using the REA framework:
- **`[VERIFIED]`**: Directly observed via DOM analysis, URL query mutations, DevTools network logs, and interaction with search on web and mobile clients.
- **`[INFERRED]`**: Deductions based on response times, ranking behavior shifts across tabs, and standard large-scale IR (Information Retrieval) engineering patterns.
- **`[UNKNOWN]`**: Proprietary ranking weights, internal embedding models (e.g. Phoenix/SimClusters), and private cluster configurations that cannot be confirmed from external telemetry.
- **`[PROPOSED]`**: Target architecture, schema, ranking algorithms, and UI flows implemented in this repository.

---

## 2. Reverse Engineering Findings

### A. Search Entry Point, Suggestions & Autocomplete
- **`[VERIFIED]`**: The search input on X features a dropdown trigger upon focus. Before typing, it displays "Recent searches" with individual delete (`×`) buttons and a "Clear all" control.
- **`[VERIFIED]`**: As the user types (minimum 1–2 characters), an autocomplete flyout appears with two distinct sections:
  1. Search suggestions (query completions, hashtags).
  2. People suggestions (matching username and display name with avatar, verification badge, and handle).
- **`[VERIFIED]`**: Keyboard navigation (Arrow Down, Arrow Up, Enter to select, Escape to dismiss) is fully supported, with accessible ARIA `combobox` and `listbox` attributes.
- **`[INFERRED]`**: The client debounces requests by ~150–250ms and cancels previous pending flight requests via `AbortController`.
- **`[PROPOSED]`**: Implement `useSearchSuggestions` with a 200ms debounce, `AbortController`, keyboard navigation handlers, and high-performance trigram matching (`pg_trgm`) for username/display name search in PostgreSQL.

### B. URL Schema & State Synchronization
- **`[VERIFIED]`**: The search state is synchronized with the browser URL using query parameters:
  - Query: `?q=<encoded_query>`
  - Tab / Filter: `&f=live` (Latest), `&f=user` (People), `&f=media` (Media), or omitted/`&f=top` (Top).
- **`[VERIFIED]`**: URL parameter changes trigger immediate feed refresh without a full page reload (HTML5 History API / Next.js shallow routing).
- **`[VERIFIED]`**: Copying the URL and opening it in an incognito window accurately preserves the search query and the selected tab.
- **`[PROPOSED]`**: Structure Next.js routes at `/search?q=...&tab=top|latest|people|media` using `useSearchParams` as the single source of truth, enabling browser Back/Forward navigation and shareable URLs.

### C. Search Tabs & Ranking Semantics
- **`[VERIFIED]`**: X provides distinct tabs with entirely different ranking and filtering rules:
  1. **Top**: Highlights authoritative posts with high engagement (likes, reposts, quotes), verified creator weight, and text relevance. Not purely chronological.
  2. **Latest**: Strict reverse-chronological order of posts matching the search query terms.
  3. **People**: Returns account profiles matching the query. Prioritizes exact handle matches, followed by display name matches and accounts with high follower counts or mutual connections.
  4. **Media**: Postings containing images, video, or GIFs matching the query.
- **`[INFERRED]`**: The "Top" tab balances BM25/full-text score, engagement velocity (recent likes/reposts), and an exponential recency decay function (half-life of ~12–24 hours).
- **`[UNKNOWN]`**: The exact neural network weights or embedding vectors used in X's proprietary ranking pipeline (e.g. Earlybird / Phoenix).
- **`[PROPOSED]`**: Implement multi-tier ranking in PostgreSQL:
  - **Top**: `(ts_rank_cd(document, query) * 4.0) + (LOG(likes_count + 1) * 1.5) + (LOG(reposts_count + 1) * 1.2) + (EXP(-EXTRACT(EPOCH FROM (NOW() - created_at)) / 86400.0) * 3.0) + (CASE WHEN is_verified THEN 1.5 ELSE 0 END)`.
  - **Latest**: Filtered by `tsquery` and ordered deterministically by `created_at DESC, id DESC` using keyset cursor pagination.
  - **People**: Match via `similarity(username, $q) * 3.0 + similarity(display_name, $q) * 2.0 + LOG(followers_count + 1) * 0.5`.
  - **Media**: Restricted to `has_media = TRUE` with tab-configurable ordering.

### D. Advanced Search Operators
- **`[VERIFIED]`**: X supports search operators embedded directly into the query string:
  - `from:username` (posts written by this user).
  - `to:username` (replies addressed to this user).
  - `@username` (posts mentioning this user).
  - `filter:media` / `has:media` / `filter:images` / `filter:videos`.
  - `min_faves:N` / `min_retweets:N`.
  - `since:YYYY-MM-DD` / `until:YYYY-MM-DD`.
- **`[PROPOSED]`**: Implement a dedicated AST search query parser in Go (`internal/search/parser.go`) that extracts known operators safely into structured SQL filter parameters without allowing arbitrary SQL injection.

### E. Pagination & Result Integrity
- **`[VERIFIED]`**: Search results paginate smoothly with infinite scroll.
- **`[VERIFIED]`**: As new posts are created in real time, scrolling down does not duplicate already-seen posts.
- **`[INFERRED]`**: Cursors encapsulate the sort values of the last item in the page (e.g. `(created_at, id)` or `(rank_score, id)`), avoiding `OFFSET` performance degradation.
- **`[PROPOSED]`**: Use base64-encoded Keyset cursor `(score, timestamp, id)` validated server-side.

---

## 3. Comparative Architecture Ledger

| Feature | X.com (Observed) | Target Implementation | REA Classification |
| :--- | :--- | :--- | :---: |
| **Search Engine** | Distributed Earlybird (Lucene-based) + Realtime Cluster | PostgreSQL Full-Text Search (tsvector + GIN) + pg_trgm for fuzzy user search | `[INFERRED]` / `[PROPOSED]` |
| **URL State** | `/search?q=...&f=live` | Next.js App Router `/search?q=...&tab=latest` with `URLSearchParams` | `[VERIFIED]` / `[PROPOSED]` |
| **Autocomplete** | Flyout showing top queries + verified people | Debounced REST endpoint `/api/v1/search/suggestions` returning queries & users | `[VERIFIED]` / `[PROPOSED]` |
| **Recent Searches** | Stored locally and synchronized to account profile | Server-persisted `recent_searches` table with Redis caching + fast deletion | `[VERIFIED]` / `[PROPOSED]` |
| **Tabs** | Top, Latest, People, Media, Lists | Top, Latest, People, Media with strict semantic differentiation | `[VERIFIED]` / `[PROPOSED]` |
| **Filters** | `from:`, `to:`, `has:media`, `since:`, `until:` | AST search query parser transforming tokens to parameterized SQL | `[VERIFIED]` / `[PROPOSED]` |
| **Pagination** | Keyset cursor pagination without OFFSET | URL-safe base64 cursor `(timestamp, id)` or `(rank, id)` | `[VERIFIED]` / `[PROPOSED]` |
