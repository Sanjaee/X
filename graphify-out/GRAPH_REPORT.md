# Graph Report - .  (2026-11-22)

## Corpus Check
- cluster-only mode — file stats not available

## Summary
- 1411 nodes · 3682 edges · 191 communities (53 shown, 138 thin omitted)
- Extraction: 97% EXTRACTED · 3% INFERRED · 0% AMBIGUOUS · INFERRED: 108 edges (avg confidence: 0.8)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `f9df5377`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- writeJSON
- PostgresRepository
- models.go
- Context
- Context
- MockCandidateProvider
- dependencies
- .CheckoutSubscription
- [postId]/page.tsx
- MemoryRepository
- api.ts
- index.ts
- Context
- Context
- utils.ts
- search/page.tsx
- compilerOptions
- Hub
- User
- Post
- totp.go
- Post
- PostModal.tsx
- UserNameWithRole.tsx
- MediaService
- User
- Repository
- NewMemoryRepository
- FeedSkeleton.tsx
- admin.go
- T
- OtherSkeletons.tsx
- qrcode.tsx
- handler.go
- VideoControls.tsx
- NewHandler
- Engine
- Duration
- Role
- setupTestApp
- .GetActiveSubscription
- MembershipPlan
- .GetTrendingTopics
- setupTestRouter
- setupRecommendationRouter
- community.go
- route.ts
- TopLoadingBar.tsx
- .GetMediaObjectByID
- .GetActiveSubscription
- Conversation
- Notification
- middleware.ts
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
7. `User` - 36 edges
8. `Handler` - 26 edges
9. `cn()` - 25 edges
10. `Post` - 23 edges

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

## Communities (191 total, 138 thin omitted)

### Community 0 - "writeJSON"
Cohesion: 0.08
Nodes (10): Handler, Request, ResponseWriter, Handler, Request, ResponseWriter, Context, Subscription (+2 more)

### Community 1 - "PostgresRepository"
Cohesion: 0.07
Nodes (5): DB, Context, Message, PostgresRepository, UUID

### Community 2 - "models.go"
Cohesion: 0.07
Nodes (66): AdminOverview, AdminPostFilter, AuditLog, Community, CommunityInvitation, CommunityJoinRequest, CommunityMember, CommunityPost (+58 more)

### Community 3 - "Context"
Cohesion: 0.06
Nodes (17): AdminOverview, AdminPostFilter, Config, Context, MembershipEntitlement, CryptoCurrency, CryptoGateway, CryptoInvoice (+9 more)

### Community 5 - "MockCandidateProvider"
Cohesion: 0.06
Nodes (34): B, Candidate, CandidateSource, DiversityReranker, Features, FeedCursor, FeedMetrics, FeedOptions (+26 more)

### Community 6 - "dependencies"
Cohesion: 0.05
Nodes (42): dependencies, clsx, lucide-react, next, next-auth, react, react-dom, tailwind-merge (+34 more)

### Community 7 - ".CheckoutSubscription"
Cohesion: 0.12
Nodes (25): CryptoInvoice, CryptoOperation, CheckoutRequest, Logger, contextKey, responseWriterWrapper, ReadWriter, expired() (+17 more)

### Community 8 - "[postId]/page.tsx"
Cohesion: 0.10
Nodes (23): StagedFile, ConfirmModal(), ConfirmModalProps, PostModalProps, PostCardProps, MediaDetailSidebar(), MediaDetailSidebarProps, MediaViewerContext (+15 more)

### Community 9 - "MemoryRepository"
Cohesion: 0.07
Nodes (15): AuditLog, PaymentOrder, FollowRequest, MediaObject, Permission, RWMutex, Community, CommunityJoinRequest (+7 more)

### Community 10 - "api.ts"
Cohesion: 0.10
Nodes (26): AuthModalProps, VerifiedBadge(), VerifiedBadgeProps, ComposerContext, ComposerContextValue, ComposerProvider(), useComposer(), PostModal() (+18 more)

### Community 11 - "index.ts"
Cohesion: 0.09
Nodes (24): AdminTab, TIER_PRESENTATION, TierBadge(), TierBadgeProps, TierKey, PostComposer(), AdminOverview, AuditLog (+16 more)

### Community 12 - "Context"
Cohesion: 0.17
Nodes (9): Community, CommunityJoinRequest, CommunityMember, CommunityRule, Context, Post, PostgresRepository, Time (+1 more)

### Community 13 - "Context"
Cohesion: 0.17
Nodes (9): Community, CommunityJoinRequest, CommunityMember, CommunityRule, Context, Post, MemoryRepository, Time (+1 more)

### Community 14 - "utils.ts"
Cohesion: 0.12
Nodes (17): metadata, AppShell(), MobileBottomNav(), MobileBottomNavProps, RouteTransition(), AppProviders(), RealtimeContext, RealtimeContextType (+9 more)

### Community 15 - "search/page.tsx"
Cohesion: 0.11
Nodes (20): encodeFeedCursor(), FeedResponse, HomePage(), ExploreCategory, SearchPageContent(), SearchPostsResponse, SearchTab, SearchUsersResponse (+12 more)

### Community 16 - "compilerOptions"
Cohesion: 0.07
Nodes (26): compilerOptions, allowJs, esModuleInterop, incremental, isolatedModules, jsx, lib, module (+18 more)

### Community 17 - "Hub"
Cohesion: 0.16
Nodes (13): RawMessage, Client, Hub, InboundEvent, OutboundEvent, Conn, Context, Request (+5 more)

### Community 18 - "User"
Cohesion: 0.25
Nodes (9): SettingsSection, AuthModal(), PostCard(), LeftNavRail(), LeftNavRailProps, RightSidebar(), RightSidebarProps, useCurrentUser() (+1 more)

### Community 19 - "Post"
Cohesion: 0.17
Nodes (3): FeedCursor, Rows, Post

### Community 20 - "totp.go"
Cohesion: 0.20
Nodes (16): mfaTokenPayload, buildOTPAuthURL(), checkTOTPWindow(), computeTOTPCode(), createMFAToken(), fetchNetworkTime(), generateDeterministicBackupCodes(), generateRecoverySeedPhrase() (+8 more)

### Community 22 - "PostModal.tsx"
Cohesion: 0.22
Nodes (13): MediaUrlInputModal(), MediaUrlInputModalProps, MobileComposerModal(), MobileComposerModalProps, StagedAttachment, PostComposerProps, StagedAttachment, StagedAttachment (+5 more)

### Community 23 - "UserNameWithRole.tsx"
Cohesion: 0.16
Nodes (13): EditProfileModal(), EditProfileModalProps, NAME_STYLE_PRESETS, TwoFactorManager(), TwoFactorManagerProps, buildCustomNameInlineStyle(), parseCustomNameStyle(), UserNameWithRole() (+5 more)

### Community 24 - "MediaService"
Cohesion: 0.18
Nodes (12): Config, Image, MediaService, Reader, getEnv(), getEnvInt(), Load(), Client (+4 more)

### Community 26 - "Repository"
Cohesion: 0.21
Nodes (14): Handler, Repository, setupTestFullAdminRouter(), TestAdminPostModerationAndAuditLogging(), TestOwnerAdminOverviewSecurity(), AuthorizeResource(), User, UUID (+6 more)

### Community 27 - "NewMemoryRepository"
Cohesion: 0.28
Nodes (12): main(), Handler, User, registerAndLoginUser(), setupTestRBACRouter(), TestConcurrentPostingCannotBypassQuota(), TestMediaUploadLimitsByTier(), TestNewUserDefaultMemberRoleAndAdminProtection() (+4 more)

### Community 28 - "FeedSkeleton.tsx"
Cohesion: 0.19
Nodes (4): FeedSkeleton(), PostCardSkeleton(), PostDetailSkeleton(), ProfileSkeleton()

### Community 29 - "admin.go"
Cohesion: 0.16
Nodes (13): AdminDeletePostRequest, AssignMembershipRequest, CreateRoleRequest, SetPermissionsRequest, SetPlanEntitlementsRequest, UpdatePlanRequest, UpdatePostModerationRequest, UpdateRoleRequest (+5 more)

### Community 30 - "T"
Cohesion: 0.50
Nodes (12): Mux, createTestPost(), Post, registerTestUserInInteractions(), setupInteractionsTestRouter(), TestBookmarkFlowAndLikesTab(), TestCounterReconciliation(), TestLikeIdempotencyAndConcurrency() (+4 more)

### Community 31 - "OtherSkeletons.tsx"
Cohesion: 0.23
Nodes (3): ExploreSkeleton(), MessagesSkeleton(), NotificationsSkeleton()

### Community 32 - "qrcode.tsx"
Cohesion: 0.24
Nodes (11): ALIGNMENT_PATTERNS, EC_CODEWORDS_L, encodeQRMatrix(), GF_EXP, GF_LOG, gfMul(), NUM_BLOCKS_L, QRCodeSVG() (+3 more)

### Community 33 - "handler.go"
Cohesion: 0.20
Nodes (11): CreatePostRequest, FeedFeedbackRequest, FeedImpressionsRequest, GoogleAuthRequest, LoginRequest, RegisterRequest, SimulatePaymentRequest, UpdatePostRequest (+3 more)

### Community 34 - "VideoControls.tsx"
Cohesion: 0.31
Nodes (7): formatTime(), SPEED_OPTIONS, VideoControls(), VideoControlsProps, formatTime(), VideoProgressBar(), VideoProgressBarProps

### Community 35 - "NewHandler"
Cohesion: 0.22
Nodes (8): Client, CryptoGateway, Hub, MediaService, Repository, NewHandler(), T, TestSearchAPIEndpoints()

### Community 36 - "Engine"
Cohesion: 0.31
Nodes (7): DiversityReranker, FeedOptions, FeedResult, CandidateProvider, Engine, Context, NewEngine()

### Community 37 - "Duration"
Cohesion: 0.25
Nodes (5): Duration, memoryLimiter, Mutex, Time, RateLimiter()

### Community 39 - "setupTestApp"
Cohesion: 0.47
Nodes (8): createTestUser(), Mux, T, User, setupTestApp(), TestDirectMessaging_IdempotencyAndFlow(), TestFollowRequest_PrivateAccount(), TestSocialGraph_FollowUnfollowBlockMute()

### Community 42 - ".GetTrendingTopics"
Cohesion: 0.29
Nodes (3): categorizeTopic(), NewPostgresRepository(), TrendingTopic

### Community 43 - "setupTestRouter"
Cohesion: 0.52
Nodes (6): Mux, T, setupTestRouter(), TestAuthAndPostingFlow(), TestHealthCheck(), TestSubscriptionsAndBadges()

### Community 44 - "setupRecommendationRouter"
Cohesion: 0.67
Nodes (6): registerTestUser(), setupRecommendationRouter(), TestRealtimeNewPostsIndicator(), TestRecommendationAndFeedAPIs(), TestTrendingAnalyticsAndSpamResistance(), TestWhoToFollowAndDismissal()

### Community 45 - "community.go"
Cohesion: 0.33
Nodes (5): CreateCommunityRequest, ReviewJoinRequestPayload, UpdateCommunityRequest, UpdateMemberRolePayload, CommunityRule

### Community 46 - "route.ts"
Cohesion: 0.60
Nodes (3): getDynamicAuthOptions(), handler(), authOptions

## Knowledge Gaps
- **131 isolated node(s):** `name`, `version`, `private`, `dev`, `build` (+126 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **138 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `MemoryRepository` connect `MemoryRepository` to `models.go`, `Context`, `Role`, `.GetActiveSubscription`, `MembershipPlan`, `Conversation`, `Notification`, `Post`, `User`, `NewMemoryRepository`?**
  _High betweenness centrality (0.124) - this node is a cross-community bridge._
- **Why does `NewMemoryRepository()` connect `NewMemoryRepository` to `models.go`, `NewHandler`, `setupTestApp`, `MemoryRepository`, `setupTestRouter`, `setupRecommendationRouter`, `Repository`, `T`?**
  _High betweenness centrality (0.064) - this node is a cross-community bridge._
- **Why does `Handler` connect `writeJSON` to `NewMemoryRepository`, `handler.go`, `NewHandler`, `totp.go`?**
  _High betweenness centrality (0.055) - this node is a cross-community bridge._
- **Are the 31 inferred relationships involving `writeJSON()` (e.g. with `.AdminAssignUserMembership()` and `.AdminCreateRole()`) actually correct?**
  _`writeJSON()` has 31 INFERRED edges - model-reasoned connections that need verification._
- **What connects `name`, `version`, `private` to the rest of the system?**
  _131 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `writeJSON` be split into smaller, more focused modules?**
  _Cohesion score 0.07940305308726361 - nodes in this community are weakly interconnected._
- **Should `PostgresRepository` be split into smaller, more focused modules?**
  _Cohesion score 0.06512345679012346 - nodes in this community are weakly interconnected._