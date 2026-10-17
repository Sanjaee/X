# Graph Report - .  (2026-10-17)

## Corpus Check
- cluster-only mode — file stats not available

## Summary
- 1132 nodes · 2948 edges · 114 communities (38 shown, 76 thin omitted)
- Extraction: 99% EXTRACTED · 1% INFERRED · 0% AMBIGUOUS · INFERRED: 41 edges (avg confidence: 0.8)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `af2553c4`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- Handler
- PostgresRepository
- T
- MockCandidateProvider
- models.go
- Context
- apiRequest
- MemoryRepository
- Context
- dependencies
- [postId]/page.tsx
- search/page.tsx
- compilerOptions
- PlisioService
- middleware.go
- Hub
- Post
- Post
- utils.ts
- index.ts
- MediaService
- admin.go
- ComposerContext.tsx
- handler.go
- VideoControls.tsx
- .GetActiveSubscription
- Role
- AuditLog
- .GetPlanEntitlements
- .GetBillingHistory
- .GetBillingHistory
- .GetTrendingTopics
- .GetActiveSubscription
- route.ts
- .GetNotifications
- .GetActiveWeights
- middleware.ts
- .GetAdminPosts
- .GetRolePermissions
- .GetMediaObjectByID
- .GetVerificationApplication
- .GetUserTopicPreferences
- .CreateOutboxEvent
- .CreateSession
- .GetSubscriptionPlanByID
- test_all_apis.py
- next.config.js
- next-env.d.ts
- tailwind.config.ts
- .GetTrendingTopics
- .ReconcileCounters
- .RecordFeedFeedback
- backup-db.sh
- monitoring.sh
- Conversation
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
- github.com/x-reconstruction/api
- Post
- RankingWeights
- RecentSearch
- Repository
- SearchFilterParams
- Request
- ResponseWriter
- Client
- Request
- ResponseWriter
- Mux
- T
- Client
- Context
- MediaObject
- UUID
- Handler
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
- Time
- UserMembership
- Session
- SocialRelationship
- Subscription
- SubscriptionPlan
- TrendingTopic
- User
- UserTopicPreference
- VerificationApplication
- VerificationBadge

## God Nodes (most connected - your core abstractions)
1. `MemoryRepository` - 169 edges
2. `PostgresRepository` - 142 edges
3. `Handler` - 122 edges
4. `writeJSON()` - 81 edges
5. `writeError()` - 78 edges
6. `apiRequest()` - 36 edges
7. `User` - 30 edges
8. `cn()` - 24 edges
9. `Post` - 18 edges
10. `NewMemoryRepository()` - 17 edges

## Surprising Connections (you probably didn't know these)
- `setupTestRouter()` --calls--> `AuthMiddleware()`  [INFERRED]
  services/api/internal/handler/handler_test.go → services/api/internal/middleware/middleware.go
- `setupTestApp()` --calls--> `AuthMiddleware()`  [INFERRED]
  services/api/internal/handler/social_messaging_test.go → services/api/internal/middleware/middleware.go
- `setupTestApp()` --calls--> `NewHub()`  [INFERRED]
  services/api/internal/handler/social_messaging_test.go → services/api/internal/realtime/hub.go
- `PhotoCanvasProps` --references--> `MediaObject`  [EXTRACTED]
  apps/web/src/components/media/PhotoCanvas.tsx → apps/web/src/types/index.ts
- `main()` --calls--> `Handler`  [INFERRED]
  services/api/cmd/server/main.go → services/api/internal/handler/handler.go

## Import Cycles
- None detected.

## Communities (114 total, 76 thin omitted)

### Community 0 - "Handler"
Cohesion: 0.09
Nodes (6): Handler, Request, ResponseWriter, Context, writeError(), writeJSON()

### Community 1 - "PostgresRepository"
Cohesion: 0.07
Nodes (17): DB, PostgresRepository, Context, Conversation, Duration, FollowRequest, MediaObject, Message (+9 more)

### Community 2 - "T"
Cohesion: 0.07
Nodes (60): Hub, MediaService, Mux, PlisioService, Repository, main(), setupTestFullAdminRouter(), TestAdminPostModerationAndAuditLogging() (+52 more)

### Community 3 - "MockCandidateProvider"
Cohesion: 0.05
Nodes (43): B, DiversityReranker, RankingWeights, FeedOptions, FeedResult, Candidate, CandidateProvider, CandidateSource (+35 more)

### Community 4 - "models.go"
Cohesion: 0.08
Nodes (59): AdminOverview, AuditLog, Conversation, CounterMismatchDetail, CounterReconciliationResult, DatabaseHealth, Entitlements, FeedFeedback (+51 more)

### Community 5 - "Context"
Cohesion: 0.07
Nodes (6): Context, Conversation, Duration, SocialRelationship, UserRecommendation, UUID

### Community 6 - "apiRequest"
Cohesion: 0.13
Nodes (32): AdminDashboardPage(), InvoiceContent(), InvoicePage(), WhiteLabelInvoice, PremiumPage(), FollowersPage(), FollowingPage(), AuthModal() (+24 more)

### Community 7 - "MemoryRepository"
Cohesion: 0.08
Nodes (13): MemoryRepository, FeedFeedback, FeedImpression, FollowRequest, Message, OutboxEvent, RankingWeights, RecentSearch (+5 more)

### Community 8 - "Context"
Cohesion: 0.08
Nodes (10): AdminOverview, AuditLog, Context, MembershipEntitlement, MembershipPlan, Permission, Role, User (+2 more)

### Community 9 - "dependencies"
Cohesion: 0.05
Nodes (42): dependencies, clsx, lucide-react, next, next-auth, react, react-dom, tailwind-merge (+34 more)

### Community 10 - "[postId]/page.tsx"
Cohesion: 0.12
Nodes (23): StagedFile, MobileComposerModalProps, StagedAttachment, PostComposer(), PostComposerProps, StagedAttachment, PostModal(), PostModalProps (+15 more)

### Community 11 - "search/page.tsx"
Cohesion: 0.10
Nodes (24): encodeFeedCursor(), FeedResponse, HomePage(), ExploreCategory, SearchPageContent(), SearchPostsResponse, SearchTab, SearchUsersResponse (+16 more)

### Community 12 - "compilerOptions"
Cohesion: 0.07
Nodes (26): compilerOptions, allowJs, esModuleInterop, incremental, isolatedModules, jsx, lib, module (+18 more)

### Community 13 - "PlisioService"
Cohesion: 0.14
Nodes (16): Config, PlisioCryptoInfo, PlisioInvoiceResponse, PlisioOperationsResponse, PlisioService, PlisioWebhookPayload, PlisioWhiteLabelInvoice, getEnv() (+8 more)

### Community 14 - "middleware.go"
Cohesion: 0.10
Nodes (19): Duration, Logger, contextKey, memoryLimiter, responseWriterWrapper, Mutex, ReadWriter, AuthMiddleware() (+11 more)

### Community 15 - "Hub"
Cohesion: 0.16
Nodes (13): Conn, RawMessage, Client, Hub, InboundEvent, OutboundEvent, RWMutex, Context (+5 more)

### Community 16 - "Post"
Cohesion: 0.16
Nodes (4): Rows, FeedCursor, Post, SearchFilterParams

### Community 17 - "Post"
Cohesion: 0.16
Nodes (4): AdminPostFilter, FeedCursor, Post, SearchFilterParams

### Community 18 - "utils.ts"
Cohesion: 0.16
Nodes (10): MediaDetailSidebar(), MediaDetailSidebarProps, PhotoCanvas(), PhotoCanvasProps, PhotoNavigation(), PhotoNavigationProps, StandalonePhotoView(), formatFullTimestamp() (+2 more)

### Community 19 - "index.ts"
Cohesion: 0.14
Nodes (17): AdminTab, TIER_PRESENTATION, TierBadge(), TierBadgeProps, TierKey, AdminOverview, AuditLog, FollowRequest (+9 more)

### Community 20 - "MediaService"
Cohesion: 0.22
Nodes (10): Client, Config, Context, Image, MediaService, MediaObject, Reader, NewMediaService() (+2 more)

### Community 21 - "admin.go"
Cohesion: 0.16
Nodes (13): AdminDeletePostRequest, AssignMembershipRequest, CreateRoleRequest, SetPermissionsRequest, SetPlanEntitlementsRequest, UpdatePlanRequest, UpdatePostModerationRequest, UpdateRoleRequest (+5 more)

### Community 22 - "ComposerContext.tsx"
Cohesion: 0.18
Nodes (8): metadata, ComposerContext, ComposerContextValue, ComposerProvider(), MediaViewerProvider(), PhotoViewerModal(), AppProviders(), RealtimeProvider()

### Community 23 - "handler.go"
Cohesion: 0.18
Nodes (12): CheckoutRequest, CreatePostRequest, FeedFeedbackRequest, FeedImpressionsRequest, GoogleAuthRequest, LoginRequest, RegisterRequest, SimulatePaymentRequest (+4 more)

### Community 24 - "VideoControls.tsx"
Cohesion: 0.31
Nodes (7): formatTime(), SPEED_OPTIONS, VideoControls(), VideoControlsProps, formatTime(), VideoProgressBar(), VideoProgressBarProps

### Community 25 - ".GetActiveSubscription"
Cohesion: 0.22
Nodes (3): Entitlements, Subscription, UserMembership

### Community 27 - "AuditLog"
Cohesion: 0.18
Nodes (3): AdminOverview, AuditLog, CounterReconciliationResult

### Community 31 - ".GetTrendingTopics"
Cohesion: 0.33
Nodes (3): categorizeTopic(), TrendingTopic, NewPostgresRepository()

## Knowledge Gaps
- **113 isolated node(s):** `name`, `version`, `private`, `dev`, `build` (+108 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **76 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `PostgresRepository` connect `PostgresRepository` to `.GetActiveSubscription`, `.GetActiveWeights`, `.GetAdminPosts`, `Context`, `.CreateOutboxEvent`, `.CreateSession`, `.GetSubscriptionPlanByID`, `Post`, `.ReconcileCounters`, `.RecordFeedFeedback`, `.GetBillingHistory`, `.GetTrendingTopics`?**
  _High betweenness centrality (0.247) - this node is a cross-community bridge._
- **Why does `main()` connect `T` to `Handler`, `.GetTrendingTopics`?**
  _High betweenness centrality (0.195) - this node is a cross-community bridge._
- **Why does `NewPostgresRepository()` connect `.GetTrendingTopics` to `PostgresRepository`, `T`?**
  _High betweenness centrality (0.160) - this node is a cross-community bridge._
- **What connects `name`, `version`, `private` to the rest of the system?**
  _113 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `Handler` be split into smaller, more focused modules?**
  _Cohesion score 0.09070464767616192 - nodes in this community are weakly interconnected._
- **Should `PostgresRepository` be split into smaller, more focused modules?**
  _Cohesion score 0.06666666666666667 - nodes in this community are weakly interconnected._
- **Should `T` be split into smaller, more focused modules?**
  _Cohesion score 0.06846635367762129 - nodes in this community are weakly interconnected._