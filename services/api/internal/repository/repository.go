package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/x-reconstruction/api/internal/domain"
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

	// Search
	SearchPosts(ctx context.Context, query string, viewerID *uuid.UUID, cursor string, limit int) ([]domain.Post, string, error)
	SearchUsers(ctx context.Context, query string, viewerID *uuid.UUID, limit int) ([]domain.User, error)

	// Notifications
	CreateNotification(ctx context.Context, n *domain.Notification) error
	GetNotifications(ctx context.Context, userID uuid.UUID, cursor string, limit int) ([]domain.Notification, string, error)
	MarkNotificationsRead(ctx context.Context, userID uuid.UUID) error

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

	// Health
	Ping(ctx context.Context) error
}

