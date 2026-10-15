package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
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

	// Static uploads file server
	fileServer := http.FileServer(http.Dir(cfg.StorageDir))
	r.Handle("/uploads/*", http.StripPrefix("/uploads", fileServer))

	// Health check & observability endpoints
	r.Get("/health/live", h.HealthLive)
	r.Get("/health/ready", h.HealthReady)
	r.Get("/health/metrics", h.HealthMetrics)

	// API v1 Routes
	r.Route("/api/v1", func(v1 chi.Router) {
		// WebSocket Realtime Gateway (Section 4G & 10)
		v1.Get("/ws", h.ServeWebSocket)

		// Public Auth routes
		v1.Post("/auth/register", h.Register)
		v1.Post("/auth/login", h.Login)
		v1.Post("/auth/google", h.GoogleAuth)

		// Public Webhook for Plisio Cryptocurrency Gateway (Section 6F)
		v1.Post("/webhooks/payments/plisio", h.HandlePlisioWebhook)

		// Publicly readable endpoints (with optional viewer identity context)
		v1.Group(func(pub chi.Router) {
			pub.Use(middleware.AuthMiddleware(repo, true)) // optional auth
			pub.Get("/feed", h.GetFeed)
			pub.Get("/users/{username}", h.GetUserByUsername)
			pub.Get("/users/{username}/posts", h.GetUserPosts)
			pub.Get("/users/{username}/badges", h.GetUserBadges)
			pub.Get("/posts/{postId}", h.GetPostByID)
			pub.Get("/posts/{postId}/replies", h.GetPostReplies)
			pub.Get("/search", h.Search)
			pub.Get("/search/posts", h.SearchPosts)
			pub.Get("/search/users", h.SearchUsers)
			pub.Get("/search/suggestions", h.GetSearchSuggestions)
			pub.Get("/subscription-plans", h.GetSubscriptionPlans)
			pub.Get("/subscriptions/cryptos", h.GetCryptoCurrencies)

			// Social Graph public reads
			pub.Get("/users/{username}/followers", h.GetFollowers)
			pub.Get("/users/{username}/following", h.GetFollowing)

			// Recommendation & Discovery
			pub.Get("/explore/trending", h.GetTrendingTopics)
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
			auth.Delete("/posts/{postId}", h.DeletePost)

			// Interactions
			auth.Post("/posts/{postId}/likes", h.LikePost)
			auth.Delete("/posts/{postId}/likes", h.UnlikePost)
			auth.Post("/posts/{postId}/reposts", h.RepostPost)
			auth.Delete("/posts/{postId}/reposts", h.UndoRepost)
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
			auth.Get("/me/topic-preferences", h.GetTopicPreferences)
			auth.Put("/me/topic-preferences", h.UpdateTopicPreference)
			auth.Post("/me/personalization/reset", h.ResetPersonalization)

			// Search History
			auth.Get("/search/recent", h.GetRecentSearches)
			auth.Delete("/search/recent", h.ClearRecentSearches)
			auth.Delete("/search/recent/{searchId}", h.DeleteRecentSearch)
		})
	})


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
