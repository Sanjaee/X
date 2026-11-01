# Graph Report - .  (2026-11-02)

## Corpus Check
- cluster-only mode — file stats not available

## Summary
- 1249 nodes · 3279 edges · 143 communities (40 shown, 103 thin omitted)
- Extraction: 97% EXTRACTED · 3% INFERRED · 0% AMBIGUOUS · INFERRED: 106 edges (avg confidence: 0.8)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `d9d949ef`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- Request
- models.go
- T
- Context
- MockCandidateProvider
- Context
- MemoryRepository
- PostgresRepository
- dependencies
- [postId]/page.tsx
- Hub
- search/page.tsx
- CryptoGateway
- index.ts
- api.ts
- utils.ts
- compilerOptions
- Post
- Post
- UpgradeModal.tsx
- FeedSkeleton.tsx
- admin.go
- OtherSkeletons.tsx
- UserNameWithRole.tsx
- handler.go
- VideoControls.tsx
- .GetActiveSubscription
- Role
- PostComposer.tsx
- PaymentOrder
- .GetPlanEntitlements
- .GetBillingHistory
- NewPostgresRepository
- route.ts
- TopLoadingBar.tsx
- .GetActiveSubscription
- middleware.ts
- Duration
- .GetMessages
- .CreateOutboxEvent
- AuditLog
- .DismissUserRecommendation
- .GetVerificationApplication
- .GetMediaObjectByID
- .GetMessages
- .GetUserTopicPreferences
- .GetUserBadges
- .CreateSession
- .GetAdminPosts
- .GetSubscriptionPlanByID
- test_all_apis.py
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
- .GetAdminOverview

## God Nodes (most connected - your core abstractions)
1. `MemoryRepository` - 175 edges
2. `PostgresRepository` - 149 edges
3. `Handler` - 104 edges
4. `writeJSON()` - 78 edges
5. `writeError()` - 76 edges
6. `GetUserFromContext()` - 54 edges
7. `apiRequest()` - 34 edges
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

## Communities (143 total, 103 thin omitted)

### Community 0 - "Request"
Cohesion: 0.07
Nodes (20): CryptoInvoice, CryptoOperation, CheckoutRequest, Handler, Request, ResponseWriter, Context, ResponseWriter (+12 more)

### Community 1 - "models.go"
Cohesion: 0.06
Nodes (69): Config, AdminOverview, AdminPostFilter, AuditLog, Conversation, CounterMismatchDetail, CounterReconciliationResult, DatabaseHealth (+61 more)

### Community 2 - "T"
Cohesion: 0.06
Nodes (63): CryptoGateway, Hub, MediaService, Mux, Repository, main(), setupTestFullAdminRouter(), TestAdminPostModerationAndAuditLogging() (+55 more)

### Community 3 - "Context"
Cohesion: 0.05
Nodes (12): Context, Conversation, FollowRequest, MediaObject, Notification, RecentSearch, SocialRelationship, User (+4 more)

### Community 4 - "MockCandidateProvider"
Cohesion: 0.05
Nodes (43): B, DiversityReranker, RankingWeights, FeedOptions, FeedResult, Candidate, CandidateProvider, CandidateSource (+35 more)

### Community 5 - "Context"
Cohesion: 0.07
Nodes (6): Context, Conversation, Notification, SocialRelationship, UserRecommendation, UUID

### Community 6 - "MemoryRepository"
Cohesion: 0.07
Nodes (14): Permission, MemoryRepository, FeedFeedback, FeedImpression, FollowRequest, OutboxEvent, RankingWeights, RecentSearch (+6 more)

### Community 7 - "PostgresRepository"
Cohesion: 0.11
Nodes (12): DB, PostgresRepository, AdminOverview, AuditLog, Context, MembershipEntitlement, MembershipPlan, Permission (+4 more)

### Community 8 - "dependencies"
Cohesion: 0.05
Nodes (42): dependencies, clsx, lucide-react, next, next-auth, react, react-dom, tailwind-merge (+34 more)

### Community 9 - "[postId]/page.tsx"
Cohesion: 0.10
Nodes (25): StagedFile, PostModal(), PostModalProps, StagedAttachment, PostCard(), PostCardProps, MediaDetailSidebar(), MediaDetailSidebarProps (+17 more)

### Community 10 - "Hub"
Cohesion: 0.08
Nodes (24): Conn, Logger, contextKey, responseWriterWrapper, RawMessage, ReadWriter, Client, Hub (+16 more)

### Community 11 - "search/page.tsx"
Cohesion: 0.11
Nodes (25): encodeFeedCursor(), FeedResponse, HomePage(), ExploreCategory, SearchPageContent(), SearchPostsResponse, SearchTab, SearchUsersResponse (+17 more)

### Community 12 - "CryptoGateway"
Cohesion: 0.12
Nodes (18): Client, Config, memoryLimiter, Mutex, CryptoCurrency, CryptoGateway, CryptoInvoice, CryptoOperation (+10 more)

### Community 13 - "index.ts"
Cohesion: 0.10
Nodes (22): AdminTab, ConfirmModal(), ConfirmModalProps, TIER_PRESENTATION, TierBadge(), TierBadgeProps, TierKey, AdminOverview (+14 more)

### Community 14 - "api.ts"
Cohesion: 0.22
Nodes (14): AuthModal(), AuthModalProps, RightSidebarProps, EditProfileModal(), EditProfileModalProps, apiRequest(), getApiBase(), getStoredToken() (+6 more)

### Community 15 - "utils.ts"
Cohesion: 0.13
Nodes (16): ComposerContext, ComposerContextValue, ComposerProvider(), useComposer(), LeftNavRailProps, RouteTransition(), RealtimeContext, RealtimeContextType (+8 more)

### Community 16 - "compilerOptions"
Cohesion: 0.07
Nodes (26): compilerOptions, allowJs, esModuleInterop, incremental, isolatedModules, jsx, lib, module (+18 more)

### Community 17 - "Post"
Cohesion: 0.16
Nodes (4): FeedCursor, Post, Rows, SearchFilterParams

### Community 18 - "Post"
Cohesion: 0.16
Nodes (4): AdminPostFilter, FeedCursor, Post, SearchFilterParams

### Community 19 - "UpgradeModal.tsx"
Cohesion: 0.11
Nodes (16): metadata, AppShell(), MediaViewerProvider(), PhotoViewerModal(), AppProviders(), CryptoCurrencyOption, num(), ParsedRolePlan (+8 more)

### Community 20 - "FeedSkeleton.tsx"
Cohesion: 0.19
Nodes (4): FeedSkeleton(), PostCardSkeleton(), PostDetailSkeleton(), ProfileSkeleton()

### Community 21 - "admin.go"
Cohesion: 0.16
Nodes (13): AdminDeletePostRequest, AssignMembershipRequest, CreateRoleRequest, SetPermissionsRequest, SetPlanEntitlementsRequest, UpdatePlanRequest, UpdatePostModerationRequest, UpdateRoleRequest (+5 more)

### Community 22 - "OtherSkeletons.tsx"
Cohesion: 0.23
Nodes (3): ExploreSkeleton(), MessagesSkeleton(), NotificationsSkeleton()

### Community 23 - "UserNameWithRole.tsx"
Cohesion: 0.20
Nodes (8): RecentSearchItem, SearchBarDropdown(), SearchBarDropdownProps, SuggestionData, UserNameWithRoleProps, TrendingTopic, getRoleBadge(), getRoleNameClass()

### Community 24 - "handler.go"
Cohesion: 0.22
Nodes (10): CreatePostRequest, FeedFeedbackRequest, FeedImpressionsRequest, GoogleAuthRequest, LoginRequest, RegisterRequest, SimulatePaymentRequest, UpdatePostRequest (+2 more)

### Community 25 - "VideoControls.tsx"
Cohesion: 0.31
Nodes (7): formatTime(), SPEED_OPTIONS, VideoControls(), VideoControlsProps, formatTime(), VideoProgressBar(), VideoProgressBarProps

### Community 26 - ".GetActiveSubscription"
Cohesion: 0.22
Nodes (3): Entitlements, Subscription, UserMembership

### Community 28 - "PostComposer.tsx"
Cohesion: 0.43
Nodes (5): PostComposerProps, StagedAttachment, filterValidMediaFiles(), isValidMediaFile(), X_ALLOWED_MEDIA_EXTENSIONS

### Community 32 - "NewPostgresRepository"
Cohesion: 0.29
Nodes (3): categorizeTopic(), NewPostgresRepository(), TrendingTopic

### Community 33 - "route.ts"
Cohesion: 0.60
Nodes (3): getDynamicAuthOptions(), handler(), authOptions

## Knowledge Gaps
- **127 isolated node(s):** `name`, `version`, `private`, `dev`, `build` (+122 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **103 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `MemoryRepository` connect `MemoryRepository` to `T`, `Context`, `AuditLog`, `.DismissUserRecommendation`, `.GetVerificationApplication`, `.GetMediaObjectByID`, `.GetMessages`, `.GetAdminOverview`, `.GetUserTopicPreferences`, `.ReconcileCounters`, `Hub`, `Post`, `.GetActiveSubscription`, `Role`, `.GetPlanEntitlements`, `.GetBillingHistory`?**
  _High betweenness centrality (0.131) - this node is a cross-community bridge._
- **Why does `NewMemoryRepository()` connect `T` to `MemoryRepository`?**
  _High betweenness centrality (0.098) - this node is a cross-community bridge._
- **Why does `Handler` connect `Request` to `handler.go`, `T`, `Hub`?**
  _High betweenness centrality (0.086) - this node is a cross-community bridge._
- **What connects `name`, `version`, `private` to the rest of the system?**
  _127 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `Request` be split into smaller, more focused modules?**
  _Cohesion score 0.07298509018871858 - nodes in this community are weakly interconnected._
- **Should `models.go` be split into smaller, more focused modules?**
  _Cohesion score 0.06060606060606061 - nodes in this community are weakly interconnected._
- **Should `T` be split into smaller, more focused modules?**
  _Cohesion score 0.06486486486486487 - nodes in this community are weakly interconnected._