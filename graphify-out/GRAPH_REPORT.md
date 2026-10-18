# Graph Report - .  (2026-10-18)

## Corpus Check
- cluster-only mode — file stats not available

## Summary
- 1188 nodes · 3113 edges · 180 communities (44 shown, 136 thin omitted)
- Extraction: 98% EXTRACTED · 2% INFERRED · 0% AMBIGUOUS · INFERRED: 49 edges (avg confidence: 0.8)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `1fa7d2ad`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- Handler
- MemoryRepository
- PostgresRepository
- MockCandidateProvider
- apiRequest
- Context
- models.go
- index.ts
- dependencies
- PlisioService
- Hub
- app/page.tsx
- compilerOptions
- utils.ts
- Post
- LeftNavRail.tsx
- User
- FeedSkeleton.tsx
- admin.go
- NewMemoryRepository
- search/page.tsx
- handler.go
- OtherSkeletons.tsx
- T
- Repository
- .GetFollowingPosts
- main
- VideoControls.tsx
- NewHandler
- Role
- setupTestApp
- IsOwnerAccount
- responseWriterWrapper
- setupTestRouter
- setupRecommendationRouter
- AuditLog
- NewPostgresRepository
- TopLoadingBar.tsx
- ParseQuery
- route.ts
- .GetAdminPosts
- Conversation
- PaymentOrder
- setupTestFullAdminRouter
- middleware.ts
- UserTopicPreference
- VerificationBadge
- .DismissUserRecommendation
- test_all_apis.py
- next.config.js
- next-env.d.ts
- tailwind.config.ts
- backup-db.sh
- monitoring.sh
- AdminOverview
- AdminPostFilter
- AuditLog
- Config
- Conversation
- CounterReconciliationResult
- Entitlements
- FeedCursor
- FeedFeedback
- FeedImpression
- FollowRequest
- Invoice
- MembershipEntitlement
- MembershipPlan
- Message
- Notification
- OutboxEvent
- PaymentOrder
- Permission
- github.com/x-reconstruction/api
- Post
- RankingWeights
- RecentSearch
- Role
- SearchFilterParams
- MembershipEntitlement
- Message
- Request
- ResponseWriter
- Client
- Context
- MediaObject
- Request
- ResponseWriter
- UUID
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
- Conversation
- CounterReconciliationResult
- Entitlements
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
- PaymentOrder
- Permission
- Post
- RankingWeights
- RecentSearch
- Role
- RWMutex
- SearchFilterParams
- Session
- SocialRelationship
- Subscription
- SubscriptionPlan
- TrendingTopic
- User
- UserMembership
- UserRecommendation
- UserTopicPreference
- VerificationApplication
- VerificationBadge
- Conversation
- CounterReconciliationResult
- Entitlements
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
- TrendingTopic
- User
- UserRecommendation
- UserTopicPreference
- VerificationApplication
- VerificationBadge
- Session
- SocialRelationship
- Subscription
- SubscriptionPlan
- SystemHealth
- Time
- TrendingTopic
- User
- UserMembership
- UserRecommendation
- UserTopicPreference
- VerificationApplication
- VerificationBadge

## God Nodes (most connected - your core abstractions)
1. `MemoryRepository` - 171 edges
2. `PostgresRepository` - 145 edges
3. `Handler` - 128 edges
4. `writeJSON()` - 81 edges
5. `writeError()` - 78 edges
6. `Post` - 45 edges
7. `apiRequest()` - 43 edges
8. `User` - 36 edges
9. `useCurrentUser()` - 32 edges
10. `User` - 30 edges

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

## Communities (180 total, 136 thin omitted)

### Community 0 - "Handler"
Cohesion: 0.09
Nodes (6): Context, Handler, Request, ResponseWriter, writeError(), writeJSON()

### Community 1 - "MemoryRepository"
Cohesion: 0.06
Nodes (7): Entitlements, Notification, MemoryRepository, Context, Duration, Time, UUID

### Community 2 - "PostgresRepository"
Cohesion: 0.06
Nodes (5): DB, PostgresRepository, Context, Time, UUID

### Community 3 - "MockCandidateProvider"
Cohesion: 0.05
Nodes (42): B, DiversityReranker, FeedOptions, FeedResult, Candidate, CandidateProvider, CandidateSource, DiversityReranker (+34 more)

### Community 4 - "apiRequest"
Cohesion: 0.13
Nodes (36): BookmarksPage(), ExplorePage(), InvoiceContent(), InvoicePage(), WhiteLabelInvoice, PremiumPage(), FollowersPage(), FollowingPage() (+28 more)

### Community 5 - "Context"
Cohesion: 0.08
Nodes (11): AdminOverview, AuditLog, Context, MembershipEntitlement, MembershipPlan, Permission, Post, Role (+3 more)

### Community 6 - "models.go"
Cohesion: 0.10
Nodes (38): CounterMismatchDetail, CounterReconciliationResult, DatabaseHealth, FeedFeedback, FeedImpression, Follow, Hashtag, Invoice (+30 more)

### Community 7 - "index.ts"
Cohesion: 0.08
Nodes (34): AdminTab, MessagesPage(), NotificationsPage(), ConfirmModal(), ConfirmModalProps, TIER_PRESENTATION, TierBadge(), TierBadgeProps (+26 more)

### Community 8 - "dependencies"
Cohesion: 0.05
Nodes (42): dependencies, clsx, lucide-react, next, next-auth, react, react-dom, tailwind-merge (+34 more)

### Community 9 - "PlisioService"
Cohesion: 0.09
Nodes (24): Config, Image, MediaService, PlisioCryptoInfo, PlisioInvoiceResponse, PlisioOperationsResponse, PlisioService, PlisioWebhookPayload (+16 more)

### Community 10 - "Hub"
Cohesion: 0.12
Nodes (17): Duration, memoryLimiter, Mutex, RawMessage, Client, Hub, InboundEvent, OutboundEvent (+9 more)

### Community 11 - "app/page.tsx"
Cohesion: 0.15
Nodes (22): encodeFeedCursor(), FeedResponse, HomePage(), StagedFile, MobileComposerModal(), MobileComposerModalProps, StagedAttachment, PostComposer() (+14 more)

### Community 12 - "compilerOptions"
Cohesion: 0.07
Nodes (26): compilerOptions, allowJs, esModuleInterop, incremental, isolatedModules, jsx, lib, module (+18 more)

### Community 13 - "utils.ts"
Cohesion: 0.14
Nodes (13): MediaDetailSidebar(), MediaDetailSidebarProps, PhotoCanvas(), PhotoCanvasProps, PhotoNavigation(), PhotoNavigationProps, StandalonePhotoView(), VideoPlayer() (+5 more)

### Community 14 - "Post"
Cohesion: 0.16
Nodes (4): NewPostsCheckResponse, Post, Rows, FeedCursor

### Community 15 - "LeftNavRail.tsx"
Cohesion: 0.14
Nodes (13): metadata, ComposerContext, ComposerContextValue, ComposerProvider(), useComposer(), AppShell(), LeftNavRail(), LeftNavRailProps (+5 more)

### Community 16 - "User"
Cohesion: 0.12
Nodes (4): FollowRequest, SearchSuggestionsResponse, User, UserRecommendation

### Community 17 - "FeedSkeleton.tsx"
Cohesion: 0.19
Nodes (4): FeedSkeleton(), PostCardSkeleton(), PostDetailSkeleton(), ProfileSkeleton()

### Community 18 - "admin.go"
Cohesion: 0.16
Nodes (13): AdminDeletePostRequest, AssignMembershipRequest, CreateRoleRequest, SetPermissionsRequest, SetPlanEntitlementsRequest, UpdatePlanRequest, UpdatePostModerationRequest, UpdateRoleRequest (+5 more)

### Community 19 - "NewMemoryRepository"
Cohesion: 0.34
Nodes (11): Handler, User, registerAndLoginUser(), setupTestRBACRouter(), TestConcurrentPostingCannotBypassQuota(), TestMediaUploadLimitsByTier(), TestNewUserDefaultMemberRoleAndAdminProtection(), TestOwnerBootstrapAndNonDemotionRule() (+3 more)

### Community 20 - "search/page.tsx"
Cohesion: 0.19
Nodes (9): ExploreCategory, SearchPageContent(), SearchPostsResponse, SearchTab, SearchUsersResponse, VirtualizedFeed(), VirtualizedFeedProps, feedKeys (+1 more)

### Community 21 - "handler.go"
Cohesion: 0.18
Nodes (12): CheckoutRequest, CreatePostRequest, FeedFeedbackRequest, FeedImpressionsRequest, GoogleAuthRequest, LoginRequest, RegisterRequest, SimulatePaymentRequest (+4 more)

### Community 22 - "OtherSkeletons.tsx"
Cohesion: 0.23
Nodes (3): ExploreSkeleton(), MessagesSkeleton(), NotificationsSkeleton()

### Community 23 - "T"
Cohesion: 0.52
Nodes (11): Mux, createTestPost(), Post, registerTestUserInInteractions(), setupInteractionsTestRouter(), TestCounterReconciliation(), TestLikeIdempotencyAndConcurrency(), TestProfileTabsFiltering() (+3 more)

### Community 24 - "Repository"
Cohesion: 0.27
Nodes (9): contextKey, Repository, AuthMiddleware(), AuthMiddlewareWithCache(), GetUserFromContext(), GetViewerID(), Client, Context (+1 more)

### Community 26 - "main"
Cohesion: 0.24
Nodes (6): Logger, main(), ConcurrencyLimiter(), RequestTimeoutMiddleware(), LoggerMiddleware(), RecoveryMiddleware()

### Community 27 - "VideoControls.tsx"
Cohesion: 0.31
Nodes (7): formatTime(), SPEED_OPTIONS, VideoControls(), VideoControlsProps, formatTime(), VideoProgressBar(), VideoProgressBarProps

### Community 28 - "NewHandler"
Cohesion: 0.22
Nodes (8): Client, Hub, MediaService, PlisioService, Repository, NewHandler(), T, TestSearchAPIEndpoints()

### Community 30 - "setupTestApp"
Cohesion: 0.47
Nodes (8): createTestUser(), Mux, T, User, setupTestApp(), TestDirectMessaging_IdempotencyAndFlow(), TestFollowRequest_PrivateAccount(), TestSocialGraph_FollowUnfollowBlockMute()

### Community 31 - "IsOwnerAccount"
Cohesion: 0.43
Nodes (7): AuthorizeResource(), User, UUID, IsOwnerAccount(), RequireAnyPermission(), RequireOwner(), RequirePermission()

### Community 32 - "responseWriterWrapper"
Cohesion: 0.33
Nodes (4): Conn, responseWriterWrapper, ReadWriter, ResponseWriter

### Community 33 - "setupTestRouter"
Cohesion: 0.52
Nodes (6): Mux, T, setupTestRouter(), TestAuthAndPostingFlow(), TestHealthCheck(), TestSubscriptionsAndBadges()

### Community 34 - "setupRecommendationRouter"
Cohesion: 0.67
Nodes (6): registerTestUser(), setupRecommendationRouter(), TestRealtimeNewPostsIndicator(), TestRecommendationAndFeedAPIs(), TestTrendingAnalyticsAndSpamResistance(), TestWhoToFollowAndDismissal()

### Community 38 - "ParseQuery"
Cohesion: 0.40
Nodes (3): ParseQuery(), T, TestParseQuery()

### Community 43 - "setupTestFullAdminRouter"
Cohesion: 0.83
Nodes (3): setupTestFullAdminRouter(), TestAdminPostModerationAndAuditLogging(), TestOwnerAdminOverviewSecurity()

## Knowledge Gaps
- **115 isolated node(s):** `name`, `version`, `private`, `dev`, `build` (+110 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **136 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `MemoryRepository` connect `MemoryRepository` to `AuditLog`, `models.go`, `.GetAdminPosts`, `Conversation`, `PaymentOrder`, `Hub`, `UserTopicPreference`, `Post`, `VerificationBadge`, `User`, `NewMemoryRepository`, `.GetFollowingPosts`, `Role`?**
  _High betweenness centrality (0.167) - this node is a cross-community bridge._
- **Why does `Handler` connect `Handler` to `setupTestFullAdminRouter`, `handler.go`, `Repository`, `main`, `NewHandler`, `IsOwnerAccount`?**
  _High betweenness centrality (0.138) - this node is a cross-community bridge._
- **Why does `main()` connect `main` to `Handler`, `NewPostgresRepository`, `PlisioService`, `NewMemoryRepository`, `Repository`, `NewHandler`?**
  _High betweenness centrality (0.129) - this node is a cross-community bridge._
- **What connects `name`, `version`, `private` to the rest of the system?**
  _115 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `Handler` be split into smaller, more focused modules?**
  _Cohesion score 0.09055472263868065 - nodes in this community are weakly interconnected._
- **Should `MemoryRepository` be split into smaller, more focused modules?**
  _Cohesion score 0.05937353903693315 - nodes in this community are weakly interconnected._
- **Should `PostgresRepository` be split into smaller, more focused modules?**
  _Cohesion score 0.06101231190150479 - nodes in this community are weakly interconnected._