package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/x-reconstruction/api/internal/domain"
	"github.com/x-reconstruction/api/internal/recommendation"
)

type Repository interface {
	// User & Profile
	CreateUser(ctx context.Context, user *domain.User) error
	GetUserByID(ctx context.Context, id uuid.UUID, viewerID *uuid.UUID) (*domain.User, error)
	GetUserByUsername(ctx context.Context, username string, viewerID *uuid.UUID) (*domain.User, error)
	GetUserByIdentifier(ctx context.Context, identifier string) (*domain.User, error)
	UpdateProfile(ctx context.Context, user *domain.User) error

	// Sessions
	CreateSession(ctx context.Context, session *domain.Session) error
	GetSessionByTokenHash(ctx context.Context, tokenHash string) (*domain.Session, error)
	DeleteSession(ctx context.Context, tokenHash string) error

	// Posts & Feed
	CreatePost(ctx context.Context, post *domain.Post) error
	GetPostByID(ctx context.Context, postID uuid.UUID, viewerID *uuid.UUID) (*domain.Post, error)
	DeletePost(ctx context.Context, postID, authorID uuid.UUID) error
	GetFeed(ctx context.Context, feedType string, viewerID *uuid.UUID, cursor string, limit int) ([]domain.Post, string, error)
	GetUserPosts(ctx context.Context, username string, tab string, viewerID *uuid.UUID, cursor string, limit int) ([]domain.Post, string, error)
	GetPostReplies(ctx context.Context, postID uuid.UUID, viewerID *uuid.UUID, cursor string, limit int) ([]domain.Post, string, error)

	// Engagements
	LikePost(ctx context.Context, postID, userID uuid.UUID) (bool, int, error)
	UnlikePost(ctx context.Context, postID, userID uuid.UUID) (bool, int, error)
	RepostPost(ctx context.Context, postID, userID uuid.UUID) (bool, int, error)
	UndoRepost(ctx context.Context, postID, userID uuid.UUID) (bool, int, error)
	BookmarkPost(ctx context.Context, postID, userID uuid.UUID) (bool, int, error)
	RemoveBookmark(ctx context.Context, postID, userID uuid.UUID) (bool, int, error)

	// Social Graph
	FollowUser(ctx context.Context, followerID, followingID uuid.UUID) (bool, int, error)
	UnfollowUser(ctx context.Context, followerID, followingID uuid.UUID) (bool, int, error)
	GetFollowers(ctx context.Context, userID, viewerID uuid.UUID, limit int, cursor string) ([]domain.User, string, error)
	GetFollowing(ctx context.Context, userID, viewerID uuid.UUID, limit int, cursor string) ([]domain.User, string, error)
	GetMutualConnections(ctx context.Context, userID, viewerID uuid.UUID, limit int) ([]domain.User, error)
	GetSocialRelationship(ctx context.Context, viewerID, targetID uuid.UUID) (*domain.SocialRelationship, error)
	BlockUser(ctx context.Context, blockerID, blockedID uuid.UUID) error
	UnblockUser(ctx context.Context, blockerID, blockedID uuid.UUID) error
	MuteUser(ctx context.Context, muterID, mutedID uuid.UUID) error
	UnmuteUser(ctx context.Context, muterID, mutedID uuid.UUID) error
	IsBlocked(ctx context.Context, userA, userB uuid.UUID) (bool, error)

	// Follow Requests
	CreateFollowRequest(ctx context.Context, req *domain.FollowRequest) error
	GetFollowRequests(ctx context.Context, targetID uuid.UUID) ([]domain.FollowRequest, error)
	RespondFollowRequest(ctx context.Context, requestID, targetID uuid.UUID, accept bool) error
	CancelFollowRequest(ctx context.Context, requesterID, targetID uuid.UUID) error

	// Direct Messaging
	GetConversations(ctx context.Context, userID uuid.UUID, limit int, cursor string) ([]domain.Conversation, error)
	GetOrCreateConversation(ctx context.Context, userA, userB uuid.UUID) (*domain.Conversation, error)
	GetConversationByID(ctx context.Context, convID, userID uuid.UUID) (*domain.Conversation, error)
	GetMessages(ctx context.Context, convID, userID uuid.UUID, limit int, beforeID *uuid.UUID) ([]domain.Message, error)
	CreateMessage(ctx context.Context, msg *domain.Message) error
	MarkConversationRead(ctx context.Context, convID, userID uuid.UUID, messageID *uuid.UUID) error

	// Search
	SearchPosts(ctx context.Context, query string, viewerID *uuid.UUID, cursor string, limit int) ([]domain.Post, string, error)
	SearchPostsAdvanced(ctx context.Context, params domain.SearchFilterParams, viewerID *uuid.UUID) ([]domain.Post, string, error)
	SearchUsers(ctx context.Context, query string, viewerID *uuid.UUID, limit int) ([]domain.User, error)
	SearchUsersAdvanced(ctx context.Context, query string, viewerID *uuid.UUID, cursor string, limit int) ([]domain.User, string, error)
	GetSearchSuggestions(ctx context.Context, query string, viewerID *uuid.UUID, limit int) ([]string, []domain.User, error)
	GetRecentSearches(ctx context.Context, userID uuid.UUID, limit int) ([]domain.RecentSearch, error)
	AddRecentSearch(ctx context.Context, userID uuid.UUID, query string) error
	DeleteRecentSearch(ctx context.Context, userID uuid.UUID, searchID uuid.UUID) error
	ClearRecentSearches(ctx context.Context, userID uuid.UUID) error

	// Notifications
	CreateNotification(ctx context.Context, n *domain.Notification) error
	GetNotifications(ctx context.Context, userID uuid.UUID, cursor string, limit int) ([]domain.Notification, string, error)
	GetNotificationsWithTab(ctx context.Context, userID uuid.UUID, tab string, cursor string, limit int) ([]domain.Notification, string, error)
	GetUnreadNotificationsCount(ctx context.Context, userID uuid.UUID) (int, error)
	MarkNotificationsRead(ctx context.Context, userID uuid.UUID) error
	MarkNotificationRead(ctx context.Context, notifID, userID uuid.UUID) error

	// Outbox Events
	CreateOutboxEvent(ctx context.Context, event *domain.OutboxEvent) error
	GetPendingOutboxEvents(ctx context.Context, batchSize int) ([]domain.OutboxEvent, error)
	MarkOutboxEventPublished(ctx context.Context, id uuid.UUID) error

	// Media
	CreateMediaObject(ctx context.Context, media *domain.MediaObject) error
	GetMediaObjectByID(ctx context.Context, id uuid.UUID) (*domain.MediaObject, error)

	// Subscriptions & Entitlements (Section 6A)
	GetSubscriptionPlans(ctx context.Context) ([]domain.SubscriptionPlan, error)
	GetSubscriptionPlanByID(ctx context.Context, id string) (*domain.SubscriptionPlan, error)
	GetActiveSubscription(ctx context.Context, userID uuid.UUID) (*domain.Subscription, error)
	CreateSubscription(ctx context.Context, sub *domain.Subscription) error
	CancelSubscription(ctx context.Context, subID, userID uuid.UUID) error
	GetUserEntitlements(ctx context.Context, userID uuid.UUID) (*domain.Entitlements, error)

	// Billing & Orders
	CreatePaymentOrder(ctx context.Context, order *domain.PaymentOrder) error
	GetPaymentOrderByID(ctx context.Context, orderID uuid.UUID) (*domain.PaymentOrder, error)
	UpdatePaymentOrderStatus(ctx context.Context, orderID uuid.UUID, status, invoiceURL string) error
	GetBillingHistory(ctx context.Context, userID uuid.UUID) ([]domain.PaymentOrder, []domain.Invoice, error)
	CreateInvoice(ctx context.Context, inv *domain.Invoice) error

	// Verification & Badges
	GetUserBadges(ctx context.Context, userID uuid.UUID) ([]domain.VerificationBadge, error)
	AssignVerificationBadge(ctx context.Context, badge *domain.VerificationBadge) error
	CreateVerificationApplication(ctx context.Context, app *domain.VerificationApplication) error
	GetVerificationApplication(ctx context.Context, userID uuid.UUID) (*domain.VerificationApplication, error)
	UpdateUserVerification(ctx context.Context, userID uuid.UUID, isVerified bool, verifiedType string) error

	// Recommendation & Ranking Candidates
	GetFollowingCandidates(ctx context.Context, viewerID uuid.UUID, limit int) ([]domain.Post, error)
	GetSocialEngagementCandidates(ctx context.Context, viewerID uuid.UUID, limit int) ([]domain.Post, error)
	GetTopicCandidates(ctx context.Context, viewerID *uuid.UUID, topics []string, limit int) ([]domain.Post, error)
	GetTrendingCandidates(ctx context.Context, limit int) ([]domain.Post, error)
	GetFreshCandidates(ctx context.Context, limit int) ([]domain.Post, error)
	GetExplorationCandidates(ctx context.Context, limit int) ([]domain.Post, error)
	GetFollowingPosts(ctx context.Context, viewerID uuid.UUID, cursor *recommendation.FeedCursor, limit int) ([]domain.Post, string, error)
	GetViewerInteractions(ctx context.Context, viewerID uuid.UUID) (
		followedAuthors map[uuid.UUID]bool,
		mutualAuthors map[uuid.UUID]bool,
		likedPosts map[uuid.UUID]bool,
		mutedAuthors map[uuid.UUID]bool,
		blockedAuthors map[uuid.UUID]bool,
		hiddenPosts map[uuid.UUID]bool,
		userTopics map[string]float64,
		err error,
	)
	GetActiveWeights(ctx context.Context) (domain.RankingWeights, string, error)

	// Feed Feedback & Personalization
	RecordFeedFeedback(ctx context.Context, feedback domain.FeedFeedback) error
	RecordFeedImpressions(ctx context.Context, impressions []domain.FeedImpression) error
	GetUserTopicPreferences(ctx context.Context, userID uuid.UUID) ([]domain.UserTopicPreference, error)
	UpdateUserTopicPreference(ctx context.Context, pref domain.UserTopicPreference) error
	ResetPersonalization(ctx context.Context, userID uuid.UUID) error
	GetTrendingTopics(ctx context.Context, limit int) ([]domain.TrendingTopic, error)

	// Health
	Ping(ctx context.Context) error
}

