# LostMedia Production Attack Surface Analysis

**Date**: October 17, 2026  
**Target**: `https://lostmedia.pro/`  

---

## 1. Attack Surface Enumeration

The attack surface consists of entry points exposed to unauthenticated internet traffic, authenticated users, and third-party webhook integrations.

```
                      [ INTERNET ]
                           │
       ┌───────────────────┼───────────────────┐
       ▼                   ▼                   ▼
 [ Unauthenticated ]  [ Authenticated ]   [ Webhooks / Inbound ]
 - Landing page /     - Post creation     - /api/v1/webhooks/
 - /auth/register     - Post deletion       payments/plisio
 - /auth/login        - Likes / Reposts   - NextAuth OAuth
 - /auth/google       - Bookmarks           callbacks
 - /feed (public)     - Direct Messages
 - /search (public)   - Media Upload
 - /users/{username}  - Profile updates
 - /uploads/*         - Subscriptions
 - /health/*          - Verification
```

---

## 2. API Endpoint Classification & Exposure Matrix

| Method | Endpoint | Auth Required | Exposure Level | Critical Assets Accessed | Vulnerability Potential |
| :--- | :--- | :--- | :--- | :--- | :--- |
| `POST` | `/api/v1/auth/register` | None | Public | `users`, `user_profiles` | User enumeration, spam registration |
| `POST` | `/api/v1/auth/login` | None | Public | `users`, `sessions` | Credential stuffing, brute force |
| `POST` | `/api/v1/auth/google` | None | Public | `users`, OAuth tokens | Account takeover if email spoofed |
| `POST` | `/api/v1/webhooks/payments/plisio` | None (Signature) | Public | `payment_orders`, `subscriptions` | Webhook forgery, free privilege upgrade |
| `GET` | `/uploads/*` | None | Public | File System `/app/data/uploads` | Directory listing, stored file leakage |
| `GET` | `/health/metrics` | None | Public | Go Runtime Heap, Goroutines | System reconnaissance, denial of service profiling |
| `POST` | `/api/v1/subscriptions/simulate-payment` | Auth | Authenticated | `payment_orders`, `subscriptions` | Logic bypass, unauthorized badge grant |
| `POST` | `/api/v1/media/upload` | Auth | Authenticated | Local disk, Cloudflare R2 | Arbitrary file upload, resource exhaustion |
| `PATCH`| `/api/v1/me/profile` | Auth | Authenticated | `user_profiles` | Stored XSS via website/bio, prototype pollution |
| `POST` | `/api/v1/posts` | Auth | Authenticated | `posts`, `post_media` | Spam injection, XSS, database load |
| `DELETE`|`/api/v1/posts/{postId}` | Auth | Authenticated | `posts` | IDOR (verified protected via author_id check) |
| `GET` | `/api/v1/conversations/{convId}` | Auth | Authenticated | `conversations`, `messages` | IDOR (verified protected via membership check) |

---

## 3. High-Risk Attack Vectors Identified

### Vector 1: Business Logic Bypass in Subscription System (`SimulatePayment`)
- **Path**: `POST /api/v1/subscriptions/simulate-payment`
- **Threat Actor**: Any registered user.
- **Mechanism**: The endpoint was accessible without an environment check, allowing any user to simulate and activate real subscriptions without spending cryptocurrency.

### Vector 2: Cryptographic Signature Omission in Payment Webhook
- **Path**: `POST /api/v1/webhooks/payments/plisio`
- **Threat Actor**: Unauthenticated remote attacker.
- **Mechanism**: `VerifyHash` was only validated if non-empty; submitting an empty hash allowed signature verification to be bypassed.

### Vector 3: Stored Cross-Site Scripting (XSS) via Profile Website
- **Path**: `PATCH /api/v1/me/profile` -> `GET /{username}`
- **Threat Actor**: Malicious authenticated user.
- **Mechanism**: Profile website accepted `javascript:` URIs and rendered them into `<a href="...">` without scheme sanitization.

### Vector 4: Information Disclosure via Open Directory Listing
- **Path**: `GET /uploads/`
- **Threat Actor**: Unauthenticated internet crawler / attacker.
- **Mechanism**: Standard `http.FileServer` returns raw directory listings, disclosing user file paths.

### Vector 5: Global Rate Limiting Collision (Collateral DoS)
- **Path**: All requests to `x-proxy`.
- **Threat Actor**: Heavy automated scraper or single malicious user.
- **Mechanism**: `x-proxy` lacked reverse proxy real-IP restoration from `global-nginx`, causing all visitors to share one 30 req/s pool.
