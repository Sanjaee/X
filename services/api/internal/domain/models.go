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

	Media        []MediaObject `json:"media"`
	IsLiked      bool          `json:"is_liked"`
	IsReposted   bool          `json:"is_reposted"`
	IsBookmarked bool          `json:"is_bookmarked"`
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
	Type        string     `json:"type"` // 'like', 'reply', 'repost', 'mention', 'follow'
	PostID      *uuid.UUID `json:"post_id,omitempty"`
	Post        *Post      `json:"post,omitempty"`
	IsRead      bool       `json:"is_read"`
	CreatedAt   time.Time  `json:"created_at"`
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
}

// Message represents a single chat entry
type Message struct {
	ID             uuid.UUID    `json:"id"`
	ConversationID uuid.UUID    `json:"conversation_id"`
	SenderID       uuid.UUID    `json:"sender_id"`
	Sender         *User        `json:"sender,omitempty"`
	Content        string       `json:"content"`
	MediaID        *uuid.UUID   `json:"media_id,omitempty"`
	Media          *MediaObject `json:"media,omitempty"`
	CreatedAt      time.Time    `json:"created_at"`
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

