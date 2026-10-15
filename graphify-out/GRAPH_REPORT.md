# Graph Report - .  (2026-10-15)

## Corpus Check
- cluster-only mode — file stats not available

## Summary
- 775 nodes · 2201 edges · 40 communities (25 shown, 15 thin omitted)
- Extraction: 98% EXTRACTED · 2% INFERRED · 0% AMBIGUOUS · INFERRED: 41 edges (avg confidence: 0.8)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `71002d29`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- apiRequest
- MemoryRepository
- Handler
- .GenerateFeed
- dependencies
- models.go
- Context
- PostgresRepository
- compilerOptions
- Client
- PlisioService
- MediaService
- Post
- NewHandler
- User
- handler.go
- setupTestApp
- responseWriterWrapper
- VideoControls.tsx
- middleware.go
- setupTestRouter
- setupRecommendationRouter
- .GetActiveSubscription
- route.ts
- .GetNotifications
- .GetConversationByID
- next-env.d.ts
- tailwind.config.ts
- .GetSocialRelationship
- github.com/x-reconstruction/api
- Context
- Request
- ResponseWriter
- UUID

## God Nodes (most connected - your core abstractions)
1. `MemoryRepository` - 123 edges
2. `PostgresRepository` - 99 edges
3. `Handler` - 92 edges
4. `writeJSON()` - 75 edges
5. `writeError()` - 72 edges
6. `apiRequest()` - 35 edges
7. `cn()` - 27 edges
8. `User` - 23 edges
9. `compilerOptions` - 16 edges
10. `Post` - 16 edges

## Surprising Connections (you probably didn't know these)
- `NewMemoryRepository()` --calls--> `DefaultRankingWeights()`  [INFERRED]
  services/api/internal/repository/memory.go → services/api/internal/domain/models.go
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

## Communities (40 total, 15 thin omitted)

### Community 0 - "apiRequest"
Cohesion: 0.06
Nodes (79): BookmarksPage(), ExplorePage(), metadata, MessagesPage(), NotificationsPage(), FeedResponse, HomePage(), InvoiceContent() (+71 more)

### Community 1 - "MemoryRepository"
Cohesion: 0.05
Nodes (17): MemoryRepository, Context, Conversation, Entitlements, FollowRequest, MediaObject, Message, Notification (+9 more)

### Community 2 - "Handler"
Cohesion: 0.13
Nodes (8): Context, Handler, Request, ResponseWriter, writeError(), writeJSON(), ConcurrencyLimiter(), RequestTimeoutMiddleware()

### Community 3 - ".GenerateFeed"
Cohesion: 0.06
Nodes (43): B, RankingWeights, Candidate, CandidateProvider, CandidateSource, DiversityReranker, Engine, Features (+35 more)

### Community 4 - "dependencies"
Cohesion: 0.05
Nodes (42): dependencies, clsx, lucide-react, next, next-auth, react, react-dom, tailwind-merge (+34 more)

### Community 5 - "models.go"
Cohesion: 0.12
Nodes (38): Conversation, Entitlements, FeedFeedback, FeedImpression, Follow, FollowRequest, Hashtag, Invoice (+30 more)

### Community 7 - "PostgresRepository"
Cohesion: 0.08
Nodes (8): DB, PostgresRepository, FollowRequest, MediaObject, Message, SubscriptionPlan, VerificationBadge, NewPostgresRepository()

### Community 8 - "compilerOptions"
Cohesion: 0.07
Nodes (26): compilerOptions, allowJs, esModuleInterop, incremental, isolatedModules, jsx, lib, module (+18 more)

### Community 9 - "Client"
Cohesion: 0.17
Nodes (13): Client, Hub, InboundEvent, OutboundEvent, Conn, Context, RawMessage, Request (+5 more)

### Community 10 - "PlisioService"
Cohesion: 0.18
Nodes (12): PlisioCryptoInfo, PlisioInvoiceResponse, PlisioOperationsResponse, PlisioService, PlisioWebhookPayload, PlisioWhiteLabelInvoice, GetSupportedCryptos(), Context (+4 more)

### Community 11 - "MediaService"
Cohesion: 0.19
Nodes (12): Config, Image, MediaService, Reader, getEnv(), getEnvInt(), Load(), Context (+4 more)

### Community 13 - "NewHandler"
Cohesion: 0.17
Nodes (9): Client, Hub, MediaService, PlisioService, Repository, NewHandler(), T, TestSearchAPIEndpoints() (+1 more)

### Community 15 - "handler.go"
Cohesion: 0.22
Nodes (10): CheckoutRequest, CreatePostRequest, FeedFeedbackRequest, FeedImpressionsRequest, GoogleAuthRequest, LoginRequest, RegisterRequest, SimulatePaymentRequest (+2 more)

### Community 16 - "setupTestApp"
Cohesion: 0.36
Nodes (9): Repository, createTestUser(), Mux, T, User, setupTestApp(), TestDirectMessaging_IdempotencyAndFlow(), TestFollowRequest_PrivateAccount() (+1 more)

### Community 17 - "responseWriterWrapper"
Cohesion: 0.33
Nodes (4): responseWriterWrapper, ReadWriter, Conn, ResponseWriter

### Community 18 - "VideoControls.tsx"
Cohesion: 0.31
Nodes (7): formatTime(), SPEED_OPTIONS, VideoControls(), VideoControlsProps, formatTime(), VideoProgressBar(), VideoProgressBarProps

### Community 19 - "middleware.go"
Cohesion: 0.14
Nodes (15): Duration, Logger, contextKey, memoryLimiter, Mutex, AuthMiddleware(), GetUserFromContext(), GetViewerID() (+7 more)

### Community 21 - "setupTestRouter"
Cohesion: 0.52
Nodes (6): Mux, T, setupTestRouter(), TestAuthAndPostingFlow(), TestHealthCheck(), TestSubscriptionsAndBadges()

### Community 22 - "setupRecommendationRouter"
Cohesion: 0.60
Nodes (5): Mux, T, registerTestUser(), setupRecommendationRouter(), TestRecommendationAndFeedAPIs()

## Knowledge Gaps
- **77 isolated node(s):** `name`, `version`, `private`, `dev`, `build` (+72 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **15 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `NewMemoryRepository()` connect `NewHandler` to `MemoryRepository`, `.GenerateFeed`, `setupTestApp`, `setupTestRouter`, `setupRecommendationRouter`?**
  _High betweenness centrality (0.248) - this node is a cross-community bridge._
- **Why does `setupTestApp()` connect `setupTestApp` to `Client`, `Handler`, `middleware.go`, `NewHandler`?**
  _High betweenness centrality (0.215) - this node is a cross-community bridge._
- **Why does `MemoryRepository` connect `MemoryRepository` to `NewHandler`, `.GenerateFeed`, `models.go`?**
  _High betweenness centrality (0.198) - this node is a cross-community bridge._
- **What connects `name`, `version`, `private` to the rest of the system?**
  _77 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `apiRequest` be split into smaller, more focused modules?**
  _Cohesion score 0.05991735537190083 - nodes in this community are weakly interconnected._
- **Should `MemoryRepository` be split into smaller, more focused modules?**
  _Cohesion score 0.053407855922993325 - nodes in this community are weakly interconnected._
- **Should `Handler` be split into smaller, more focused modules?**
  _Cohesion score 0.1290491118077325 - nodes in this community are weakly interconnected._