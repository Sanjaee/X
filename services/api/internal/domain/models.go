package domain

import (
	"time"

	"github.com/google/uuid"
)

// User represents the core identity
type User struct {
	ID           uuid.UUID  `json:"id"`
	Username     string     `json:"username"`
	Email        string     `json:"email,omitempty"`
	PasswordHash string     `json:"-"`
	Role         string     `json:"role"`
	IsVerified   bool       `json:"is_verified"`
	VerifiedType string     `json:"verified_type"`
	IsPrivate    bool       `json:"is_private"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty"`

	// Joined profile fields
	DisplayName    string `json:"display_name"`
	Bio            string `json:"bio"`
	AvatarURL      string `json:"avatar_url"`
	BannerURL      string `json:"banner_url"`
	Location       string `json:"location"`
	Website        string `json:"website"`
	FollowersCount int    `json:"followers_count"`
	FollowingCount int    `json:"following_count"`
	PostsCount     int    `json:"posts_count"`

	// Viewer relational context
	IsFollowing  bool `json:"is_following,omitempty"`
	IsFollowedBy bool `json:"is_followed_by,omitempty"`
}

// Session represents an active authenticated session
type Session struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	TokenHash string    `json:"-"`
	IPAddress string    `json:"ip_address"`
	UserAgent string    `json:"user_agent"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

// MediaObject represents an uploaded media asset
type MediaObject struct {
	ID              uuid.UUID `json:"id"`
	UploaderID      uuid.UUID `json:"uploader_id"`
	StorageKey      string    `json:"storage_key"`
	URL             string    `json:"url"`
	MediaType       string    `json:"media_type"` // 'image', 'video', 'gif'
	MimeType        string    `json:"mime_type"`
	FileSizeBytes   int64     `json:"file_size_bytes"`
	Width           *int      `json:"width,omitempty"`
	Height          *int      `json:"height,omitempty"`
	DurationSeconds *float64  `json:"duration_seconds,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
}

// Post represents a microblogging post/tweet/reply
type Post struct {
	ID           uuid.UUID  `json:"id"`
	AuthorID     uuid.UUID  `json:"author_id"`
	Author       *User      `json:"author,omitempty"`
	Content      string     `json:"content"`
	ReplyToID    *uuid.UUID `json:"reply_to_id,omitempty"`
	ParentPost   *Post      `json:"parent_post,omitempty"`
	QuotePostID  *uuid.UUID `json:"quote_post_id,omitempty"`
	QuotePost    *Post      `json:"quote_post,omitempty"`
	RepliesCount int        `json:"replies_count"`
	RepostsCount int        `json:"reposts_count"`
	LikesCount   int        `json:"likes_count"`
	ViewsCount   int64      `json:"views_count"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty"`

	Media            []MediaObject `json:"media"`
	Visibility       string        `json:"visibility"`
	ModerationStatus string        `json:"moderation_status"`
	IsLiked          bool          `json:"is_liked"`
	IsReposted       bool          `json:"is_reposted"`
	IsBookmarked     bool          `json:"is_bookmarked"`
}

// Follow represents a directional social connection
type Follow struct {
	FollowerID  uuid.UUID `json:"follower_id"`
	FollowingID uuid.UUID `json:"following_id"`
	CreatedAt   time.Time `json:"created_at"`
}

// Notification represents a user activity alert
type Notification struct {
	ID          uuid.UUID  `json:"id"`
	RecipientID uuid.UUID  `json:"recipient_id"`
	ActorID     uuid.UUID  `json:"actor_id"`
	Actor       *User      `json:"actor,omitempty"`
	Type        string     `json:"type"` // 'like', 'reply', 'repost', 'mention', 'follow', 'follow_request', 'dm'
	PostID      *uuid.UUID `json:"post_id,omitempty"`
	Post        *Post      `json:"post,omitempty"`
	EntityType  string     `json:"entity_type,omitempty"`
	EntityID    *uuid.UUID `json:"entity_id,omitempty"`
	DedupeKey   string     `json:"dedupe_key,omitempty"`
	IsRead      bool       `json:"is_read"`
	ReadAt      *time.Time `json:"read_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

// FollowRequest represents a pending connection request for private accounts
type FollowRequest struct {
	ID          uuid.UUID `json:"id"`
	RequesterID uuid.UUID `json:"requester_id"`
	Requester   *User     `json:"requester,omitempty"`
	TargetID    uuid.UUID `json:"target_id"`
	Target      *User     `json:"target,omitempty"`
	Status      string    `json:"status"` // 'pending', 'accepted', 'rejected', 'cancelled'
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// SocialRelationship summarizes bilateral state between two accounts
type SocialRelationship struct {
	TargetUserID           uuid.UUID `json:"target_user_id"`
	IsFollowing            bool      `json:"is_following"`
	IsFollowedBy           bool      `json:"is_followed_by"`
	IsMutual               bool      `json:"is_mutual"`
	IsBlocked              bool      `json:"is_blocked"`
	IsBlockedBy            bool      `json:"is_blocked_by"`
	IsMuted                bool      `json:"is_muted"`
	HasPendingFollowRequest bool     `json:"has_pending_follow_request"`
	IsPrivate              bool      `json:"is_private"`
}

// Conversation represents a DM chat channel
type Conversation struct {
	ID          uuid.UUID `json:"id"`
	IsGroup     bool      `json:"is_group"`
	Name        string    `json:"name,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	LastMessage *Message  `json:"last_message,omitempty"`
	Members     []User    `json:"members,omitempty"`
	OtherMember *User     `json:"other_member,omitempty"`
	UnreadCount int       `json:"unread_count"`
}

// Message represents a single chat entry
type Message struct {
	ID              uuid.UUID         `json:"id"`
	ConversationID  uuid.UUID         `json:"conversation_id"`
	SenderID        uuid.UUID         `json:"sender_id"`
	Sender          *User             `json:"sender,omitempty"`
	ClientMessageID string            `json:"client_message_id,omitempty"`
	Content         string            `json:"content"`
	MessageType     string            `json:"message_type"` // 'text', 'media', 'post_share'
	ReplyToID       *uuid.UUID        `json:"reply_to_id,omitempty"`
	ReplyTo         *Message          `json:"reply_to,omitempty"`
	MediaID         *uuid.UUID        `json:"media_id,omitempty"`
	Media           *MediaObject      `json:"media,omitempty"`
	Reactions       []MessageReaction `json:"reactions,omitempty"`
	EditedAt        *time.Time        `json:"edited_at,omitempty"`
	DeletedAt       *time.Time        `json:"deleted_at,omitempty"`
	CreatedAt       time.Time         `json:"created_at"`
}

// MessageReaction represents an emoji reaction on a message
type MessageReaction struct {
	ID        uuid.UUID `json:"id"`
	MessageID uuid.UUID `json:"message_id"`
	UserID    uuid.UUID `json:"user_id"`
	Emoji     string    `json:"emoji"`
	CreatedAt time.Time `json:"created_at"`
}

// NotificationPreferences holds user notification settings
type NotificationPreferences struct {
	UserID        uuid.UUID `json:"user_id"`
	EmailDMs      bool      `json:"email_dms"`
	EmailFollows  bool      `json:"email_follows"`
	PushDMs       bool      `json:"push_dms"`
	PushMentions  bool      `json:"push_mentions"`
	PushLikes     bool      `json:"push_likes"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// OutboxEvent represents an atomic event record for guaranteed asynchronous delivery
type OutboxEvent struct {
	ID            uuid.UUID              `json:"id"`
	EventType     string                 `json:"event_type"`
	AggregateType string                 `json:"aggregate_type"`
	AggregateID   uuid.UUID              `json:"aggregate_id"`
	Payload       map[string]interface{} `json:"payload"`
	Status        string                 `json:"status"` // 'pending', 'published', 'failed'
	RetryCount    int                    `json:"retry_count"`
	CreatedAt     time.Time              `json:"created_at"`
	PublishedAt   *time.Time             `json:"published_at,omitempty"`
}

// UserSettings represents configuration preferences
type UserSettings struct {
	UserID             uuid.UUID `json:"user_id"`
	Theme              string    `json:"theme"`
	AccentColor        string    `json:"accent_color"`
	AllowDMsFrom       string    `json:"allow_dms_from"`
	EmailNotifications bool      `json:"email_notifications"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// SubscriptionPlan defines a catalog offering (Section 6A.A)
type SubscriptionPlan struct {
	ID                  string    `json:"id"`
	Name                string    `json:"name"`
	Description         string    `json:"description"`
	MonthlyPriceCents   int       `json:"monthly_price_cents"`
	AnnualPriceCents    int       `json:"annual_price_cents"`
	Currency            string    `json:"currency"`
	MaxPostLength       int       `json:"max_post_length"`
	CanGetVerifiedBadge bool      `json:"can_get_verified_badge"`
	CanEditPosts        bool      `json:"can_edit_posts"`
	CanUndoPosts        bool      `json:"can_undo_posts"`
	MaxVideoMinutes     int       `json:"max_video_minutes"`
	MaxMediaCount       int       `json:"max_media_count"`
	ReducedAds          bool      `json:"reduced_ads"`
	IsActive            bool      `json:"is_active"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// Subscription represents a user's active billing contract
type Subscription struct {
	ID                 uuid.UUID         `json:"id"`
	UserID             uuid.UUID         `json:"user_id"`
	PlanID             string            `json:"plan_id"`
	Plan               *SubscriptionPlan `json:"plan,omitempty"`
	BillingCycle       string            `json:"billing_cycle"` // 'monthly' or 'annual'
	Status             string            `json:"status"`        // 'pending', 'active', 'canceled', 'expired', 'suspended'
	CurrentPeriodStart time.Time         `json:"current_period_start"`
	CurrentPeriodEnd   time.Time         `json:"current_period_end"`
	CancelAtPeriodEnd  bool              `json:"cancel_at_period_end"`
	CreatedAt          time.Time         `json:"created_at"`
	UpdatedAt          time.Time         `json:"updated_at"`
}

// Entitlements represents the resolved active benefits for a user (Section 6A.B)
type Entitlements struct {
	PlanID           string `json:"plan_id"`
	MaxPostLength    int    `json:"max_post_length"`
	HasVerifiedBadge bool   `json:"has_verified_badge"`
	BadgeType        string `json:"badge_type"`
	CanEditPosts     bool   `json:"can_edit_posts"`
	CanUndoPosts     bool   `json:"can_undo_posts"`
	MaxVideoMinutes  int    `json:"max_video_minutes"`
	MaxMediaCount    int    `json:"max_media_count"`
	ReducedAds       bool   `json:"reduced_ads"`
}

// PaymentOrder represents a checkout request
type PaymentOrder struct {
	ID                 uuid.UUID              `json:"id"`
	UserID             uuid.UUID              `json:"user_id"`
	PlanID             string                 `json:"plan_id"`
	BillingCycle       string                 `json:"billing_cycle"`
	AmountCents        int                    `json:"amount_cents"`
	Currency           string                 `json:"currency"`
	Provider           string                 `json:"provider"`
	ProviderOrderID    string                 `json:"provider_order_id,omitempty"`
	ProviderInvoiceURL string                 `json:"provider_invoice_url,omitempty"`
	Status             string                 `json:"status"`
	Metadata           map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt          time.Time              `json:"created_at"`
	UpdatedAt          time.Time              `json:"updated_at"`
}

// Invoice represents a billing receipt
type Invoice struct {
	ID          uuid.UUID `json:"id"`
	OrderID     uuid.UUID `json:"order_id"`
	UserID      uuid.UUID `json:"user_id"`
	AmountCents int       `json:"amount_cents"`
	Currency    string    `json:"currency"`
	Status      string    `json:"status"`
	PDFURL      string    `json:"pdf_url,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// VerificationBadge represents a displayed badge on an account (Section 6A.C & D)
type VerificationBadge struct {
	ID           uuid.UUID  `json:"id"`
	UserID       uuid.UUID  `json:"user_id"`
	BadgeType    string     `json:"badge_type"` // 'blue', 'gold', 'grey', 'staff'
	DisplayLabel string     `json:"display_label"`
	Description  string     `json:"description"`
	ColorToken   string     `json:"color_token"`
	Status       string     `json:"status"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// VerificationApplication represents an application or review record (Section 6A.E)
type VerificationApplication struct {
	ID          uuid.UUID  `json:"id"`
	UserID      uuid.UUID  `json:"user_id"`
	BadgeType   string     `json:"badge_type"`
	Category    string     `json:"category"`
	DocumentURL string     `json:"document_url,omitempty"`
	Status      string     `json:"status"`
	ReviewNotes string     `json:"review_notes,omitempty"`
	ReviewedBy  *uuid.UUID `json:"reviewed_by,omitempty"`
	ReviewedAt  *time.Time `json:"reviewed_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

// FeedImpression tracks post exposure on client viewport
type FeedImpression struct {
	ID          uuid.UUID  `json:"id"`
	UserID      *uuid.UUID `json:"user_id,omitempty"`
	PostID      uuid.UUID  `json:"post_id"`
	FeedType    string     `json:"feed_type"`
	DwellTimeMs int        `json:"dwell_time_ms"`
	CreatedAt   time.Time  `json:"created_at"`
}

// FeedFeedback represents user negative or positive feedback signals
type FeedFeedback struct {
	ID           uuid.UUID `json:"id"`
	UserID       uuid.UUID `json:"user_id"`
	PostID       uuid.UUID `json:"post_id"`
	FeedbackType string    `json:"feedback_type"` // 'hide', 'not_interested', 'show_less', 'report'
	Topic        string    `json:"topic,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

// UserTopicPreference stores explicitly followed or algorithmically inferred topics
type UserTopicPreference struct {
	ID            uuid.UUID `json:"id"`
	UserID        uuid.UUID `json:"user_id"`
	Topic         string    `json:"topic"`
	AffinityScore float64   `json:"affinity_score"`
	IsExplicit    bool      `json:"is_explicit"`
	IsMuted       bool      `json:"is_muted"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// PostFeatures stores precomputed content quality, topic vectors, and velocity
type PostFeatures struct {
	PostID        uuid.UUID `json:"post_id"`
	QualityScore  float64   `json:"quality_score"`
	PrimaryTopic  string    `json:"primary_topic"`
	VelocityScore float64   `json:"velocity_score"`
	HasMedia      bool      `json:"has_media"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// RankingWeights defines the configurable, versionable weights for the linear scoring model
type RankingWeights struct {
	RelationshipWeight     float64 `json:"relationship_weight"`
	FreshnessWeight        float64 `json:"freshness_weight"`
	EngagementWeight       float64 `json:"engagement_weight"`
	TopicWeight            float64 `json:"topic_weight"`
	QualityWeight          float64 `json:"quality_weight"`
	ExplorationWeight      float64 `json:"exploration_weight"`
	NegativePenaltyWeight  float64 `json:"negative_penalty_weight"`
	RepetitionWeight       float64 `json:"repetition_weight"`
	FreshnessHalfLifeHours float64 `json:"freshness_half_life_hours"`
}

// DefaultRankingWeights returns calibrated default production weights
func DefaultRankingWeights() RankingWeights {
	return RankingWeights{
		RelationshipWeight:     2.5,
		FreshnessWeight:        3.0,
		EngagementWeight:       2.0,
		TopicWeight:            1.8,
		QualityWeight:          1.2,
		ExplorationWeight:      1.0,
		NegativePenaltyWeight:  5.0,
		RepetitionWeight:       3.0,
		FreshnessHalfLifeHours: 12.0,
	}
}

// RankingExperiment represents an active or candidate versioned scoring configuration
type RankingExperiment struct {
	ID        uuid.UUID      `json:"id"`
	Name      string         `json:"name"`
	Version   string         `json:"version"`
	Weights   RankingWeights `json:"weights"`
	IsActive  bool           `json:"is_active"`
	CreatedAt time.Time      `json:"created_at"`
}

// Hashtag represents a tracked topic tag
type Hashtag struct {
	ID         uuid.UUID `json:"id"`
	Tag        string    `json:"tag"`
	PostsCount int       `json:"posts_count"`
	CreatedAt  time.Time `json:"created_at"`
}

// TrendingTopic represents an aggregated trending subject with velocity
type TrendingTopic struct {
	Category      string  `json:"category"`
	Topic         string  `json:"topic"`
	PostsCount    int     `json:"posts_count"`
	VelocityScore float64 `json:"velocity_score"`
}

// PageInfo represents cursor pagination metadata
type PageInfo struct {
	NextCursor  string `json:"next_cursor,omitempty"`
	HasNextPage bool   `json:"has_next_page"`
}

// SearchFilterParams holds parsed search criteria
type SearchFilterParams struct {
	Query        string     `json:"query"`
	RawQuery     string     `json:"raw_query"`
	Tab          string     `json:"tab"` // "top", "latest", "people", "media"
	FromUsername string     `json:"from_username,omitempty"`
	ToUsername   string     `json:"to_username,omitempty"`
	Mentions     string     `json:"mentions,omitempty"`
	HasMedia     bool       `json:"has_media,omitempty"`
	MediaType    string     `json:"media_type,omitempty"` // "image", "video"
	MinLikes     int        `json:"min_likes,omitempty"`
	SinceDate    *time.Time `json:"since_date,omitempty"`
	UntilDate    *time.Time `json:"until_date,omitempty"`
	Language     string     `json:"language,omitempty"`
	Cursor       string     `json:"cursor,omitempty"`
	Limit        int        `json:"limit,omitempty"`
}

// SearchSuggestionsResponse provides autocomplete results
type SearchSuggestionsResponse struct {
	Queries []string `json:"queries"`
	Users   []User   `json:"users"`
}

// RecentSearch represents a saved search query
type RecentSearch struct {
	ID         uuid.UUID `json:"id"`
	UserID     uuid.UUID `json:"user_id"`
	Query      string    `json:"query"`
	SearchedAt time.Time `json:"searched_at"`
}

// Role represents an authorization role in RBAC
type Role struct {
	ID          uuid.UUID `json:"id"`
	Key         string    `json:"key"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	IsSystem    bool      `json:"is_system"`
	IsActive    bool      `json:"is_active"`
	UserCount   int       `json:"user_count,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Permission represents a fine-grained authorization permission
type Permission struct {
	ID          uuid.UUID `json:"id"`
	Key         string    `json:"key"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

// RolePermission represents the mapping between a role and a permission
type RolePermission struct {
	RoleID       uuid.UUID `json:"role_id"`
	PermissionID uuid.UUID `json:"permission_id"`
	CreatedAt    time.Time `json:"created_at"`
}

// UserRole represents an assigned role to a user
type UserRole struct {
	ID         uuid.UUID  `json:"id"`
	UserID     uuid.UUID  `json:"user_id"`
	RoleID     uuid.UUID  `json:"role_id"`
	Role       *Role      `json:"role,omitempty"`
	AssignedBy *uuid.UUID `json:"assigned_by,omitempty"`
	AssignedAt time.Time  `json:"assigned_at"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	Reason     string     `json:"reason,omitempty"`
}

// MembershipPlan represents a tiered monetization / quota plan
type MembershipPlan struct {
	ID               string                  `json:"id"`
	Key              string                  `json:"key"`
	Name             string                  `json:"name"`
	Description      string                  `json:"description"`
	PriceAmountCents int                     `json:"price_amount_cents"`
	Currency         string                  `json:"currency"`
	BillingInterval  string                  `json:"billing_interval"`
	IsActive         bool                    `json:"is_active"`
	Entitlements     []MembershipEntitlement `json:"entitlements,omitempty"`
	CreatedAt        time.Time               `json:"created_at"`
	UpdatedAt        time.Time               `json:"updated_at"`
}

// MembershipEntitlement represents a quota or limit rule for a plan
type MembershipEntitlement struct {
	ID             uuid.UUID `json:"id"`
	PlanID         string    `json:"plan_id"`
	EntitlementKey string    `json:"entitlement_key"`
	LimitValue     int64     `json:"limit_value"`
	Unit           string    `json:"unit"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// UserMembership represents a user's active membership tier
type UserMembership struct {
	ID        uuid.UUID       `json:"id"`
	UserID    uuid.UUID       `json:"user_id"`
	PlanID    string          `json:"plan_id"`
	Plan      *MembershipPlan `json:"plan,omitempty"`
	Status    string          `json:"status"` // 'active', 'expired', 'canceled'
	StartsAt  time.Time       `json:"starts_at"`
	ExpiresAt *time.Time      `json:"expires_at,omitempty"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

// UserUsageCounter tracks period usage for posts, media, etc.
type UserUsageCounter struct {
	ID           uuid.UUID `json:"id"`
	UserID       uuid.UUID `json:"user_id"`
	PeriodKey    string    `json:"period_key"`
	PeriodStart  time.Time `json:"period_start"`
	PostsCreated int       `json:"posts_created"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// AuditLog records sensitive administrative and moderation operations
type AuditLog struct {
	ID               uuid.UUID              `json:"id"`
	ActorUserID      *uuid.UUID             `json:"actor_user_id,omitempty"`
	ActorUsername    string                 `json:"actor_username,omitempty"`
	Action           string                 `json:"action"`
	TargetType       string                 `json:"target_type"`
	TargetID         string                 `json:"target_id"`
	Reason           string                 `json:"reason"`
	MetadataRedacted map[string]interface{} `json:"metadata_redacted"`
	CreatedAt        time.Time              `json:"created_at"`
}

// AdminOverview represents system KPIs and operational status
type AdminOverview struct {
	TotalUsers            int64              `json:"total_users"`
	ActiveUsers24h        int64              `json:"active_users_24h"`
	NewUsersToday         int64              `json:"new_users_today"`
	TotalPosts            int64              `json:"total_posts"`
	PostsToday            int64              `json:"posts_today"`
	TotalMediaCount       int64              `json:"total_media_count"`
	EstimatedStorageBytes int64              `json:"estimated_storage_bytes"`
	FlaggedPostsCount     int64              `json:"flagged_posts_count"`
	SystemHealth          SystemHealth       `json:"system_health"`
	RecentAuditLogs       []AuditLog         `json:"recent_audit_logs"`
	RecentSecurityEvents  []AuditLog         `json:"recent_security_events"`
}

type SystemHealth struct {
	Database DatabaseHealth `json:"database"`
	Redis    RedisHealth    `json:"redis"`
	RabbitMQ ServiceStatus  `json:"rabbitmq"`
	Runtime  RuntimeMetrics `json:"runtime"`
}

type DatabaseHealth struct {
	Status           string  `json:"status"` // "healthy", "degraded", "down"
	PingMs           float64 `json:"ping_ms"`
	OpenConnections  int     `json:"open_connections"`
	InUseConnections int     `json:"in_use_connections"`
	IdleConnections  int     `json:"idle_connections"`
}

type RedisHealth struct {
	Status          string  `json:"status"`
	PingMs          float64 `json:"ping_ms"`
	UsedMemoryHuman string  `json:"used_memory_human"`
}

type ServiceStatus struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

type RuntimeMetrics struct {
	Goroutines      int    `json:"goroutines"`
	AllocBytes      uint64 `json:"alloc_bytes"`
	TotalAllocBytes uint64 `json:"total_alloc_bytes"`
	SysBytes        uint64 `json:"sys_bytes"`
	NumGC           uint32 `json:"num_gc"`
	UptimeSeconds   int64  `json:"uptime_seconds"`
}

// AdminPostFilter specifies parameters for filtering posts in the admin panel
type AdminPostFilter struct {
	Query            string `json:"query"`
	AuthorID         string `json:"author_id"`
	ModerationStatus string `json:"moderation_status"`
	Visibility       string `json:"visibility"`
	Limit            int    `json:"limit"`
	Offset           int    `json:"offset"`
}




