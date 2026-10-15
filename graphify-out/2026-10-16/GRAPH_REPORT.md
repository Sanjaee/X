# Graph Report - .  (2026-10-16)

## Corpus Check
- cluster-only mode — file stats not available

## Summary
- 828 nodes · 2271 edges · 71 communities (32 shown, 39 thin omitted)
- Extraction: 98% EXTRACTED · 2% INFERRED · 0% AMBIGUOUS · INFERRED: 36 edges (avg confidence: 0.8)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `acec0121`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- index.ts
- MemoryRepository
- Handler
- .GenerateFeed
- dependencies
- Context
- models.go
- compilerOptions
- Client
- PostgresRepository
- PlisioService
- Post
- MediaService
- middleware.go
- setupTestApp
- VideoControls.tsx
- handler.go
- RateLimiter
- responseWriterWrapper
- NewMemoryRepository
- setupTestRouter
- NewHandler
- .GetBillingHistory
- setupRecommendationRouter
- .GetActiveSubscription
- route.ts
- .GetMediaObjectByID
- .GetFollowRequests
- .GetMessages
- .CreateOutboxEvent
- .GetUserBadges
- .CreateSession
- .GetVerificationApplication
- .GetUserTopicPreferences
- test_all_apis.py
- next-env.d.ts
- tailwind.config.ts
- .RecordFeedFeedback
- .RecordFeedImpressions
- .GetActiveWeights
- .GetSocialRelationship
- .GetTrendingTopics
- monitoring.sh
- github.com/x-reconstruction/api
- Context
- Request
- ResponseWriter
- UUID
- Context
- Conversation
- Entitlements
- FollowRequest
- MediaObject
- Message
- Notification
- Post
- SocialRelationship
- Subscription
- SubscriptionPlan
- User
- UUID
- VerificationBadge

## God Nodes (most connected - your core abstractions)
1. `MemoryRepository` - 123 edges
2. `PostgresRepository` - 99 edges
3. `Handler` - 92 edges
4. `writeJSON()` - 75 edges
5. `writeError()` - 72 edges
6. `apiRequest()` - 35 edges
7. `User` - 26 edges
8. `cn()` - 25 edges
9. `Post` - 17 edges
10. `compilerOptions` - 16 edges

## Surprising Connections (you probably didn't know these)
- `BenchmarkScoringEngine()` --calls--> `DefaultRankingWeights()`  [INFERRED]
  services/api/internal/recommendation/engine_test.go → services/api/internal/domain/models.go
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

## Communities (71 total, 39 thin omitted)

### Community 0 - "index.ts"
Cohesion: 0.05
Nodes (89): BookmarksPage(), ExplorePage(), metadata, FeedResponse, InvoiceContent(), InvoicePage(), WhiteLabelInvoice, PremiumPage() (+81 more)

### Community 1 - "MemoryRepository"
Cohesion: 0.05
Nodes (18): PaymentOrder, MemoryRepository, Context, Conversation, Entitlements, FollowRequest, MediaObject, Message (+10 more)

### Community 2 - "Handler"
Cohesion: 0.11
Nodes (7): Handler, Request, ResponseWriter, writeError(), writeJSON(), ConcurrencyLimiter(), RequestTimeoutMiddleware()

### Community 3 - ".GenerateFeed"
Cohesion: 0.06
Nodes (42): B, RankingWeights, Candidate, CandidateProvider, CandidateSource, DiversityReranker, Engine, Features (+34 more)

### Community 4 - "dependencies"
Cohesion: 0.05
Nodes (42): dependencies, clsx, lucide-react, next, next-auth, react, react-dom, tailwind-merge (+34 more)

### Community 5 - "Context"
Cohesion: 0.10
Nodes (5): Context, Conversation, Notification, RecentSearch, UUID

### Community 6 - "models.go"
Cohesion: 0.13
Nodes (37): Conversation, Entitlements, FeedFeedback, FeedImpression, Follow, FollowRequest, Hashtag, Invoice (+29 more)

### Community 7 - "compilerOptions"
Cohesion: 0.07
Nodes (26): compilerOptions, allowJs, esModuleInterop, incremental, isolatedModules, jsx, lib, module (+18 more)

### Community 8 - "Client"
Cohesion: 0.17
Nodes (13): Client, Hub, InboundEvent, OutboundEvent, Conn, Context, RawMessage, Request (+5 more)

### Community 9 - "PostgresRepository"
Cohesion: 0.16
Nodes (5): DB, PostgresRepository, NewPostgresRepository(), SubscriptionPlan, User

### Community 10 - "PlisioService"
Cohesion: 0.18
Nodes (12): PlisioCryptoInfo, PlisioInvoiceResponse, PlisioOperationsResponse, PlisioService, PlisioWebhookPayload, PlisioWhiteLabelInvoice, GetSupportedCryptos(), Context (+4 more)

### Community 11 - "Post"
Cohesion: 0.17
Nodes (4): FeedCursor, Post, Rows, SearchFilterParams

### Community 12 - "MediaService"
Cohesion: 0.19
Nodes (12): Config, Image, MediaService, Reader, getEnv(), getEnvInt(), Load(), Context (+4 more)

### Community 13 - "middleware.go"
Cohesion: 0.23
Nodes (10): Logger, contextKey, AuthMiddleware(), GetUserFromContext(), GetViewerID(), Context, User, UUID (+2 more)

### Community 14 - "setupTestApp"
Cohesion: 0.36
Nodes (9): Repository, createTestUser(), Mux, T, User, setupTestApp(), TestDirectMessaging_IdempotencyAndFlow(), TestFollowRequest_PrivateAccount() (+1 more)

### Community 15 - "VideoControls.tsx"
Cohesion: 0.31
Nodes (7): formatTime(), SPEED_OPTIONS, VideoControls(), VideoControlsProps, formatTime(), VideoProgressBar(), VideoProgressBarProps

### Community 16 - "handler.go"
Cohesion: 0.22
Nodes (8): CheckoutRequest, FeedFeedbackRequest, FeedImpressionsRequest, GoogleAuthRequest, LoginRequest, RegisterRequest, SimulatePaymentRequest, VerificationAppRequest

### Community 17 - "RateLimiter"
Cohesion: 0.33
Nodes (5): Duration, memoryLimiter, Mutex, Time, RateLimiter()

### Community 18 - "responseWriterWrapper"
Cohesion: 0.33
Nodes (4): responseWriterWrapper, ReadWriter, Conn, ResponseWriter

### Community 19 - "NewMemoryRepository"
Cohesion: 0.29
Nodes (4): DefaultRankingWeights(), T, TestSearchAPIEndpoints(), NewMemoryRepository()

### Community 20 - "setupTestRouter"
Cohesion: 0.52
Nodes (6): Mux, T, setupTestRouter(), TestAuthAndPostingFlow(), TestHealthCheck(), TestSubscriptionsAndBadges()

### Community 21 - "NewHandler"
Cohesion: 0.33
Nodes (6): Client, Hub, MediaService, PlisioService, Repository, NewHandler()

### Community 23 - "setupRecommendationRouter"
Cohesion: 0.60
Nodes (5): Mux, T, registerTestUser(), setupRecommendationRouter(), TestRecommendationAndFeedAPIs()

## Knowledge Gaps
- **92 isolated node(s):** `name`, `version`, `private`, `dev`, `build` (+87 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **39 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Handler` connect `Handler` to `middleware.go`, `setupTestApp`, `handler.go`, `RateLimiter`, `NewHandler`?**
  _High betweenness centrality (0.234) - this node is a cross-community bridge._
- **Why does `NewMemoryRepository()` connect `NewMemoryRepository` to `MemoryRepository`, `setupTestRouter`, `setupTestApp`, `setupRecommendationRouter`?**
  _High betweenness centrality (0.224) - this node is a cross-community bridge._
- **Why does `setupTestApp()` connect `setupTestApp` to `Handler`, `Client`, `middleware.go`, `NewMemoryRepository`, `NewHandler`?**
  _High betweenness centrality (0.213) - this node is a cross-community bridge._
- **What connects `name`, `version`, `private` to the rest of the system?**
  _92 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `index.ts` be split into smaller, more focused modules?**
  _Cohesion score 0.051387461459403906 - nodes in this community are weakly interconnected._
- **Should `MemoryRepository` be split into smaller, more focused modules?**
  _Cohesion score 0.05308924485125858 - nodes in this community are weakly interconnected._
- **Should `Handler` be split into smaller, more focused modules?**
  _Cohesion score 0.11226944667201283 - nodes in this community are weakly interconnected._