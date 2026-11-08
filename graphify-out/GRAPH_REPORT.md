# Graph Report - .  (2026-11-09)

## Corpus Check
- cluster-only mode — file stats not available

## Summary
- 1307 nodes · 3448 edges · 156 communities (40 shown, 116 thin omitted)
- Extraction: 97% EXTRACTED · 3% INFERRED · 0% AMBIGUOUS · INFERRED: 117 edges (avg confidence: 0.8)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `83a04718`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- writeJSON
- T
- models.go
- PostgresRepository
- MockCandidateProvider
- Context
- Hub
- Context
- app/page.tsx
- .CheckoutSubscription
- MemoryRepository
- dependencies
- [postId]/page.tsx
- index.ts
- UserNameWithRole.tsx
- api.ts
- totp.go
- compilerOptions
- Post
- PostComposer.tsx
- Post
- User
- FeedSkeleton.tsx
- admin.go
- OtherSkeletons.tsx
- qrcode.tsx
- User
- handler.go
- VideoControls.tsx
- .GetActiveSubscription
- Role
- PaymentOrder
- .GetPlanEntitlements
- NewPostgresRepository
- route.ts
- search/page.tsx
- TopLoadingBar.tsx
- .GetActiveSubscription
- .GetNotifications
- .GetNotifications
- middleware.ts
- .DismissUserRecommendation
- .GetUserBadges
- .DismissUserRecommendation
- .GetFollowRequests
- .GetMediaObjectByID
- .GetMessages
- .GetSubscriptionPlanByID
- test_all_apis.py
- next.config.js
- next-env.d.ts
- tailwind.config.ts
- .GetAdminOverview
- .GetTrendingTopics
- backup-db.sh
- monitoring.sh
- AdminOverview
- AdminPostFilter
- AuditLog
- Conn
- Conversation
- CounterReconciliationResult
- Entitlements
- FeedCursor
- FeedFeedback
- FeedImpression
- FollowRequest
- Invoice
- MediaObject
- MembershipEntitlement
- MembershipPlan
- Message
- Notification
- OutboxEvent
- Permission
- github.com/x-reconstruction/api
- PlisioService
- Post
- RankingWeights
- RecentSearch
- Role
- SearchFilterParams
- Client
- Context
- MediaObject
- PaymentOrder
- Repository
- Subscription
- PaymentOrder
- Mux
- T
- Request
- ResponseWriter
- Time
- User
- UUID
- MediaObject
- Handler
- Client
- Repository
- Client
- Context
- Time
- UUID
- Client
- Context
- RawMessage
- UUID
- RawMessage
- AdminPostFilter
- CounterReconciliationResult
- FeedFeedback
- FeedImpression
- Invoice
- OutboxEvent
- PaymentOrder
- RankingWeights
- RecentSearch
- RWMutex
- SearchFilterParams
- Session
- UserTopicPreference
- VerificationApplication
- CounterReconciliationResult
- FeedFeedback
- FeedImpression
- Invoice
- OutboxEvent
- PaymentOrder
- RankingWeights
- RecentSearch
- SearchFilterParams
- Session
- UserTopicPreference
- VerificationApplication
- AdminPostFilter
- Session
- SocialRelationship
- Subscription
- SubscriptionPlan
- TrendingTopic
- UserMembership
- UserRecommendation
- UserTopicPreference
- VerificationApplication
- VerificationBadge

## God Nodes (most connected - your core abstractions)
1. `MemoryRepository` - 176 edges
2. `PostgresRepository` - 114 edges
3. `writeJSON()` - 109 edges
4. `writeError()` - 108 edges
5. `Handler` - 89 edges
6. `PostgresRepository` - 36 edges
7. `User` - 35 edges
8. `apiRequest()` - 34 edges
9. `Handler` - 26 edges
10. `cn()` - 22 edges

## Surprising Connections (you probably didn't know these)
- `setupTestRouter()` --calls--> `AuthMiddleware()`  [INFERRED]
  services/api/internal/handler/handler_test.go → services/api/internal/middleware/middleware.go
- `setupTestApp()` --calls--> `AuthMiddleware()`  [INFERRED]
  services/api/internal/handler/social_messaging_test.go → services/api/internal/middleware/middleware.go
- `setupTestApp()` --calls--> `NewHub()`  [INFERRED]
  services/api/internal/handler/social_messaging_test.go → services/api/internal/realtime/hub.go
- `RateLimiter()` --calls--> `GetUserFromContext()`  [INFERRED]
  services/api/internal/middleware/ratelimit.go → services/api/internal/middleware/middleware.go
- `main()` --calls--> `Load()`  [INFERRED]
  services/api/cmd/server/main.go → services/api/internal/config/config.go

## Import Cycles
- None detected.

## Communities (156 total, 116 thin omitted)

### Community 0 - "writeJSON"
Cohesion: 0.09
Nodes (10): Context, PaymentOrder, Handler, Request, ResponseWriter, Handler, Request, ResponseWriter (+2 more)

### Community 1 - "T"
Cohesion: 0.07
Nodes (62): Client, CryptoGateway, Hub, MediaService, Mux, Repository, Repository, setupTestFullAdminRouter() (+54 more)

### Community 2 - "models.go"
Cohesion: 0.06
Nodes (63): Config, AdminOverview, AdminPostFilter, AuditLog, Conversation, CounterMismatchDetail, CounterReconciliationResult, DatabaseHealth (+55 more)

### Community 3 - "PostgresRepository"
Cohesion: 0.08
Nodes (8): DB, Context, Conversation, PostgresRepository, SocialRelationship, Time, UserRecommendation, UUID

### Community 4 - "MockCandidateProvider"
Cohesion: 0.05
Nodes (43): B, DiversityReranker, RankingWeights, FeedOptions, FeedResult, Candidate, CandidateProvider, CandidateSource (+35 more)

### Community 5 - "Context"
Cohesion: 0.07
Nodes (5): Context, Conversation, SocialRelationship, UserRecommendation, UUID

### Community 6 - "Hub"
Cohesion: 0.07
Nodes (29): Config, Duration, memoryLimiter, Mutex, CryptoCurrency, CryptoGateway, CryptoInvoice, CryptoOperation (+21 more)

### Community 7 - "Context"
Cohesion: 0.11
Nodes (12): AdminOverview, AuditLog, Context, MembershipEntitlement, MembershipPlan, Permission, Post, PostgresRepository (+4 more)

### Community 8 - "app/page.tsx"
Cohesion: 0.17
Nodes (16): FeedResponse, SettingsSection, useComposer(), MobileComposerModal(), MobileComposerModalProps, StagedAttachment, LeftNavRail(), MobileBottomNav() (+8 more)

### Community 9 - ".CheckoutSubscription"
Cohesion: 0.10
Nodes (29): CryptoInvoice, CryptoOperation, Handler, CheckoutRequest, Logger, contextKey, responseWriterWrapper, ReadWriter (+21 more)

### Community 10 - "MemoryRepository"
Cohesion: 0.07
Nodes (17): FeedFeedback, FeedImpression, OutboxEvent, RecentSearch, Session, UserTopicPreference, VerificationApplication, MemoryRepository (+9 more)

### Community 11 - "dependencies"
Cohesion: 0.05
Nodes (42): dependencies, clsx, lucide-react, next, next-auth, react, react-dom, tailwind-merge (+34 more)

### Community 12 - "[postId]/page.tsx"
Cohesion: 0.10
Nodes (25): StagedFile, PostModal(), PostModalProps, StagedAttachment, PostCard(), MediaDetailSidebar(), MediaDetailSidebarProps, MediaViewerContext (+17 more)

### Community 13 - "index.ts"
Cohesion: 0.08
Nodes (28): AdminTab, ConfirmModal(), ConfirmModalProps, TIER_PRESENTATION, TierBadge(), TierBadgeProps, TierKey, RealtimeContext (+20 more)

### Community 14 - "UserNameWithRole.tsx"
Cohesion: 0.15
Nodes (14): EditProfileModal(), EditProfileModalProps, NAME_STYLE_PRESETS, TwoFactorManager(), TwoFactorManagerProps, buildCustomNameInlineStyle(), parseCustomNameStyle(), UserNameWithRole() (+6 more)

### Community 15 - "api.ts"
Cohesion: 0.12
Nodes (22): metadata, AuthModal(), AuthModalProps, ComposerContext, ComposerContextValue, ComposerProvider(), AppShell(), PhotoViewerModal() (+14 more)

### Community 16 - "totp.go"
Cohesion: 0.16
Nodes (21): Handler, mfaTokenPayload, Request, ResponseWriter, buildOTPAuthURL(), checkTOTPWindow(), computeTOTPCode(), createMFAToken() (+13 more)

### Community 17 - "compilerOptions"
Cohesion: 0.07
Nodes (26): compilerOptions, allowJs, esModuleInterop, incremental, isolatedModules, jsx, lib, module (+18 more)

### Community 18 - "Post"
Cohesion: 0.17
Nodes (3): Rows, FeedCursor, Post

### Community 19 - "PostComposer.tsx"
Cohesion: 0.15
Nodes (15): encodeFeedCursor(), HomePage(), VerifiedBadge(), VerifiedBadgeProps, PostComposer(), PostComposerProps, StagedAttachment, UpgradeModalContext (+7 more)

### Community 22 - "FeedSkeleton.tsx"
Cohesion: 0.19
Nodes (4): FeedSkeleton(), PostCardSkeleton(), PostDetailSkeleton(), ProfileSkeleton()

### Community 23 - "admin.go"
Cohesion: 0.16
Nodes (13): AdminDeletePostRequest, AssignMembershipRequest, CreateRoleRequest, SetPermissionsRequest, SetPlanEntitlementsRequest, UpdatePlanRequest, UpdatePostModerationRequest, UpdateRoleRequest (+5 more)

### Community 24 - "OtherSkeletons.tsx"
Cohesion: 0.23
Nodes (3): ExploreSkeleton(), MessagesSkeleton(), NotificationsSkeleton()

### Community 25 - "qrcode.tsx"
Cohesion: 0.24
Nodes (11): ALIGNMENT_PATTERNS, EC_CODEWORDS_L, encodeQRMatrix(), GF_EXP, GF_LOG, gfMul(), NUM_BLOCKS_L, QRCodeSVG() (+3 more)

### Community 27 - "handler.go"
Cohesion: 0.22
Nodes (10): CreatePostRequest, FeedFeedbackRequest, FeedImpressionsRequest, GoogleAuthRequest, LoginRequest, RegisterRequest, SimulatePaymentRequest, UpdatePostRequest (+2 more)

### Community 28 - "VideoControls.tsx"
Cohesion: 0.31
Nodes (7): formatTime(), SPEED_OPTIONS, VideoControls(), VideoControlsProps, formatTime(), VideoProgressBar(), VideoProgressBarProps

### Community 29 - ".GetActiveSubscription"
Cohesion: 0.22
Nodes (3): Entitlements, Subscription, UserMembership

### Community 33 - "NewPostgresRepository"
Cohesion: 0.29
Nodes (3): categorizeTopic(), TrendingTopic, NewPostgresRepository()

### Community 34 - "route.ts"
Cohesion: 0.60
Nodes (3): getDynamicAuthOptions(), handler(), authOptions

### Community 35 - "search/page.tsx"
Cohesion: 0.13
Nodes (18): ExploreCategory, SearchPageContent(), SearchPostsResponse, SearchTab, SearchUsersResponse, PostCardProps, LeftNavRailProps, RightSidebarProps (+10 more)

## Knowledge Gaps
- **127 isolated node(s):** `name`, `version`, `private`, `dev`, `build` (+122 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **116 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `MemoryRepository` connect `MemoryRepository` to `.GetPlanEntitlements`, `T`, `MockCandidateProvider`, `Context`, `.GetNotifications`, `.DismissUserRecommendation`, `Post`, `User`, `.GetAdminOverview`, `.GetTrendingTopics`, `.GetActiveSubscription`, `Role`, `PaymentOrder`?**
  _High betweenness centrality (0.153) - this node is a cross-community bridge._
- **Why does `NewMemoryRepository()` connect `T` to `.CheckoutSubscription`, `MemoryRepository`, `MockCandidateProvider`?**
  _High betweenness centrality (0.111) - this node is a cross-community bridge._
- **Why does `NewHandler()` connect `T` to `writeJSON`, `.CheckoutSubscription`, `handler.go`?**
  _High betweenness centrality (0.087) - this node is a cross-community bridge._
- **Are the 31 inferred relationships involving `writeJSON()` (e.g. with `.AdminAssignUserMembership()` and `.AdminCreateRole()`) actually correct?**
  _`writeJSON()` has 31 INFERRED edges - model-reasoned connections that need verification._
- **What connects `name`, `version`, `private` to the rest of the system?**
  _127 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `writeJSON` be split into smaller, more focused modules?**
  _Cohesion score 0.09343763581051717 - nodes in this community are weakly interconnected._
- **Should `T` be split into smaller, more focused modules?**
  _Cohesion score 0.06773211567732115 - nodes in this community are weakly interconnected._