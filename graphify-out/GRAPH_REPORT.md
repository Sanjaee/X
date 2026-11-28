# Graph Report - .  (2026-11-28)

## Corpus Check
- cluster-only mode — file stats not available

## Summary
- 1413 nodes · 3680 edges · 193 communities (44 shown, 149 thin omitted)
- Extraction: 97% EXTRACTED · 3% INFERRED · 0% AMBIGUOUS · INFERRED: 107 edges (avg confidence: 0.8)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `335d8f61`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- Context
- MemoryRepository
- PostgresRepository
- T
- models.go
- [postId]/page.tsx
- MockCandidateProvider
- .CheckoutSubscription
- dependencies
- search/page.tsx
- utils.ts
- Context
- CryptoGateway
- Hub
- index.ts
- Context
- compilerOptions
- Post
- UserNameWithRole.tsx
- totp.go
- Post
- PostModal.tsx
- User
- UpgradeModal.tsx
- FeedSkeleton.tsx
- admin.go
- OtherSkeletons.tsx
- qrcode.tsx
- Role
- AuditLog
- VideoControls.tsx
- Engine
- .GetActiveSubscription
- MembershipPlan
- app/page.tsx
- .GetTrendingTopics
- MediaObject
- route.ts
- TopLoadingBar.tsx
- config.go
- .GetActiveSubscription
- Notification
- Permission
- middleware.ts
- Duration
- FollowRequest
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
- github.com/x-reconstruction/api
- PlisioService
- RankingWeights
- RecentSearch
- SearchFilterParams
- MembershipEntitlement
- CommunityRule
- Client
- Context
- MediaObject
- PaymentOrder
- Repository
- Request
- ResponseWriter
- Subscription
- UUID
- PaymentOrder
- Mux
- T
- Request
- ResponseWriter
- Time
- User
- UUID
- Client
- Context
- MediaObject
- UUID
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
4. `Handler` - 108 edges
5. `writeError()` - 108 edges
6. `apiRequest()` - 36 edges
7. `User` - 32 edges
8. `Handler` - 26 edges
9. `cn()` - 25 edges
10. `MemoryRepository` - 21 edges

## Surprising Connections (you probably didn't know these)
- `setupTestRouter()` --calls--> `AuthMiddleware()`  [INFERRED]
  services/api/internal/handler/handler_test.go → services/api/internal/middleware/middleware.go
- `setupTestApp()` --calls--> `AuthMiddleware()`  [INFERRED]
  services/api/internal/handler/social_messaging_test.go → services/api/internal/middleware/middleware.go
- `setupTestApp()` --calls--> `NewHub()`  [INFERRED]
  services/api/internal/handler/social_messaging_test.go → services/api/internal/realtime/hub.go
- `RateLimiter()` --calls--> `GetUserFromContext()`  [INFERRED]
  services/api/internal/middleware/ratelimit.go → services/api/internal/middleware/middleware.go
- `BenchmarkScoringEngine()` --calls--> `DefaultRankingWeights()`  [INFERRED]
  services/api/internal/recommendation/engine_test.go → services/api/internal/domain/models.go

## Import Cycles
- None detected.

## Communities (193 total, 149 thin omitted)

### Community 0 - "Context"
Cohesion: 0.05
Nodes (23): AdminPostFilter, Context, CreatePostRequest, FeedFeedbackRequest, FeedImpressionsRequest, GoogleAuthRequest, Handler, LoginRequest (+15 more)

### Community 1 - "MemoryRepository"
Cohesion: 0.06
Nodes (13): Conversation, PaymentOrder, RWMutex, Community, CommunityJoinRequest, CommunityMember, CommunityRule, Context (+5 more)

### Community 2 - "PostgresRepository"
Cohesion: 0.07
Nodes (4): DB, Context, PostgresRepository, UUID

### Community 3 - "T"
Cohesion: 0.07
Nodes (61): CryptoGateway, Hub, MediaService, Mux, Repository, Repository, setupTestFullAdminRouter(), TestAdminPostModerationAndAuditLogging() (+53 more)

### Community 4 - "models.go"
Cohesion: 0.07
Nodes (66): AdminOverview, AdminPostFilter, AuditLog, Community, CommunityInvitation, CommunityJoinRequest, CommunityMember, CommunityPost (+58 more)

### Community 5 - "[postId]/page.tsx"
Cohesion: 0.07
Nodes (32): metadata, SearchPostsResponse, StagedFile, ConfirmModal(), ConfirmModalProps, ComposerProvider(), PostModal(), PostCard() (+24 more)

### Community 6 - "MockCandidateProvider"
Cohesion: 0.06
Nodes (34): B, Candidate, CandidateSource, DiversityReranker, Features, FeedCursor, FeedMetrics, FeedOptions (+26 more)

### Community 7 - ".CheckoutSubscription"
Cohesion: 0.11
Nodes (28): CryptoInvoice, CryptoOperation, Handler, CheckoutRequest, Logger, contextKey, responseWriterWrapper, ReadWriter (+20 more)

### Community 8 - "dependencies"
Cohesion: 0.05
Nodes (42): dependencies, clsx, lucide-react, next, next-auth, react, react-dom, tailwind-merge (+34 more)

### Community 9 - "search/page.tsx"
Cohesion: 0.15
Nodes (22): ExploreCategory, SearchPageContent(), SearchTab, SearchUsersResponse, AuthModal(), AuthModalProps, RightSidebar(), RightSidebarProps (+14 more)

### Community 10 - "utils.ts"
Cohesion: 0.12
Nodes (22): SettingsSection, ComposerContext, ComposerContextValue, useComposer(), MobileComposerModal(), LeftNavRail(), LeftNavRailProps, MobileBottomNav() (+14 more)

### Community 11 - "Context"
Cohesion: 0.13
Nodes (13): Community, CommunityJoinRequest, CommunityMember, CommunityRule, CreateCommunityRequest, ReviewJoinRequestPayload, UpdateCommunityRequest, UpdateMemberRolePayload (+5 more)

### Community 12 - "CryptoGateway"
Cohesion: 0.10
Nodes (18): Client, Config, Image, MediaService, CryptoCurrency, CryptoGateway, CryptoInvoice, CryptoOperation (+10 more)

### Community 13 - "Hub"
Cohesion: 0.12
Nodes (17): memoryLimiter, Mutex, RawMessage, Client, Hub, InboundEvent, OutboundEvent, Time (+9 more)

### Community 14 - "index.ts"
Cohesion: 0.09
Nodes (25): AdminTab, TIER_PRESENTATION, TierBadge(), TierBadgeProps, TierKey, PostComposer(), AdminOverview, AuditLog (+17 more)

### Community 15 - "Context"
Cohesion: 0.17
Nodes (9): Community, CommunityJoinRequest, CommunityMember, CommunityRule, Context, Post, MemoryRepository, Time (+1 more)

### Community 16 - "compilerOptions"
Cohesion: 0.07
Nodes (26): compilerOptions, allowJs, esModuleInterop, incremental, isolatedModules, jsx, lib, module (+18 more)

### Community 17 - "Post"
Cohesion: 0.17
Nodes (3): FeedCursor, Rows, Post

### Community 18 - "UserNameWithRole.tsx"
Cohesion: 0.16
Nodes (13): EditProfileModal(), EditProfileModalProps, NAME_STYLE_PRESETS, TwoFactorManager(), TwoFactorManagerProps, buildCustomNameInlineStyle(), parseCustomNameStyle(), UserNameWithRole() (+5 more)

### Community 19 - "totp.go"
Cohesion: 0.20
Nodes (16): mfaTokenPayload, buildOTPAuthURL(), checkTOTPWindow(), computeTOTPCode(), createMFAToken(), fetchNetworkTime(), generateDeterministicBackupCodes(), generateRecoverySeedPhrase() (+8 more)

### Community 21 - "PostModal.tsx"
Cohesion: 0.21
Nodes (13): MediaUrlInputModal(), MediaUrlInputModalProps, MobileComposerModalProps, StagedAttachment, PostComposerProps, StagedAttachment, PostModalProps, StagedAttachment (+5 more)

### Community 23 - "UpgradeModal.tsx"
Cohesion: 0.16
Nodes (13): VerifiedBadge(), VerifiedBadgeProps, CryptoCurrencyOption, num(), ParsedRolePlan, PaymentOrderStatus, ROLE_METADATA, UpgradeModal() (+5 more)

### Community 24 - "FeedSkeleton.tsx"
Cohesion: 0.19
Nodes (4): FeedSkeleton(), PostCardSkeleton(), PostDetailSkeleton(), ProfileSkeleton()

### Community 25 - "admin.go"
Cohesion: 0.16
Nodes (13): AdminDeletePostRequest, AssignMembershipRequest, CreateRoleRequest, SetPermissionsRequest, SetPlanEntitlementsRequest, UpdatePlanRequest, UpdatePostModerationRequest, UpdateRoleRequest (+5 more)

### Community 26 - "OtherSkeletons.tsx"
Cohesion: 0.23
Nodes (3): ExploreSkeleton(), MessagesSkeleton(), NotificationsSkeleton()

### Community 27 - "qrcode.tsx"
Cohesion: 0.24
Nodes (11): ALIGNMENT_PATTERNS, EC_CODEWORDS_L, encodeQRMatrix(), GF_EXP, GF_LOG, gfMul(), NUM_BLOCKS_L, QRCodeSVG() (+3 more)

### Community 30 - "VideoControls.tsx"
Cohesion: 0.31
Nodes (7): formatTime(), SPEED_OPTIONS, VideoControls(), VideoControlsProps, formatTime(), VideoProgressBar(), VideoProgressBarProps

### Community 31 - "Engine"
Cohesion: 0.31
Nodes (7): DiversityReranker, FeedOptions, FeedResult, CandidateProvider, Engine, Context, NewEngine()

### Community 34 - "app/page.tsx"
Cohesion: 0.38
Nodes (5): encodeFeedCursor(), FeedResponse, HomePage(), feedKeys, searchKeys

### Community 35 - ".GetTrendingTopics"
Cohesion: 0.29
Nodes (3): categorizeTopic(), NewPostgresRepository(), TrendingTopic

### Community 37 - "route.ts"
Cohesion: 0.60
Nodes (3): getDynamicAuthOptions(), handler(), authOptions

### Community 39 - "config.go"
Cohesion: 0.70
Nodes (4): Config, getEnv(), getEnvInt(), Load()

## Knowledge Gaps
- **135 isolated node(s):** `name`, `version`, `private`, `dev`, `build` (+130 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **149 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `MemoryRepository` connect `MemoryRepository` to `.GetActiveSubscription`, `MembershipPlan`, `T`, `models.go`, `MediaObject`, `Notification`, `Permission`, `FollowRequest`, `Post`, `User`, `Role`, `AuditLog`?**
  _High betweenness centrality (0.139) - this node is a cross-community bridge._
- **Why does `RankingWeights` connect `models.go` to `MemoryRepository`, `MockCandidateProvider`?**
  _High betweenness centrality (0.052) - this node is a cross-community bridge._
- **Why does `NewMemoryRepository()` connect `T` to `MemoryRepository`, `models.go`?**
  _High betweenness centrality (0.047) - this node is a cross-community bridge._
- **What connects `name`, `version`, `private` to the rest of the system?**
  _135 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `Context` be split into smaller, more focused modules?**
  _Cohesion score 0.051491595242841005 - nodes in this community are weakly interconnected._
- **Should `MemoryRepository` be split into smaller, more focused modules?**
  _Cohesion score 0.060083594566353184 - nodes in this community are weakly interconnected._
- **Should `PostgresRepository` be split into smaller, more focused modules?**
  _Cohesion score 0.06793206793206794 - nodes in this community are weakly interconnected._