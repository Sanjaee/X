# REA Implementation Decisions: Photo Viewer, Video Player & Navigation State

This document formalizes the architectural decisions made for the production-grade implementation of the X-like photo viewer, video player, and seamless back navigation system.

---

## 1. Architectural Decisions Summary

| Decision Area | Chosen Strategy | Technical Rationale |
|---|---|---|
| **Viewer Display Mode** | Dual-Mode: In-App Modal Overlay + Standalone Next.js Page Route | Provides instantaneous, zero-flash experience when navigating from feed, while still supporting direct link sharing, SEO, and hard refresh. |
| **URL & History Strategy** | In-app modal uses `history.pushState` to update URL to `/[username]/status/[postId]/photo/[index]`. Gallery switching uses `history.replaceState`. | Ensures pressing browser `Back` closes the viewer in a single click and returns directly to the feed, without stepping through every gallery image. |
| **Scroll Restoration** | Composite Anchor Tracking (`sessionStorage` + `data-post-id` DOM query + Virtualizer index alignment) | Prevents the infamous "jump to top" bug. Ensures the feed remains exactly where the user clicked it, with zero skeleton flash. |
| **Feed Cache Policy** | TanStack Query `staleTime: 60s`, `gcTime: 5m`, `maxPages: 5` | Keeps loaded feed pages in memory when opening and closing the viewer, completely eliminating unnecessary refetches. |
| **Video Player** | Native HTML5 `<video>` engine with custom X-style UI overlays | Avoids heavy player bundle bloat (e.g. Video.js 300KB+), gives 100% control over design tokens, scrubbers, and autohiding controls. |
| **Video Autoplay & Concurrency** | Single-Active Audio Stream with `IntersectionObserver` threshold 0.5 | Ensures only one video audio plays at a time. Offscreen videos automatically pause to conserve GPU/decoding memory. |

---

## 2. Component Hierarchy

```
apps/web/src/components/
├── media/
│   ├── PhotoViewerModal.tsx       # Root modal overlay with dark backdrop & keyboard listeners
│   ├── PhotoCanvas.tsx            # Centered photo canvas with touch/zoom/pan controller
│   ├── PhotoNavigation.tsx        # Left/Right navigation chevrons & photo indicator
│   ├── PhotoZoomController.tsx    # Zoom in/out, pan, reset, double-click zoom
│   ├── VideoPlayer.tsx            # In-feed & fullscreen video player
│   ├── VideoControls.tsx          # Custom control bar (play/pause, volume, time, fullscreen)
│   ├── VideoProgressBar.tsx       # Scrub bar with hover timestamp tooltip
│   └── MediaViewerContext.tsx     # Global context for opening photo/video from any feed
```

---

## 3. URL & Route Scheme

1. **In-Feed Click**:
   - User clicks image #1 on tweet `post-123` by `@alice`.
   - `MediaViewerContext` mounts `PhotoViewerModal` over the current page.
   - URL updates to `/@alice/status/post-123/photo/1` (also aliased at `/status/post-123/photo/1`).
2. **Gallery Switching**:
   - User clicks Next arrow.
   - Active index increments to `2`.
   - URL updates via `window.history.replaceState(null, '', '/@alice/status/post-123/photo/2')`.
3. **Closing Viewer**:
   - User clicks `[X]`, presses `Escape`, or clicks browser `Back`.
   - Modal closes, URL returns to `/home` (or origin route), scroll anchor is smoothly restored.
4. **Direct Link / Hard Refresh**:
   - User visits `http://localhost/@alice/status/post-123/photo/1` directly.
   - Next.js renders `app/[username]/status/[postId]/photo/[photoIndex]/page.tsx`.
   - Standalone fallback page renders high-res photo, post metadata, and a Back button directing to `/@alice/status/post-123`.
