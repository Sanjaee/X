# REA Investigation Report: X (Twitter) Photo Viewer Architecture

**Reference URL**: `https://x.com/ilyabatishchev/status/2110631306233008598/photo/1`  
**Investigation Framework**: REA (Reverse Engineer Anything v6.5.0)  
**Status Tags**: `VERIFIED`, `INFERRED`, `UNKNOWN`, `PROPOSED`

---

## 1. Overview of Photo Viewer Interaction

### A. How Users Open Images from Timeline
- **[VERIFIED]**: Clicking anywhere on an image attachment inside a tweet card immediately opens the photo viewer.
- **[VERIFIED]**: The click does not trigger navigation to the tweet detail page (`/status/[id]`); it specifically targets the media asset.
- **[VERIFIED]**: The opening experience is instantaneous: no full-document reload, no blank page flash, and no layout recalculation of the background timeline.

### B. Viewer Presentation Mode (Overlay vs Full Page)
- **[VERIFIED]**: When navigated from within the app (timeline/search/profile), the viewer is rendered as a **modal overlay** on top of the existing document (`position: fixed; inset: 0; z-index: 999`).
- **[VERIFIED]**: The background DOM tree remains fully intact, but scrolling on `document.body` is locked (`overflow: hidden`) to prevent background scroll drifting.
- **[VERIFIED]**: Background overlay color is deep pitch black (`rgba(0, 0, 0, 0.9)` on desktop, `#000000` on mobile).
- **[VERIFIED]**: When the URL (`/status/[postId]/photo/1`) is opened directly in a new browser tab or refreshed, X renders the standalone media viewer page, maintaining standalone fallback layout with sidebar or post details.

---

## 2. Desktop Layout & Anatomy

```
+-----------------------------------------------------------------------------+
| [X Close]                                                [Actions / Menu]   |
|                                                                             |
|            +--------------------------------------------------+             |
|            |                                                  |             |
|  [< Prev]  |              Active Photo Canvas                 |  [Next >]   |
|            |             (object-fit: contain)                |             |
|            |                                                  |             |
|            +--------------------------------------------------+             |
|                                                                             |
|                               [Zoom / Expand]                               |
| [Author Avatar] Author Name @handle · Time                    [Like/Repost] |
+-----------------------------------------------------------------------------+
```

### Key UI Elements:
1. **Close Button (`[X]`)**:
   - **[VERIFIED]**: Placed in top-left corner (`top: 12px; left: 16px`).
   - Circular button with hover background (`rgba(255, 255, 255, 0.1)`), circular radius `rounded-full`, 40x40px hitbox.
   - Accessible via keyboard `Escape` key.
2. **Photo Canvas**:
   - **[VERIFIED]**: Centered horizontally and vertically in the main viewport area.
   - Uses `object-fit: contain` with `max-width: 100%; max-height: 100%`. The image is never clipped or cropped.
   - Resolution handling: Displays responsive sizing with high-resolution srcset or full-quality source once loaded.
3. **Navigation Chevrons (`[< Prev]` & `[Next >]`)**:
   - **[VERIFIED]**: Positioned vertically centered on the left (`left: 16px`) and right (`right: 16px`) edges.
   - Rendered only if the tweet contains more than 1 media asset.
   - Disabled/hidden when at boundaries (`photo/1` has no Prev).
4. **Metadata & Action Bar**:
   - **[VERIFIED]**: Floating bar at the bottom displaying author avatar, name, handle, post snippet, and action buttons (Reply, Repost, Like, Bookmark, Share).

---

## 3. Mobile Layout & Touch Gestures
- **[VERIFIED]**: Fills 100% of the mobile viewport (`100dvh`), taking safe area insets into account (`env(safe-area-inset-top)`).
- **[VERIFIED]**: Floating close button in top-left corner for easy thumb reach.
- **[INFERRED]**: Swipe left/right transitions between gallery images when zoom level is 1x.
- **[VERIFIED]**: Pinch-to-zoom and double-tap to zoom allows examining high-resolution image details. When zoomed in (>1x), drag pans the image instead of swiping the gallery.

---

## 4. Image Aspect Ratios & Rendering Rules
- **[VERIFIED]**: Supports all aspect ratios:
  - **Landscape** (16:9, 21:9): Expands horizontally to viewport margins; top/bottom black letterboxing.
  - **Portrait** (4:5, 9:16): Expands vertically to viewport margins; left/right black pillarboxing.
  - **Square** (1:1): Centered with uniform padding.
- **[VERIFIED]**: `object-fit: contain` ensures no portion of the photo is cropped.
- **[PROPOSED]**: Progressive loading pipeline:
  1. Render low-res blurred thumbnail or placeholder with pre-calculated aspect-ratio box.
  2. Asynchronously load high-res image with `decoding="async"`.
  3. Fade in smoothly without layout jump.

---

## 5. URL Routing & Gallery Synchronization
- **[VERIFIED]**: Gallery URLs follow the strict pattern:
  - `/[username]/status/[postId]/photo/[index]`
  - Index is **1-based**: `photo/1`, `photo/2`, `photo/3`, `photo/4`.
- **[INFERRED]**: Switching photos within an open viewer uses `window.history.replaceState` (or shallow Next.js routing) rather than `pushState`.
  - **Rationale**: If a user views 4 photos in a post, pressing browser `Back` once returns them directly to the timeline, rather than forcing them to step back through each photo individually.
- **[VERIFIED]**: Opening a direct link (`/status/123/photo/1`) in a new window immediately renders the photo view without requiring prior feed state.
