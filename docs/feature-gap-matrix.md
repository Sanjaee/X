# X/Twitter Feature Gap Matrix

**Audit Date:** 2026-10-24  
**Audit Scope:** Full-Stack (Next.js Web App, Go API Microservices, PostgreSQL Schema, Realtime WebSocket)  
**Standard Status Categories:** Complete | Partial | Broken | Missing | Unverified | Different by design  

---

## 1. Feature Gap Matrix Overview

| Feature / User Flow | Reference Behavior (X.com) | Current Implementation | Gap & Status | Priority | Dependencies | Files & Evidence |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| **Home Timeline Feeds** | For You & Following feeds with cursor pagination, stable scroll & new-post pill. | Implemented with TanStack Query, cursor encoding, and realtime new post indicator. | **Complete** (Verified end-to-end with unit & integration tests) | P0 | Auth, DB Migrations | `apps/web/src/app/page.tsx`, `services/api/internal/recommendation` |
| **Post Creation & Deletion** | Character limits based on tier, media attachments, instant optimistic updates. | Supported with character limit validation, quota enforcement, and optimistic rollback. | **Complete** (Verified end-to-end) | P0 | Media service, Auth | `apps/web/src/components/composer/PostComposer.tsx`, `internal/handler/handler.go` |
| **Post Threading & Replies** | Threaded replies with vertical connecting line between parent & replies. | Hierarchical comments and nested thread display supported on status page. | **Complete** (Verified end-to-end) | P1 | Post Composer, DB | `apps/web/src/app/[username]/status/[postId]/page.tsx` |
| **Likes & Reposts** | Idempotent toggle with atomic counts and reposter attribution header. | Atomic SQL operations with Redis cache invalidation and reposter attribution. | **Complete** (Verified with concurrency tests) | P0 | Posts DB | `services/api/internal/repository/postgres.go` |
| **Bookmark Posts** | Save posts privately; view curated list at `/bookmarks` with pagination. | Backend supports `POST/DELETE /posts/{id}/bookmarks` but **`/bookmarks` UI only has empty state hardcoded** (`useState([])`), no `GET /bookmarks` API endpoint! | **Partial / Missing Read** | **P1** | Bookmarks DB Table | `apps/web/src/app/bookmarks/page.tsx`, `services/api/internal/handler/handler.go` |
| **User Profile Tabs** | Posts, Replies, Highlights, Media, Likes tabs with scroll preservation. | Supports Posts, Replies, Reposts, and Media tabs. Missing **Likes tab** on profile. | **Partial** | **P1** | PostLikes DB Table | `apps/web/src/app/[username]/page.tsx`, `services/api/internal/repository/postgres.go` |
| **Follow / Unfollow / Graph** | Follow/unfollow, mutuals, blocking, muting, private account follow requests. | Fully implemented with mutual connections, block/mute exclusion, and private requests. | **Complete** (Verified end-to-end) | P0 | Social Graph DB | `services/api/internal/handler/handler.go` |
| **Direct Messaging (DMs)** | Realtime two-pane DM interface with text, media, read receipts, unread counter. | Fully functional with WebSocket streaming, conversation lists, and optimistic append. | **Complete** (Verified end-to-end) | P1 | WebSocket hub | `apps/web/src/app/messages/page.tsx`, `internal/realtime` |
| **Notifications** | All, Verified, Mentions tabs, badge counter, realtime push, follow request banner. | Implemented with realtime push and follow request actions. Filter tabs exist. | **Complete** (Verified end-to-end) | P1 | Notifications DB | `apps/web/src/app/notifications/page.tsx` |
| **Explore & Trending** | Categorized trending topics, query search bar, tabbed search results. | Search supports query operators (`from:`, `to:`, `filter:`), trends aggregated from DB. | **Complete** (Verified end-to-end) | P1 | Full-text indexes | `apps/web/src/app/explore/page.tsx`, `internal/search` |
| **Media Viewer Overlay** | Fullscreen media lightbox with URL route sync (`/status/{id}/photo/1`), browser back nav. | Implemented with `MediaViewerContext` and standalone photo route. | **Complete** (Verified end-to-end) | P2 | Media Router | `apps/web/src/components/media/MediaViewerModal.tsx` |
| **Settings & Account Preferences** | Display preferences (Dark/Dim/Light), account management, password change. | **Route `/settings` is completely missing** in web app. | **Missing** | **P2** | User Profile API | `apps/web/src/app/settings/page.tsx` |
| **Verification & Subscriptions** | Premium tiers (Basic, Premium, Premium+), crypto checkout via Plisio, checkmark badges. | Fully implemented with checkout, IPN webhooks, entitlement validation, and test simulation. | **Complete** (Verified end-to-end) | P1 | Payment gateway | `apps/web/src/app/premium/page.tsx`, `internal/payment` |
| **Grok AI / Exploration Route** | Quick access AI / Grok entrypoint in sidebar. | Sidebar links to `/grok`, but `/grok` route is not yet created. | **Missing** | **P2** | LLM API / UI | `apps/web/src/app/grok/page.tsx` |
| **Responsive Layout & Mobile Nav** | 3-column desktop, collapsed tablet, bottom nav on mobile with safe areas. | Mobile bottom nav, floating composer action button, drawer navigation implemented. | **Complete** (Verified end-to-end) | P1 | Layout Components | `apps/web/src/components/layout` |

---

## 2. Identified High-Priority Gaps & Remediation Plan

### Gap 1 (P1): Bookmark Management Flow is Incomplete
- **Observed Deficiency:**
  - Users can click the Bookmark button on a post (`PostCard.tsx`), which calls `POST /api/v1/posts/{postId}/bookmarks`.
  - However, navigating to `/bookmarks` displays a static empty state because `apps/web/src/app/bookmarks/page.tsx` defines `const [posts] = useState<Post[]>([]);` without calling an API.
  - Furthermore, the backend is missing a `GET /api/v1/bookmarks` endpoint to return the user's bookmarked posts with cursor pagination.
- **Remediation Plan:**
  1. Add `GetBookmarks` repository method in `services/api/internal/repository/postgres.go` and `memory.go`.
  2. Add `GET /api/v1/bookmarks` endpoint in `services/api/internal/handler/handler.go` and register route in `services/api/cmd/server/main.go`.
  3. Connect `apps/web/src/app/bookmarks/page.tsx` to fetch bookmarked posts using TanStack Query / `apiRequest`.

### Gap 2 (P1): User Profile Missing "Likes" Tab
- **Observed Deficiency:**
  - X profiles allow users to view liked posts on `/{username}` under the Likes tab.
  - The repository currently implements `posts`, `replies`, `reposts`, and `media` tabs, but lacks `likes` tab support in `GetUserPosts`.
- **Remediation Plan:**
  1. Add `case "likes", "liked":` in `GetUserPosts` in `services/api/internal/repository/postgres.go` querying `post_likes`.
  2. Add `Likes` tab button to `apps/web/src/app/[username]/page.tsx`.

### Gap 3 (P2): Missing `/settings` Preferences Route
- **Observed Deficiency:**
  - When users click Settings icons or links, no `/settings` route exists, causing Next.js 404.
- **Remediation Plan:**
  1. Create `apps/web/src/app/settings/page.tsx` with Display Theme toggles (Dark, Dim, Light) and Account Information update forms.

### Gap 4 (P2): Missing `/grok` Page
- **Observed Deficiency:**
  - Left navigation rail includes a "Grok" button pointing to `/grok`, which currently returns a 404.
- **Remediation Plan:**
  1. Create `apps/web/src/app/grok/page.tsx` matching X's Grok AI interface style.
