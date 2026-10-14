# REA Investigation: X/Twitter Social Graph

## 1. Executive Summary
This document records the reverse engineering findings regarding X's observable social graph mechanics (following, followers, mutual connections, private accounts, and protective controls such as blocks and mutes). Findings are cataloged with REA evidentiary classifications: **VERIFIED**, **INFERRED**, **UNKNOWN**, and **PROPOSED**.

---

## 2. REA Evidentiary Matrix

| Feature Flow | Classification | Observed Behavior / Architectural Reality | Proposed Independent Implementation |
|---|---|---|---|
| **Follow Public Account** | **VERIFIED** | Instant toggle from black "Follow" to white outlined "Following" with optimistic UI. Counter increments by 1. Rollback on failure. | Idempotent `POST /api/v1/users/{userId}/follow` writing to `follows` table; Redis counter cache updated; outbox event emitted. |
| **Unfollow Account** | **VERIFIED** | Hovering on "Following" turns red text "Unfollow". Clicking prompts confirmation modal: "Unfollow @username? Their posts will no longer show up...". | `DELETE /api/v1/users/{userId}/follow`. Deletes row or sets inactive; decrements follower counters; invalidates feed cache. |
| **Followers & Following Lists** | **VERIFIED** | Modal and dedicated tab displaying user cards (avatar, display name, handle, verified badge, bio, follow button, and "Follows you" pill). Infinite scrolling pagination. | Cursor-based pagination (`created_at`, `user_id`) to prevent duplicate items during live feed shifts. |
| **Mutual Connections** | **VERIFIED** | Displays "Follows you" badge on handles that follow the viewer. Profiles showcase "Followed by [User A], [User B], and X others you follow". | Computed via SQL set intersection (`follows f1 JOIN follows f2 ON f1.following_id = f2.follower_id`) with 1-hour Redis cache. |
| **Private Account / Follow Request** | **VERIFIED** | Private accounts (lock icon 🔒) display "Follow" which transitions to "Pending" / "Requested". Owner receives notification and can "Accept" or "Decline". | `follow_requests` table with statuses `pending`, `accepted`, `rejected`, `cancelled`. Private posts gated by acceptance check. |
| **Hard Block vs. Soft Mute** | **VERIFIED** | **Block:** Mutual interaction severed. Blocked user cannot view profile, feed, or send DMs. Follower relationship automatically dissolved in both directions.<br>**Mute:** Silent suppression. Muted user is unaware; their posts and notifications are hidden from the viewer's feed. | Separate `blocks` and `mutes` tables. Authorization layer rejects blocked user access at the gateway level. Mutes applied at query projection level. |
| **Counter Fanout Strategy** | **INFERRED** | High-follower accounts do not perform live `COUNT(*)` across millions of rows; materialized counters with eventual consistency or cache warming are utilized. | Dual-write: row update triggers PostgreSQL increment, plus Redis atomic `HINCRBY` for hot profiles. |
| **Internal Social Graph DB** | **UNKNOWN** | X historically utilized custom graph services (FlockDB / Manhattan graph partitions). | Clean relational PostgreSQL model with indexed foreign keys and Redis presence/counters. |

---

## 3. UI Component Specifications & States

### Follow / Following Button
- **Default State (Unfollowed):** Solid background (black in dark mode, white text "Follow").
- **Hover State (Unfollowed):** Slight opacity elevation (90%).
- **Pending Request State (Private Account):** Neutral border, muted text "Requested".
- **Following State:** Transparent background, subtle border, white text "Following".
- **Hover on Following State:** Border turns red border, text turns red "Unfollow".
- **Confirmation Modal:** Shown when unfollowing to avoid accidental clicks.

### Relationship Badge ("Follows you")
- Muted grey pill rendered alongside `@username` when `follows` table contains `follower_id = target` AND `following_id = viewer`.

### Blocked Profile View
- If viewer has blocked user: "You blocked @username. You cannot see their posts or interact with them. [Unblock]".
- If user has blocked viewer: "You are blocked. You cannot follow or view @username's posts."

---

## 4. Acceptance Criteria for Implementation
1. Double-clicking "Follow" must not trigger concurrent duplicate entries (guaranteed by unique constraint `(follower_id, following_id)`).
2. Users cannot follow themselves (enforced via database check `follower_id != following_id`).
3. Blocking automatically dissolves existing follow records in both directions.
4. Blocked users cannot retrieve the blocker's profile posts or send DMs (HTTP 403 Forbidden).
5. Private account posts are completely hidden from non-followers unless a follow request is accepted.
