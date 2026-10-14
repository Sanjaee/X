# UI Specification & Design System Reference

**Status:** Phase 1 Complete  
**Reference Origin:** X Web Production Design System (`jetfuel` / `x-web`)

---

## 1. Grid & Spatial Dimensions

```
+------------------+------------------------+-----------------------+
|  Left Sidebar    |  Main Timeline Feed    |  Right Sidebar        |
|  88px / 275px    |  600px                 |  350px - 420px        |
+------------------+------------------------+-----------------------+
```

| Region | Desktop (>=1280px) | Laptop (1000px-1279px) | Tablet (500px-999px) | Mobile (<500px) |
| :--- | :--- | :--- | :--- | :--- |
| **Left Rail Width** | `275px` (labeled) | `88px` (icon-only) | `88px` (icon-only) | Hidden (Drawer) |
| **Main Feed Width** | `600px` | `600px` | `100%` (max `600px`) | `100%` |
| **Right Rail Width**| `420px` (min `258px`) | `350px` | Hidden | Hidden |
| **Top Header Height** | `53px` (sticky) | `53px` (sticky) | `53px` (sticky) | `53px` (sticky) |
| **Mobile Bottom Bar** | N/A | N/A | N/A | `53px` (fixed bottom) |
| **Max Content Container** | `1280px` centered | `1050px` | `700px` | `100%` |

---

## 2. Color Palette & Token Definitions

### Theme Variables

```css
/* Light Theme */
:root[data-theme="light"] {
  --background: #ffffff;
  --foreground: #0f1419;
  --card: #f7f9f9;
  --border: #eff3f4;
  --muted: #536471;
  --accent-blue: #1d9bf0;
  --accent-blue-hover: #1a8cd8;
  --like-pink: #f91880;
  --like-pink-bg: rgba(249, 24, 128, 0.1);
  --repost-green: #00ba7c;
  --repost-green-bg: rgba(0, 186, 124, 0.1);
  --hover-subtle: rgba(15, 20, 25, 0.1);
}

/* Dark Theme (Lights Out / Pitch Black) */
:root[data-theme="dark"] {
  --background: #000000;
  --foreground: #e7e9ea;
  --card: #16181c;
  --border: #2f3336;
  --muted: #71767b;
  --accent-blue: #1d9bf0;
  --accent-blue-hover: #1a8cd8;
  --like-pink: #f91880;
  --like-pink-bg: rgba(249, 24, 128, 0.1);
  --repost-green: #00ba7c;
  --repost-green-bg: rgba(0, 186, 124, 0.1);
  --hover-subtle: rgba(231, 233, 234, 0.1);
}
```

---

## 3. Typography & Hierarchy

| Element | Font Size | Line Height | Weight | Color (Dark/Light) |
| :--- | :--- | :--- | :--- | :--- |
| **Hero Heading** | `31px` / `2rem` | `36px` | `800` (Bold) | `--foreground` |
| **Page Title (Header)**| `20px` / `1.25rem`| `24px` | `700` (Bold) | `--foreground` |
| **Card Display Name** | `15px` | `20px` | `700` (Bold) | `--foreground` |
| **Handle & Timestamp**| `15px` | `20px` | `400` (Regular) | `--muted` |
| **Post Body Text** | `15px` | `20px` | `400` (Regular) | `--foreground` |
| **Focal Post Body** | `23px` / `1.4rem` | `28px` | `400` (Regular) | `--foreground` |
| **Engagement Numbers** | `13px` | `16px` | `400` (Regular) | `--muted` |
| **Button Text** | `15px` | `20px` | `700` (Bold) | Inverted / Accent |

---

## 4. Components & Interactive States

### 1. Avatars
- Post Avatar: `40px` circular (`size-10 rounded-full`).
- Large Profile Avatar: `136px` circular (`size-[136px] rounded-full border-4 border-background`).
- Mobile Profile Avatar: `75px` circular.
- Mini Drawer Avatar: `32px` circular.

### 2. Action Buttons & Engagement Pills
- **Reply**: Hover background `rgba(29, 155, 240, 0.1)`, icon & text color `#1d9bf0`.
- **Repost**: Hover background `rgba(0, 186, 124, 0.1)`, icon & text color `#00ba7c`. Active state holds green.
- **Like**: Hover background `rgba(249, 24, 128, 0.1)`, icon & text color `#f91880`. Active state holds solid pink/red heart with subtle scale animation.
- **Bookmark**: Hover background `rgba(29, 155, 240, 0.1)`, icon & text color `#1d9bf0`.
- **Share**: Hover background `rgba(29, 155, 240, 0.1)`, icon & text color `#1d9bf0`.

### 3. Media Grids
- **1 Image**: Aspect ratio preserved (up to 16:9 or max height 510px), `rounded-2xl border border-border overflow-hidden`.
- **2 Images**: 2 equal columns, height `288px`, rounded corners.
- **3 Images**: 1 large on left (span 2 rows), 2 stacked on right.
- **4 Images**: 2x2 symmetrical grid, height `288px`.

### 4. Modals and Dropdowns
- Modal overlay: `rgba(91, 112, 131, 0.4)` backdrop blur.
- Dialog box: `max-w-[600px] w-full rounded-2xl bg-background border border-border shadow-2xl`.
- Dropdown menu: `rounded-xl bg-background border border-border shadow-lg py-2 min-w-[200px] z-50`.
