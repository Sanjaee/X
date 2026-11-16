# Graph Report - .  (2026-11-16)

## Corpus Check
- cluster-only mode — file stats not available

## Summary
- 1312 nodes · 3469 edges · 163 communities (48 shown, 115 thin omitted)
- Extraction: 97% EXTRACTED · 3% INFERRED · 0% AMBIGUOUS · INFERRED: 117 edges (avg confidence: 0.8)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `fb721562`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- writeJSON
- PostgresRepository
- MockCandidateProvider
- models.go
- Context
- PostgresRepository
- dependencies
- MemoryRepository
- [postId]/page.tsx
- index.ts
- app/page.tsx
- api.ts
- PostModal.tsx
- compilerOptions
- CryptoGateway
- totp.go
- Hub
- Post
- .CheckoutSubscription
- main
- UserNameWithRole.tsx
- Post
- T
- search/page.tsx
- MediaService
- User
- NewHandler
- FeedSkeleton.tsx
- admin.go
- NewMemoryRepository
- Repository
- OtherSkeletons.tsx
- qrcode.tsx
- User
- handler.go
- VideoControls.tsx
- .GetActiveSubscription
- setupTestApp
- UpgradeModalContext.tsx
- Role
- .GetPlanEntitlements
- NewPostgresRepository
- route.ts
- TopLoadingBar.tsx
- .DismissUserRecommendation
- .GetActiveSubscription
- .GetNotifications
- .GetNotifications
- middleware.ts
- .GetUserBadges
- .DismissUserRecommendation
- .GetFollowRequests
- .GetMediaObjectByID
- .GetMessages
- .GetSubscriptionPlanByID
- test_all_apis.py
- next.config.js
- next-env.d.ts
- tailwind.config.ts
- backup-db.sh
- monitoring.sh
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
- Message
- Notification
- OutboxEvent
- github.com/x-reconstruction/api
- PlisioService
- RankingWeights
- RecentSearch
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
- AdminOverview
- AdminPostFilter
- AuditLog
- Context
- MembershipEntitlement
- MembershipPlan
- Permission
- Post
- Role
- User
- UserMembership
- UUID
- Session
- SocialRelationship
- Subscription
- SubscriptionPlan
- TrendingTopic
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
- `setupTestRouter()` --calls--> `NewMemoryRepository()`  [INFERRED]
  services/api/internal/handler/handler_test.go → services/api/internal/repository/memory.go
- `TestSearchAPIEndpoints()` --calls--> `NewMemoryRepository()`  [INFERRED]
  services/api/internal/handler/search_test.go → services/api/internal/repository/memory.go
- `setupTestApp()` --calls--> `NewHandler()`  [INFERRED]
  services/api/internal/handler/social_messaging_test.go → services/api/internal/handler/handler.go
- `setupTestApp()` --calls--> `AuthMiddleware()`  [INFERRED]
  services/api/internal/handler/social_messaging_test.go → services/api/internal/middleware/middleware.go

## Import Cycles
- None detected.

## Communities (163 total, 115 thin omitted)

### Community 0 - "writeJSON"
Cohesion: 0.09
Nodes (10): Context, PaymentOrder, Handler, Request, ResponseWriter, Handler, Request, ResponseWriter (+2 more)

### Community 1 - "PostgresRepository"
Cohesion: 0.08
Nodes (8): DB, Context, Conversation, PostgresRepository, SocialRelationship, Time, UserRecommendation, UUID

### Community 2 - "MockCandidateProvider"
Cohesion: 0.05
Nodes (43): B, DiversityReranker, RankingWeights, FeedOptions, FeedResult, Candidate, CandidateProvider, CandidateSource (+35 more)

### Community 3 - "models.go"
Cohesion: 0.08
Nodes (59): AdminOverview, AdminPostFilter, AuditLog, Conversation, CounterMismatchDetail, CounterReconciliationResult, DatabaseHealth, Entitlements (+51 more)

### Community 4 - "Context"
Cohesion: 0.06
Nodes (5): PaymentOrder, Context, SocialRelationship, UserRecommendation, UUID

### Community 5 - "PostgresRepository"
Cohesion: 0.08
Nodes (12): AdminOverview, AdminPostFilter, AuditLog, MembershipEntitlement, MembershipPlan, Permission, Post, PostgresRepository (+4 more)

### Community 6 - "dependencies"
Cohesion: 0.05
Nodes (42): dependencies, clsx, lucide-react, next, next-auth, react, react-dom, tailwind-merge (+34 more)

### Community 7 - "MemoryRepository"
Cohesion: 0.07
Nodes (12): MemoryRepository, RWMutex, AuditLog, Conversation, FollowRequest, MediaObject, Message, Permission (+4 more)

### Community 8 - "[postId]/page.tsx"
Cohesion: 0.11
Nodes (23): StagedFile, PostModal(), PostCard(), PostCardProps, MediaDetailSidebar(), MediaDetailSidebarProps, MediaViewerContext, MediaViewerContextType (+15 more)

### Community 9 - "index.ts"
Cohesion: 0.08
Nodes (28): AdminTab, ConfirmModal(), ConfirmModalProps, TIER_PRESENTATION, TierBadge(), TierBadgeProps, TierKey, RealtimeContext (+20 more)

### Community 10 - "app/page.tsx"
Cohesion: 0.16
Nodes (18): encodeFeedCursor(), FeedResponse, HomePage(), SettingsSection, AuthModal(), useComposer(), MobileComposerModal(), LeftNavRail() (+10 more)

### Community 11 - "api.ts"
Cohesion: 0.12
Nodes (22): metadata, AuthModalProps, ComposerContext, ComposerContextValue, ComposerProvider(), AppShell(), MediaViewerProvider(), PhotoViewerModal() (+14 more)

### Community 12 - "PostModal.tsx"
Cohesion: 0.20
Nodes (14): MediaUrlInputModal(), MediaUrlInputModalProps, MobileComposerModalProps, StagedAttachment, PostComposer(), PostComposerProps, StagedAttachment, PostModalProps (+6 more)

### Community 13 - "compilerOptions"
Cohesion: 0.07
Nodes (26): compilerOptions, allowJs, esModuleInterop, incremental, isolatedModules, jsx, lib, module (+18 more)

### Community 14 - "CryptoGateway"
Cohesion: 0.13
Nodes (16): Config, Duration, memoryLimiter, Mutex, CryptoCurrency, CryptoGateway, CryptoInvoice, CryptoOperation (+8 more)

### Community 15 - "totp.go"
Cohesion: 0.17
Nodes (19): Handler, mfaTokenPayload, Request, ResponseWriter, buildOTPAuthURL(), checkTOTPWindow(), computeTOTPCode(), createMFAToken() (+11 more)

### Community 16 - "Hub"
Cohesion: 0.16
Nodes (13): RawMessage, Client, Hub, InboundEvent, OutboundEvent, Conn, Context, Request (+5 more)

### Community 17 - "Post"
Cohesion: 0.17
Nodes (3): Rows, FeedCursor, Post

### Community 18 - ".CheckoutSubscription"
Cohesion: 0.27
Nodes (12): CryptoInvoice, CryptoOperation, CheckoutRequest, expired(), Handler, Request, ResponseWriter, UUID (+4 more)

### Community 19 - "main"
Cohesion: 0.13
Nodes (17): Handler, Logger, contextKey, responseWriterWrapper, ReadWriter, main(), AuthMiddleware(), AuthMiddlewareWithCache() (+9 more)

### Community 20 - "UserNameWithRole.tsx"
Cohesion: 0.15
Nodes (14): EditProfileModal(), EditProfileModalProps, NAME_STYLE_PRESETS, TwoFactorManager(), TwoFactorManagerProps, buildCustomNameInlineStyle(), parseCustomNameStyle(), UserNameWithRole() (+6 more)

### Community 22 - "T"
Cohesion: 0.32
Nodes (18): Mux, createTestPost(), Post, registerTestUserInInteractions(), setupInteractionsTestRouter(), TestBookmarkFlowAndLikesTab(), TestCounterReconciliation(), TestLikeIdempotencyAndConcurrency() (+10 more)

### Community 23 - "search/page.tsx"
Cohesion: 0.13
Nodes (16): ExploreCategory, SearchPageContent(), SearchPostsResponse, SearchTab, SearchUsersResponse, RightSidebarProps, RecentSearchItem, SearchBarDropdown() (+8 more)

### Community 24 - "MediaService"
Cohesion: 0.18
Nodes (12): Config, Image, MediaService, Reader, getEnv(), getEnvInt(), Load(), Client (+4 more)

### Community 26 - "NewHandler"
Cohesion: 0.17
Nodes (14): Client, CryptoGateway, Hub, MediaService, Repository, NewHandler(), Mux, T (+6 more)

### Community 27 - "FeedSkeleton.tsx"
Cohesion: 0.19
Nodes (4): FeedSkeleton(), PostCardSkeleton(), PostDetailSkeleton(), ProfileSkeleton()

### Community 28 - "admin.go"
Cohesion: 0.16
Nodes (13): AdminDeletePostRequest, AssignMembershipRequest, CreateRoleRequest, SetPermissionsRequest, SetPlanEntitlementsRequest, UpdatePlanRequest, UpdatePostModerationRequest, UpdateRoleRequest (+5 more)

### Community 29 - "NewMemoryRepository"
Cohesion: 0.34
Nodes (11): Handler, User, registerAndLoginUser(), setupTestRBACRouter(), TestConcurrentPostingCannotBypassQuota(), TestMediaUploadLimitsByTier(), TestNewUserDefaultMemberRoleAndAdminProtection(), TestOwnerBootstrapAndNonDemotionRule() (+3 more)

### Community 30 - "Repository"
Cohesion: 0.25
Nodes (11): Repository, setupTestFullAdminRouter(), TestAdminPostModerationAndAuditLogging(), TestOwnerAdminOverviewSecurity(), AuthorizeResource(), User, UUID, IsOwnerAccount() (+3 more)

### Community 31 - "OtherSkeletons.tsx"
Cohesion: 0.23
Nodes (3): ExploreSkeleton(), MessagesSkeleton(), NotificationsSkeleton()

### Community 32 - "qrcode.tsx"
Cohesion: 0.24
Nodes (11): ALIGNMENT_PATTERNS, EC_CODEWORDS_L, encodeQRMatrix(), GF_EXP, GF_LOG, gfMul(), NUM_BLOCKS_L, QRCodeSVG() (+3 more)

### Community 34 - "handler.go"
Cohesion: 0.22
Nodes (10): CreatePostRequest, FeedFeedbackRequest, FeedImpressionsRequest, GoogleAuthRequest, LoginRequest, RegisterRequest, SimulatePaymentRequest, UpdatePostRequest (+2 more)

### Community 35 - "VideoControls.tsx"
Cohesion: 0.31
Nodes (7): formatTime(), SPEED_OPTIONS, VideoControls(), VideoControlsProps, formatTime(), VideoProgressBar(), VideoProgressBarProps

### Community 36 - ".GetActiveSubscription"
Cohesion: 0.22
Nodes (3): Entitlements, Subscription, UserMembership

### Community 37 - "setupTestApp"
Cohesion: 0.47
Nodes (8): createTestUser(), Mux, T, User, setupTestApp(), TestDirectMessaging_IdempotencyAndFlow(), TestFollowRequest_PrivateAccount(), TestSocialGraph_FollowUnfollowBlockMute()

### Community 38 - "UpgradeModalContext.tsx"
Cohesion: 0.32
Nodes (6): VerifiedBadge(), VerifiedBadgeProps, UpgradeModalContext, UpgradeModalContextType, UpgradeModalProvider(), useUpgradeModal()

### Community 41 - "NewPostgresRepository"
Cohesion: 0.29
Nodes (3): categorizeTopic(), TrendingTopic, NewPostgresRepository()

### Community 42 - "route.ts"
Cohesion: 0.60
Nodes (3): getDynamicAuthOptions(), handler(), authOptions

## Knowledge Gaps
- **131 isolated node(s):** `name`, `version`, `private`, `dev`, `build` (+126 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **115 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `MemoryRepository` connect `MemoryRepository` to `MockCandidateProvider`, `models.go`, `Context`, `.GetActiveSubscription`, `Role`, `.GetPlanEntitlements`, `.DismissUserRecommendation`, `.GetNotifications`, `Post`, `User`, `NewMemoryRepository`?**
  _High betweenness centrality (0.143) - this node is a cross-community bridge._
- **Why does `NewHandler()` connect `NewHandler` to `writeJSON`, `handler.go`, `setupTestApp`, `main`, `T`, `NewMemoryRepository`?**
  _High betweenness centrality (0.118) - this node is a cross-community bridge._
- **Why does `NewMemoryRepository()` connect `NewMemoryRepository` to `MockCandidateProvider`, `setupTestApp`, `MemoryRepository`, `main`, `T`, `NewHandler`, `Repository`?**
  _High betweenness centrality (0.109) - this node is a cross-community bridge._
- **What connects `name`, `version`, `private` to the rest of the system?**
  _131 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `writeJSON` be split into smaller, more focused modules?**
  _Cohesion score 0.09343763581051717 - nodes in this community are weakly interconnected._
- **Should `PostgresRepository` be split into smaller, more focused modules?**
  _Cohesion score 0.07785547785547786 - nodes in this community are weakly interconnected._
- **Should `MockCandidateProvider` be split into smaller, more focused modules?**
  _Cohesion score 0.05048076923076923 - nodes in this community are weakly interconnected._