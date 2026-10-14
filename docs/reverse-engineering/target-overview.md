# Target Overview: X (Twitter) Public Architecture Investigation

**Status:** Phase 1 Complete  
**Date:** October 2026  
**Investigated Origin:** `https://x.com/`  
**Investigation Framework:** REA (Reverse Engineer Anything v6.5.0) & Passive CDP / Network / DOM Inspection

---

## 1. System Summary

X (formerly Twitter) is a global real-time social networking and microblogging platform. It facilitates short-form text publishing, multimedia sharing (images, animated GIFs, videos, audio/Spaces), conversational threads, bookmarking, reposts (retweets), direct messaging, and recommendation-driven feeds.

The target was investigated in public access mode without bypassing authentication, paywalls, or administrative barriers. Public endpoints investigated include:
- Root landing (`https://x.com/`)
- Public user profile (`https://x.com/NASA`)
- Public post status thread (`https://x.com/NASA/status/2110371804959191288`)
- Explore / Onboarding landing (`https://x.com/explore`)
- Web Application Manifest (`https://x.com/manifest.json`)
- Production Asset CDN (`https://abs.twimg.com/`)

---

## 2. Target Routing and URL Schema

The public target exhibits a clean, resource-oriented URI structure:

| URL Pattern | Function / View | Access Level | Observed Behavior |
| :--- | :--- | :--- | :--- |
| `/` | Landing / Login / Signup (Logged out) or Home Feed (Logged in) | Public (Logged out) | Renders hero call-to-action ("Happening now"), OAuth/Phone login, and footer navigation. |
| `/{username}` | Public Profile Page | Public | Renders banner, avatar, bio, follower/following stats, pinned post, and user timeline. |
| `/{username}/status/{postId}` | Post Detail & Thread | Public | Displays main focal post, parent/reply thread, view count, engagement metrics, and sidebar. |
| `/{username}/with_replies` | User Posts & Replies | Public / Auth | Sub-tab on user profile displaying replies. |
| `/{username}/media` | User Media Gallery | Public / Auth | Sub-tab on user profile displaying photos/videos. |
| `/{username}/reposts` | User Reposts | Public / Auth | Sub-tab on user profile displaying reposted content. |
| `/explore` | Explore & Trends | Public (Redirects/Skeletons) | Trend topics, search bar, news cards. |
| `/search?q={query}` | Global Search | Public / Auth | Multi-tab search results (Top, Latest, People, Media, Lists). |
| `/compose/post` | New Post Composer | Authenticated | Modal dialog or inline composer on desktop. |
| `/notifications` | Activity Notifications | Authenticated | Notification feed (All, Verified, Mentions). |
| `/messages` | Direct Messages (DMs) | Authenticated | Split view: conversation list on left, chat thread on right. |
| `/settings` | Account & Privacy Preferences | Authenticated | Master settings panel. |

---

## 3. Scope of Reconstruction

Our project objective is to build an independent, full-stack microblogging platform matching X's visual layout, design system, interactive behaviors, and responsive structure:

1. **Frontend**: Next.js App Router (React 19+, Server Components, TypeScript, Tailwind CSS, Lucide icons).
2. **Backend**: Go (`net/http` / Chi router, REST API under `/api/v1`, WebSocket for real-time events, OpenAPI contract).
3. **Database**: PostgreSQL (relational normalized schema, SQL migrations, index strategy).
4. **Caching & Async**: Redis (caching, rate limiting) & optional RabbitMQ (background job worker).
5. **Media Storage**: S3-compatible object storage (MinIO for local dev, AWS/Cloudflare R2 for production).
6. **Deployment**: Docker Compose, Nginx reverse proxy, multi-stage builds.
