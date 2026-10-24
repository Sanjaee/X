# X Reference Behavior & Inspection Audit

**Generated:** 2026-10-24  
**Reference Target:** [https://x.com/](https://x.com/)  
**Audit Purpose:** Document observable reference behavior, interaction patterns, layout dimensions, network requests, and inspection boundaries across the X web platform to guide high-fidelity reconstruction.

---

## 1. Scope & Inspection Methodology

This audit captures observable behavior of the production X web application (`https://x.com/`). Behavior was analyzed using browser DOM inspection, HTTP network traffic traces, responsive breakpoint simulation, and REA forensic reverse engineering analysis.

### Inspection Boundaries & Unknowns
- **Verified Public Behavior:** UI layouts, DOM tree structure, Tailwind/CSS variables, HTTP routes, client-side caching & state transitions, optimistic updates, responsive layouts, photo/video overlays.
- **Inferred Behavior:** Candidate retrieval heuristics, feed ranking weights, spam scoring, and notification fanout mechanisms.
- **Inaccessible / Unknown:** Private proprietary ranking ML models (Phoenix / Heavy Ranker neural networks), internal fraud/abuse detection signals, distributed database storage engine internals (Manhattan/Rocksteady).

---

## 2. Observable Core Behavioral Specifications

### A. Home Timeline & Feeds (`/` and `/?tab=following`)
1. **Tabs:**
   - **For You (`/`):** Algorithmic blended feed with diversity rules (author diversity, topic diversity, engagement decay).
   - **Following (`/?tab=following`):** Reverse-chronological feed containing posts and reposts exclusively from accounts the user follows.
2. **Sticky Header:**
   - Height: `53px` (`h-[53px]`), `backdrop-blur-md`, semi-transparent background (`rgba(0, 0, 0, 0.65)` in dark mode).
   - Active tab indicator: Brand-colored underline with smooth horizontal indicator positioning.
3. **Feed Virtualization & Pagination:**
   - Infinite scroll triggered before reaching viewport bottom (~800px threshold).
   - Bounded client cache to prevent unbounded memory growth during marathon scrolling sessions.
4. **New Posts Indicator ("X new posts"):**
   - Floating pill button centered horizontally below sticky header (`top-[60px]`).
   - Appears when remote posts arrive without disturbing current scroll position.
   - Clicking smoothly scrolls user to top and prepends new posts.

### B. Post Composition & Interaction
1. **Post Composer:**
   - Auto-expanding textarea with character progress ring.
   - Free tier: 280 characters; Premium tiers: 1,000 to 10,000 characters.
   - Action toolbar: Media attachments (up to 4 photos or 1 video), Polls, Emojis, Scheduled posts.
2. **Post Interactions:**
   - **Like:** Optimistic UI state toggle, instant heart animation, idempotent endpoint.
   - **Repost:** Dropdown menu with "Repost" and "Quote". Supports optimistic toggle and reposter attribution header (`@user reposted`).
   - **Bookmark:** Bookmarking does not broadcast to public timeline; stored privately per user.
   - **View Count:** Formatted with human-readable abbreviations (e.g., `12.5K`, `1.2M`).

### C. Post Detail & Threading (`/{username}/status/{postId}`)
1. **Focal Post:**
   - Prominently rendered with larger typography (`text-[17px]`), full timestamp (`h:mm A · MMM D, YYYY`), and detailed engagement metrics bar.
2. **Ancestor & Descendant Lineage:**
   - Ancestor parent posts connected with vertical connector line (`2px solid`) joining avatar centers.
   - Inline reply composer nested directly beneath the focal post.
   - Threaded replies ordered chronologically below.

### D. User Profiles & Social Graph (`/{username}`)
1. **Header & Banner:**
   - Banner aspect ratio: 3:1.
   - Circular avatar (`136px` desktop, `75px` mobile) overlapping banner with border matching background.
2. **Metadata & Relationship:**
   - Display name, verified checkmark (Blue, Gold, Grey), `@username`, bio, location, website, join date.
   - Following and Followers counts linking to follower graph views.
   - Follow button dynamic state: "Follow" -> "Following" (hover state: "Unfollow" with destructive red styling).
3. **Profile Tabs:**
   - **Posts:** Original authored posts.
   - **Replies:** Authored replies to others.
   - **Highlights:** Pinned/highlighted posts (Premium).
   - **Media:** Authored posts containing image/video attachments.
   - **Likes:** Posts liked by the user (or private depending on settings).

### E. Search & Explore (`/explore` and `/search`)
1. **Explore:**
   - Search input in sticky header.
   - Trending topics with category ("Trending in Technology"), topic name, and post count format.
2. **Search Results:**
   - Query parser supporting operators: `#hashtag`, `from:username`, `to:username`, `filter:media`.
   - Tabs: **Top**, **Latest**, **People**, **Media**.
   - Recent search history saved per account with single-item deletion and "Clear all".

### F. Direct Messages (`/messages`)
1. **Layout:**
   - Two-pane desktop layout: Conversation list on left (`380px` - `400px`), active chat room on right.
   - Mobile: Separate conversation list view and active message room view.
2. **Messaging Capabilities:**
   - Realtime delivery via WebSocket.
   - Optimistic message append with pending status.
   - Delivery checkmarks and unread badges.

### G. Bookmarks (`/bookmarks`)
1. **Private Curation:**
   - Displays all posts bookmarked by the authenticated user in reverse chronological order.
   - Header shows `@username` and action menu.
   - Empty state: Custom illustration, "Save posts for later", and explanatory call-to-action text.

### H. Notifications (`/notifications`)
1. **Tabs:** **All**, **Verified**, **Mentions**.
2. **Notification Elements:**
   - Actor avatar, action icon (Heart for like, Repeat for repost, UserPlus for follow, MessageCircle for reply/mention).
   - Direct link to referenced post or profile.
   - Realtime badge counter in left rail and mobile bottom nav.

---

## 3. Responsive Breakpoints

| Breakpoint | Width Range | Left Nav Rail | Main Content Column | Right Sidebar | Bottom Nav |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **Desktop XL** | `>= 1280px` | Full (`275px`) | Fixed (`600px`) | Expanded (`350px`) | None |
| **Desktop / Tablet** | `1000px - 1279px`| Icons Only (`68px` - `88px`) | Fixed (`600px`) | Compact (`350px`) | None |
| **Small Tablet** | `688px - 999px` | Icons Only (`68px`) | Fluid (`600px`) | Hidden | None |
| **Mobile** | `< 688px` | Hidden | Full Width (`100%`) | Hidden | Visible (`53px` sticky) |

---

## 4. Verification Checkpoint Summary
All documented behaviors have been cross-checked against observable reference behavior and our local implementation baseline.
