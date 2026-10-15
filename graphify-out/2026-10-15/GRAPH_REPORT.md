# Graph Report - .  (2026-10-15)

## Corpus Check
- cluster-only mode — file stats not available

## Summary
- 780 nodes · 2204 edges · 41 communities (23 shown, 18 thin omitted)
- Extraction: 98% EXTRACTED · 2% INFERRED · 0% AMBIGUOUS · INFERRED: 41 edges (avg confidence: 0.8)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `3c7b95c5`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- apiRequest
- MemoryRepository
- Handler
- NewHandler
- .GenerateFeed
- PostgresRepository
- models.go
- dependencies
- compilerOptions
- Client
- PlisioService
- MediaService
- Post
- User
- handler.go
- VideoControls.tsx
- .GetUserBadges
- .GetMessages
- .GetActiveSubscription
- route.ts
- .GetNotifications
- .GetConversationByID
- .GetFollowRequests
- .GetMediaObjectByID
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
6. `apiRequest()` - 36 edges
7. `cn()` - 27 edges
8. `User` - 23 edges
9. `compilerOptions` - 16 edges
10. `Post` - 16 edges

## Surprising Connections (you probably didn't know these)
- `BenchmarkScoringEngine()` --calls--> `DefaultRankingWeights()`  [INFERRED]
  services/api/internal/recommendation/engine_test.go → services/api/internal/domain/models.go
- `NewMemoryRepository()` --calls--> `DefaultRankingWeights()`  [INFERRED]
  services/api/internal/repository/memory.go → services/api/internal/domain/models.go
- `setupTestApp()` --calls--> `NewHub()`  [INFERRED]
  services/api/internal/handler/social_messaging_test.go → services/api/internal/realtime/hub.go
- `ExplorePage()` --references--> `react`  [EXTRACTED]
  apps/web/src/app/explore/page.tsx → apps/web/package.json
- `BookmarksPage()` --calls--> `apiRequest()`  [EXTRACTED]
  apps/web/src/app/bookmarks/page.tsx → apps/web/src/lib/api.ts

## Import Cycles
- None detected.

## Communities (41 total, 18 thin omitted)

### Community 0 - "apiRequest"
Cohesion: 0.06
Nodes (80): BookmarksPage(), ExplorePage(), metadata, MessagesPage(), NotificationsPage(), FeedResponse, HomePage(), InvoiceContent() (+72 more)

### Community 1 - "MemoryRepository"
Cohesion: 0.05
Nodes (17): MemoryRepository, Context, Conversation, Entitlements, FollowRequest, MediaObject, Message, Notification (+9 more)

### Community 2 - "Handler"
Cohesion: 0.13
Nodes (8): Context, Handler, Request, ResponseWriter, writeError(), writeJSON(), ConcurrencyLimiter(), RequestTimeoutMiddleware()

### Community 3 - "NewHandler"
Cohesion: 0.05
Nodes (48): Client, Duration, Hub, Logger, MediaService, contextKey, memoryLimiter, responseWriterWrapper (+40 more)

### Community 4 - ".GenerateFeed"
Cohesion: 0.06
Nodes (41): B, Candidate, CandidateProvider, CandidateSource, DiversityReranker, Engine, Features, FeedCursor (+33 more)

### Community 5 - "PostgresRepository"
Cohesion: 0.10
Nodes (4): DB, PostgresRepository, Context, UUID

### Community 6 - "models.go"
Cohesion: 0.11
Nodes (40): Conversation, Entitlements, FeedFeedback, FeedImpression, Follow, FollowRequest, Hashtag, Invoice (+32 more)

### Community 7 - "dependencies"
Cohesion: 0.05
Nodes (42): dependencies, clsx, lucide-react, next, next-auth, react, react-dom, tailwind-merge (+34 more)

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

### Community 14 - "handler.go"
Cohesion: 0.22
Nodes (10): CheckoutRequest, CreatePostRequest, FeedFeedbackRequest, FeedImpressionsRequest, GoogleAuthRequest, LoginRequest, RegisterRequest, SimulatePaymentRequest (+2 more)

### Community 15 - "VideoControls.tsx"
Cohesion: 0.31
Nodes (7): formatTime(), SPEED_OPTIONS, VideoControls(), VideoControlsProps, formatTime(), VideoProgressBar(), VideoProgressBarProps

## Knowledge Gaps
- **77 isolated node(s):** `name`, `version`, `private`, `dev`, `build` (+72 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **18 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `NewMemoryRepository()` connect `NewHandler` to `MemoryRepository`, `models.go`?**
  _High betweenness centrality (0.244) - this node is a cross-community bridge._
- **Why does `setupTestApp()` connect `NewHandler` to `Client`, `Handler`?**
  _High betweenness centrality (0.212) - this node is a cross-community bridge._
- **Why does `MemoryRepository` connect `MemoryRepository` to `NewHandler`, `models.go`?**
  _High betweenness centrality (0.196) - this node is a cross-community bridge._
- **What connects `name`, `version`, `private` to the rest of the system?**
  _77 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `apiRequest` be split into smaller, more focused modules?**
  _Cohesion score 0.059206069638260396 - nodes in this community are weakly interconnected._
- **Should `MemoryRepository` be split into smaller, more focused modules?**
  _Cohesion score 0.053407855922993325 - nodes in this community are weakly interconnected._
- **Should `Handler` be split into smaller, more focused modules?**
  _Cohesion score 0.1290491118077325 - nodes in this community are weakly interconnected._