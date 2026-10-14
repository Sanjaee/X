# Frontend Architecture & Reverse Engineering Findings

**Status:** Phase 1 Complete  
**Evidence Source:** Live DOM, CSS bundle (`styles-BPdMqb99.css`), scripts, and manifest from `https://x.com/`

---

## 1. UI Framework and Runtime

### Evidence & Verification
- **Framework & Routing**:
  - **VERIFIED**: X Web client utilizes **TanStack Router** (`$_TSR.router`) with streaming server-side rendering (SSR). The HTML payload includes hydration hooks:
    ```javascript
    self.$_TSR = { h() { this.hydrated = !0, this.c() }, e() { this.streamEnded = !0, this.c() } ... }
    $_TSR.router = ($R => $R[0] = { manifest: void 0, matches: $R[1] = ... })
    ```
  - **VERIFIED**: Streaming serialization uses **Seroval** (`__SEROVAL_STREAM__: true`).
  - **VERIFIED**: The client bootstrap triggers a custom document event `x-web:bootstrap-data`.
  - **VERIFIED**: Client error monitoring is handled by **Sentry** (`meta name="sentry-trace"`, `sentry-release="e69bbfac36e935aa477fe462405a498955108e53"`).

### Client State & Feature Flags
- **VERIFIED**: Initial bootstrap state is embedded in `window.__INITIAL_DATA__`:
  - `appMetadata`: Environment (`prod`), country (`ID`), session flags (`hasMultiAccountSession`, `isTwoffice`).
  - `featureSwitchPayload`: Granular boolean toggles (e.g., `responsive_web_verified_organizations_handle_form_enabled`, `rweb_client_transaction_id_enabled`, `subscriptions_premium_tiers_order_variant`).

---

## 2. Styling System and Design Tokens

### Styling Engine
- **VERIFIED**: Production stylesheet header explicitly identifies **Tailwind CSS v4.3.3**:
  ```css
  /*! tailwindcss v4.3.3 | MIT License | https://tailwindcss.com */
  ```
- **VERIFIED**: Internal platform codebase is designated as `jetfuel` / `x-web` (`.jf-element`, `.jetfuel-style-root`).

### Layout Metrics & Breakpoints
The following CSS layout custom properties are extracted verbatim from `styles-BPdMqb99.css`:

```css
:root {
  --layout-width-rail-icon: 88px;       /* Collapsed navigation sidebar */
  --layout-width-rail-labeled: 275px;   /* Expanded navigation sidebar */
  --layout-width-primary: 600px;        /* Center timeline feed width */
  --layout-width-right: 420px;          /* Right sidebar (search, trends, who to follow) */
  --layout-min-primary: 420px;          /* Minimum timeline width */
  --layout-min-right: 258px;            /* Minimum right sidebar width */
  --sticky-header-height: 53px;         /* Top navigation / tab bar height */
  --spacing: 4px;
  --spacing-horizontal: 16px;
}
```

### Color Palette & Theme Modes
- **Theme Modes**: Supported via `data-theme="light"` and `data-theme="dark"`.
  - Light background: `#ffffff` (`hsl(0, 0%, 100%)`)
  - Dark background: `#000000` (`hsl(0, 0%, 0%)` - AMOLED/Pitch Black mode)
  - Dark modal/card background: `hsl(0, 0%, 8%)`
- **Brand Blue**: `hsl(204, 88%, 53%)` / `#1d9bf0`.
- **Engagement Colors**:
  - Heart / Like: `oklch(65.6% .241 354.308)` (Pink/Red `#f91880`)
  - Retweet / Repost: Green `hsl(160, 100%, 36%)` / `#00ba7c`
  - Reply: Blue `hsl(204, 88%, 53%)` / `#1d9bf0`
  - Bookmark: Blue `hsl(204, 88%, 53%)`
  - Analytics / Views: Neutral gray `hsl(206, 14%, 49%)`
- **Custom Accent Colors**: Supported via `root.dataset.accent` from `xweb.display` in localStorage (`yellow500`, `magenta500`, `purple500`, `orange500`, `green500`).

### Typography
- **Primary Font**: `TwitterChirp`, with fallbacks to `-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif`.
- **Preloaded Subsets**:
  - `Chirp-Regular.woff2` (400)
  - `Chirp-Medium.woff2` (500)
  - `Chirp-Bold.woff2` (700)
- **Base Font Size**: `15px` (`calc(15 / 16 * 1rem)`) with `20px` line-height.

---

## 3. Component Hierarchy & Reusable Patterns

From inspection of the timeline, post detail, and profile views:

```
AppLayout
├── LeftNavRail (88px icon / 275px labeled)
│   ├── Logo (X SVG)
│   ├── NavItems [Home, Explore, Notifications, Messages, Grok, Bookmarks, Profile, More]
│   ├── PostActionCTA (Floating Pill Button)
│   └── UserAccountSwitcher (Bottom avatar + menu)
├── MainTimelineContainer (600px border-l border-r border-border)
│   ├── StickyHeader (53px backdrop-blur-md)
│   │   ├── Title / BackButton
│   │   └── TabBar [For you, Following]
│   ├── PostComposer (Inline)
│   │   ├── Avatar (40px circular)
│   │   ├── AutoGrowingTextarea
│   │   ├── MediaAttachmentPreview (Grid 1-4 items)
│   │   ├── Toolbar [Image, GIF, Poll, Emoji, Schedule, Location]
│   │   └── SubmitButton ("Post")
│   └── FeedList (Virtual/Infinite Scroll)
│       └── PostCard (Standard timeline cell)
│           ├── AvatarColumn (40px)
│           └── ContentColumn
│               ├── HeaderRow (DisplayName, VerifiedBadge, @Handle, · , RelativeTimestamp, MoreMenu)
│               ├── PostBody (Parsed text, links, hashtags, mentions)
│               ├── MediaGrid (Images/Video container with rounded borders)
│               └── EngagementBar (Reply, Repost, Like, ViewCount, Bookmark, Share)
└── RightSidebarContainer (420px max, 258px min)
    ├── SearchBox (Pill rounded-full with search icon)
    ├── PremiumPromotionCard
    ├── TrendingCard ("What's happening")
    │   └── TrendItem (Category, Topic, PostCount, MoreMenu)
    ├── WhoToFollowCard
    │   └── UserRecommendationRow (Avatar, Name, Handle, FollowButton)
    └── FooterLegalLinks (About, Terms, Privacy, © 2026 X Corp)
```

---

## 4. Proposed Implementation Architecture

For our full-stack reconstruction, we will map these findings into our chosen modern stack:
- **Framework**: Next.js 15+ App Router with React Server Components (RSC) by default.
- **Styling**: Tailwind CSS v3/v4 configured with exact tokens matching X:
  - Widths: `w-[88px]`, `w-[275px]`, `w-[600px]`, `w-[420px]`.
  - Colors: Brand blue `#1d9bf0`, dark `#000000`, card `#0f1419` / `#16181c`, borders `#2f3336` (dark) / `#eff3f4` (light).
  - Font: Inter / Outfit modern clean typography.
- **Client Cache**: TanStack Query (React Query) for feed caching, cursor pagination, and optimistic updates.
- **Icons**: Lucide React matching X SVG geometry (Heart, Repeat2, MessageCircle, Bookmark, Share2, BarChart2).
