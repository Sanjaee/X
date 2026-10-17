# Graph Report - .  (2026-10-17)

## Corpus Check
- cluster-only mode — file stats not available

## Summary
- 1043 nodes · 2783 edges · 100 communities (44 shown, 56 thin omitted)
- Extraction: 98% EXTRACTED · 2% INFERRED · 0% AMBIGUOUS · INFERRED: 52 edges (avg confidence: 0.8)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `c67c91bb`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- Handler
- .GenerateFeed
- Context
- models.go
- PostgresRepository
- dependencies
- PostgresRepository
- MemoryRepository
- apiRequest
- User
- Handler
- compilerOptions
- PlisioService
- index.ts
- .GetUserBadges
- Client
- search/page.tsx
- api.ts
- messages/page.tsx
- middleware.go
- Post
- Post
- User
- MediaService
- NewMemoryRepository
- admin.go
- handler.go
- StandalonePhotoView.tsx
- User
- Repository
- VideoControls.tsx
- .GetActiveSubscription
- Role
- setupTestApp
- NewHandler
- responseWriterWrapper
- .GetPlanEntitlements
- setupTestRouter
- setupRecommendationRouter
- .GetActiveSubscription
- route.ts
- .GetNotifications
- setupTestFullAdminRouter
- middleware.ts
- .GetFollowRequests
- .GetMediaObjectByID
- .GetMessages
- test_all_apis.py
- next.config.js
- next-env.d.ts
- tailwind.config.ts
- .GetSocialRelationship
- backup-db.sh
- monitoring.sh
- Context
- FeedFeedback
- FeedImpression
- Invoice
- Message
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
- Handler
- Conversation
- Entitlements
- FeedCursor
- FollowRequest
- MediaObject
- Notification
- RWMutex
- SocialRelationship
- Subscription
- SubscriptionPlan
- VerificationBadge
- Session
- TrendingTopic
- User
- UserTopicPreference
- UUID
- VerificationApplication
- .ensureSearchSchema
- .GetSubscriptionPlanByID

## God Nodes (most connected - your core abstractions)
1. `MemoryRepository` - 164 edges
2. `PostgresRepository` - 99 edges
3. `Handler` - 86 edges
4. `writeJSON()` - 76 edges
5. `writeError()` - 73 edges
6. `apiRequest()` - 38 edges
7. `PostgresRepository` - 36 edges
8. `Handler` - 31 edges
9. `User` - 30 edges
10. `cn()` - 25 edges

## Surprising Connections (you probably didn't know these)
- `SuggestionData` --references--> `User`  [EXTRACTED]
  apps/web/src/components/search/SearchBarDropdown.tsx → apps/web/src/types/index.ts
- `setupTestRouter()` --calls--> `NewHandler()`  [INFERRED]
  services/api/internal/handler/handler_test.go → services/api/internal/handler/handler.go
- `setupTestRouter()` --calls--> `AuthMiddleware()`  [INFERRED]
  services/api/internal/handler/handler_test.go → services/api/internal/middleware/middleware.go
- `setupTestRouter()` --calls--> `NewMemoryRepository()`  [INFERRED]
  services/api/internal/handler/handler_test.go → services/api/internal/repository/memory.go
- `setupRecommendationRouter()` --calls--> `NewHandler()`  [INFERRED]
  services/api/internal/handler/recommendation_test.go → services/api/internal/handler/handler.go

## Import Cycles
- None detected.

## Communities (100 total, 56 thin omitted)

### Community 0 - "Handler"
Cohesion: 0.14
Nodes (6): Context, Handler, Request, ResponseWriter, writeError(), writeJSON()

### Community 1 - ".GenerateFeed"
Cohesion: 0.06
Nodes (43): B, RankingWeights, Candidate, CandidateProvider, CandidateSource, DiversityReranker, Engine, Features (+35 more)

### Community 2 - "Context"
Cohesion: 0.06
Nodes (4): Conversation, Context, UUID, SocialRelationship

### Community 3 - "models.go"
Cohesion: 0.09
Nodes (55): AdminOverview, AuditLog, Conversation, DatabaseHealth, Entitlements, FeedFeedback, FeedImpression, Follow (+47 more)

### Community 4 - "PostgresRepository"
Cohesion: 0.10
Nodes (13): AdminPostFilter, PostgresRepository, AdminOverview, AuditLog, Context, MembershipEntitlement, MembershipPlan, Permission (+5 more)

### Community 5 - "dependencies"
Cohesion: 0.05
Nodes (42): dependencies, clsx, lucide-react, next, next-auth, react, react-dom, tailwind-merge (+34 more)

### Community 6 - "PostgresRepository"
Cohesion: 0.09
Nodes (5): DB, Context, Conversation, PostgresRepository, UUID

### Community 7 - "MemoryRepository"
Cohesion: 0.08
Nodes (12): FollowRequest, MediaObject, Notification, MemoryRepository, RWMutex, AdminOverview, AuditLog, Message (+4 more)

### Community 8 - "apiRequest"
Cohesion: 0.17
Nodes (22): BookmarksPage(), ExplorePage(), InvoiceContent(), InvoicePage(), WhiteLabelInvoice, PremiumPage(), FollowersPage(), FollowingPage() (+14 more)

### Community 9 - "User"
Cohesion: 0.16
Nodes (19): AuthModalProps, PostCard(), PostCardProps, MediaDetailSidebar(), MediaDetailSidebarProps, MediaViewerContext, MediaViewerContextType, MediaViewerProvider() (+11 more)

### Community 10 - "Handler"
Cohesion: 0.19
Nodes (3): Handler, Request, ResponseWriter

### Community 11 - "compilerOptions"
Cohesion: 0.07
Nodes (26): compilerOptions, allowJs, esModuleInterop, incremental, isolatedModules, jsx, lib, module (+18 more)

### Community 12 - "PlisioService"
Cohesion: 0.14
Nodes (16): Config, PlisioCryptoInfo, PlisioInvoiceResponse, PlisioOperationsResponse, PlisioService, PlisioWebhookPayload, PlisioWhiteLabelInvoice, getEnv() (+8 more)

### Community 13 - "index.ts"
Cohesion: 0.12
Nodes (20): AdminDashboardPage(), AdminTab, TIER_PRESENTATION, TierBadge(), TierBadgeProps, TierKey, AdminOverview, AuditLog (+12 more)

### Community 15 - "Client"
Cohesion: 0.17
Nodes (13): Client, Hub, InboundEvent, OutboundEvent, Conn, Context, RawMessage, Request (+5 more)

### Community 16 - "search/page.tsx"
Cohesion: 0.13
Nodes (15): FeedResponse, ExploreCategory, SearchPageContent(), SearchPostsResponse, SearchTab, SearchUsersResponse, MobileComposerModal(), MobileComposerModalProps (+7 more)

### Community 17 - "api.ts"
Cohesion: 0.13
Nodes (13): metadata, ComposerContext, ComposerContextValue, ComposerProvider(), PostModal(), PostModalProps, StagedAttachment, PhotoViewerModal() (+5 more)

### Community 18 - "messages/page.tsx"
Cohesion: 0.17
Nodes (13): MobileBottomNav(), MobileBottomNavProps, MobileDrawer(), MobileDrawerProps, RealtimeContext, RealtimeContextType, ToastNotification, TypingEvent (+5 more)

### Community 19 - "middleware.go"
Cohesion: 0.14
Nodes (15): Duration, Logger, contextKey, memoryLimiter, Mutex, AuthMiddleware(), GetUserFromContext(), GetViewerID() (+7 more)

### Community 20 - "Post"
Cohesion: 0.18
Nodes (3): Rows, FeedCursor, Post

### Community 23 - "MediaService"
Cohesion: 0.22
Nodes (10): Config, Image, MediaService, Reader, Client, Context, MediaObject, UUID (+2 more)

### Community 24 - "NewMemoryRepository"
Cohesion: 0.36
Nodes (12): Handler, User, registerAndLoginUser(), setupTestRBACRouter(), TestConcurrentPostingCannotBypassQuota(), TestMediaUploadLimitsByTier(), TestNewUserDefaultMemberRoleAndAdminProtection(), TestOwnerBootstrapAndNonDemotionRule() (+4 more)

### Community 25 - "admin.go"
Cohesion: 0.16
Nodes (13): AdminDeletePostRequest, AssignMembershipRequest, CreateRoleRequest, SetPermissionsRequest, SetPlanEntitlementsRequest, UpdatePlanRequest, UpdatePostModerationRequest, UpdateRoleRequest (+5 more)

### Community 26 - "handler.go"
Cohesion: 0.18
Nodes (12): CheckoutRequest, CreatePostRequest, FeedFeedbackRequest, FeedImpressionsRequest, GoogleAuthRequest, LoginRequest, RegisterRequest, SimulatePaymentRequest (+4 more)

### Community 27 - "StandalonePhotoView.tsx"
Cohesion: 0.26
Nodes (4): PhotoCanvas(), PhotoCanvasProps, StandalonePhotoView(), MediaObject

### Community 29 - "Repository"
Cohesion: 0.27
Nodes (9): Repository, main(), AuthorizeResource(), User, UUID, IsOwnerAccount(), RequireAnyPermission(), RequireOwner() (+1 more)

### Community 30 - "VideoControls.tsx"
Cohesion: 0.31
Nodes (7): formatTime(), SPEED_OPTIONS, VideoControls(), VideoControlsProps, formatTime(), VideoProgressBar(), VideoProgressBarProps

### Community 31 - ".GetActiveSubscription"
Cohesion: 0.22
Nodes (3): Entitlements, UserMembership, Subscription

### Community 33 - "setupTestApp"
Cohesion: 0.47
Nodes (8): createTestUser(), Mux, T, User, setupTestApp(), TestDirectMessaging_IdempotencyAndFlow(), TestFollowRequest_PrivateAccount(), TestSocialGraph_FollowUnfollowBlockMute()

### Community 34 - "NewHandler"
Cohesion: 0.25
Nodes (7): Client, Hub, MediaService, PlisioService, NewHandler(), T, TestSearchAPIEndpoints()

### Community 35 - "responseWriterWrapper"
Cohesion: 0.33
Nodes (4): responseWriterWrapper, ReadWriter, Conn, ResponseWriter

### Community 37 - "setupTestRouter"
Cohesion: 0.52
Nodes (6): Mux, T, setupTestRouter(), TestAuthAndPostingFlow(), TestHealthCheck(), TestSubscriptionsAndBadges()

### Community 38 - "setupRecommendationRouter"
Cohesion: 0.60
Nodes (5): Mux, T, registerTestUser(), setupRecommendationRouter(), TestRecommendationAndFeedAPIs()

### Community 42 - "setupTestFullAdminRouter"
Cohesion: 0.83
Nodes (3): setupTestFullAdminRouter(), TestAdminPostModerationAndAuditLogging(), TestOwnerAdminOverviewSecurity()

## Knowledge Gaps
- **109 isolated node(s):** `name`, `version`, `private`, `dev`, `build` (+104 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **56 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `NewMemoryRepository()` connect `NewMemoryRepository` to `setupTestApp`, `NewHandler`, `.GenerateFeed`, `setupTestRouter`, `setupRecommendationRouter`, `MemoryRepository`, `setupTestFullAdminRouter`, `Repository`?**
  _High betweenness centrality (0.204) - this node is a cross-community bridge._
- **Why does `MemoryRepository` connect `MemoryRepository` to `Role`, `.GenerateFeed`, `Context`, `models.go`, `.GetPlanEntitlements`, `Post`, `User`, `NewMemoryRepository`, `.GetActiveSubscription`?**
  _High betweenness centrality (0.185) - this node is a cross-community bridge._
- **Why does `DefaultRankingWeights()` connect `.GenerateFeed` to `NewMemoryRepository`, `models.go`?**
  _High betweenness centrality (0.116) - this node is a cross-community bridge._
- **What connects `name`, `version`, `private` to the rest of the system?**
  _109 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `Handler` be split into smaller, more focused modules?**
  _Cohesion score 0.1392156862745098 - nodes in this community are weakly interconnected._
- **Should `.GenerateFeed` be split into smaller, more focused modules?**
  _Cohesion score 0.05888376856118792 - nodes in this community are weakly interconnected._
- **Should `Context` be split into smaller, more focused modules?**
  _Cohesion score 0.05989110707803993 - nodes in this community are weakly interconnected._