# REA Investigation Report: X (Twitter) Video Player Architecture

**Investigation Framework**: REA (Reverse Engineer Anything v6.5.0)  
**Status Tags**: `VERIFIED`, `INFERRED`, `UNKNOWN`, `PROPOSED`

---

## 1. Overview & Playback Lifecycle

### A. Timeline Video vs Viewer Video
- **[VERIFIED]**: Within the timeline feed, videos are embedded directly inside the tweet card with a fixed or intrinsic aspect ratio (typically 16:9 for landscape, 1:1 for square, 9:16 for portrait).
- **[VERIFIED]**: Video player in feed uses a clean overlay with:
  - Play/Pause toggle.
  - Mute/Unmute audio button (bottom right corner).
  - Current timestamp and duration.
  - Minimal scrub progress bar.
- **[VERIFIED]**: Clicking the video or the expand/fullscreen icon transitions to the dedicated Video Viewer (fullscreen modal overlay with dark backdrop), matching the photo viewer layout.

---

## 2. In-Feed Video Autoplay & Resource Bounds

### A. Autoplay Policy
- **[VERIFIED]**: Videos autoplay **only when muted** (`muted = true`). Browsers strictly block unmuted autoplay without user gesture.
- **[VERIFIED]**: Unmuting one video immediately mutes or pauses any other playing audio on the page. Only **one video audio stream** is active at any time.

### B. Viewport Intersection & Performance Bounds
- **[VERIFIED]**: Utilizes `IntersectionObserver` to trigger playback only when the video element is at least 50% visible in the viewport (`threshold: 0.5`).
- **[VERIFIED]**: When scrolled out of view, playback pauses immediately (`video.pause()`) to save GPU, battery, and decoding memory.
- **[PROPOSED]**: Hard limit of active media decoders: At most 1 video element active/playing simultaneously across the entire feed.

---

## 3. Video Controls & Scrubbing Anatomy

```
+-----------------------------------------------------------------------------+
|                                                                             |
|                                [Video Surface]                              |
|                            (object-contain, centered)                       |
|                                                                             |
| [Play/Pause]  [Time: 0:14 / 1:30]  [====O======== Progress]  [Vol] [FS]     |
+-----------------------------------------------------------------------------+
```

### Controls Specifications:
1. **Play/Pause Button**:
   - Toggle icon: Lucide `Play` / `Pause`.
   - Accessible via keyboard `Space` key.
2. **Seek & Progress Bar**:
   - Scrubber track with loaded buffer indicator and current progress fill.
   - Hover preview tooltip showing the timestamp at cursor position.
   - Click or drag scrub updating `video.currentTime` smoothly without stalling UI.
3. **Volume & Mute Slider**:
   - Audio state indicator (`Volume2`, `VolumeX`).
   - Hover slider adjusting volume from `0.0` to `1.0`.
   - Global volume memory saved to `localStorage`.
4. **Playback Speed Menu**:
   - Speeds: `0.5x`, `1.0x` (Normal), `1.25x`, `1.5x`, `2.0x`.
5. **Fullscreen Toggle**:
   - Native HTML5 Fullscreen API (`element.requestFullscreen()`).
   - Graceful fallback for mobile Safari iOS (`video.webkitEnterFullscreen()`).
6. **Autohide Behavior**:
   - **[VERIFIED]**: When playing, controls fade out after 2.5 seconds of inactivity.
   - Moving mouse, tapping screen, or pressing keys immediately restores control visibility.

---

## 4. Buffering, Error Handling & Fallbacks
- **[VERIFIED]**: Poster image displayed before video data is ready.
- **[VERIFIED]**: Centered spinning loader displayed during network stall/buffering (`waiting` event).
- **[VERIFIED]**: Actionable error card displayed on load failure with a **Retry** button.
- **[PROPOSED]**: Video streaming via standard HTTP range requests supported by S3/R2 storage, enabling instant scrubbing to unbuffered sections.
