# Graph Report - .  (2026-10-19)

## Corpus Check
- cluster-only mode — file stats not available

## Summary
- 1189 nodes · 3103 edges · 126 communities (48 shown, 78 thin omitted)
- Extraction: 98% EXTRACTED · 2% INFERRED · 0% AMBIGUOUS · INFERRED: 47 edges (avg confidence: 0.8)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `235681ff`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- Handler
- Context
- models.go
- MockCandidateProvider
- MemoryRepository
- Context
- Context
- User
- dependencies
- PostgresRepository
- PlisioService
- [postId]/page.tsx
- apiRequest
- Hub
- compilerOptions
- index.ts
- Post
- Post
- app/page.tsx
- FeedSkeleton.tsx
- admin.go
- NewMemoryRepository
- search/page.tsx
- handler.go
- OtherSkeletons.tsx
- T
- ComposerContext.tsx
- Repository
- main
- VideoControls.tsx
- NewHandler
- Engine
- .GetActiveSubscription
- Role
- setupTestApp
- IsOwnerAccount
- responseWriterWrapper
- .GetPlanEntitlements
- setupTestRouter
- setupRecommendationRouter
- .GetBillingHistory
- .GetBillingHistory
- TopLoadingBar.tsx
- .GetActiveSubscription
- route.ts
- AuditLog
- .GetNotifications
- setupTestFullAdminRouter
- middleware.ts
- .GetRolePermissions
- .GetFollowRequests
- .GetMessages
- .GetVerificationApplication
- test_all_apis.py
- .GetAdminOverview
- next.config.js
- next-env.d.ts
- tailwind.config.ts
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
1. `MemoryRepository` - 171 edges
2. `PostgresRepository` - 145 edges
3. `Handler` - 128 edges
4. `writeJSON()` - 81 edges
5. `writeError()` - 78 edges
6. `apiRequest()` - 39 edges
7. `User` - 30 edges
8. `useCurrentUser()` - 29 edges
9. `cn()` - 22 edges
10. `Post` - 18 edges

## Surprising Connections (you probably didn't know these)
- `setupTestRouter()` --calls--> `NewHandler()`  [INFERRED]
  services/api/internal/handler/handler_test.go → services/api/internal/handler/handler.go
- `setupTestRouter()` --calls--> `AuthMiddleware()`  [INFERRED]
  services/api/internal/handler/handler_test.go → services/api/internal/middleware/middleware.go
- `setupTestRouter()` --calls--> `NewMemoryRepository()`  [INFERRED]
  services/api/internal/handler/handler_test.go → services/api/internal/repository/memory.go
- `TestSearchAPIEndpoints()` --calls--> `NewMemoryRepository()`  [INFERRED]
  services/api/internal/handler/search_test.go → services/api/internal/repository/memory.go
- `setupTestApp()` --calls--> `NewHandler()`  [INFERRED]
  services/api/internal/handler/social_messaging_test.go → services/api/internal/handler/handler.go

## Import Cycles
- None detected.

## Communities (126 total, 78 thin omitted)

### Community 0 - "Handler"
Cohesion: 0.09
Nodes (6): Handler, Request, ResponseWriter, Context, writeError(), writeJSON()

### Community 1 - "Context"
Cohesion: 0.06
Nodes (14): Context, Conversation, Duration, FollowRequest, MediaObject, Message, Notification, RecentSearch (+6 more)

### Community 2 - "models.go"
Cohesion: 0.08
Nodes (58): AdminOverview, AuditLog, Conversation, CounterMismatchDetail, CounterReconciliationResult, DatabaseHealth, Entitlements, FeedFeedback (+50 more)

### Community 3 - "MockCandidateProvider"
Cohesion: 0.06
Nodes (34): B, Candidate, CandidateSource, DiversityReranker, Features, FeedCursor, FeedMetrics, FeedOptions (+26 more)

### Community 4 - "MemoryRepository"
Cohesion: 0.06
Nodes (16): MemoryRepository, CounterReconciliationResult, FeedFeedback, FeedImpression, MediaObject, OutboxEvent, RankingWeights, RecentSearch (+8 more)

### Community 5 - "Context"
Cohesion: 0.07
Nodes (6): Context, Conversation, Duration, SocialRelationship, UserRecommendation, UUID

### Community 6 - "Context"
Cohesion: 0.08
Nodes (12): AdminPostFilter, AdminOverview, AuditLog, Context, MembershipEntitlement, MembershipPlan, Permission, Post (+4 more)

### Community 7 - "User"
Cohesion: 0.11
Nodes (26): WhiteLabelInvoice, VerifiedBadge(), VerifiedBadgeProps, useComposer(), LeftNavRail(), LeftNavRailProps, MobileBottomNav(), MobileBottomNavProps (+18 more)

### Community 8 - "dependencies"
Cohesion: 0.05
Nodes (42): dependencies, clsx, lucide-react, next, next-auth, react, react-dom, tailwind-merge (+34 more)

### Community 9 - "PostgresRepository"
Cohesion: 0.07
Nodes (13): DB, PostgresRepository, categorizeTopic(), CounterReconciliationResult, FeedFeedback, FeedImpression, OutboxEvent, RankingWeights (+5 more)

### Community 10 - "PlisioService"
Cohesion: 0.09
Nodes (24): Config, Image, MediaService, PlisioCryptoInfo, PlisioInvoiceResponse, PlisioOperationsResponse, PlisioService, PlisioWebhookPayload (+16 more)

### Community 11 - "[postId]/page.tsx"
Cohesion: 0.11
Nodes (22): StagedFile, PostModal(), PostCardProps, MediaDetailSidebar(), MediaDetailSidebarProps, MediaViewerContext, MediaViewerContextType, MediaViewerProvider() (+14 more)

### Community 12 - "apiRequest"
Cohesion: 0.16
Nodes (25): BookmarksPage(), ExplorePage(), PremiumPage(), FollowersPage(), FollowingPage(), PostDetailPage(), AuthModal(), AuthModalProps (+17 more)

### Community 13 - "Hub"
Cohesion: 0.12
Nodes (17): Duration, memoryLimiter, Mutex, RawMessage, Client, Hub, InboundEvent, OutboundEvent (+9 more)

### Community 14 - "compilerOptions"
Cohesion: 0.07
Nodes (26): compilerOptions, allowJs, esModuleInterop, incremental, isolatedModules, jsx, lib, module (+18 more)

### Community 15 - "index.ts"
Cohesion: 0.11
Nodes (20): AdminTab, ConfirmModal(), ConfirmModalProps, TIER_PRESENTATION, TierBadge(), TierBadgeProps, TierKey, AdminOverview (+12 more)

### Community 16 - "Post"
Cohesion: 0.16
Nodes (4): Rows, FeedCursor, Post, SearchFilterParams

### Community 17 - "Post"
Cohesion: 0.16
Nodes (4): AdminPostFilter, FeedCursor, Post, SearchFilterParams

### Community 18 - "app/page.tsx"
Cohesion: 0.17
Nodes (15): encodeFeedCursor(), FeedResponse, HomePage(), MobileComposerModal(), MobileComposerModalProps, StagedAttachment, PostComposer(), PostComposerProps (+7 more)

### Community 19 - "FeedSkeleton.tsx"
Cohesion: 0.19
Nodes (4): FeedSkeleton(), PostCardSkeleton(), PostDetailSkeleton(), ProfileSkeleton()

### Community 20 - "admin.go"
Cohesion: 0.16
Nodes (13): AdminDeletePostRequest, AssignMembershipRequest, CreateRoleRequest, SetPermissionsRequest, SetPlanEntitlementsRequest, UpdatePlanRequest, UpdatePostModerationRequest, UpdateRoleRequest (+5 more)

### Community 21 - "NewMemoryRepository"
Cohesion: 0.34
Nodes (11): Handler, User, registerAndLoginUser(), setupTestRBACRouter(), TestConcurrentPostingCannotBypassQuota(), TestMediaUploadLimitsByTier(), TestNewUserDefaultMemberRoleAndAdminProtection(), TestOwnerBootstrapAndNonDemotionRule() (+3 more)

### Community 22 - "search/page.tsx"
Cohesion: 0.19
Nodes (9): ExploreCategory, SearchPageContent(), SearchPostsResponse, SearchTab, SearchUsersResponse, VirtualizedFeed(), VirtualizedFeedProps, feedKeys (+1 more)

### Community 23 - "handler.go"
Cohesion: 0.18
Nodes (12): CheckoutRequest, CreatePostRequest, FeedFeedbackRequest, FeedImpressionsRequest, GoogleAuthRequest, LoginRequest, RegisterRequest, SimulatePaymentRequest (+4 more)

### Community 24 - "OtherSkeletons.tsx"
Cohesion: 0.23
Nodes (3): ExploreSkeleton(), MessagesSkeleton(), NotificationsSkeleton()

### Community 25 - "T"
Cohesion: 0.52
Nodes (11): Mux, createTestPost(), Post, registerTestUserInInteractions(), setupInteractionsTestRouter(), TestCounterReconciliation(), TestLikeIdempotencyAndConcurrency(), TestProfileTabsFiltering() (+3 more)

### Community 26 - "ComposerContext.tsx"
Cohesion: 0.22
Nodes (6): metadata, ComposerContext, ComposerContextValue, ComposerProvider(), AppShell(), AppProviders()

### Community 27 - "Repository"
Cohesion: 0.27
Nodes (9): contextKey, Repository, AuthMiddleware(), AuthMiddlewareWithCache(), GetUserFromContext(), GetViewerID(), Client, Context (+1 more)

### Community 28 - "main"
Cohesion: 0.24
Nodes (6): Logger, main(), ConcurrencyLimiter(), RequestTimeoutMiddleware(), LoggerMiddleware(), RecoveryMiddleware()

### Community 29 - "VideoControls.tsx"
Cohesion: 0.31
Nodes (7): formatTime(), SPEED_OPTIONS, VideoControls(), VideoControlsProps, formatTime(), VideoProgressBar(), VideoProgressBarProps

### Community 30 - "NewHandler"
Cohesion: 0.22
Nodes (8): Client, Hub, MediaService, PlisioService, Repository, NewHandler(), T, TestSearchAPIEndpoints()

### Community 31 - "Engine"
Cohesion: 0.31
Nodes (7): DiversityReranker, FeedOptions, FeedResult, CandidateProvider, Engine, Context, NewEngine()

### Community 32 - ".GetActiveSubscription"
Cohesion: 0.22
Nodes (3): Entitlements, Subscription, UserMembership

### Community 34 - "setupTestApp"
Cohesion: 0.47
Nodes (8): createTestUser(), Mux, T, User, setupTestApp(), TestDirectMessaging_IdempotencyAndFlow(), TestFollowRequest_PrivateAccount(), TestSocialGraph_FollowUnfollowBlockMute()

### Community 35 - "IsOwnerAccount"
Cohesion: 0.43
Nodes (7): AuthorizeResource(), User, UUID, IsOwnerAccount(), RequireAnyPermission(), RequireOwner(), RequirePermission()

### Community 36 - "responseWriterWrapper"
Cohesion: 0.33
Nodes (4): Conn, responseWriterWrapper, ReadWriter, ResponseWriter

### Community 38 - "setupTestRouter"
Cohesion: 0.52
Nodes (6): Mux, T, setupTestRouter(), TestAuthAndPostingFlow(), TestHealthCheck(), TestSubscriptionsAndBadges()

### Community 39 - "setupRecommendationRouter"
Cohesion: 0.67
Nodes (6): registerTestUser(), setupRecommendationRouter(), TestRealtimeNewPostsIndicator(), TestRecommendationAndFeedAPIs(), TestTrendingAnalyticsAndSpamResistance(), TestWhoToFollowAndDismissal()

### Community 47 - "setupTestFullAdminRouter"
Cohesion: 0.83
Nodes (3): setupTestFullAdminRouter(), TestAdminPostModerationAndAuditLogging(), TestOwnerAdminOverviewSecurity()

## Knowledge Gaps
- **117 isolated node(s):** `name`, `version`, `private`, `dev`, `build` (+112 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **78 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `PostgresRepository` connect `PostgresRepository` to `Context`, `Context`, `.GetBillingHistory`, `.GetActiveSubscription`, `Post`, `.GetVerificationApplication`?**
  _High betweenness centrality (0.176) - this node is a cross-community bridge._
- **Why does `main()` connect `main` to `Handler`, `PostgresRepository`, `PlisioService`, `NewMemoryRepository`, `Repository`, `NewHandler`?**
  _High betweenness centrality (0.153) - this node is a cross-community bridge._
- **Why does `Handler` connect `Handler` to `IsOwnerAccount`, `setupTestFullAdminRouter`, `handler.go`, `Repository`, `main`, `NewHandler`?**
  _High betweenness centrality (0.118) - this node is a cross-community bridge._
- **What connects `name`, `version`, `private` to the rest of the system?**
  _117 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `Handler` be split into smaller, more focused modules?**
  _Cohesion score 0.09070464767616192 - nodes in this community are weakly interconnected._
- **Should `Context` be split into smaller, more focused modules?**
  _Cohesion score 0.05625 - nodes in this community are weakly interconnected._
- **Should `models.go` be split into smaller, more focused modules?**
  _Cohesion score 0.08469945355191257 - nodes in this community are weakly interconnected._