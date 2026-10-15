# Graph Report - .  (2026-10-15)

## Corpus Check
- cluster-only mode — file stats not available

## Summary
- 732 nodes · 2196 edges · 31 communities (17 shown, 14 thin omitted)
- Extraction: 95% EXTRACTED · 5% INFERRED · 0% AMBIGUOUS · INFERRED: 108 edges (avg confidence: 0.8)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `74a133bd`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- MemoryRepository
- apiRequest
- Handler
- .GenerateFeed
- PostgresRepository
- dependencies
- handler.go
- models.go
- PlisioService
- compilerOptions
- Client
- middleware.go
- Post
- User
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

## God Nodes (most connected - your core abstractions)
1. `MemoryRepository` - 123 edges
2. `PostgresRepository` - 99 edges
3. `Handler` - 92 edges
4. `writeJSON()` - 75 edges
5. `writeError()` - 72 edges
6. `GetUserFromContext()` - 54 edges
7. `apiRequest()` - 35 edges
8. `cn()` - 29 edges
9. `User` - 23 edges
10. `GetViewerID()` - 18 edges

## Surprising Connections (you probably didn't know these)
- `NewMemoryRepository()` --calls--> `DefaultRankingWeights()`  [INFERRED]
  services/api/internal/repository/memory.go → services/api/internal/domain/models.go
- `setupTestApp()` --calls--> `NewHub()`  [INFERRED]
  services/api/internal/handler/social_messaging_test.go → services/api/internal/realtime/hub.go
- `RateLimiter()` --calls--> `GetUserFromContext()`  [INFERRED]
  services/api/internal/middleware/ratelimit.go → services/api/internal/middleware/middleware.go
- `ExplorePage()` --references--> `react`  [EXTRACTED]
  apps/web/src/app/explore/page.tsx → apps/web/package.json
- `BookmarksPage()` --calls--> `apiRequest()`  [EXTRACTED]
  apps/web/src/app/bookmarks/page.tsx → apps/web/src/lib/api.ts

## Import Cycles
- None detected.

## Communities (31 total, 14 thin omitted)

### Community 0 - "MemoryRepository"
Cohesion: 0.05
Nodes (17): MemoryRepository, Context, Conversation, Entitlements, FollowRequest, MediaObject, Message, Notification (+9 more)

### Community 1 - "apiRequest"
Cohesion: 0.08
Nodes (68): BookmarksPage(), ExplorePage(), metadata, MessagesPage(), NotificationsPage(), FeedResponse, HomePage(), InvoiceContent() (+60 more)

### Community 2 - "Handler"
Cohesion: 0.13
Nodes (14): Handler, Context, Request, ResponseWriter, writeError(), writeJSON(), ConcurrencyLimiter(), GetActiveRequests() (+6 more)

### Community 3 - ".GenerateFeed"
Cohesion: 0.06
Nodes (43): B, RankingWeights, Candidate, CandidateProvider, CandidateSource, DiversityReranker, Engine, Features (+35 more)

### Community 4 - "PostgresRepository"
Cohesion: 0.10
Nodes (4): DB, PostgresRepository, Context, UUID

### Community 5 - "dependencies"
Cohesion: 0.05
Nodes (42): dependencies, clsx, lucide-react, next, next-auth, react, react-dom, tailwind-merge (+34 more)

### Community 6 - "handler.go"
Cohesion: 0.08
Nodes (35): CheckoutRequest, CreatePostRequest, FeedFeedbackRequest, FeedImpressionsRequest, GoogleAuthRequest, LoginRequest, RegisterRequest, SimulatePaymentRequest (+27 more)

### Community 7 - "models.go"
Cohesion: 0.12
Nodes (38): Conversation, Entitlements, FeedFeedback, FeedImpression, Follow, FollowRequest, Hashtag, Invoice (+30 more)

### Community 8 - "PlisioService"
Cohesion: 0.09
Nodes (24): Config, Image, MediaService, PlisioCryptoInfo, PlisioInvoiceResponse, PlisioOperationsResponse, PlisioService, PlisioWebhookPayload (+16 more)

### Community 9 - "compilerOptions"
Cohesion: 0.07
Nodes (26): compilerOptions, allowJs, esModuleInterop, incremental, isolatedModules, jsx, lib, module (+18 more)

### Community 10 - "Client"
Cohesion: 0.17
Nodes (13): Client, Hub, InboundEvent, OutboundEvent, Conn, Context, RawMessage, Request (+5 more)

### Community 11 - "middleware.go"
Cohesion: 0.12
Nodes (13): Duration, Logger, contextKey, memoryLimiter, responseWriterWrapper, Mutex, ReadWriter, Conn (+5 more)

## Knowledge Gaps
- **71 isolated node(s):** `name`, `version`, `private`, `dev`, `build` (+66 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **14 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `MemoryRepository` connect `MemoryRepository` to `.GenerateFeed`, `handler.go`, `models.go`?**
  _High betweenness centrality (0.211) - this node is a cross-community bridge._
- **Why does `NewMemoryRepository()` connect `handler.go` to `MemoryRepository`, `.GenerateFeed`?**
  _High betweenness centrality (0.205) - this node is a cross-community bridge._
- **Why does `Handler` connect `Handler` to `PlisioService`, `Client`, `middleware.go`, `handler.go`?**
  _High betweenness centrality (0.161) - this node is a cross-community bridge._
- **What connects `name`, `version`, `private` to the rest of the system?**
  _71 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `MemoryRepository` be split into smaller, more focused modules?**
  _Cohesion score 0.053407855922993325 - nodes in this community are weakly interconnected._
- **Should `apiRequest` be split into smaller, more focused modules?**
  _Cohesion score 0.08075601374570447 - nodes in this community are weakly interconnected._
- **Should `Handler` be split into smaller, more focused modules?**
  _Cohesion score 0.13160355306217858 - nodes in this community are weakly interconnected._