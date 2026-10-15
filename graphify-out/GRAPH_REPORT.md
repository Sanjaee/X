# Graph Report - .  (2026-10-16)

## Corpus Check
- cluster-only mode — file stats not available

## Summary
- 800 nodes · 2265 edges · 44 communities (24 shown, 20 thin omitted)
- Extraction: 98% EXTRACTED · 2% INFERRED · 0% AMBIGUOUS · INFERRED: 41 edges (avg confidence: 0.8)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `f66560b1`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- index.ts
- MemoryRepository
- Handler
- NewHandler
- PostgresRepository
- .GenerateFeed
- models.go
- dependencies
- compilerOptions
- Client
- PlisioService
- MediaService
- Post
- MockCandidateProvider
- handler.go
- User
- VideoControls.tsx
- .GetUserBadges
- .GetMessages
- .GetActiveSubscription
- route.ts
- .GetNotifications
- .GetConversationByID
- .GetFollowRequests
- .GetMediaObjectByID
- test_all_apis.py
- next-env.d.ts
- tailwind.config.ts
- .GetSocialRelationship
- monitoring.sh
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
7. `cn()` - 26 edges
8. `User` - 26 edges
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

## Communities (44 total, 20 thin omitted)

### Community 0 - "index.ts"
Cohesion: 0.05
Nodes (86): BookmarksPage(), ExplorePage(), metadata, FeedResponse, InvoiceContent(), InvoicePage(), WhiteLabelInvoice, PremiumPage() (+78 more)

### Community 1 - "MemoryRepository"
Cohesion: 0.05
Nodes (17): MemoryRepository, Context, Conversation, Entitlements, FollowRequest, MediaObject, Message, Notification (+9 more)

### Community 2 - "Handler"
Cohesion: 0.13
Nodes (8): Context, Handler, Request, ResponseWriter, writeError(), writeJSON(), ConcurrencyLimiter(), RequestTimeoutMiddleware()

### Community 3 - "NewHandler"
Cohesion: 0.05
Nodes (48): Client, Duration, Hub, Logger, MediaService, contextKey, memoryLimiter, responseWriterWrapper (+40 more)

### Community 4 - "PostgresRepository"
Cohesion: 0.10
Nodes (4): DB, PostgresRepository, Context, UUID

### Community 5 - ".GenerateFeed"
Cohesion: 0.07
Nodes (37): B, Candidate, CandidateProvider, CandidateSource, DiversityReranker, Engine, Features, FeedCursor (+29 more)

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

### Community 13 - "MockCandidateProvider"
Cohesion: 0.40
Nodes (4): MockCandidateProvider, Context, Post, UUID

### Community 14 - "handler.go"
Cohesion: 0.20
Nodes (11): CheckoutRequest, CreatePostRequest, FeedFeedbackRequest, FeedImpressionsRequest, GoogleAuthRequest, LoginRequest, RegisterRequest, SimulatePaymentRequest (+3 more)

### Community 16 - "VideoControls.tsx"
Cohesion: 0.31
Nodes (7): formatTime(), SPEED_OPTIONS, VideoControls(), VideoControlsProps, formatTime(), VideoProgressBar(), VideoProgressBarProps

## Knowledge Gaps
- **89 isolated node(s):** `name`, `version`, `private`, `dev`, `build` (+84 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **20 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `NewMemoryRepository()` connect `NewHandler` to `MemoryRepository`, `models.go`?**
  _High betweenness centrality (0.233) - this node is a cross-community bridge._
- **Why does `setupTestApp()` connect `NewHandler` to `Client`, `Handler`?**
  _High betweenness centrality (0.202) - this node is a cross-community bridge._
- **Why does `MemoryRepository` connect `MemoryRepository` to `NewHandler`, `models.go`?**
  _High betweenness centrality (0.187) - this node is a cross-community bridge._
- **What connects `name`, `version`, `private` to the rest of the system?**
  _89 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `index.ts` be split into smaller, more focused modules?**
  _Cohesion score 0.053703703703703705 - nodes in this community are weakly interconnected._
- **Should `MemoryRepository` be split into smaller, more focused modules?**
  _Cohesion score 0.053407855922993325 - nodes in this community are weakly interconnected._
- **Should `Handler` be split into smaller, more focused modules?**
  _Cohesion score 0.1290491118077325 - nodes in this community are weakly interconnected._