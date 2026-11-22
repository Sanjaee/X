# Graph Report - .  (2026-11-22)

## Corpus Check
- cluster-only mode — file stats not available

## Summary
- 1411 nodes · 3680 edges · 186 communities (46 shown, 140 thin omitted)
- Extraction: 97% EXTRACTED · 3% INFERRED · 0% AMBIGUOUS · INFERRED: 108 edges (avg confidence: 0.8)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `71f2071f`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- writeJSON
- T
- PostgresRepository
- models.go
- Context
- Context
- MockCandidateProvider
- MemoryRepository
- dependencies
- .CheckoutSubscription
- utils.ts
- index.ts
- search/page.tsx
- [postId]/page.tsx
- api.ts
- Hub
- Context
- Context
- compilerOptions
- Post
- PostCard.tsx
- totp.go
- Post
- app/page.tsx
- UserNameWithRole.tsx
- MediaService
- User
- FeedSkeleton.tsx
- admin.go
- OtherSkeletons.tsx
- qrcode.tsx
- handler.go
- VideoControls.tsx
- Engine
- Role
- .GetActiveSubscription
- MembershipPlan
- .GetTrendingTopics
- community.go
- route.ts
- TopLoadingBar.tsx
- .GetActiveSubscription
- Conversation
- Notification
- middleware.ts
- Duration
- .GetMediaObjectByID
- .GetMessages
- test_all_apis.py
- next.config.js
- next-env.d.ts
- tailwind.config.ts
- .GetSocialRelationship
- .GetWhoToFollow
- .MarkOutboxEventFailed
- backup-db.sh
- monitoring.sh
- Conn
- CounterReconciliationResult
- FeedFeedback
- FeedImpression
- Invoice
- Message
- OutboxEvent
- PaymentOrder
- github.com/x-reconstruction/api
- PlisioService
- RankingWeights
- RecentSearch
- SearchFilterParams
- MembershipEntitlement
- Client
- PaymentOrder
- Repository
- Request
- ResponseWriter
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
- Community
- CommunityJoinRequest
- CommunityMember
- CommunityRule
- Conversation
- CounterReconciliationResult
- Duration
- Entitlements
- FeedCursor
- FeedFeedback
- FeedImpression
- FollowRequest
- Invoice
- Notification
- OutboxEvent
- PaymentOrder
- RankingWeights
- RecentSearch
- SearchFilterParams
- Session
- SocialRelationship
- SubscriptionPlan
- TrendingTopic
- User
- UserRecommendation
- UserTopicPreference
- VerificationApplication
- VerificationBadge
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
5. `Handler` - 107 edges
6. `apiRequest()` - 36 edges
7. `User` - 35 edges
8. `Handler` - 26 edges
9. `cn()` - 25 edges
10. `Post` - 22 edges

## Surprising Connections (you probably didn't know these)
- `setupTestApp()` --calls--> `NewHub()`  [INFERRED]
  services/api/internal/handler/social_messaging_test.go → services/api/internal/realtime/hub.go
- `RateLimiter()` --calls--> `GetUserFromContext()`  [INFERRED]
  services/api/internal/middleware/ratelimit.go → services/api/internal/middleware/middleware.go
- `BenchmarkScoringEngine()` --calls--> `DefaultRankingWeights()`  [INFERRED]
  services/api/internal/recommendation/engine_test.go → services/api/internal/domain/models.go
- `PhotoCanvasProps` --references--> `MediaObject`  [EXTRACTED]
  apps/web/src/components/media/PhotoCanvas.tsx → apps/web/src/types/index.ts
- `main()` --calls--> `Handler`  [INFERRED]
  services/api/cmd/server/main.go → services/api/internal/handler/handler.go

## Import Cycles
- None detected.

## Communities (186 total, 140 thin omitted)

### Community 0 - "writeJSON"
Cohesion: 0.08
Nodes (10): Handler, Request, ResponseWriter, Handler, Request, ResponseWriter, Context, Subscription (+2 more)

### Community 1 - "T"
Cohesion: 0.06
Nodes (66): Client, CryptoGateway, Handler, Hub, MediaService, Mux, Repository, Repository (+58 more)

### Community 2 - "PostgresRepository"
Cohesion: 0.07
Nodes (4): DB, Context, PostgresRepository, UUID

### Community 3 - "models.go"
Cohesion: 0.07
Nodes (66): AdminOverview, AdminPostFilter, AuditLog, Community, CommunityInvitation, CommunityJoinRequest, CommunityMember, CommunityPost (+58 more)

### Community 4 - "Context"
Cohesion: 0.06
Nodes (17): AdminOverview, AdminPostFilter, Config, Context, MembershipEntitlement, CryptoCurrency, CryptoGateway, CryptoInvoice (+9 more)

### Community 6 - "MockCandidateProvider"
Cohesion: 0.06
Nodes (34): B, Candidate, CandidateSource, DiversityReranker, Features, FeedCursor, FeedMetrics, FeedOptions (+26 more)

### Community 7 - "MemoryRepository"
Cohesion: 0.07
Nodes (15): AuditLog, PaymentOrder, FollowRequest, MediaObject, Permission, RWMutex, Community, CommunityJoinRequest (+7 more)

### Community 8 - "dependencies"
Cohesion: 0.05
Nodes (42): dependencies, clsx, lucide-react, next, next-auth, react, react-dom, tailwind-merge (+34 more)

### Community 9 - ".CheckoutSubscription"
Cohesion: 0.12
Nodes (25): CryptoInvoice, CryptoOperation, CheckoutRequest, Logger, contextKey, responseWriterWrapper, ReadWriter, expired() (+17 more)

### Community 10 - "utils.ts"
Cohesion: 0.11
Nodes (20): metadata, SettingsSection, AppShell(), MobileBottomNav(), MobileBottomNavProps, MobileDrawer(), MobileDrawerProps, RouteTransition() (+12 more)

### Community 11 - "index.ts"
Cohesion: 0.08
Nodes (25): AdminTab, TIER_PRESENTATION, TierBadge(), TierBadgeProps, TierKey, PostComposer(), AdminOverview, AuditLog (+17 more)

### Community 12 - "search/page.tsx"
Cohesion: 0.16
Nodes (17): ExploreCategory, SearchPageContent(), SearchTab, SearchUsersResponse, PostCard(), LeftNavRail(), LeftNavRailProps, RightSidebar() (+9 more)

### Community 13 - "[postId]/page.tsx"
Cohesion: 0.13
Nodes (23): SearchPostsResponse, StagedFile, MediaUrlInputModal(), MediaUrlInputModalProps, MobileComposerModal(), MobileComposerModalProps, StagedAttachment, PostComposerProps (+15 more)

### Community 14 - "api.ts"
Cohesion: 0.12
Nodes (23): AuthModal(), AuthModalProps, ComposerContext, ComposerContextValue, ComposerProvider(), useComposer(), PostModal(), PhotoNavigation() (+15 more)

### Community 15 - "Hub"
Cohesion: 0.12
Nodes (17): memoryLimiter, Mutex, RawMessage, Client, Hub, InboundEvent, OutboundEvent, Time (+9 more)

### Community 16 - "Context"
Cohesion: 0.17
Nodes (9): Community, CommunityJoinRequest, CommunityMember, CommunityRule, Context, Post, PostgresRepository, Time (+1 more)

### Community 17 - "Context"
Cohesion: 0.17
Nodes (9): Community, CommunityJoinRequest, CommunityMember, CommunityRule, Context, Post, MemoryRepository, Time (+1 more)

### Community 18 - "compilerOptions"
Cohesion: 0.07
Nodes (26): compilerOptions, allowJs, esModuleInterop, incremental, isolatedModules, jsx, lib, module (+18 more)

### Community 19 - "Post"
Cohesion: 0.17
Nodes (3): FeedCursor, Rows, Post

### Community 20 - "PostCard.tsx"
Cohesion: 0.13
Nodes (11): ConfirmModal(), ConfirmModalProps, PostCardProps, MediaDetailSidebar(), MediaDetailSidebarProps, PhotoCanvas(), PhotoCanvasProps, StandalonePhotoView() (+3 more)

### Community 21 - "totp.go"
Cohesion: 0.20
Nodes (16): mfaTokenPayload, buildOTPAuthURL(), checkTOTPWindow(), computeTOTPCode(), createMFAToken(), fetchNetworkTime(), generateDeterministicBackupCodes(), generateRecoverySeedPhrase() (+8 more)

### Community 23 - "app/page.tsx"
Cohesion: 0.15
Nodes (13): encodeFeedCursor(), FeedResponse, HomePage(), VerifiedBadge(), VerifiedBadgeProps, UpgradeModalContext, UpgradeModalContextType, UpgradeModalProvider() (+5 more)

### Community 24 - "UserNameWithRole.tsx"
Cohesion: 0.16
Nodes (13): EditProfileModal(), EditProfileModalProps, NAME_STYLE_PRESETS, TwoFactorManager(), TwoFactorManagerProps, buildCustomNameInlineStyle(), parseCustomNameStyle(), UserNameWithRole() (+5 more)

### Community 25 - "MediaService"
Cohesion: 0.18
Nodes (12): Config, Image, MediaService, Reader, getEnv(), getEnvInt(), Load(), Client (+4 more)

### Community 27 - "FeedSkeleton.tsx"
Cohesion: 0.19
Nodes (4): FeedSkeleton(), PostCardSkeleton(), PostDetailSkeleton(), ProfileSkeleton()

### Community 28 - "admin.go"
Cohesion: 0.16
Nodes (13): AdminDeletePostRequest, AssignMembershipRequest, CreateRoleRequest, SetPermissionsRequest, SetPlanEntitlementsRequest, UpdatePlanRequest, UpdatePostModerationRequest, UpdateRoleRequest (+5 more)

### Community 29 - "OtherSkeletons.tsx"
Cohesion: 0.23
Nodes (3): ExploreSkeleton(), MessagesSkeleton(), NotificationsSkeleton()

### Community 30 - "qrcode.tsx"
Cohesion: 0.24
Nodes (11): ALIGNMENT_PATTERNS, EC_CODEWORDS_L, encodeQRMatrix(), GF_EXP, GF_LOG, gfMul(), NUM_BLOCKS_L, QRCodeSVG() (+3 more)

### Community 31 - "handler.go"
Cohesion: 0.20
Nodes (11): CreatePostRequest, FeedFeedbackRequest, FeedImpressionsRequest, GoogleAuthRequest, LoginRequest, RegisterRequest, SimulatePaymentRequest, UpdatePostRequest (+3 more)

### Community 32 - "VideoControls.tsx"
Cohesion: 0.31
Nodes (7): formatTime(), SPEED_OPTIONS, VideoControls(), VideoControlsProps, formatTime(), VideoProgressBar(), VideoProgressBarProps

### Community 33 - "Engine"
Cohesion: 0.31
Nodes (7): DiversityReranker, FeedOptions, FeedResult, CandidateProvider, Engine, Context, NewEngine()

### Community 37 - ".GetTrendingTopics"
Cohesion: 0.29
Nodes (3): categorizeTopic(), NewPostgresRepository(), TrendingTopic

### Community 38 - "community.go"
Cohesion: 0.33
Nodes (5): CreateCommunityRequest, ReviewJoinRequestPayload, UpdateCommunityRequest, UpdateMemberRolePayload, CommunityRule

### Community 39 - "route.ts"
Cohesion: 0.60
Nodes (3): getDynamicAuthOptions(), handler(), authOptions

## Knowledge Gaps
- **132 isolated node(s):** `name`, `version`, `private`, `dev`, `build` (+127 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **140 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `MemoryRepository` connect `MemoryRepository` to `T`, `Role`, `models.go`, `MembershipPlan`, `Context`, `.GetActiveSubscription`, `Conversation`, `Notification`, `Post`, `User`?**
  _High betweenness centrality (0.128) - this node is a cross-community bridge._
- **Why does `NewMemoryRepository()` connect `T` to `models.go`, `MemoryRepository`?**
  _High betweenness centrality (0.064) - this node is a cross-community bridge._
- **Why does `PaymentOrder` connect `MemoryRepository` to `writeJSON`, `.CheckoutSubscription`, `PostgresRepository`, `models.go`?**
  _High betweenness centrality (0.055) - this node is a cross-community bridge._
- **Are the 31 inferred relationships involving `writeJSON()` (e.g. with `.AdminAssignUserMembership()` and `.AdminCreateRole()`) actually correct?**
  _`writeJSON()` has 31 INFERRED edges - model-reasoned connections that need verification._
- **What connects `name`, `version`, `private` to the rest of the system?**
  _132 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `writeJSON` be split into smaller, more focused modules?**
  _Cohesion score 0.07940305308726361 - nodes in this community are weakly interconnected._
- **Should `T` be split into smaller, more focused modules?**
  _Cohesion score 0.06227106227106227 - nodes in this community are weakly interconnected._