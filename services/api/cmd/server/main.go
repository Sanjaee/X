package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/redis/go-redis/v9"
	"github.com/x-reconstruction/api/internal/config"
	"github.com/x-reconstruction/api/internal/handler"
	"github.com/x-reconstruction/api/internal/media"
	"github.com/x-reconstruction/api/internal/middleware"
	"github.com/x-reconstruction/api/internal/payment"
	"github.com/x-reconstruction/api/internal/realtime"
	"github.com/x-reconstruction/api/internal/repository"
)


func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	cfg := config.Load()
	logger.Info("starting_service", slog.String("port", cfg.Port), slog.String("db_url", cfg.DatabaseURL))

	// Connect to Database with graceful fallback
	var repo repository.Repository
	pgRepo, err := repository.NewPostgresRepository(cfg.DatabaseURL)
	if err != nil {
		logger.Warn("postgres_not_available_using_memory_repository", slog.String("error", err.Error()))
		repo = repository.NewMemoryRepository()
	} else {
		logger.Info("connected_to_postgres_successfully")
		repo = pgRepo
	}

	// Connect to Redis
	var rdb *redis.Client
	opt, err := redis.ParseURL(cfg.RedisURL)
	if err == nil {
		rdb = redis.NewClient(opt)
		ctxPing, cancelPing := context.WithTimeout(context.Background(), 2*time.Second)
		if err := rdb.Ping(ctxPing).Err(); err != nil {
			logger.Warn("redis_not_available_using_memory_realtime", slog.String("error", err.Error()))
		} else {
			logger.Info("connected_to_redis_successfully")
		}
		cancelPing()
	} else {
		logger.Warn("invalid_redis_url", slog.String("error", err.Error()))
	}

	// Initialize Realtime Hub
	hub := realtime.NewHub(rdb)
	ctxHub, cancelHub := context.WithCancel(context.Background())
	defer cancelHub()
	go hub.Run(ctxHub)

	mediaSvc := media.NewMediaService(cfg)
	plisioSvc := payment.NewPlisioService(cfg)
	h := handler.NewHandler(repo, cfg.StorageDir, mediaSvc, plisioSvc, hub, rdb)

	// Start Outbox Worker (Transactional Outbox pattern for asynchronous event fanout)
	go h.StartOutboxWorker(ctxHub)

	r := chi.NewRouter()

	// Base middleware
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(middleware.LoggerMiddleware(logger))
	r.Use(middleware.RecoveryMiddleware(logger))

	// Resiliency & Backpressure middleware (4GB VPS Tuning)
	r.Use(middleware.ConcurrencyLimiter(cfg.MaxConcurrentRequests))
	r.Use(middleware.RateLimiter(rdb, cfg.RateLimitPerMin))
	r.Use(middleware.RequestTimeoutMiddleware(15))

	// CORS configuration
	allowedOrigins := []string{
		"http://localhost:3000", "http://127.0.0.1:3000",
		"http://localhost:80", "http://127.0.0.1:80", "http://localhost",
		"https://lostmedia.pro", "https://www.lostmedia.pro", "http://lostmedia.pro",
	}
	if cfg.AppURL != "" {
		allowedOrigins = append(allowedOrigins, cfg.AppURL)
	}

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link", "Set-Cookie"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// LM-SEC-003 Remediation: Safe static uploads file server (directory listing strictly disabled)
	uploadFS := http.Dir(cfg.StorageDir)
	r.Get("/uploads/*", func(w http.ResponseWriter, r *http.Request) {
		subpath := strings.TrimPrefix(r.URL.Path, "/uploads")
		if subpath == "" || strings.HasSuffix(subpath, "/") {
			http.NotFound(w, r)
			return
		}
		f, err := uploadFS.Open(subpath)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		defer f.Close()
		stat, err := f.Stat()
		if err != nil || stat.IsDir() {
			http.NotFound(w, r)
			return
		}
		http.StripPrefix("/uploads", http.FileServer(uploadFS)).ServeHTTP(w, r)
	})

	// Health check & observability endpoints
	r.Get("/health/live", h.HealthLive)
	r.Get("/health/ready", h.HealthReady)

	// LM-SEC-009 Remediation: Restrict runtime metrics endpoint to localhost or internal queries
	r.Get("/health/metrics", func(w http.ResponseWriter, r *http.Request) {
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			host = r.RemoteAddr
		}
		// Allow loopback (127.0.0.1, ::1) or authenticated header
		if host != "127.0.0.1" && host != "::1" && host != "localhost" && r.Header.Get("X-Internal-Token") != cfg.SessionSecret {
			http.Error(w, `{"error":"forbidden: internal metrics only"}`, http.StatusForbidden)
			return
		}
		h.HealthMetrics(w, r)
	})

	// API Routes (mounted on both /api/v1 and /api for seamless frontend compatibility)
	registerAPIRoutes := func(v1 chi.Router) {
		// WebSocket Realtime Gateway (Section 4G & 10)
		v1.Get("/ws", h.ServeWebSocket)

		// LM-SEC-013 Remediation: Dedicated strict rate limiting for public authentication routes (10 req/min)
		authLimiter := middleware.RateLimiter(rdb, 10)
		v1.Group(func(authPub chi.Router) {
			authPub.Use(authLimiter)
			authPub.Post("/auth/register", h.Register)
			authPub.Post("/auth/login", h.Login)
		})
		v1.Post("/auth/google", h.GoogleAuth)

		// Public Webhook for Plisio Cryptocurrency Gateway (Section 6F)
		v1.Post("/webhooks/payments/plisio", h.HandlePlisioWebhook)

		// Publicly readable endpoints (with optional viewer identity context)
		v1.Group(func(pub chi.Router) {
			pub.Use(middleware.AuthMiddleware(repo, true)) // optional auth
			pub.Get("/feed", h.GetFeed)
			pub.Get("/feed/new-posts", h.GetNewPosts)
			pub.Get("/users/{username}", h.GetUserByUsername)
			pub.Get("/users/{username}/posts", h.GetUserPosts)
			pub.Get("/users/{username}/badges", h.GetUserBadges)
			pub.Get("/posts/{postId}", h.GetPostByID)
			pub.Get("/posts/{postId}/replies", h.GetPostReplies)
			pub.Get("/posts/{postId}/quotes", h.GetPostQuotes)
			pub.Get("/search", h.Search)
			pub.Get("/search/posts", h.SearchPosts)
			pub.Get("/search/users", h.SearchUsers)
			pub.Get("/search/suggestions", h.GetSearchSuggestions)
			pub.Get("/subscription-plans", h.GetSubscriptionPlans)
			pub.Get("/membership/plans", h.GetMembershipPlansCatalog)
			pub.Get("/subscriptions/cryptos", h.GetCryptoCurrencies)

			// Social Graph public reads
			pub.Get("/users/{username}/followers", h.GetFollowers)
			pub.Get("/users/{username}/following", h.GetFollowing)

			// Recommendation & Discovery
			pub.Get("/explore/trending", h.GetTrendingTopics)
			pub.Get("/users/recommendations", h.GetWhoToFollow)
			pub.Post("/feed/impressions", h.RecordFeedImpressions)
		})

		// Authenticated protected routes
		v1.Group(func(auth chi.Router) {
			auth.Use(middleware.AuthMiddleware(repo, false)) // strictly required auth

			auth.Post("/auth/logout", h.Logout)
			auth.Get("/me", h.GetMe)
			auth.Patch("/me/profile", h.UpdateProfile)

			// Posts
			auth.Post("/posts", h.CreatePost)
			auth.Patch("/posts/{postId}", h.PatchPost)
			auth.Delete("/posts/{postId}", h.DeletePost)

			// Membership
			auth.Get("/membership/me", h.GetMyMembershipStatus)

			// Interactions (Idempotent Like, Repost, Bookmark)
			auth.Post("/posts/{postId}/likes", h.LikePost)
			auth.Put("/posts/{postId}/likes", h.LikePost)
			auth.Post("/posts/{postId}/like", h.LikePost)
			auth.Put("/posts/{postId}/like", h.LikePost)
			auth.Delete("/posts/{postId}/likes", h.UnlikePost)
			auth.Delete("/posts/{postId}/like", h.UnlikePost)

			auth.Post("/posts/{postId}/reposts", h.RepostPost)
			auth.Post("/posts/{postId}/repost", h.RepostPost)
			auth.Delete("/posts/{postId}/reposts", h.UndoRepost)
			auth.Delete("/posts/{postId}/repost", h.UndoRepost)

			auth.Post("/posts/{postId}/bookmarks", h.BookmarkPost)
			auth.Delete("/posts/{postId}/bookmarks", h.RemoveBookmark)

			// Social Graph (Follow, Block, Mute, Mutuals, Relationships)
			auth.Post("/users/{username}/follow", h.FollowUser)
			auth.Delete("/users/{username}/follow", h.UnfollowUser)
			auth.Get("/users/{username}/mutuals", h.GetMutualConnections)
			auth.Get("/users/{username}/relationship", h.GetSocialRelationship)
			auth.Post("/users/{username}/block", h.BlockUser)
			auth.Delete("/users/{username}/block", h.UnblockUser)
			auth.Post("/users/{username}/mute", h.MuteUser)
			auth.Delete("/users/{username}/mute", h.UnmuteUser)

			// Private Account Follow Requests
			auth.Post("/users/{username}/follow-request", h.CreateFollowRequest)
			auth.Delete("/users/{username}/follow-request", h.CancelFollowRequest)
			auth.Post("/users/{username}/follow-request/respond", h.RespondFollowRequestByUsername)
			auth.Get("/follow-requests", h.GetFollowRequests)
			auth.Post("/follow-requests/{requestId}/respond", h.RespondFollowRequest)

			// Realtime Direct Messaging (Conversations & Messages)
			auth.Get("/conversations", h.GetConversations)
			auth.Post("/conversations", h.CreateConversation)
			auth.Get("/conversations/{conversationId}", h.GetConversationByID)
			auth.Get("/conversations/{conversationId}/messages", h.GetMessages)
			auth.Post("/conversations/{conversationId}/messages", h.SendMessage)
			auth.Post("/conversations/{conversationId}/read", h.MarkConversationRead)

			// Notifications
			auth.Get("/notifications", h.GetNotifications)
			auth.Patch("/notifications/read", h.MarkNotificationsRead)
			auth.Get("/notifications/unread-count", h.GetUnreadNotificationsCount)
			auth.Post("/notifications/{notificationId}/read", h.MarkNotificationRead)

			// Media Upload (with automatic compression and R2 storage)
			auth.Post("/media/upload", h.UploadMedia)

			// Subscriptions & Monetization (Section 6A)
			auth.Get("/subscriptions/me", h.GetMySubscription)
			auth.Post("/subscriptions/checkout", h.CheckoutSubscription)
			auth.Get("/subscriptions/orders/{orderId}", h.GetSubscriptionOrder)
			auth.Post("/subscriptions/{subscriptionId}/cancel", h.CancelSubscription)
			auth.Post("/subscriptions/simulate-payment", h.SimulatePayment)
			auth.Get("/billing/history", h.GetBillingHistory)

			// Verification Application & Badges
			auth.Get("/verification/me", h.GetMyVerification)
			auth.Post("/verification/applications", h.ApplyVerification)

			// Feed Feedback & Personalization Controls
			auth.Post("/feed/feedback", h.RecordFeedFeedback)
			auth.Post("/users/recommendations/{userId}/dismiss", h.DismissUserRecommendation)
			auth.Get("/me/topic-preferences", h.GetTopicPreferences)
			auth.Put("/me/topic-preferences", h.UpdateTopicPreference)
			auth.Post("/me/personalization/reset", h.ResetPersonalization)

			// Search History
			auth.Get("/search/recent", h.GetRecentSearches)
			auth.Delete("/search/recent", h.ClearRecentSearches)
			auth.Delete("/search/recent/{searchId}", h.DeleteRecentSearch)

			// Admin & Governance Endpoints (RBAC, Memberships, Quotas, Audit Logs, Posts, Monitoring)
			adminPerms := middleware.RequireAnyPermission(repo, "role.read", "role.create", "role.update", "permission.manage", "user.read", "audit.read", "membership.plan.manage", "posts.moderate", "posts.delete.any", "monitoring.read")
			auth.Group(func(admin chi.Router) {
				admin.Use(adminPerms)

				// Overview & Monitoring
				admin.Get("/admin/overview", h.AdminGetOverview)
				admin.Get("/admin/monitoring/overview", h.AdminGetMonitoringOverview)
				admin.Get("/admin/security/events", h.AdminGetSecurityEvents)

				// Roles & Permissions
				admin.Get("/admin/roles", h.AdminGetRoles)
				admin.Post("/admin/roles", h.AdminCreateRole)
				admin.Get("/admin/roles/{id}", h.AdminGetRoleByID)
				admin.Patch("/admin/roles/{id}", h.AdminUpdateRole)
				admin.Delete("/admin/roles/{id}", h.AdminDeleteRole)
				admin.Get("/admin/permissions", h.AdminGetPermissions)
				admin.Get("/admin/roles/{id}/permissions", h.AdminGetRolePermissions)
				admin.Put("/admin/roles/{id}/permissions", h.AdminSetRolePermissions)

				// Membership Plans & Entitlements
				admin.Get("/admin/memberships", h.AdminGetMembershipPlans)
				admin.Patch("/admin/memberships/{id}", h.AdminUpdateMembershipPlan)
				admin.Put("/admin/plans/{id}/entitlements", h.AdminSetPlanEntitlements)
				admin.Post("/admin/users/{id}/membership", h.AdminAssignUserMembership)
				admin.Delete("/admin/users/{id}/membership/{planId}", h.AdminRevokeUserMembership)

				// User Administration
				admin.Get("/admin/users", h.AdminGetUsers)
				admin.Get("/admin/users/{id}", h.AdminGetUserByID)
				admin.Patch("/admin/users/{id}/roles", h.AdminUpdateUserRoles)
				admin.Patch("/admin/users/{id}/status", h.AdminUpdateUserStatus)

				// Post Moderation & Management
				admin.Get("/admin/posts", h.AdminGetPosts)
				admin.Patch("/admin/posts/{id}/moderation", h.AdminUpdatePostModeration)
				admin.Delete("/admin/posts/{id}", h.AdminDeletePost)

				// Audit Logs
				admin.Get("/admin/audit-logs", h.AdminGetAuditLogs)

				// System Maintenance
				admin.Post("/admin/maintenance/reconcile-counters", h.ReconcileCounters)
			})
		})
	}

	r.Route("/api/v1", registerAPIRoutes)
	r.Route("/api", registerAPIRoutes)


	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Server runner
	go func() {
		logger.Info("server_listening", slog.String("address", server.Addr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server_listen_error", slog.String("error", err.Error()))
		}
	}()

	// Graceful shutdown listener
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	logger.Info("shutting_down_gracefully")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("server_shutdown_failed", slog.String("error", err.Error()))
	}
	logger.Info("server_stopped")
}
