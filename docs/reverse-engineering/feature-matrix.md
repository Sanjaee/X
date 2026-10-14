# Feature Matrix & Behavioral Specification

**Status:** Phase 1 Complete  
**Investigation Order:** A through J (Home Timeline to State Handling)

---

## Feature A: Home Timeline

### 1. Observed Behavior
- Two main tabs: **For you** (algorithmic recommendation) and **Following** (chronological reverse stream).
- Sticky top header (`53px` height) with backdrop blur (`backdrop-blur-md`).
- Feed items rendered as cards with left avatar (`40px`), author name, handle, relative timestamp (`5m`, `18h`), text, media grid, and engagement metrics (replies, reposts, likes, views, bookmarks, share).
- Infinite scroll triggering cursor-based pagination before user hits the viewport bottom.

### 2. Reproduction Steps
1. Navigate to `/` (or `/home`).
2. Verify sticky header and active tab indicator (pill or underline).
3. Scroll down; observe seamless addition of older posts.

### 3. Relevant Evidence
- Extracted CSS: `--layout-width-primary: 600px`, `--sticky-header-height: 53px`.
- SVG icons: `icon-heart-stroke`, `icon-retweet-stroke`, `icon-reply-stroke`, `icon-bookmark-stroke`.

### 4. Implementation Requirements for Our Reconstruction
- `GET /api/v1/feed?type=for_you|following&cursor={cursor}&limit=20`
- Cursor pagination based on `(created_at, id)`.
- Client-side virtualized list or TanStack Query `useInfiniteQuery`.

---

## Feature B: Post Composer and Media Attachments

### 1. Observed Behavior
- Located at top of timeline (inline) and available via modal dialog (`/compose/post`).
- Auto-expanding textarea up to max height before scrolling.
- Circular progress bar tracking character count (default 280 characters).
- Toolbar with action buttons: Image upload (up to 4 photos or 1 video), GIF selector, Poll creator, Emoji picker, Schedule post.
- Drag-and-drop file upload support.
- Primary "Post" button disabled until valid text or media is provided.

### 2. Relevant Evidence
- `manifest.json`: `share_target.files`: `name: externalMedia`, accept `image/jpeg`, `image/png`, `image/gif`, `video/mp4`.
- Extracted CSS: `.x-button`, disabled styling with `opacity: 0.5`.

### 3. Implementation Requirements for Our Reconstruction
- `POST /api/v1/posts` with `{ content, media_ids: [], reply_to_id: null }`.
- S3 upload pipeline: `POST /api/v1/media/upload` returning `{ media_id, url }`.
- Client-side image preview and removal before publishing.

---

## Feature C: Post Detail and Reply Threads

### 1. Observed Behavior
- URL route: `/{username}/status/{postId}`.
- Focal post is emphasized with larger typography (`text-[17px]` to `text-[20px]`), full timestamp (`2:07 PM · Oct 14, 2026`), and full view count.
- Vertical connecting line (`2px solid`) linking parent post, reply composer, and child replies.
- Inline reply composer positioned directly beneath the focal post.
- Hierarchical replies rendered in descending chronological order.

### 2. Relevant Evidence
- Investigated live on `https://x.com/NASA/status/2110371804959191288`: Verified nested replies by `@canraptor_` and `@tobyllionair`.
- Formatters: `window.__xClientTextFormatters.run(..., "formatFullTimestamp", ...)`.

### 3. Implementation Requirements for Our Reconstruction
- `GET /api/v1/posts/{postId}` returning post details and parent ancestor chain.
- `GET /api/v1/posts/{postId}/replies?cursor={cursor}` returning threaded replies.
- UI rendering of tree branch lines connecting avatars.

---

## Feature D: User Profiles and Follow Relationships

### 1. Observed Behavior
- URL route: `/{username}`.
- Header: Banner image (aspect 3:1), overlapping circular avatar (`136px` desktop, `75px` mobile) with white/black border ring.
- User metadata: Display Name, Verification Badge, `@username`, Bio, Location, Website link, Joined Date.
- Social counters: Following count (`/following`) and Followers count (`/verified_followers` or `/followers`).
- Profile navigation tabs: **Posts**, **Replies**, **Highlights**, **Articles**, **Media**, **Reposts**.
- "Follow" / "Following" button with hover state changing to "Unfollow" (destructive red outline).

### 2. Relevant Evidence
- Investigated live on `https://x.com/NASA`: Verified banner layout, follower counts (`92.3M Followers`), and tab routes (`/with_replies`, `/media`, `/reposts`).

### 3. Implementation Requirements for Our Reconstruction
- `GET /api/v1/users/{username}`
- `POST /api/v1/users/{username}/follow` & `DELETE /api/v1/users/{username}/follow`
- `GET /api/v1/users/{username}/posts?tab=posts|replies|media`
- Optimistic update of follow state and follower count.

---

## Feature E: Search and Explore

### 1. Observed Behavior
- Search input in top right sidebar and dedicated `/explore` view.
- Search input displays clear button (`x`) when text is present.
- Search result tabs: **Top**, **Latest**, **People**, **Media**, **Lists**.
- Trending topics list: Category (e.g., "Trending in Technology"), Topic Title, Post Count (e.g., "45.2K posts"), three-dot action menu.

### 2. Relevant Evidence
- Investigated on `https://x.com/explore`: Skeletons, trend cards, search action in schema.org LD-JSON.

### 3. Implementation Requirements for Our Reconstruction
- `GET /api/v1/search?q={query}&type=top|latest|users|hashtags`
- PostgreSQL `to_tsvector` and `ts_query` full-text search indexes on post text and usernames.
- Trending hashtags aggregated via periodic background query or Redis counters.

---

## Feature F: Notifications

### 1. Observed Behavior
- Navigation icon displays red unread badge with counter.
- Tabs: **All**, **Verified**, **Mentions**.
- Notification items indicate actor avatar, icon type (Heart for like, Repeat for repost, UserPlus for follow), actor name, and referenced post snippet.
- Clicking marks notification as read.

### 2. Implementation Requirements for Our Reconstruction
- `GET /api/v1/notifications?cursor={cursor}`
- `PATCH /api/v1/notifications/read`
- Real-time notification dispatch via WebSocket connection.

---

## Feature G: Direct Messaging (DMs)

### 1. Observed Behavior
- Two-pane layout on desktop: Conversation list on left (`380px`), Active chat thread on right.
- Real-time message streaming with status indicator (sent, delivered, read).
- Support for text messages and image attachments.

### 2. Implementation Requirements for Our Reconstruction
- `GET /api/v1/conversations`
- `GET /api/v1/conversations/{id}/messages`
- `POST /api/v1/conversations/{id}/messages`
- WebSocket message broadcast with room subscription.

---

## Feature H: Settings and Account Preferences

### 1. Observed Behavior
- Settings panel covering: Account Info, Password Change, Display / Dark Mode, Privacy & Safety, Muted/Blocked Accounts.
- Real-time theme toggle (Light, Dim, Lights Out/Black).
- Accent color selection (Blue, Yellow, Pink, Purple, Orange, Green).

### 2. Implementation Requirements for Our Reconstruction
- `PATCH /api/v1/me/profile`
- `PATCH /api/v1/me/settings`
- Client-side persistence of theme in `localStorage` and `cookie`.

---

## Feature I: Responsive Mobile and Desktop Layouts

### 1. Observed Behavior
- **Desktop (>= 1280px)**: 3-column layout.
  - Left rail: `275px` (full labeled nav)
  - Main timeline: `600px`
  - Right sidebar: `350px` - `420px`
- **Tablet (1000px - 1279px)**:
  - Left rail collapses to icon-only (`88px`)
  - Main timeline: `600px`
  - Right sidebar: `350px`
- **Small Tablet / Landscape Mobile (500px - 999px)**:
  - Right sidebar hidden.
  - Left rail icon-only (`88px`).
  - Timeline takes remaining width (up to `600px`).
- **Mobile (< 500px)**:
  - Left sidebar hidden.
  - Bottom navigation bar (`53px` height, icons: Home, Search, Grok/Explore, Notifications, Messages).
  - Floating Action Button (FAB) for new post in bottom right corner.
  - Top mobile header with user avatar (opens drawer), logo center, and settings/sparkle icon right.

### 2. Implementation Requirements for Our Reconstruction
- Mobile bottom navigation bar with safe-area padding.
- Responsive breakpoints matching Tailwind `sm:`, `md:`, `lg:`, `xl:`.

---

## Feature J: Loading, Empty, Error, and Authentication States

### 1. Observed Behavior
- Loading states: Shimmer skeleton cards imitating avatar, header bar, and text lines.
- SVG animated spinner matching `#1d9bf0`.
- Empty states: Friendly illustration, clear headline ("No posts yet"), and call-to-action button.
- Error states: "Something went wrong. Try reloading", retry button with exponential backoff.
- Unauthenticated overlay: Floating bottom banner ("Don't miss what's happening - People on X are the first to know") with Sign up / Log in buttons.

### 2. Implementation Requirements for Our Reconstruction
- Reusable UI skeleton components using Tailwind `animate-pulse`.
- Global error boundary and React Query retry logic.
- Guest bottom modal / banner.

---

## Feature K: Media Upload, Image Compression & Cloudflare R2 Pipeline

### 1. Observed Behavior
- Supported formats: JPEG, PNG, WebP, animated GIF, MP4/MOV videos.
- X image compression pipeline:
  - Strips EXIF metadata to ensure user privacy and reduce byte footprint.
  - Progressive image re-encoding with ~85% JPEG quality factor balancing sharpness and bandwidth.
  - Bilinear downsampling for dimensions exceeding 2048px bounding box while preserving aspect ratio.
  - Delivery through global edge CDN with immutable asset caching.

### 2. Evidence and Implementation
- Go Media Processing Service (`internal/media/media.go`):
  - Pure Go bilinear downsampling algorithm and JPEG quality optimization.
  - Direct integration with Cloudflare R2 object storage (`lostmedia` bucket) via S3-compatible API.
  - Public CDN delivery via `pub-a638ff524f744502b575203fb48d39ae.r2.dev`.
  - Multi-tier fallback to local filesystem storage if R2 is offline.

---

## Feature L: Subscriptions, Entitlement Engine, Verified Badges & Plisio Payments (Section 6A)

### 1. Observed Behavior
- Tiered catalog: **Basic** ($3/mo), **Premium** ($8/mo), and **Premium+** ($16/mo) with Annual billing (12-17% discount).
- Verified Checkmarks:
  - Blue: Verified individual / paid Premium subscription.
  - Gold: Verified organization / business.
  - Grey: Government agency or multilateral organization (e.g., NASA).
  - Staff: Platform employee.
  - Verification dialog popover explaining verification details and subscription backing.
- Entitlement gates enforced strictly on the backend:
  - Post length limit: Free (280 chars), Basic (1,000 chars), Premium (4,000 chars), Premium+ (10,000 chars).
  - Edit post window: 1 hour for paid tiers.
  - Video upload duration: Free (2 mins), Basic (10 mins), Premium (60 mins), Premium+ (3 hours).
  - Ad-reduction: 50% fewer ads for Premium, 100% ad-free for Premium+.
- Cryptocurrency payment checkout via Plisio gateway (`PLISIO_API_KEY`).

### 2. Implementation Architecture
- Entitlement Engine: Server-side authorization in `CreatePost` and `UploadMedia` evaluating `GetUserEntitlements`.
- Plisio Payment Integration (`internal/payment/plisio.go`):
  - Invoice generation: `POST /api/v1/subscriptions/checkout`
  - Authoritative IPN webhook: `POST /api/v1/webhooks/payments/plisio`
  - Automated state machine: `pending` -> `active` -> `cancel_at_period_end` -> `expired`.
  - Idempotent processing avoiding duplicate activations.
  - Test simulation: `POST /api/v1/subscriptions/simulate-payment` for automated zero-cost testing.
- UI Suite: Dedicated `/premium` view with comparison table, billing toggle, crypto checkout, and verified badge popover.

