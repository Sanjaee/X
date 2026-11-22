# Graph Report - .  (2026-11-22)

## Corpus Check
- cluster-only mode — file stats not available

## Summary
- 1408 nodes · 3714 edges · 172 communities (48 shown, 124 thin omitted)
- Extraction: 97% EXTRACTED · 3% INFERRED · 0% AMBIGUOUS · INFERRED: 110 edges (avg confidence: 0.8)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `31d63bea`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- writeJSON
- T
- PostgresRepository
- models.go
- PostgresRepository
- MockCandidateProvider
- Context
- .CheckoutSubscription
- dependencies
- MemoryRepository
- [postId]/page.tsx
- Handler
- index.ts
- Context
- compilerOptions
- search/page.tsx
- CryptoGateway
- Hub
- utils.ts
- User
- api.ts
- Post
- Post
- PostModal.tsx
- UserNameWithRole.tsx
- MediaService
- User
- UpgradeModal.tsx
- FeedSkeleton.tsx
- admin.go
- OtherSkeletons.tsx
- qrcode.tsx
- User
- handler.go
- VideoControls.tsx
- .GetActiveSubscription
- Role
- MembershipPlan
- .GetTrendingTopics
- community.go
- route.ts
- TopLoadingBar.tsx
- .GetActiveSubscription
- .GetNotifications
- .GetNotifications
- middleware.ts
- Duration
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
- .GetTrendingTopics
- backup-db.sh
- monitoring.sh
- Conn
- CounterReconciliationResult
- FeedFeedback
- FeedImpression
- Invoice
- Message
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
- AdminOverview
- AdminPostFilter
- AuditLog
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
- MembershipPlan
- Notification
- OutboxEvent
- PaymentOrder
- Permission
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
- UserTopicPreference
- VerificationApplication

## God Nodes (most connected - your core abstractions)
1. `MemoryRepository` - 180 edges
2. `PostgresRepository` - 114 edges
3. `writeJSON()` - 109 edges
4. `writeError()` - 108 edges
5. `Handler` - 89 edges
6. `PostgresRepository` - 57 edges
7. `apiRequest()` - 36 edges
8. `User` - 34 edges
9. `Handler` - 26 edges
10. `cn()` - 25 edges

## Surprising Connections (you probably didn't know these)
- `setupTestRouter()` --calls--> `AuthMiddleware()`  [INFERRED]
  services/api/internal/handler/handler_test.go → services/api/internal/middleware/middleware.go
- `setupTestApp()` --calls--> `AuthMiddleware()`  [INFERRED]
  services/api/internal/handler/social_messaging_test.go → services/api/internal/middleware/middleware.go
- `setupTestApp()` --calls--> `NewHub()`  [INFERRED]
  services/api/internal/handler/social_messaging_test.go → services/api/internal/realtime/hub.go
- `RateLimiter()` --calls--> `GetUserFromContext()`  [INFERRED]
  services/api/internal/middleware/ratelimit.go → services/api/internal/middleware/middleware.go
- `SearchUsersResponse` --references--> `User`  [EXTRACTED]
  apps/web/src/app/search/page.tsx → apps/web/src/types/index.ts

## Import Cycles
- None detected.

## Communities (172 total, 124 thin omitted)

### Community 0 - "writeJSON"
Cohesion: 0.09
Nodes (10): Context, PaymentOrder, Handler, Request, ResponseWriter, Handler, Request, ResponseWriter (+2 more)

### Community 1 - "T"
Cohesion: 0.07
Nodes (62): Client, CryptoGateway, Hub, MediaService, Mux, Repository, Repository, setupTestFullAdminRouter() (+54 more)

### Community 2 - "PostgresRepository"
Cohesion: 0.06
Nodes (14): AdminOverview, AdminPostFilter, MembershipEntitlement, Post, PostgresRepository, Community, CommunityJoinRequest, CommunityMember (+6 more)

### Community 3 - "models.go"
Cohesion: 0.08
Nodes (65): AdminOverview, AdminPostFilter, AuditLog, Community, CommunityInvitation, CommunityJoinRequest, CommunityMember, CommunityPost (+57 more)

### Community 4 - "PostgresRepository"
Cohesion: 0.08
Nodes (8): DB, Context, Conversation, PostgresRepository, SocialRelationship, Time, UserRecommendation, UUID

### Community 5 - "MockCandidateProvider"
Cohesion: 0.05
Nodes (43): B, DiversityReranker, RankingWeights, FeedOptions, FeedResult, Candidate, CandidateProvider, CandidateSource (+35 more)

### Community 6 - "Context"
Cohesion: 0.06
Nodes (5): Conversation, Context, UUID, SocialRelationship, UserRecommendation

### Community 7 - ".CheckoutSubscription"
Cohesion: 0.11
Nodes (28): CryptoInvoice, CryptoOperation, Handler, CheckoutRequest, Logger, contextKey, responseWriterWrapper, ReadWriter (+20 more)

### Community 8 - "dependencies"
Cohesion: 0.05
Nodes (42): dependencies, clsx, lucide-react, next, next-auth, react, react-dom, tailwind-merge (+34 more)

### Community 9 - "MemoryRepository"
Cohesion: 0.07
Nodes (15): AuditLog, PaymentOrder, FollowRequest, MediaObject, Permission, RWMutex, Community, CommunityJoinRequest (+7 more)

### Community 10 - "[postId]/page.tsx"
Cohesion: 0.10
Nodes (23): SearchPostsResponse, StagedFile, ConfirmModal(), ConfirmModalProps, PostCardProps, MediaDetailSidebar(), MediaDetailSidebarProps, MediaViewerContext (+15 more)

### Community 11 - "Handler"
Cohesion: 0.13
Nodes (20): Handler, mfaTokenPayload, Request, ResponseWriter, main(), buildOTPAuthURL(), checkTOTPWindow(), computeTOTPCode() (+12 more)

### Community 12 - "index.ts"
Cohesion: 0.09
Nodes (23): AdminTab, TIER_PRESENTATION, TierBadge(), TierBadgeProps, TierKey, AdminOverview, AuditLog, Community (+15 more)

### Community 13 - "Context"
Cohesion: 0.17
Nodes (9): Community, CommunityJoinRequest, CommunityMember, CommunityRule, Context, Post, MemoryRepository, Time (+1 more)

### Community 14 - "compilerOptions"
Cohesion: 0.07
Nodes (26): compilerOptions, allowJs, esModuleInterop, incremental, isolatedModules, jsx, lib, module (+18 more)

### Community 15 - "search/page.tsx"
Cohesion: 0.12
Nodes (19): encodeFeedCursor(), FeedResponse, HomePage(), ExploreCategory, SearchPageContent(), SearchTab, SearchUsersResponse, MobileDrawer() (+11 more)

### Community 16 - "CryptoGateway"
Cohesion: 0.13
Nodes (15): Config, memoryLimiter, Mutex, CryptoCurrency, CryptoGateway, CryptoInvoice, CryptoOperation, WebhookPayload (+7 more)

### Community 17 - "Hub"
Cohesion: 0.16
Nodes (13): RawMessage, Client, Hub, InboundEvent, OutboundEvent, Conn, Context, Request (+5 more)

### Community 18 - "utils.ts"
Cohesion: 0.15
Nodes (15): MobileComposerModal(), MobileBottomNav(), MobileBottomNavProps, RouteTransition(), RealtimeContext, RealtimeContextType, RealtimeProvider(), ToastNotification (+7 more)

### Community 19 - "User"
Cohesion: 0.25
Nodes (9): SettingsSection, AuthModal(), PostCard(), LeftNavRail(), LeftNavRailProps, RightSidebar(), RightSidebarProps, useCurrentUser() (+1 more)

### Community 20 - "api.ts"
Cohesion: 0.14
Nodes (16): metadata, AuthModalProps, ComposerContext, ComposerContextValue, ComposerProvider(), useComposer(), PostModal(), AppShell() (+8 more)

### Community 21 - "Post"
Cohesion: 0.17
Nodes (3): Rows, FeedCursor, Post

### Community 23 - "PostModal.tsx"
Cohesion: 0.20
Nodes (14): MediaUrlInputModal(), MediaUrlInputModalProps, MobileComposerModalProps, StagedAttachment, PostComposer(), PostComposerProps, StagedAttachment, PostModalProps (+6 more)

### Community 24 - "UserNameWithRole.tsx"
Cohesion: 0.16
Nodes (13): EditProfileModal(), EditProfileModalProps, NAME_STYLE_PRESETS, TwoFactorManager(), TwoFactorManagerProps, buildCustomNameInlineStyle(), parseCustomNameStyle(), UserNameWithRole() (+5 more)

### Community 25 - "MediaService"
Cohesion: 0.18
Nodes (12): Config, Image, MediaService, Reader, getEnv(), getEnvInt(), Load(), Client (+4 more)

### Community 27 - "UpgradeModal.tsx"
Cohesion: 0.16
Nodes (13): VerifiedBadge(), VerifiedBadgeProps, CryptoCurrencyOption, num(), ParsedRolePlan, PaymentOrderStatus, ROLE_METADATA, UpgradeModal() (+5 more)

### Community 28 - "FeedSkeleton.tsx"
Cohesion: 0.19
Nodes (4): FeedSkeleton(), PostCardSkeleton(), PostDetailSkeleton(), ProfileSkeleton()

### Community 29 - "admin.go"
Cohesion: 0.16
Nodes (13): AdminDeletePostRequest, AssignMembershipRequest, CreateRoleRequest, SetPermissionsRequest, SetPlanEntitlementsRequest, UpdatePlanRequest, UpdatePostModerationRequest, UpdateRoleRequest (+5 more)

### Community 30 - "OtherSkeletons.tsx"
Cohesion: 0.23
Nodes (3): ExploreSkeleton(), MessagesSkeleton(), NotificationsSkeleton()

### Community 31 - "qrcode.tsx"
Cohesion: 0.24
Nodes (11): ALIGNMENT_PATTERNS, EC_CODEWORDS_L, encodeQRMatrix(), GF_EXP, GF_LOG, gfMul(), NUM_BLOCKS_L, QRCodeSVG() (+3 more)

### Community 33 - "handler.go"
Cohesion: 0.22
Nodes (10): CreatePostRequest, FeedFeedbackRequest, FeedImpressionsRequest, GoogleAuthRequest, LoginRequest, RegisterRequest, SimulatePaymentRequest, UpdatePostRequest (+2 more)

### Community 34 - "VideoControls.tsx"
Cohesion: 0.31
Nodes (7): formatTime(), SPEED_OPTIONS, VideoControls(), VideoControlsProps, formatTime(), VideoProgressBar(), VideoProgressBarProps

### Community 35 - ".GetActiveSubscription"
Cohesion: 0.22
Nodes (3): Entitlements, Subscription, UserMembership

### Community 38 - ".GetTrendingTopics"
Cohesion: 0.29
Nodes (3): categorizeTopic(), TrendingTopic, NewPostgresRepository()

### Community 39 - "community.go"
Cohesion: 0.33
Nodes (5): CreateCommunityRequest, ReviewJoinRequestPayload, UpdateCommunityRequest, UpdateMemberRolePayload, CommunityRule

### Community 40 - "route.ts"
Cohesion: 0.60
Nodes (3): getDynamicAuthOptions(), handler(), authOptions

## Knowledge Gaps
- **133 isolated node(s):** `name`, `version`, `private`, `dev`, `build` (+128 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **124 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `MemoryRepository` connect `MemoryRepository` to `T`, `models.go`, `.GetActiveSubscription`, `MembershipPlan`, `MockCandidateProvider`, `Context`, `Role`, `.GetTrendingTopics`, `.GetNotifications`, `Duration`, `Post`, `User`?**
  _High betweenness centrality (0.197) - this node is a cross-community bridge._
- **Why does `NewMemoryRepository()` connect `T` to `MemoryRepository`, `Handler`, `MockCandidateProvider`?**
  _High betweenness centrality (0.074) - this node is a cross-community bridge._
- **Why does `PaymentOrder` connect `MemoryRepository` to `models.go`, `PostgresRepository`, `.CheckoutSubscription`?**
  _High betweenness centrality (0.067) - this node is a cross-community bridge._
- **What connects `name`, `version`, `private` to the rest of the system?**
  _133 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `writeJSON` be split into smaller, more focused modules?**
  _Cohesion score 0.09343763581051717 - nodes in this community are weakly interconnected._
- **Should `T` be split into smaller, more focused modules?**
  _Cohesion score 0.06773211567732115 - nodes in this community are weakly interconnected._
- **Should `PostgresRepository` be split into smaller, more focused modules?**
  _Cohesion score 0.05835010060362173 - nodes in this community are weakly interconnected._