# Graph Report - .  (2026-11-02)

## Corpus Check
- cluster-only mode — file stats not available

## Summary
- 1248 nodes · 3273 edges · 146 communities (40 shown, 106 thin omitted)
- Extraction: 97% EXTRACTED · 3% INFERRED · 0% AMBIGUOUS · INFERRED: 106 edges (avg confidence: 0.8)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `a655b08e`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- Request
- T
- Context
- MockCandidateProvider
- models.go
- Context
- MemoryRepository
- PostgresRepository
- Hub
- dependencies
- [postId]/page.tsx
- search/page.tsx
- index.ts
- CryptoGateway
- UpgradeModal.tsx
- compilerOptions
- api.ts
- Post
- Post
- utils.ts
- MediaService
- FeedSkeleton.tsx
- admin.go
- OtherSkeletons.tsx
- handler.go
- VideoControls.tsx
- .GetActiveSubscription
- Role
- PaymentOrder
- .GetPlanEntitlements
- .GetBillingHistory
- NewPostgresRepository
- UserNameWithRole.tsx
- route.ts
- TopLoadingBar.tsx
- .GetActiveSubscription
- AuditLog
- middleware.ts
- Duration
- .GetMessages
- .CreateOutboxEvent
- .DismissUserRecommendation
- .GetMediaObjectByID
- .GetMessages
- .GetVerificationApplication
- .GetUserTopicPreferences
- .GetUserBadges
- .CreateSession
- .GetAdminPosts
- .GetSubscriptionPlanByID
- test_all_apis.py
- .GetAdminOverview
- next.config.js
- next-env.d.ts
- tailwind.config.ts
- .ReconcileCounters
- .RecordFeedFeedback
- .RecordFeedImpressions
- .GetActiveWeights
- .ReconcileCounters
- backup-db.sh
- monitoring.sh
- CryptoInvoice
- CryptoOperation
- github.com/x-reconstruction/api
- PlisioService
- Repository
- MembershipEntitlement
- Message
- Request
- ResponseWriter
- MediaObject
- PaymentOrder
- Request
- Subscription
- PaymentOrder
- Request
- ResponseWriter
- Mux
- T
- MediaObject
- Handler
- Conn
- User
- Client
- Context
- Time
- UUID
- Client
- Context
- RawMessage
- UUID
- Conn
- RawMessage
- RWMutex
- AdminOverview
- AdminPostFilter
- AuditLog
- MembershipEntitlement
- MembershipPlan
- Permission
- Role
- RWMutex
- UserMembership
- Context
- Conversation
- CounterReconciliationResult
- Duration
- Entitlements
- FeedCursor
- FeedFeedback
- FeedImpression
- FollowRequest
- Invoice
- MediaObject
- Message
- Notification
- OutboxEvent
- PaymentOrder
- Post
- RankingWeights
- RecentSearch
- SearchFilterParams
- Session
- SocialRelationship
- Subscription
- SubscriptionPlan
- Time
- TrendingTopic
- User
- UserRecommendation
- UserTopicPreference
- UUID
- VerificationApplication
- VerificationBadge
- SystemHealth

## God Nodes (most connected - your core abstractions)
1. `MemoryRepository` - 175 edges
2. `PostgresRepository` - 149 edges
3. `Handler` - 104 edges
4. `writeJSON()` - 78 edges
5. `writeError()` - 76 edges
6. `GetUserFromContext()` - 54 edges
7. `apiRequest()` - 33 edges
8. `User` - 27 edges
9. `cn()` - 22 edges
10. `GetViewerID()` - 22 edges

## Surprising Connections (you probably didn't know these)
- `setupTestRouter()` --calls--> `AuthMiddleware()`  [INFERRED]
  services/api/internal/handler/handler_test.go → services/api/internal/middleware/middleware.go
- `setupTestApp()` --calls--> `AuthMiddleware()`  [INFERRED]
  services/api/internal/handler/social_messaging_test.go → services/api/internal/middleware/middleware.go
- `setupTestApp()` --calls--> `NewHub()`  [INFERRED]
  services/api/internal/handler/social_messaging_test.go → services/api/internal/realtime/hub.go
- `RateLimiter()` --calls--> `GetUserFromContext()`  [INFERRED]
  services/api/internal/middleware/ratelimit.go → services/api/internal/middleware/middleware.go
- `main()` --calls--> `Handler`  [INFERRED]
  services/api/cmd/server/main.go → services/api/internal/handler/handler.go

## Import Cycles
- None detected.

## Communities (146 total, 106 thin omitted)

### Community 0 - "Request"
Cohesion: 0.09
Nodes (11): Handler, Request, ResponseWriter, Context, ResponseWriter, writeError(), writeJSON(), GetUserFromContext() (+3 more)

### Community 1 - "T"
Cohesion: 0.06
Nodes (63): CryptoGateway, Hub, MediaService, Mux, Repository, main(), setupTestFullAdminRouter(), TestAdminPostModerationAndAuditLogging() (+55 more)

### Community 2 - "Context"
Cohesion: 0.05
Nodes (12): Context, Conversation, FollowRequest, MediaObject, Notification, RecentSearch, SocialRelationship, User (+4 more)

### Community 3 - "MockCandidateProvider"
Cohesion: 0.05
Nodes (43): B, DiversityReranker, RankingWeights, FeedOptions, FeedResult, Candidate, CandidateProvider, CandidateSource (+35 more)

### Community 4 - "models.go"
Cohesion: 0.09
Nodes (57): AdminOverview, AdminPostFilter, AuditLog, Conversation, CounterMismatchDetail, CounterReconciliationResult, DatabaseHealth, Entitlements (+49 more)

### Community 5 - "Context"
Cohesion: 0.07
Nodes (6): Context, Conversation, Notification, SocialRelationship, UserRecommendation, UUID

### Community 6 - "MemoryRepository"
Cohesion: 0.07
Nodes (14): Permission, MemoryRepository, FeedFeedback, FeedImpression, FollowRequest, OutboxEvent, RankingWeights, RecentSearch (+6 more)

### Community 7 - "PostgresRepository"
Cohesion: 0.11
Nodes (12): DB, PostgresRepository, AdminOverview, AuditLog, Context, MembershipEntitlement, MembershipPlan, Permission (+4 more)

### Community 8 - "Hub"
Cohesion: 0.07
Nodes (28): Conn, Logger, contextKey, memoryLimiter, responseWriterWrapper, Mutex, RawMessage, ReadWriter (+20 more)

### Community 9 - "dependencies"
Cohesion: 0.05
Nodes (42): dependencies, clsx, lucide-react, next, next-auth, react, react-dom, tailwind-merge (+34 more)

### Community 10 - "[postId]/page.tsx"
Cohesion: 0.10
Nodes (24): StagedFile, PostModal(), PostModalProps, StagedAttachment, PostCard(), PostCardProps, MediaDetailSidebar(), MediaDetailSidebarProps (+16 more)

### Community 11 - "search/page.tsx"
Cohesion: 0.12
Nodes (22): encodeFeedCursor(), FeedResponse, HomePage(), ExploreCategory, SearchPageContent(), SearchPostsResponse, SearchTab, SearchUsersResponse (+14 more)

### Community 12 - "index.ts"
Cohesion: 0.07
Nodes (31): AdminTab, ConfirmModal(), ConfirmModalProps, TIER_PRESENTATION, TierBadge(), TierBadgeProps, TierKey, RealtimeContext (+23 more)

### Community 13 - "CryptoGateway"
Cohesion: 0.11
Nodes (19): Client, Config, CheckoutRequest, CryptoCurrency, CryptoGateway, CryptoInvoice, CryptoOperation, WebhookPayload (+11 more)

### Community 14 - "UpgradeModal.tsx"
Cohesion: 0.09
Nodes (24): metadata, VerifiedBadge(), VerifiedBadgeProps, PostComposer(), PostComposerProps, StagedAttachment, AppShell(), MediaViewerProvider() (+16 more)

### Community 15 - "compilerOptions"
Cohesion: 0.07
Nodes (26): compilerOptions, allowJs, esModuleInterop, incremental, isolatedModules, jsx, lib, module (+18 more)

### Community 16 - "api.ts"
Cohesion: 0.17
Nodes (19): AuthModalProps, ComposerContext, ComposerContextValue, ComposerProvider(), RightSidebarProps, EditProfileModal(), EditProfileModalProps, RecentSearchItem (+11 more)

### Community 17 - "Post"
Cohesion: 0.16
Nodes (4): FeedCursor, Post, Rows, SearchFilterParams

### Community 18 - "Post"
Cohesion: 0.16
Nodes (4): AdminPostFilter, FeedCursor, Post, SearchFilterParams

### Community 19 - "utils.ts"
Cohesion: 0.25
Nodes (8): useComposer(), LeftNavRail(), LeftNavRailProps, RouteTransition(), UserNameWithRole(), useCurrentUser(), cn(), formatFullTimestamp()

### Community 20 - "MediaService"
Cohesion: 0.18
Nodes (12): Config, Image, MediaService, Reader, getEnv(), getEnvInt(), Load(), Client (+4 more)

### Community 21 - "FeedSkeleton.tsx"
Cohesion: 0.19
Nodes (4): FeedSkeleton(), PostCardSkeleton(), PostDetailSkeleton(), ProfileSkeleton()

### Community 22 - "admin.go"
Cohesion: 0.16
Nodes (13): AdminDeletePostRequest, AssignMembershipRequest, CreateRoleRequest, SetPermissionsRequest, SetPlanEntitlementsRequest, UpdatePlanRequest, UpdatePostModerationRequest, UpdateRoleRequest (+5 more)

### Community 23 - "OtherSkeletons.tsx"
Cohesion: 0.23
Nodes (3): ExploreSkeleton(), MessagesSkeleton(), NotificationsSkeleton()

### Community 24 - "handler.go"
Cohesion: 0.22
Nodes (10): CreatePostRequest, FeedFeedbackRequest, FeedImpressionsRequest, GoogleAuthRequest, LoginRequest, RegisterRequest, SimulatePaymentRequest, UpdatePostRequest (+2 more)

### Community 25 - "VideoControls.tsx"
Cohesion: 0.31
Nodes (7): formatTime(), SPEED_OPTIONS, VideoControls(), VideoControlsProps, formatTime(), VideoProgressBar(), VideoProgressBarProps

### Community 26 - ".GetActiveSubscription"
Cohesion: 0.22
Nodes (3): Entitlements, Subscription, UserMembership

### Community 31 - "NewPostgresRepository"
Cohesion: 0.29
Nodes (3): categorizeTopic(), NewPostgresRepository(), TrendingTopic

### Community 32 - "UserNameWithRole.tsx"
Cohesion: 0.47
Nodes (3): UserNameWithRoleProps, getRoleBadge(), getRoleNameClass()

### Community 33 - "route.ts"
Cohesion: 0.60
Nodes (3): getDynamicAuthOptions(), handler(), authOptions

## Knowledge Gaps
- **127 isolated node(s):** `name`, `version`, `private`, `dev`, `build` (+122 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **106 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `MemoryRepository` connect `MemoryRepository` to `T`, `Context`, `AuditLog`, `Hub`, `.DismissUserRecommendation`, `.GetMediaObjectByID`, `.GetMessages`, `.GetVerificationApplication`, `.GetUserTopicPreferences`, `Post`, `.GetAdminOverview`, `.GetActiveSubscription`, `Role`, `.ReconcileCounters`, `.GetPlanEntitlements`, `.GetBillingHistory`?**
  _High betweenness centrality (0.136) - this node is a cross-community bridge._
- **Why does `NewMemoryRepository()` connect `T` to `MemoryRepository`?**
  _High betweenness centrality (0.097) - this node is a cross-community bridge._
- **Why does `Handler` connect `Request` to `handler.go`, `T`, `CryptoGateway`, `Hub`?**
  _High betweenness centrality (0.086) - this node is a cross-community bridge._
- **What connects `name`, `version`, `private` to the rest of the system?**
  _127 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `Request` be split into smaller, more focused modules?**
  _Cohesion score 0.08870523415977961 - nodes in this community are weakly interconnected._
- **Should `T` be split into smaller, more focused modules?**
  _Cohesion score 0.06486486486486487 - nodes in this community are weakly interconnected._
- **Should `Context` be split into smaller, more focused modules?**
  _Cohesion score 0.053313587560162905 - nodes in this community are weakly interconnected._