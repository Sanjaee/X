# Graph Report - .  (2026-11-01)

## Corpus Check
- cluster-only mode — file stats not available

## Summary
- 1249 nodes · 3278 edges · 141 communities (41 shown, 100 thin omitted)
- Extraction: 97% EXTRACTED · 3% INFERRED · 0% AMBIGUOUS · INFERRED: 106 edges (avg confidence: 0.8)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `fada0f55`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- Request
- Context
- T
- models.go
- Context
- MockCandidateProvider
- MemoryRepository
- [postId]/page.tsx
- PostgresRepository
- search/page.tsx
- Hub
- dependencies
- CryptoGateway
- index.ts
- api.ts
- UpgradeModal.tsx
- compilerOptions
- Post
- Post
- PostComposer.tsx
- FeedSkeleton.tsx
- admin.go
- OtherSkeletons.tsx
- UserNameWithRole.tsx
- handler.go
- VideoControls.tsx
- Engine
- .GetActiveSubscription
- PaymentOrder
- .GetPlanEntitlements
- .GetBillingHistory
- Role
- NewPostgresRepository
- route.ts
- TopLoadingBar.tsx
- .GetActiveSubscription
- utils.ts
- middleware.ts
- Duration
- .CreateOutboxEvent
- .GetRolePermissions
- .DismissUserRecommendation
- .GetFollowRequests
- .GetMediaObjectByID
- .GetMessages
- .GetUserTopicPreferences
- .GetUserBadges
- .CreateSession
- .GetAdminPosts
- .GetSubscriptionPlanByID
- .GetUserTopicPreferences
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
- Client
- Config
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
- `BenchmarkScoringEngine()` --calls--> `DefaultRankingWeights()`  [INFERRED]
  services/api/internal/recommendation/engine_test.go → services/api/internal/domain/models.go

## Import Cycles
- None detected.

## Communities (141 total, 100 thin omitted)

### Community 0 - "Request"
Cohesion: 0.07
Nodes (20): CryptoInvoice, CryptoOperation, CheckoutRequest, Handler, Request, ResponseWriter, Context, ResponseWriter (+12 more)

### Community 1 - "Context"
Cohesion: 0.05
Nodes (13): Context, Conversation, FollowRequest, MediaObject, Message, Notification, RecentSearch, SocialRelationship (+5 more)

### Community 2 - "T"
Cohesion: 0.06
Nodes (63): CryptoGateway, Hub, MediaService, Mux, Repository, main(), setupTestFullAdminRouter(), TestAdminPostModerationAndAuditLogging() (+55 more)

### Community 3 - "models.go"
Cohesion: 0.08
Nodes (59): AdminOverview, AdminPostFilter, AuditLog, Conversation, CounterMismatchDetail, CounterReconciliationResult, DatabaseHealth, Entitlements (+51 more)

### Community 4 - "Context"
Cohesion: 0.07
Nodes (6): Context, Conversation, Notification, SocialRelationship, UserRecommendation, UUID

### Community 5 - "MockCandidateProvider"
Cohesion: 0.06
Nodes (34): B, Candidate, CandidateSource, DiversityReranker, Features, FeedCursor, FeedMetrics, FeedOptions (+26 more)

### Community 6 - "MemoryRepository"
Cohesion: 0.07
Nodes (15): AdminOverview, AuditLog, MemoryRepository, FeedFeedback, FeedImpression, OutboxEvent, RankingWeights, RecentSearch (+7 more)

### Community 7 - "[postId]/page.tsx"
Cohesion: 0.10
Nodes (25): StagedFile, PostModal(), PostModalProps, StagedAttachment, PostCard(), PostCardProps, MediaDetailSidebar(), MediaDetailSidebarProps (+17 more)

### Community 8 - "PostgresRepository"
Cohesion: 0.11
Nodes (12): DB, PostgresRepository, AdminOverview, AuditLog, Context, MembershipEntitlement, MembershipPlan, Permission (+4 more)

### Community 9 - "search/page.tsx"
Cohesion: 0.11
Nodes (25): encodeFeedCursor(), FeedResponse, HomePage(), ExploreCategory, SearchPageContent(), SearchPostsResponse, SearchTab, SearchUsersResponse (+17 more)

### Community 10 - "Hub"
Cohesion: 0.07
Nodes (28): Conn, Logger, contextKey, memoryLimiter, responseWriterWrapper, Mutex, RawMessage, ReadWriter (+20 more)

### Community 11 - "dependencies"
Cohesion: 0.05
Nodes (42): dependencies, clsx, lucide-react, next, next-auth, react, react-dom, tailwind-merge (+34 more)

### Community 12 - "CryptoGateway"
Cohesion: 0.09
Nodes (27): Config, Image, MediaService, CryptoCurrency, CryptoGateway, CryptoInvoice, CryptoOperation, WebhookPayload (+19 more)

### Community 13 - "index.ts"
Cohesion: 0.10
Nodes (22): AdminTab, ConfirmModal(), ConfirmModalProps, TIER_PRESENTATION, TierBadge(), TierBadgeProps, TierKey, AdminOverview (+14 more)

### Community 14 - "api.ts"
Cohesion: 0.22
Nodes (14): AuthModal(), AuthModalProps, RightSidebarProps, EditProfileModal(), EditProfileModalProps, apiRequest(), getApiBase(), getStoredToken() (+6 more)

### Community 15 - "UpgradeModal.tsx"
Cohesion: 0.11
Nodes (16): metadata, AppShell(), MediaViewerProvider(), PhotoViewerModal(), AppProviders(), CryptoCurrencyOption, num(), ParsedRolePlan (+8 more)

### Community 16 - "compilerOptions"
Cohesion: 0.07
Nodes (26): compilerOptions, allowJs, esModuleInterop, incremental, isolatedModules, jsx, lib, module (+18 more)

### Community 17 - "Post"
Cohesion: 0.16
Nodes (4): FeedCursor, Post, Rows, SearchFilterParams

### Community 18 - "Post"
Cohesion: 0.16
Nodes (4): AdminPostFilter, FeedCursor, Post, SearchFilterParams

### Community 19 - "PostComposer.tsx"
Cohesion: 0.43
Nodes (5): PostComposerProps, StagedAttachment, filterValidMediaFiles(), isValidMediaFile(), X_ALLOWED_MEDIA_EXTENSIONS

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

### Community 26 - "Engine"
Cohesion: 0.31
Nodes (7): DiversityReranker, FeedOptions, FeedResult, CandidateProvider, Engine, Context, NewEngine()

### Community 27 - ".GetActiveSubscription"
Cohesion: 0.22
Nodes (3): Entitlements, Subscription, UserMembership

### Community 32 - "NewPostgresRepository"
Cohesion: 0.29
Nodes (3): categorizeTopic(), NewPostgresRepository(), TrendingTopic

### Community 33 - "route.ts"
Cohesion: 0.60
Nodes (3): getDynamicAuthOptions(), handler(), authOptions

### Community 37 - "utils.ts"
Cohesion: 0.13
Nodes (16): ComposerContext, ComposerContextValue, ComposerProvider(), useComposer(), LeftNavRailProps, RouteTransition(), RealtimeContext, RealtimeContextType (+8 more)

## Knowledge Gaps
- **127 isolated node(s):** `name`, `version`, `private`, `dev`, `build` (+122 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **100 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `MemoryRepository` connect `MemoryRepository` to `T`, `Context`, `.GetRolePermissions`, `.DismissUserRecommendation`, `.GetFollowRequests`, `.GetMediaObjectByID`, `.GetMessages`, `.GetUserTopicPreferences`, `Hub`, `Post`, `.ReconcileCounters`, `.GetActiveSubscription`, `.GetPlanEntitlements`, `.GetBillingHistory`, `Role`?**
  _High betweenness centrality (0.158) - this node is a cross-community bridge._
- **Why does `CryptoGateway` connect `CryptoGateway` to `Duration`, `Hub`?**
  _High betweenness centrality (0.149) - this node is a cross-community bridge._
- **Why does `PostgresRepository` connect `PostgresRepository` to `NewPostgresRepository`, `Context`, `.GetActiveSubscription`, `Duration`, `.CreateOutboxEvent`, `.GetActiveWeights`, `.GetUserBadges`, `Post`, `.CreateSession`, `.GetAdminPosts`, `.GetSubscriptionPlanByID`, `.GetUserTopicPreferences`, `.ReconcileCounters`, `.RecordFeedFeedback`, `.RecordFeedImpressions`, `PaymentOrder`?**
  _High betweenness centrality (0.142) - this node is a cross-community bridge._
- **What connects `name`, `version`, `private` to the rest of the system?**
  _127 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `Request` be split into smaller, more focused modules?**
  _Cohesion score 0.07298509018871858 - nodes in this community are weakly interconnected._
- **Should `Context` be split into smaller, more focused modules?**
  _Cohesion score 0.05225225225225225 - nodes in this community are weakly interconnected._
- **Should `T` be split into smaller, more focused modules?**
  _Cohesion score 0.06486486486486487 - nodes in this community are weakly interconnected._