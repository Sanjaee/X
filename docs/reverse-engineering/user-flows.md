# User Flows & Interaction State Progression

**Status:** Phase 1 Complete  
**Scope:** Core User Journeys for Microblogging Platform

---

## 1. Authentication & Onboarding Flow

```mermaid
graph TD
    A[Visitor arrives at /] --> B{Authenticated?}
    B -- Yes --> C[Redirect to /home Timeline]
    B -- No --> D[Display Landing Hero & Auth Modals]
    D --> E[Click 'Sign Up']
    E --> F[Enter Name, Email/Username, Password]
    F --> G[Submit to /api/v1/auth/register]
    G -- Success --> H[Set HttpOnly Session Cookie & Redirect /home]
    G -- Failure --> I[Display Field Validation Errors]
    D --> J[Click 'Log In']
    J --> K[Enter Credentials]
    K --> L[Submit to /api/v1/auth/login]
    L -- Success --> H
    L -- Failure --> M[Display Invalid Credentials Alert]
```

---

## 2. Post Creation & Media Attachment Flow (X Observable Model)

```mermaid
sequenceDiagram
    autonumber
    actor User
    participant Composer as Frontend Composer (State)
    participant MediaAPI as Media Service (/media/upload)
    participant R2 as Cloudflare R2 CDN
    participant PostAPI as Posts API (/posts)
    participant DB as PostgreSQL

    User->>Composer: Types post content (< 280 or < 4,000 chars for Premium)
    User->>Composer: Selects image(s)/video (1-4 files)
    Composer->>Composer: URL.createObjectURL(file) -> stage in local state
    Note over Composer: Previews rendered immediately with 0ms network latency; user can delete/modify attachments
    User->>Composer: Clicks "Post" button
    Composer->>MediaAPI: POST /api/v1/media/upload (multipart)
    MediaAPI->>MediaAPI: Image processing (strip EXIF, resize >2048px, JPEG 85% quality)
    MediaAPI->>R2: S3 PutObject to lostmedia bucket
    R2-->>MediaAPI: Upload acknowledged
    MediaAPI-->>Composer: 201 Created { media_id, url }
    Composer->>PostAPI: POST /api/v1/posts { content, media_ids }
    PostAPI->>DB: INSERT into posts & post_media (in transaction)
    DB-->>PostAPI: Post entity created
    PostAPI-->>Composer: 201 Created { post }
    Composer->>Composer: Revoke ObjectURLs, reset textarea & prepend to timeline
```

---

## 5. Subscription Checkout, Plisio Gateway & Verification Flow (Section 6A)

```mermaid
sequenceDiagram
    autonumber
    actor User
    participant Web as Next.js Web App (/premium)
    participant Invoice as Checkout Invoice (/premium/invoice)
    participant API as Go Backend API
    participant Plisio as Plisio Crypto Gateway
    participant DB as PostgreSQL

    User->>Web: Selects Plan (e.g. Premium) & Billing Cycle
    User->>Web: Clicks "Subscribe with Crypto (Plisio)"
    Web->>API: POST /api/v1/subscriptions/checkout { plan_id, billing_cycle }
    API->>DB: INSERT into payment_orders (status: 'pending')
    API->>Plisio: GET /invoices/new (api_key=PLISIO_API_KEY)
    Plisio-->>API: 200 { invoice_url, txn_id }
    API-->>Web: 201 Created { order_id, invoice_url }
    Web->>Invoice: Redirect to /premium/invoice?order_id={id}
    Note over Invoice: Renders order details, crypto options (BTC, USDT, ETH) and instant test mode
    alt Real Crypto Settlement
        Plisio->>API: POST /api/v1/webhooks/payments/plisio (IPN callback)
    else Dev Simulation Test
        User->>Invoice: Clicks "Confirm Simulated Payment"
        Invoice->>API: POST /api/v1/subscriptions/simulate-payment { order_id }
    end
    API->>DB: UPDATE subscriptions (status: 'active', period_end: NOW() + 30d/365d)
    API->>DB: UPDATE users SET is_verified = TRUE, verified_type = 'blue'
    API->>DB: INSERT into verification_badges & invoices
    API-->>Invoice: 200 OK { is_verified: true, status: 'completed' }
    Invoice-->>User: Displays verified celebration & Blue Checkmark badge everywhere
```


---

## 3. Social Interaction & Optimistic Updates Flow

```mermaid
sequenceDiagram
    autonumber
    actor User
    participant UI as Timeline Card
    participant ClientCache as TanStack Query Cache
    participant API as Like Endpoint

    User->>UI: Clicks Heart (Like) Icon
    UI->>ClientCache: 1. Snapshot previous state
    UI->>ClientCache: 2. Optimistically update: is_liked = true, likes_count + 1
    UI->>UI: Play heart pop animation immediately
    UI->>API: POST /api/v1/posts/{id}/likes
    alt Success (200 OK)
        API-->>UI: Confirmed { is_liked: true, likes_count: N }
    else Network / Server Error (5xx / 4xx)
        API-->>UI: Error response
        UI->>ClientCache: Rollback to snapshot (is_liked = false, likes_count - 1)
        UI->>UI: Show toast notification ("Failed to like post")
    end
```

---

## 4. Reply Thread & Detail Navigation Flow

1. User clicks on a post card background (excluding action buttons).
2. Browser navigates to `/{username}/status/{postId}`.
3. Top header displays back arrow button with history return.
4. Focal post rendered prominently in center.
5. Ancestor parent post (if this post is a reply) loaded above with vertical connecting line.
6. Inline reply box rendered below focal post.
7. Nested descendant replies loaded beneath in thread list.

---

## 5. Follow & Social Graph Flow

1. User views profile `/{username}`.
2. If self: Shows "Edit profile" button.
3. If other:
   - Not following: White/black pill "Follow".
   - Following: Transparent outline pill "Following".
4. When hovering "Following": Text changes to "Unfollow" with red text and border (`destructive`).
5. Clicking triggers `POST` or `DELETE /api/v1/users/{username}/follow` with optimistic state update.
