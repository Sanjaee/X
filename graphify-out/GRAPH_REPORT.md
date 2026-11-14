# Graph Report - .  (2026-11-14)

## Corpus Check
- cluster-only mode — file stats not available

## Summary
- 1312 nodes · 3470 edges · 158 communities (48 shown, 110 thin omitted)
- Extraction: 97% EXTRACTED · 3% INFERRED · 0% AMBIGUOUS · INFERRED: 117 edges (avg confidence: 0.8)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `22baf4d2`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- writeJSON
- T
- models.go
- Context
- CryptoGateway
- Context
- Context
- .CheckoutSubscription
- dependencies
- [postId]/page.tsx
- index.ts
- MemoryRepository
- api.ts
- PostgresRepository
- notifications/page.tsx
- compilerOptions
- app/page.tsx
- Hub
- Post
- search/page.tsx
- UserNameWithRole.tsx
- Post
- MediaService
- User
- FeedSkeleton.tsx
- engine_test.go
- admin.go
- MockCandidateProvider
- OtherSkeletons.tsx
- qrcode.tsx
- User
- handler.go
- types.go
- VideoControls.tsx
- Engine
- .GetActiveSubscription
- LeftNavRail.tsx
- RankingWeights
- Role
- .GetPlanEntitlements
- .DismissUserRecommendation
- EncodeCursor
- route.ts
- TopLoadingBar.tsx
- EvaluateFeed
- .GetActiveSubscription
- .GetNotifications
- .GetNotifications
- middleware.ts
- .GetConversationByID
- .GetFollowRequests
- .GetUserBadges
- test_all_apis.py
- next.config.js
- next-env.d.ts
- tailwind.config.ts
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
7. `apiRequest()` - 34 edges
8. `User` - 33 edges
9. `Handler` - 26 edges
10. `cn()` - 23 edges

## Surprising Connections (you probably didn't know these)
- `setupTestRouter()` --calls--> `AuthMiddleware()`  [INFERRED]
  services/api/internal/handler/handler_test.go → services/api/internal/middleware/middleware.go
- `setupTestApp()` --calls--> `AuthMiddleware()`  [INFERRED]
  services/api/internal/handler/social_messaging_test.go → services/api/internal/middleware/middleware.go
- `setupTestApp()` --calls--> `NewHub()`  [INFERRED]
  services/api/internal/handler/social_messaging_test.go → services/api/internal/realtime/hub.go
- `RateLimiter()` --calls--> `GetUserFromContext()`  [INFERRED]
  services/api/internal/middleware/ratelimit.go → services/api/internal/middleware/middleware.go
- `TestFeedEvaluator()` --calls--> `EvaluateFeed()`  [INFERRED]
  services/api/internal/recommendation/engine_test.go → services/api/internal/recommendation/evaluator.go

## Import Cycles
- None detected.

## Communities (158 total, 110 thin omitted)

### Community 0 - "writeJSON"
Cohesion: 0.09
Nodes (10): Context, PaymentOrder, Handler, Request, ResponseWriter, Handler, Request, ResponseWriter (+2 more)

### Community 1 - "T"
Cohesion: 0.07
Nodes (62): Client, CryptoGateway, Hub, MediaService, Mux, Repository, Repository, setupTestFullAdminRouter() (+54 more)

### Community 2 - "models.go"
Cohesion: 0.08
Nodes (58): AdminOverview, AuditLog, Conversation, CounterMismatchDetail, CounterReconciliationResult, DatabaseHealth, Entitlements, FeedFeedback (+50 more)

### Community 3 - "Context"
Cohesion: 0.06
Nodes (5): PaymentOrder, Context, SocialRelationship, UserRecommendation, UUID

### Community 4 - "CryptoGateway"
Cohesion: 0.07
Nodes (37): Config, Duration, Handler, mfaTokenPayload, memoryLimiter, Mutex, CryptoCurrency, CryptoGateway (+29 more)

### Community 5 - "Context"
Cohesion: 0.07
Nodes (5): Context, SocialRelationship, Time, UserRecommendation, UUID

### Community 6 - "Context"
Cohesion: 0.10
Nodes (13): AdminPostFilter, AdminOverview, AuditLog, Context, MembershipEntitlement, MembershipPlan, Permission, Post (+5 more)

### Community 7 - ".CheckoutSubscription"
Cohesion: 0.10
Nodes (29): CryptoInvoice, CryptoOperation, Handler, CheckoutRequest, Logger, contextKey, responseWriterWrapper, ReadWriter (+21 more)

### Community 8 - "dependencies"
Cohesion: 0.05
Nodes (42): dependencies, clsx, lucide-react, next, next-auth, react, react-dom, tailwind-merge (+34 more)

### Community 9 - "[postId]/page.tsx"
Cohesion: 0.12
Nodes (21): StagedFile, PostCard(), PostCardProps, MediaDetailSidebar(), MediaDetailSidebarProps, MediaViewerContext, MediaViewerContextType, useMediaViewer() (+13 more)

### Community 10 - "index.ts"
Cohesion: 0.12
Nodes (19): AdminTab, TIER_PRESENTATION, TierBadge(), TierBadgeProps, TierKey, AdminOverview, AuditLog, Entitlements (+11 more)

### Community 11 - "MemoryRepository"
Cohesion: 0.07
Nodes (12): MemoryRepository, RWMutex, AuditLog, Conversation, FollowRequest, MediaObject, Message, Permission (+4 more)

### Community 12 - "api.ts"
Cohesion: 0.11
Nodes (29): SettingsSection, AuthModal(), AuthModalProps, ComposerContext, ComposerContextValue, ComposerProvider(), LeftNavRail(), RightSidebar() (+21 more)

### Community 13 - "PostgresRepository"
Cohesion: 0.09
Nodes (9): DB, categorizeTopic(), Duration, MediaObject, Message, PostgresRepository, SubscriptionPlan, TrendingTopic (+1 more)

### Community 14 - "notifications/page.tsx"
Cohesion: 0.10
Nodes (20): VerifiedBadge(), VerifiedBadgeProps, MobileBottomNav(), MobileBottomNavProps, MobileDrawer(), MobileDrawerProps, UpgradeModalContext, UpgradeModalContextType (+12 more)

### Community 15 - "compilerOptions"
Cohesion: 0.07
Nodes (26): compilerOptions, allowJs, esModuleInterop, incremental, isolatedModules, jsx, lib, module (+18 more)

### Community 16 - "app/page.tsx"
Cohesion: 0.14
Nodes (21): encodeFeedCursor(), FeedResponse, HomePage(), MediaUrlInputModal(), MediaUrlInputModalProps, MobileComposerModal(), MobileComposerModalProps, StagedAttachment (+13 more)

### Community 17 - "Hub"
Cohesion: 0.16
Nodes (13): RawMessage, Client, Hub, InboundEvent, OutboundEvent, Conn, Context, Request (+5 more)

### Community 18 - "Post"
Cohesion: 0.17
Nodes (3): Rows, FeedCursor, Post

### Community 19 - "search/page.tsx"
Cohesion: 0.18
Nodes (11): ExploreCategory, SearchPageContent(), SearchPostsResponse, SearchTab, SearchUsersResponse, RecentSearchItem, SearchBarDropdown(), SearchBarDropdownProps (+3 more)

### Community 20 - "UserNameWithRole.tsx"
Cohesion: 0.21
Nodes (10): EditProfileModal(), EditProfileModalProps, NAME_STYLE_PRESETS, buildCustomNameInlineStyle(), parseCustomNameStyle(), UserNameWithRole(), UserNameWithRoleProps, CustomNameStyle (+2 more)

### Community 22 - "MediaService"
Cohesion: 0.18
Nodes (12): Config, Image, MediaService, Reader, getEnv(), getEnvInt(), Load(), Client (+4 more)

### Community 24 - "FeedSkeleton.tsx"
Cohesion: 0.19
Nodes (4): FeedSkeleton(), PostCardSkeleton(), PostDetailSkeleton(), ProfileSkeleton()

### Community 25 - "engine_test.go"
Cohesion: 0.20
Nodes (11): B, DiversityReranker, NewDiversityReranker(), BenchmarkDiversityReranker(), BenchmarkScoringEngine(), T, TestDiversityReranking(), TestFeedEvaluator() (+3 more)

### Community 26 - "admin.go"
Cohesion: 0.16
Nodes (13): AdminDeletePostRequest, AssignMembershipRequest, CreateRoleRequest, SetPermissionsRequest, SetPlanEntitlementsRequest, UpdatePlanRequest, UpdatePostModerationRequest, UpdateRoleRequest (+5 more)

### Community 27 - "MockCandidateProvider"
Cohesion: 0.40
Nodes (4): MockCandidateProvider, Context, Post, UUID

### Community 28 - "OtherSkeletons.tsx"
Cohesion: 0.23
Nodes (3): ExploreSkeleton(), MessagesSkeleton(), NotificationsSkeleton()

### Community 29 - "qrcode.tsx"
Cohesion: 0.24
Nodes (11): ALIGNMENT_PATTERNS, EC_CODEWORDS_L, encodeQRMatrix(), GF_EXP, GF_LOG, gfMul(), NUM_BLOCKS_L, QRCodeSVG() (+3 more)

### Community 31 - "handler.go"
Cohesion: 0.22
Nodes (10): CreatePostRequest, FeedFeedbackRequest, FeedImpressionsRequest, GoogleAuthRequest, LoginRequest, RegisterRequest, SimulatePaymentRequest, UpdatePostRequest (+2 more)

### Community 32 - "types.go"
Cohesion: 0.25
Nodes (9): Candidate, CandidateSource, Features, FeedCursor, FeedOptions, FeedResult, Post, Time (+1 more)

### Community 33 - "VideoControls.tsx"
Cohesion: 0.31
Nodes (7): formatTime(), SPEED_OPTIONS, VideoControls(), VideoControlsProps, formatTime(), VideoProgressBar(), VideoProgressBarProps

### Community 34 - "Engine"
Cohesion: 0.31
Nodes (7): DiversityReranker, FeedOptions, FeedResult, CandidateProvider, Engine, Context, NewEngine()

### Community 35 - ".GetActiveSubscription"
Cohesion: 0.22
Nodes (3): Entitlements, Subscription, UserMembership

### Community 36 - "LeftNavRail.tsx"
Cohesion: 0.18
Nodes (10): metadata, ConfirmModal(), ConfirmModalProps, useComposer(), AppShell(), LeftNavRailProps, RouteTransition(), AppProviders() (+2 more)

### Community 37 - "RankingWeights"
Cohesion: 0.29
Nodes (3): RankingWeights, DefaultRankingWeights(), CalculateLinearScore()

### Community 41 - "EncodeCursor"
Cohesion: 0.40
Nodes (5): DecodeCursor(), EncodeCursor(), Time, UUID, TestCursorRoundtrip()

### Community 42 - "route.ts"
Cohesion: 0.60
Nodes (3): getDynamicAuthOptions(), handler(), authOptions

### Community 44 - "EvaluateFeed"
Cohesion: 0.50
Nodes (4): FeedMetrics, EvaluateFeed(), Post, Time

## Knowledge Gaps
- **131 isolated node(s):** `name`, `version`, `private`, `dev`, `build` (+126 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **110 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `MemoryRepository` connect `MemoryRepository` to `T`, `models.go`, `.GetActiveSubscription`, `Context`, `RankingWeights`, `Role`, `.GetPlanEntitlements`, `.DismissUserRecommendation`, `.GetNotifications`, `Post`, `User`?**
  _High betweenness centrality (0.135) - this node is a cross-community bridge._
- **Why does `NewMemoryRepository()` connect `T` to `MemoryRepository`, `RankingWeights`, `.CheckoutSubscription`?**
  _High betweenness centrality (0.096) - this node is a cross-community bridge._
- **Why does `NewHandler()` connect `T` to `writeJSON`, `handler.go`, `.CheckoutSubscription`?**
  _High betweenness centrality (0.084) - this node is a cross-community bridge._
- **Are the 31 inferred relationships involving `writeJSON()` (e.g. with `.AdminAssignUserMembership()` and `.AdminCreateRole()`) actually correct?**
  _`writeJSON()` has 31 INFERRED edges - model-reasoned connections that need verification._
- **What connects `name`, `version`, `private` to the rest of the system?**
  _131 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `writeJSON` be split into smaller, more focused modules?**
  _Cohesion score 0.09343763581051717 - nodes in this community are weakly interconnected._
- **Should `T` be split into smaller, more focused modules?**
  _Cohesion score 0.06773211567732115 - nodes in this community are weakly interconnected._