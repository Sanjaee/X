# Graph Report - .  (2026-10-24)

## Corpus Check
- cluster-only mode — file stats not available

## Summary
- 1196 nodes · 3153 edges · 131 communities (38 shown, 93 thin omitted)
- Extraction: 99% EXTRACTED · 1% INFERRED · 0% AMBIGUOUS · INFERRED: 40 edges (avg confidence: 0.8)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `feb831a4`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- Handler
- T
- Context
- models.go
- MockCandidateProvider
- Context
- MemoryRepository
- Hub
- PostgresRepository
- apiRequest
- dependencies
- [postId]/page.tsx
- PlisioService
- index.ts
- compilerOptions
- Post
- Post
- utils.ts
- User
- FeedSkeleton.tsx
- PostModal.tsx
- admin.go
- handler.go
- OtherSkeletons.tsx
- search/page.tsx
- VideoControls.tsx
- Engine
- .GetActiveSubscription
- .GetPlanEntitlements
- Role
- NewPostgresRepository
- .GetBillingHistory
- .GetBillingHistory
- TopLoadingBar.tsx
- .GetActiveSubscription
- route.ts
- AuditLog
- .GetNotifications
- middleware.ts
- .GetAdminPosts
- .DismissUserRecommendation
- .GetFollowRequests
- .GetMediaObjectByID
- .GetUserBadges
- .DismissUserRecommendation
- .GetMessages
- .CreateOutboxEvent
- .CreateSession
- .GetVerificationApplication
- .GetSubscriptionPlanByID
- .GetUserTopicPreferences
- test_all_apis.py
- .GetAdminOverview
- next.config.js
- next-env.d.ts
- tailwind.config.ts
- .GetRecentSearches
- .ReconcileCounters
- .GetActiveWeights
- .ReconcileCounters
- .RecordFeedFeedback
- .RecordFeedImpressions
- backup-db.sh
- monitoring.sh
- Config
- Context
- Conversation
- CounterReconciliationResult
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
- github.com/x-reconstruction/api
- Post
- RankingWeights
- RecentSearch
- Repository
- SearchFilterParams
- MembershipEntitlement
- Message
- Request
- ResponseWriter
- Client
- Request
- ResponseWriter
- Mux
- T
- MediaObject
- Handler
- Conn
- User
- Conn
- RawMessage
- RWMutex
- AdminOverview
- AuditLog
- MembershipEntitlement
- MembershipPlan
- Permission
- Role
- RWMutex
- UserMembership
- Session
- SocialRelationship
- Subscription
- SubscriptionPlan
- SystemHealth
- Time
- TrendingTopic
- User
- UserRecommendation
- UserTopicPreference
- UUID
- VerificationApplication
- VerificationBadge

## God Nodes (most connected - your core abstractions)
1. `MemoryRepository` - 172 edges
2. `PostgresRepository` - 146 edges
3. `Handler` - 129 edges
4. `writeJSON()` - 82 edges
5. `writeError()` - 79 edges
6. `apiRequest()` - 39 edges
7. `User` - 30 edges
8. `useCurrentUser()` - 27 edges
9. `cn()` - 23 edges
10. `Post` - 18 edges

## Surprising Connections (you probably didn't know these)
- `setupTestRouter()` --calls--> `AuthMiddleware()`  [INFERRED]
  services/api/internal/handler/handler_test.go → services/api/internal/middleware/middleware.go
- `setupTestApp()` --calls--> `AuthMiddleware()`  [INFERRED]
  services/api/internal/handler/social_messaging_test.go → services/api/internal/middleware/middleware.go
- `setupTestApp()` --calls--> `NewHub()`  [INFERRED]
  services/api/internal/handler/social_messaging_test.go → services/api/internal/realtime/hub.go
- `BenchmarkScoringEngine()` --calls--> `DefaultRankingWeights()`  [INFERRED]
  services/api/internal/recommendation/engine_test.go → services/api/internal/domain/models.go
- `main()` --calls--> `Handler`  [INFERRED]
  services/api/cmd/server/main.go → services/api/internal/handler/handler.go

## Import Cycles
- None detected.

## Communities (131 total, 93 thin omitted)

### Community 0 - "Handler"
Cohesion: 0.08
Nodes (8): Handler, Request, ResponseWriter, Context, writeError(), writeJSON(), ConcurrencyLimiter(), RequestTimeoutMiddleware()

### Community 1 - "T"
Cohesion: 0.07
Nodes (62): Client, Hub, MediaService, Mux, PlisioService, Repository, main(), setupTestFullAdminRouter() (+54 more)

### Community 2 - "Context"
Cohesion: 0.06
Nodes (11): Context, Conversation, FollowRequest, MediaObject, Notification, RecentSearch, SocialRelationship, Time (+3 more)

### Community 3 - "models.go"
Cohesion: 0.08
Nodes (58): AdminOverview, AuditLog, Conversation, CounterMismatchDetail, CounterReconciliationResult, DatabaseHealth, Entitlements, FeedFeedback (+50 more)

### Community 4 - "MockCandidateProvider"
Cohesion: 0.06
Nodes (34): B, Candidate, CandidateSource, DiversityReranker, Features, FeedCursor, FeedMetrics, FeedOptions (+26 more)

### Community 5 - "Context"
Cohesion: 0.07
Nodes (5): Context, Conversation, SocialRelationship, UserRecommendation, UUID

### Community 6 - "MemoryRepository"
Cohesion: 0.07
Nodes (15): Permission, MemoryRepository, FeedFeedback, FeedImpression, Message, OutboxEvent, RankingWeights, Session (+7 more)

### Community 7 - "Hub"
Cohesion: 0.06
Nodes (32): Conn, Duration, Logger, contextKey, memoryLimiter, responseWriterWrapper, Mutex, RawMessage (+24 more)

### Community 8 - "PostgresRepository"
Cohesion: 0.11
Nodes (12): DB, PostgresRepository, AdminOverview, AuditLog, Context, MembershipEntitlement, MembershipPlan, Permission (+4 more)

### Community 9 - "apiRequest"
Cohesion: 0.16
Nodes (27): ExplorePage(), encodeFeedCursor(), FeedResponse, HomePage(), PremiumPage(), SettingsSection, FollowersPage(), FollowingPage() (+19 more)

### Community 10 - "dependencies"
Cohesion: 0.05
Nodes (42): dependencies, clsx, lucide-react, next, next-auth, react, react-dom, tailwind-merge (+34 more)

### Community 11 - "[postId]/page.tsx"
Cohesion: 0.10
Nodes (24): StagedFile, ConfirmModal(), ConfirmModalProps, PostCard(), PostCardProps, MediaDetailSidebar(), MediaDetailSidebarProps, MediaViewerContext (+16 more)

### Community 12 - "PlisioService"
Cohesion: 0.09
Nodes (24): Config, Image, MediaService, PlisioCryptoInfo, PlisioInvoiceResponse, PlisioOperationsResponse, PlisioService, PlisioWebhookPayload (+16 more)

### Community 13 - "index.ts"
Cohesion: 0.09
Nodes (26): AdminTab, TIER_PRESENTATION, TierBadge(), TierBadgeProps, TierKey, RealtimeContext, RealtimeContextType, ToastNotification (+18 more)

### Community 14 - "compilerOptions"
Cohesion: 0.07
Nodes (26): compilerOptions, allowJs, esModuleInterop, incremental, isolatedModules, jsx, lib, module (+18 more)

### Community 15 - "Post"
Cohesion: 0.16
Nodes (4): Rows, FeedCursor, Post, SearchFilterParams

### Community 16 - "Post"
Cohesion: 0.16
Nodes (4): AdminPostFilter, FeedCursor, Post, SearchFilterParams

### Community 17 - "utils.ts"
Cohesion: 0.14
Nodes (14): metadata, ComposerContext, ComposerContextValue, ComposerProvider(), useComposer(), PostModal(), AppShell(), LeftNavRailProps (+6 more)

### Community 18 - "User"
Cohesion: 0.14
Nodes (12): WhiteLabelInvoice, VerifiedBadge(), VerifiedBadgeProps, EditProfileModal(), EditProfileModalProps, RecentSearchItem, SearchBarDropdownProps, SuggestionData (+4 more)

### Community 19 - "FeedSkeleton.tsx"
Cohesion: 0.19
Nodes (4): FeedSkeleton(), PostCardSkeleton(), PostDetailSkeleton(), ProfileSkeleton()

### Community 20 - "PostModal.tsx"
Cohesion: 0.21
Nodes (11): MobileComposerModalProps, StagedAttachment, PostComposer(), PostComposerProps, StagedAttachment, PostModalProps, StagedAttachment, filterValidMediaFiles() (+3 more)

### Community 21 - "admin.go"
Cohesion: 0.16
Nodes (13): AdminDeletePostRequest, AssignMembershipRequest, CreateRoleRequest, SetPermissionsRequest, SetPlanEntitlementsRequest, UpdatePlanRequest, UpdatePostModerationRequest, UpdateRoleRequest (+5 more)

### Community 22 - "handler.go"
Cohesion: 0.18
Nodes (12): CheckoutRequest, CreatePostRequest, FeedFeedbackRequest, FeedImpressionsRequest, GoogleAuthRequest, LoginRequest, RegisterRequest, SimulatePaymentRequest (+4 more)

### Community 23 - "OtherSkeletons.tsx"
Cohesion: 0.23
Nodes (3): ExploreSkeleton(), MessagesSkeleton(), NotificationsSkeleton()

### Community 24 - "search/page.tsx"
Cohesion: 0.22
Nodes (8): ExploreCategory, SearchPageContent(), SearchPostsResponse, SearchTab, SearchUsersResponse, SearchBarDropdown(), feedKeys, searchKeys

### Community 25 - "VideoControls.tsx"
Cohesion: 0.31
Nodes (7): formatTime(), SPEED_OPTIONS, VideoControls(), VideoControlsProps, formatTime(), VideoProgressBar(), VideoProgressBarProps

### Community 26 - "Engine"
Cohesion: 0.31
Nodes (7): DiversityReranker, FeedOptions, FeedResult, CandidateProvider, Engine, Context, NewEngine()

### Community 27 - ".GetActiveSubscription"
Cohesion: 0.22
Nodes (3): Entitlements, Subscription, UserMembership

### Community 30 - "NewPostgresRepository"
Cohesion: 0.29
Nodes (3): categorizeTopic(), TrendingTopic, NewPostgresRepository()

## Knowledge Gaps
- **118 isolated node(s):** `name`, `version`, `private`, `dev`, `build` (+113 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **93 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `PostgresRepository` connect `PostgresRepository` to `.GetBillingHistory`, `Context`, `.GetActiveSubscription`, `.GetAdminPosts`, `.GetUserBadges`, `.DismissUserRecommendation`, `.GetMessages`, `.CreateOutboxEvent`, `Post`, `.CreateSession`, `.GetVerificationApplication`, `.GetSubscriptionPlanByID`, `.GetUserTopicPreferences`, `.RecordFeedImpressions`, `.GetActiveWeights`, `.ReconcileCounters`, `.RecordFeedFeedback`, `NewPostgresRepository`?**
  _High betweenness centrality (0.191) - this node is a cross-community bridge._
- **Why does `Handler` connect `Handler` to `T`, `handler.go`, `Hub`?**
  _High betweenness centrality (0.151) - this node is a cross-community bridge._
- **Why does `main()` connect `T` to `Handler`, `NewPostgresRepository`?**
  _High betweenness centrality (0.123) - this node is a cross-community bridge._
- **What connects `name`, `version`, `private` to the rest of the system?**
  _118 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `Handler` be split into smaller, more focused modules?**
  _Cohesion score 0.08484848484848485 - nodes in this community are weakly interconnected._
- **Should `T` be split into smaller, more focused modules?**
  _Cohesion score 0.0670122176971492 - nodes in this community are weakly interconnected._
- **Should `Context` be split into smaller, more focused modules?**
  _Cohesion score 0.05924920850293985 - nodes in this community are weakly interconnected._