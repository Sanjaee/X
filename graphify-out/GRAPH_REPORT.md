# Graph Report - .  (2026-11-01)

## Corpus Check
- cluster-only mode — file stats not available

## Summary
- 1246 nodes · 3207 edges · 147 communities (47 shown, 100 thin omitted)
- Extraction: 99% EXTRACTED · 1% INFERRED · 0% AMBIGUOUS · INFERRED: 40 edges (avg confidence: 0.8)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `26f7fc2b`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- Handler
- Context
- MockCandidateProvider
- models.go
- Context
- Request
- MemoryRepository
- PostgresRepository
- [postId]/page.tsx
- dependencies
- CryptoGateway
- api.ts
- cn
- Hub
- compilerOptions
- index.ts
- Post
- app/page.tsx
- Post
- search/page.tsx
- notifications/page.tsx
- FeedSkeleton.tsx
- Repository
- admin.go
- NewMemoryRepository
- T
- OtherSkeletons.tsx
- VideoControls.tsx
- .GetActiveSubscription
- Role
- setupTestApp
- NewHandler
- IsOwnerAccount
- responseWriterWrapper
- .GetBillingHistory
- .GetPlanEntitlements
- .GetBillingHistory
- NewPostgresRepository
- setupTestRouter
- setupRecommendationRouter
- TopLoadingBar.tsx
- .GetActiveSubscription
- route.ts
- AuditLog
- setupTestFullAdminRouter
- middleware.ts
- Duration
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
- .GetUserTopicPreferences
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
- Client
- Config
- github.com/x-reconstruction/api
- PlisioService
- MembershipEntitlement
- Message
- Request
- ResponseWriter
- MediaObject
- PaymentOrder
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
3. `Handler` - 89 edges
4. `writeJSON()` - 78 edges
5. `writeError()` - 76 edges
6. `apiRequest()` - 33 edges
7. `User` - 27 edges
8. `cn()` - 22 edges
9. `useCurrentUser()` - 22 edges
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

## Communities (147 total, 100 thin omitted)

### Community 0 - "Handler"
Cohesion: 0.11
Nodes (16): CreatePostRequest, FeedFeedbackRequest, FeedImpressionsRequest, GoogleAuthRequest, LoginRequest, RegisterRequest, SimulatePaymentRequest, UpdatePostRequest (+8 more)

### Community 1 - "Context"
Cohesion: 0.05
Nodes (13): Context, Conversation, FollowRequest, MediaObject, Message, Notification, RecentSearch, SocialRelationship (+5 more)

### Community 2 - "MockCandidateProvider"
Cohesion: 0.05
Nodes (43): B, DiversityReranker, RankingWeights, FeedOptions, FeedResult, Candidate, CandidateProvider, CandidateSource (+35 more)

### Community 3 - "models.go"
Cohesion: 0.09
Nodes (57): AdminOverview, AdminPostFilter, AuditLog, Conversation, CounterMismatchDetail, CounterReconciliationResult, DatabaseHealth, Entitlements (+49 more)

### Community 4 - "Context"
Cohesion: 0.07
Nodes (6): Context, Conversation, Notification, SocialRelationship, UserRecommendation, UUID

### Community 5 - "Request"
Cohesion: 0.09
Nodes (13): CryptoInvoice, CryptoOperation, CheckoutRequest, Handler, Request, ResponseWriter, main(), expired() (+5 more)

### Community 6 - "MemoryRepository"
Cohesion: 0.07
Nodes (14): Permission, MemoryRepository, FeedFeedback, FeedImpression, FollowRequest, OutboxEvent, RankingWeights, RecentSearch (+6 more)

### Community 7 - "PostgresRepository"
Cohesion: 0.11
Nodes (12): DB, PostgresRepository, AdminOverview, AuditLog, Context, MembershipEntitlement, MembershipPlan, Permission (+4 more)

### Community 8 - "[postId]/page.tsx"
Cohesion: 0.11
Nodes (25): StagedFile, PostModal(), PostModalProps, StagedAttachment, PostCard(), PostCardProps, MediaDetailSidebar(), MediaDetailSidebarProps (+17 more)

### Community 9 - "dependencies"
Cohesion: 0.05
Nodes (42): dependencies, clsx, lucide-react, next, next-auth, react, react-dom, tailwind-merge (+34 more)

### Community 10 - "CryptoGateway"
Cohesion: 0.09
Nodes (27): Config, Image, MediaService, CryptoCurrency, CryptoGateway, CryptoInvoice, CryptoOperation, WebhookPayload (+19 more)

### Community 11 - "api.ts"
Cohesion: 0.20
Nodes (18): SettingsSection, AuthModal(), AuthModalProps, LeftNavRail(), LeftNavRailProps, RightSidebar(), RightSidebarProps, EditProfileModal() (+10 more)

### Community 12 - "cn"
Cohesion: 0.09
Nodes (21): metadata, ComposerContext, ComposerContextValue, ComposerProvider(), useComposer(), AppShell(), RouteTransition(), MediaViewerProvider() (+13 more)

### Community 13 - "Hub"
Cohesion: 0.12
Nodes (16): memoryLimiter, Mutex, RawMessage, Client, Hub, InboundEvent, OutboundEvent, RWMutex (+8 more)

### Community 14 - "compilerOptions"
Cohesion: 0.07
Nodes (26): compilerOptions, allowJs, esModuleInterop, incremental, isolatedModules, jsx, lib, module (+18 more)

### Community 15 - "index.ts"
Cohesion: 0.10
Nodes (21): AdminTab, ConfirmModal(), ConfirmModalProps, TIER_PRESENTATION, TierBadge(), TierBadgeProps, TierKey, AdminOverview (+13 more)

### Community 16 - "Post"
Cohesion: 0.16
Nodes (4): FeedCursor, Post, Rows, SearchFilterParams

### Community 17 - "app/page.tsx"
Cohesion: 0.13
Nodes (18): encodeFeedCursor(), FeedResponse, HomePage(), VerifiedBadge(), VerifiedBadgeProps, MobileComposerModal(), MobileComposerModalProps, StagedAttachment (+10 more)

### Community 18 - "Post"
Cohesion: 0.16
Nodes (4): AdminPostFilter, FeedCursor, Post, SearchFilterParams

### Community 19 - "search/page.tsx"
Cohesion: 0.13
Nodes (16): ExploreCategory, SearchPageContent(), SearchPostsResponse, SearchTab, SearchUsersResponse, MobileDrawer(), MobileDrawerProps, RecentSearchItem (+8 more)

### Community 20 - "notifications/page.tsx"
Cohesion: 0.16
Nodes (12): MobileBottomNav(), MobileBottomNavProps, RealtimeContext, RealtimeContextType, RealtimeProvider(), ToastNotification, TypingEvent, useRealtime() (+4 more)

### Community 21 - "FeedSkeleton.tsx"
Cohesion: 0.19
Nodes (4): FeedSkeleton(), PostCardSkeleton(), PostDetailSkeleton(), ProfileSkeleton()

### Community 22 - "Repository"
Cohesion: 0.19
Nodes (12): Logger, contextKey, Repository, AuthMiddleware(), AuthMiddlewareWithCache(), GetUserFromContext(), GetViewerID(), Client (+4 more)

### Community 23 - "admin.go"
Cohesion: 0.16
Nodes (13): AdminDeletePostRequest, AssignMembershipRequest, CreateRoleRequest, SetPermissionsRequest, SetPlanEntitlementsRequest, UpdatePlanRequest, UpdatePostModerationRequest, UpdateRoleRequest (+5 more)

### Community 24 - "NewMemoryRepository"
Cohesion: 0.34
Nodes (11): Handler, User, registerAndLoginUser(), setupTestRBACRouter(), TestConcurrentPostingCannotBypassQuota(), TestMediaUploadLimitsByTier(), TestNewUserDefaultMemberRoleAndAdminProtection(), TestOwnerBootstrapAndNonDemotionRule() (+3 more)

### Community 25 - "T"
Cohesion: 0.50
Nodes (12): Mux, createTestPost(), Post, registerTestUserInInteractions(), setupInteractionsTestRouter(), TestBookmarkFlowAndLikesTab(), TestCounterReconciliation(), TestLikeIdempotencyAndConcurrency() (+4 more)

### Community 26 - "OtherSkeletons.tsx"
Cohesion: 0.23
Nodes (3): ExploreSkeleton(), MessagesSkeleton(), NotificationsSkeleton()

### Community 27 - "VideoControls.tsx"
Cohesion: 0.31
Nodes (7): formatTime(), SPEED_OPTIONS, VideoControls(), VideoControlsProps, formatTime(), VideoProgressBar(), VideoProgressBarProps

### Community 28 - ".GetActiveSubscription"
Cohesion: 0.22
Nodes (3): Entitlements, Subscription, UserMembership

### Community 30 - "setupTestApp"
Cohesion: 0.47
Nodes (8): createTestUser(), Mux, T, User, setupTestApp(), TestDirectMessaging_IdempotencyAndFlow(), TestFollowRequest_PrivateAccount(), TestSocialGraph_FollowUnfollowBlockMute()

### Community 31 - "NewHandler"
Cohesion: 0.25
Nodes (7): Hub, MediaService, Repository, Client, NewHandler(), T, TestSearchAPIEndpoints()

### Community 32 - "IsOwnerAccount"
Cohesion: 0.43
Nodes (7): AuthorizeResource(), User, UUID, IsOwnerAccount(), RequireAnyPermission(), RequireOwner(), RequirePermission()

### Community 33 - "responseWriterWrapper"
Cohesion: 0.33
Nodes (4): Conn, responseWriterWrapper, ReadWriter, ResponseWriter

### Community 37 - "NewPostgresRepository"
Cohesion: 0.29
Nodes (3): categorizeTopic(), NewPostgresRepository(), TrendingTopic

### Community 38 - "setupTestRouter"
Cohesion: 0.52
Nodes (6): Mux, T, setupTestRouter(), TestAuthAndPostingFlow(), TestHealthCheck(), TestSubscriptionsAndBadges()

### Community 39 - "setupRecommendationRouter"
Cohesion: 0.67
Nodes (6): registerTestUser(), setupRecommendationRouter(), TestRealtimeNewPostsIndicator(), TestRecommendationAndFeedAPIs(), TestTrendingAnalyticsAndSpamResistance(), TestWhoToFollowAndDismissal()

### Community 45 - "setupTestFullAdminRouter"
Cohesion: 0.83
Nodes (3): setupTestFullAdminRouter(), TestAdminPostModerationAndAuditLogging(), TestOwnerAdminOverviewSecurity()

## Knowledge Gaps
- **128 isolated node(s):** `name`, `version`, `private`, `dev`, `build` (+123 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **100 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `MemoryRepository` connect `MemoryRepository` to `.GetPlanEntitlements`, `Context`, `.GetBillingHistory`, `.ReconcileCounters`, `AuditLog`, `Hub`, `.GetAdminOverview`, `.DismissUserRecommendation`, `.GetMediaObjectByID`, `.GetMessages`, `Post`, `.GetVerificationApplication`, `.GetUserTopicPreferences`, `NewMemoryRepository`, `.GetActiveSubscription`, `Role`?**
  _High betweenness centrality (0.153) - this node is a cross-community bridge._
- **Why does `NewMemoryRepository()` connect `NewMemoryRepository` to `Request`, `setupTestRouter`, `setupRecommendationRouter`, `MemoryRepository`, `setupTestFullAdminRouter`, `T`, `setupTestApp`, `NewHandler`?**
  _High betweenness centrality (0.144) - this node is a cross-community bridge._
- **Why does `PostgresRepository` connect `PostgresRepository` to `.ReconcileCounters`, `Context`, `.GetBillingHistory`, `.GetActiveWeights`, `.RecordFeedFeedback`, `NewPostgresRepository`, `.RecordFeedImpressions`, `.GetActiveSubscription`, `Duration`, `Post`, `.CreateOutboxEvent`, `.GetUserBadges`, `.CreateSession`, `.GetAdminPosts`, `.GetSubscriptionPlanByID`, `.GetUserTopicPreferences`?**
  _High betweenness centrality (0.131) - this node is a cross-community bridge._
- **What connects `name`, `version`, `private` to the rest of the system?**
  _128 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `Handler` be split into smaller, more focused modules?**
  _Cohesion score 0.10966810966810966 - nodes in this community are weakly interconnected._
- **Should `Context` be split into smaller, more focused modules?**
  _Cohesion score 0.05225225225225225 - nodes in this community are weakly interconnected._
- **Should `MockCandidateProvider` be split into smaller, more focused modules?**
  _Cohesion score 0.05048076923076923 - nodes in this community are weakly interconnected._