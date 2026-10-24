# Implementation Roadmap & Parity Action Plan

**Plan Version:** 1.0  
**Target Completion:** P0 & P1 Parity  
**Dependencies:** Go API Service, Next.js 14 App Router, PostgreSQL Migrations  

---

## 1. Priority Execution Phases

### Batch 1: P1 Critical Completeness (Immediate Execution)
1. **Bookmark Read & Management Flow:**
   - Implement `GetBookmarks(ctx, userID, cursor, limit)` in repository layer (`postgres.go` and `memory.go`).
   - Implement `GET /api/v1/bookmarks` in HTTP handler (`handler.go`).
   - Register route in `main.go`.
   - Update `apps/web/src/app/bookmarks/page.tsx` with dynamic fetching, infinite scroll, unbookmarking callbacks, and authentic empty state.
   - Add backend tests in `internal/handler/bookmark_test.go` verifying bookmark retrieval, cursor pagination, and unbookmarking.

2. **User Profile "Likes" Tab:**
   - Add `"likes"` / `"liked"` query branch in `GetUserPosts` in `services/api/internal/repository/postgres.go` and `memory.go` querying `post_likes`.
   - Update `apps/web/src/app/[username]/page.tsx` with Likes tab, tab persistence, and scroll position restoration.
   - Add backend test verifying profile likes filtering.

### Batch 2: P2 Quality & Route Completeness
3. **Settings & Account Preferences Route (`/settings`):**
   - Create `apps/web/src/app/settings/page.tsx` with Display Theme toggles (Dark, Dim, Light) and Account Information update forms.
4. **Grok AI / Assistant Placeholder Page (`/grok`):**
   - Create `apps/web/src/app/grok/page.tsx` with conversation mockup and authentic X styling to prevent 404 navigation error.

### Batch 3: Verification, Lint & Graphify
5. Run full Go test suite (`go test ./... -v`).
6. Run Next.js TypeScript check (`npx tsc --noEmit`).
7. Run Graphify auto-update commands as per user rules:
   - `graphify . --code-only`
   - `graphify cluster-only .`
