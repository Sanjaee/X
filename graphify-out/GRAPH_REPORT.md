# Graph Report - .  (2026-11-01)

## Corpus Check
- cluster-only mode — file stats not available

## Summary
- 1214 nodes · 3171 edges · 139 communities (37 shown, 102 thin omitted)
- Extraction: 99% EXTRACTED · 1% INFERRED · 0% AMBIGUOUS · INFERRED: 41 edges (avg confidence: 0.8)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `1a6e327c`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- Handler
- PostgresRepository
- T
- models.go
- MemoryRepository
- MockCandidateProvider
- Context
- Hub
- Context
- api.ts
- dependencies
- app/page.tsx
- [postId]/page.tsx
- index.ts
- compilerOptions
- PlisioService
- Post
- Post
- MediaService
- User
- FeedSkeleton.tsx
- ComposerContext.tsx
- search/page.tsx
- admin.go
- handler.go
- OtherSkeletons.tsx
- User
- VideoControls.tsx
- Engine
- PostComposer.tsx
- .GetActiveSubscription
- .GetPlanEntitlements
- NewPostgresRepository
- Invoice
- TopLoadingBar.tsx
- .GetActiveSubscription
- route.ts
- Notification
- VerifiedBadge.tsx
- middleware.ts
- test_all_apis.py
- next.config.js
- next-env.d.ts
- tailwind.config.ts
- backup-db.sh
- monitoring.sh
- AdminOverview
- AdminPostFilter
- AuditLog
- Client
- Context
- MediaObject
- MembershipEntitlement
- MembershipPlan
- Permission
- github.com/x-reconstruction/api
- PlisioService
- Role
- MembershipEntitlement
- Message
- Request
- ResponseWriter
- Request
- ResponseWriter
- Mux
- T
- MediaObject
- Handler
- Conn
- User
- RawMessage
- Conn
- RawMessage
- RWMutex
- Conversation
- CounterReconciliationResult
- Duration
- Entitlements
- FeedCursor
- FeedFeedback
- FeedImpression
- FollowRequest
- Invoice
- Message
- Notification
- OutboxEvent
- RankingWeights
- RecentSearch
- RWMutex
- SearchFilterParams
- Session
- SocialRelationship
- SubscriptionPlan
- Time
- TrendingTopic
- UserRecommendation
- UserTopicPreference
- VerificationApplication
- VerificationBadge
- Conversation
- CounterReconciliationResult
- Duration
- Entitlements
- FeedCursor
- FeedFeedback
- FeedImpression
- FollowRequest
- Invoice
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
- VerificationApplication
- VerificationBadge
- SystemHealth
- UserMembership
- UUID

## God Nodes (most connected - your core abstractions)
1. `MemoryRepository` - 172 edges
2. `PostgresRepository` - 146 edges
3. `Handler` - 131 edges
4. `writeJSON()` - 82 edges
5. `writeError()` - 79 edges
6. `apiRequest()` - 34 edges
7. `User` - 28 edges
8. `cn()` - 23 edges
9. `useCurrentUser()` - 22 edges
10. `UserNameWithRole()` - 20 edges

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

## Communities (139 total, 102 thin omitted)

### Community 0 - "Handler"
Cohesion: 0.08
Nodes (11): Handler, Request, ResponseWriter, Context, PaymentOrder, Subscription, writeError(), writeJSON() (+3 more)

### Community 1 - "PostgresRepository"
Cohesion: 0.06
Nodes (10): CounterReconciliationResult, DB, Duration, PaymentOrder, PostgresRepository, Context, MediaObject, UUID (+2 more)

### Community 2 - "T"
Cohesion: 0.07
Nodes (61): Hub, MediaService, Mux, Repository, Repository, setupTestFullAdminRouter(), TestAdminPostModerationAndAuditLogging(), TestOwnerAdminOverviewSecurity() (+53 more)

### Community 3 - "models.go"
Cohesion: 0.08
Nodes (59): AdminOverview, AdminPostFilter, AuditLog, Conversation, CounterMismatchDetail, CounterReconciliationResult, DatabaseHealth, Entitlements (+51 more)

### Community 4 - "MemoryRepository"
Cohesion: 0.06
Nodes (19): FeedFeedback, FeedImpression, FollowRequest, Message, OutboxEvent, RankingWeights, RecentSearch, MemoryRepository (+11 more)

### Community 5 - "MockCandidateProvider"
Cohesion: 0.06
Nodes (34): B, Candidate, CandidateSource, DiversityReranker, Features, FeedCursor, FeedMetrics, FeedOptions (+26 more)

### Community 6 - "Context"
Cohesion: 0.07
Nodes (3): Conversation, Context, UUID

### Community 7 - "Hub"
Cohesion: 0.06
Nodes (31): Conn, Logger, contextKey, memoryLimiter, responseWriterWrapper, Mutex, ReadWriter, Client (+23 more)

### Community 8 - "Context"
Cohesion: 0.08
Nodes (12): AdminOverview, AdminPostFilter, AuditLog, Context, MembershipEntitlement, MembershipPlan, Permission, Post (+4 more)

### Community 9 - "api.ts"
Cohesion: 0.13
Nodes (24): WhiteLabelInvoice, AuthModalProps, RightSidebar(), RightSidebarProps, EditProfileModal(), EditProfileModalProps, RecentSearchItem, SearchBarDropdown() (+16 more)

### Community 10 - "dependencies"
Cohesion: 0.05
Nodes (42): dependencies, clsx, lucide-react, next, next-auth, react, react-dom, tailwind-merge (+34 more)

### Community 11 - "app/page.tsx"
Cohesion: 0.12
Nodes (24): encodeFeedCursor(), FeedResponse, HomePage(), SettingsSection, AuthModal(), MobileComposerModal(), MobileComposerModalProps, StagedAttachment (+16 more)

### Community 12 - "[postId]/page.tsx"
Cohesion: 0.11
Nodes (23): StagedFile, PostModal(), PostModalProps, StagedAttachment, PostCard(), PostCardProps, MediaDetailSidebar(), MediaDetailSidebarProps (+15 more)

### Community 13 - "index.ts"
Cohesion: 0.09
Nodes (26): AdminTab, CryptoCurrencyOption, PURCHASABLE_KEYS, ROLE_TIER_DEFAULTS, ConfirmModal(), ConfirmModalProps, TIER_PRESENTATION, TierBadge() (+18 more)

### Community 14 - "compilerOptions"
Cohesion: 0.07
Nodes (26): compilerOptions, allowJs, esModuleInterop, incremental, isolatedModules, jsx, lib, module (+18 more)

### Community 15 - "PlisioService"
Cohesion: 0.16
Nodes (14): Config, PlisioCryptoInfo, PlisioInvoiceResponse, PlisioOperationsResponse, PlisioService, PlisioWebhookPayload, PlisioWhiteLabelInvoice, RawMessage (+6 more)

### Community 16 - "Post"
Cohesion: 0.16
Nodes (4): FeedCursor, SearchFilterParams, AdminPostFilter, Post

### Community 18 - "MediaService"
Cohesion: 0.18
Nodes (12): Config, Image, MediaService, Reader, getEnv(), getEnvInt(), Load(), Client (+4 more)

### Community 20 - "FeedSkeleton.tsx"
Cohesion: 0.19
Nodes (4): FeedSkeleton(), PostCardSkeleton(), PostDetailSkeleton(), ProfileSkeleton()

### Community 21 - "ComposerContext.tsx"
Cohesion: 0.15
Nodes (10): metadata, ComposerContext, ComposerContextValue, ComposerProvider(), useComposer(), AppShell(), MediaViewerProvider(), PhotoViewerModal() (+2 more)

### Community 22 - "search/page.tsx"
Cohesion: 0.18
Nodes (10): ExploreCategory, SearchPageContent(), SearchPostsResponse, SearchTab, SearchUsersResponse, VirtualizedFeed(), VirtualizedFeedProps, feedKeys (+2 more)

### Community 23 - "admin.go"
Cohesion: 0.16
Nodes (13): AdminDeletePostRequest, AssignMembershipRequest, CreateRoleRequest, SetPermissionsRequest, SetPlanEntitlementsRequest, UpdatePlanRequest, UpdatePostModerationRequest, UpdateRoleRequest (+5 more)

### Community 24 - "handler.go"
Cohesion: 0.18
Nodes (12): CheckoutRequest, CreatePostRequest, FeedFeedbackRequest, FeedImpressionsRequest, GoogleAuthRequest, LoginRequest, RegisterRequest, SimulatePaymentRequest (+4 more)

### Community 25 - "OtherSkeletons.tsx"
Cohesion: 0.23
Nodes (3): ExploreSkeleton(), MessagesSkeleton(), NotificationsSkeleton()

### Community 27 - "VideoControls.tsx"
Cohesion: 0.31
Nodes (7): formatTime(), SPEED_OPTIONS, VideoControls(), VideoControlsProps, formatTime(), VideoProgressBar(), VideoProgressBarProps

### Community 28 - "Engine"
Cohesion: 0.31
Nodes (7): DiversityReranker, FeedOptions, FeedResult, CandidateProvider, Engine, Context, NewEngine()

### Community 29 - "PostComposer.tsx"
Cohesion: 0.36
Nodes (6): PostComposer(), PostComposerProps, StagedAttachment, filterValidMediaFiles(), isValidMediaFile(), X_ALLOWED_MEDIA_EXTENSIONS

### Community 32 - "NewPostgresRepository"
Cohesion: 0.29
Nodes (3): categorizeTopic(), NewPostgresRepository(), TrendingTopic

## Knowledge Gaps
- **124 isolated node(s):** `name`, `version`, `private`, `dev`, `build` (+119 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **102 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `GetUserFromContext()` connect `Hub` to `models.go`?**
  _High betweenness centrality (0.152) - this node is a cross-community bridge._
- **Why does `User` connect `models.go` to `Hub`?**
  _High betweenness centrality (0.148) - this node is a cross-community bridge._
- **Why does `MemoryRepository` connect `MemoryRepository` to `Invoice`, `T`, `Notification`, `Context`, `Hub`, `Post`, `User`, `.GetActiveSubscription`, `.GetPlanEntitlements`?**
  _High betweenness centrality (0.130) - this node is a cross-community bridge._
- **What connects `name`, `version`, `private` to the rest of the system?**
  _124 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `Handler` be split into smaller, more focused modules?**
  _Cohesion score 0.08090322580645161 - nodes in this community are weakly interconnected._
- **Should `PostgresRepository` be split into smaller, more focused modules?**
  _Cohesion score 0.06361570918532944 - nodes in this community are weakly interconnected._
- **Should `T` be split into smaller, more focused modules?**
  _Cohesion score 0.06924882629107981 - nodes in this community are weakly interconnected._