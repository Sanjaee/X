package repository

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/x-reconstruction/api/internal/domain"
	"github.com/x-reconstruction/api/internal/recommendation"
)

var (
	ErrNotFound       = errors.New("resource not found")
	ErrAlreadyExists  = errors.New("resource already exists")
	ErrUnauthorized   = errors.New("unauthorized")
	ErrInvalidRequest = errors.New("invalid request")
)

// MemoryRepository is an in-memory thread-safe implementation of Repository
type MemoryRepository struct {
	mu            sync.RWMutex
	users         map[uuid.UUID]*domain.User
	userByUname   map[string]uuid.UUID
	userByEmail   map[string]uuid.UUID
	sessions      map[string]*domain.Session
	posts         map[uuid.UUID]*domain.Post
	media         map[uuid.UUID]*domain.MediaObject
	likes         map[string]time.Time // "postID:userID"
	reposts       map[string]time.Time // "postID:userID"
	bookmarks     map[string]time.Time // "postID:userID"
	follows       map[string]time.Time // "followerID:followingID"
	blocks        map[string]time.Time // "blockerID:blockedID"
	mutes         map[string]time.Time // "muterID:mutedID"
	followReqs    map[uuid.UUID]*domain.FollowRequest
	conversations map[uuid.UUID]*domain.Conversation
	convMembers   map[uuid.UUID][]uuid.UUID
	messages      map[uuid.UUID]*domain.Message
	outboxEvents  map[uuid.UUID]*domain.OutboxEvent
	notifications map[uuid.UUID]*domain.Notification

	// Subscriptions & Verification (Section 6A)
	plans        map[string]*domain.SubscriptionPlan
	subs         map[uuid.UUID]*domain.Subscription
	orders       map[uuid.UUID]*domain.PaymentOrder
	invoices     map[uuid.UUID]*domain.Invoice
	badges       map[uuid.UUID]*domain.VerificationBadge
	applications map[uuid.UUID]*domain.VerificationApplication

	// Recommendation & Feedback
	feedFeedback         map[string]*domain.FeedFeedback
	feedImpressions      []*domain.FeedImpression
	userTopicPreferences    map[string]*domain.UserTopicPreference
	activeWeights           domain.RankingWeights
	recentSearches          map[uuid.UUID][]domain.RecentSearch
	dismissedRecommendations map[string]time.Time // "viewerID:targetID" -> dismissedUntil

	// Dynamic Roles, Memberships, Quotas, Audit
	rbacRoles       map[uuid.UUID]*domain.Role
	rbacPermissions map[uuid.UUID]*domain.Permission
	rbacRolePerms   map[uuid.UUID][]uuid.UUID // roleID -> []permissionID
	rbacUserRoles   map[uuid.UUID][]uuid.UUID // userID -> []roleID
	memPlans        map[string]*domain.MembershipPlan
	memEntitlements map[string]map[string]int64 // planID -> (key -> val)
	userMemberships map[uuid.UUID]*domain.UserMembership
	usageCounters   map[string]int // "userID:periodKey" -> count
	auditLogs       []domain.AuditLog
}

func NewMemoryRepository() *MemoryRepository {
	repo := &MemoryRepository{
		users:                make(map[uuid.UUID]*domain.User),
		userByUname:          make(map[string]uuid.UUID),
		userByEmail:          make(map[string]uuid.UUID),
		sessions:             make(map[string]*domain.Session),
		posts:                make(map[uuid.UUID]*domain.Post),
		media:                make(map[uuid.UUID]*domain.MediaObject),
		likes:                make(map[string]time.Time),
		reposts:              make(map[string]time.Time),
		bookmarks:            make(map[string]time.Time),
		follows:              make(map[string]time.Time),
		blocks:               make(map[string]time.Time),
		mutes:                make(map[string]time.Time),
		followReqs:           make(map[uuid.UUID]*domain.FollowRequest),
		conversations:        make(map[uuid.UUID]*domain.Conversation),
		convMembers:          make(map[uuid.UUID][]uuid.UUID),
		messages:             make(map[uuid.UUID]*domain.Message),
		outboxEvents:         make(map[uuid.UUID]*domain.OutboxEvent),
		notifications:        make(map[uuid.UUID]*domain.Notification),
		plans:                make(map[string]*domain.SubscriptionPlan),
		subs:                 make(map[uuid.UUID]*domain.Subscription),
		orders:               make(map[uuid.UUID]*domain.PaymentOrder),
		invoices:             make(map[uuid.UUID]*domain.Invoice),
		badges:               make(map[uuid.UUID]*domain.VerificationBadge),
		applications:         make(map[uuid.UUID]*domain.VerificationApplication),
		feedFeedback:         make(map[string]*domain.FeedFeedback),
		feedImpressions:      make([]*domain.FeedImpression, 0),
		userTopicPreferences: make(map[string]*domain.UserTopicPreference),
		activeWeights:        domain.DefaultRankingWeights(),
		recentSearches:       make(map[uuid.UUID][]domain.RecentSearch),
		dismissedRecommendations: make(map[string]time.Time),

		rbacRoles:       make(map[uuid.UUID]*domain.Role),
		rbacPermissions: make(map[uuid.UUID]*domain.Permission),
		rbacRolePerms:   make(map[uuid.UUID][]uuid.UUID),
		rbacUserRoles:   make(map[uuid.UUID][]uuid.UUID),
		memPlans:        make(map[string]*domain.MembershipPlan),
		memEntitlements: make(map[string]map[string]int64),
		userMemberships: make(map[uuid.UUID]*domain.UserMembership),
		usageCounters:   make(map[string]int),
		auditLogs:       make([]domain.AuditLog, 0),
	}
	repo.seedInitialData()
	return repo
}


func (r *MemoryRepository) Ping(ctx context.Context) error {
	return nil
}

func (r *MemoryRepository) seedInitialData() {
	r.plans["free"] = &domain.SubscriptionPlan{
		ID: "free", Name: "Free", Description: "Standard microblogging experience for everyone.",
		MonthlyPriceCents: 0, AnnualPriceCents: 0, Currency: "USD", MaxPostLength: 280,
		CanGetVerifiedBadge: false, CanEditPosts: false, CanUndoPosts: false, MaxVideoMinutes: 2, MaxMediaCount: 4, ReducedAds: false, IsActive: true,
	}
	r.plans["basic"] = &domain.SubscriptionPlan{
		ID: "basic", Name: "Basic", Description: "Essential features: edit posts, bookmark folders, and up to 1,000 characters.",
		MonthlyPriceCents: 300, AnnualPriceCents: 3000, Currency: "USD", MaxPostLength: 1000,
		CanGetVerifiedBadge: false, CanEditPosts: true, CanUndoPosts: true, MaxVideoMinutes: 10, MaxMediaCount: 4, ReducedAds: false, IsActive: true,
	}
	r.plans["premium"] = &domain.SubscriptionPlan{
		ID: "premium", Name: "Premium", Description: "The signature X experience: Blue verification badge, 4,000 characters, 60min video, 50% fewer ads.",
		MonthlyPriceCents: 800, AnnualPriceCents: 8400, Currency: "USD", MaxPostLength: 4000,
		CanGetVerifiedBadge: true, CanEditPosts: true, CanUndoPosts: true, MaxVideoMinutes: 60, MaxMediaCount: 4, ReducedAds: true, IsActive: true,
	}
	r.plans["premium_plus"] = &domain.SubscriptionPlan{
		ID: "premium_plus", Name: "Premium+", Description: "Everything in Premium with no ads in For You, largest reply boost, 10,000 characters, and exclusive creator tools.",
		MonthlyPriceCents: 1600, AnnualPriceCents: 16800, Currency: "USD", MaxPostLength: 10000,
		CanGetVerifiedBadge: true, CanEditPosts: true, CanUndoPosts: true, MaxVideoMinutes: 180, MaxMediaCount: 4, ReducedAds: true, IsActive: true,
	}

	// Seed RBAC Roles
	rRoles := []domain.Role{
		{ID: uuid.MustParse("00000000-0000-0000-0000-000000000001"), Key: "member", Name: "Member", Description: "Default standard role", IsSystem: true, IsActive: true},
		{ID: uuid.MustParse("00000000-0000-0000-0000-000000000002"), Key: "vip", Name: "VIP", Description: "VIP membership role", IsSystem: true, IsActive: true},
		{ID: uuid.MustParse("00000000-0000-0000-0000-000000000003"), Key: "mvp", Name: "MVP", Description: "MVP membership role", IsSystem: true, IsActive: true},
		{ID: uuid.MustParse("00000000-0000-0000-0000-000000000004"), Key: "mod", Name: "Moderator", Description: "Moderation role", IsSystem: true, IsActive: true},
		{ID: uuid.MustParse("00000000-0000-0000-0000-000000000005"), Key: "god", Name: "God Admin", Description: "Administrative role", IsSystem: true, IsActive: true},
		{ID: uuid.MustParse("00000000-0000-0000-0000-000000000006"), Key: "owner", Name: "Owner", Description: "Platform Owner", IsSystem: true, IsActive: true},
	}
	for _, role := range rRoles {
		copied := role
		r.rbacRoles[role.ID] = &copied
	}

	// Seed Permissions
	permKeys := []string{
		"post.create", "post.read", "post.update.own", "post.delete.own", "post.update.any", "post.delete.any", "post.pin.own", "post.pin.any",
		"media.upload.image", "media.upload.video", "media.delete.own", "media.delete.any",
		"profile.update.own", "profile.update.any", "profile.manage.colors", "profile.view.analytics",
		"moderation.post.hide", "moderation.post.restore", "moderation.user.warn", "moderation.user.suspend", "moderation.report.manage",
		"membership.view", "membership.manage.own", "membership.assign", "membership.revoke", "membership.plan.manage",
		"role.read", "role.create", "role.update", "role.delete", "permission.manage", "user.read", "user.update", "user.suspend", "audit.read", "system.settings.manage",
	}
	permMap := make(map[string]uuid.UUID)
	for i, k := range permKeys {
		pID := uuid.NewMD5(uuid.NameSpaceDNS, []byte(k))
		permMap[k] = pID
		r.rbacPermissions[pID] = &domain.Permission{
			ID:          pID,
			Key:         k,
			Name:        k,
			Description: fmt.Sprintf("Permission %s (%d)", k, i),
		}
	}

	// Map role permissions
	memberPerms := []string{"post.create", "post.read", "post.update.own", "post.delete.own", "post.pin.own", "media.upload.image", "media.delete.own", "profile.update.own", "membership.view", "membership.manage.own"}
	for _, k := range memberPerms {
		r.rbacRolePerms[rRoles[0].ID] = append(r.rbacRolePerms[rRoles[0].ID], permMap[k])
	}
	for _, k := range permKeys {
		// Owner has all permissions
		r.rbacRolePerms[rRoles[5].ID] = append(r.rbacRolePerms[rRoles[5].ID], permMap[k])
	}

	// Seed Membership Plans & Entitlements
	r.memPlans["free"] = &domain.MembershipPlan{ID: "free", Key: "free", Name: "Member", PriceAmountCents: 0, IsActive: true}
	r.memPlans["vip"] = &domain.MembershipPlan{ID: "vip", Key: "vip", Name: "VIP Member", PriceAmountCents: 100, IsActive: true}
	r.memPlans["mvp"] = &domain.MembershipPlan{ID: "mvp", Key: "mvp", Name: "MVP", PriceAmountCents: 100, IsActive: true}

	r.memEntitlements["free"] = map[string]int64{"posts_per_day": 2, "max_images_per_post": 2, "max_videos_per_post": 0, "max_post_length": 280, "max_file_size_mb": 5}
	r.memEntitlements["vip"] = map[string]int64{"posts_per_day": 10, "max_images_per_post": 4, "max_videos_per_post": 1, "max_post_length": 4000, "max_file_size_mb": 50, "can_view_analytics": 1}
	r.memEntitlements["mvp"] = map[string]int64{"posts_per_day": 30, "max_images_per_post": 4, "max_videos_per_post": 1, "max_post_length": 10000, "max_file_size_mb": 200, "can_view_analytics": 1}

	// Create sample official NASA account
	nasaID := uuid.MustParse("11111111-1111-1111-1111-111111111111")

	nasa := &domain.User{
		ID:             nasaID,
		Username:       "NASA",
		Email:          "nasa@example.com",
		DisplayName:    "NASA",
		Bio:            "Exploring the secrets of the universe for the benefit of all.",
		AvatarURL:      "https://images.unsplash.com/photo-1614728894747-a83421e2b9c9?w=150&auto=format&fit=crop&q=80",
		BannerURL:      "https://images.unsplash.com/photo-1451187580459-43490279c0fa?w=1200&auto=format&fit=crop&q=80",
		Location:       "Washington, DC",
		Website:        "https://nasa.gov",
		IsVerified:     true,
		VerifiedType:   "government",
		FollowersCount: 92300000,
		FollowingCount: 115,
		PostsCount:     42,
		CreatedAt:      time.Now().Add(-365 * 24 * time.Hour),
		UpdatedAt:      time.Now(),
	}
	r.users[nasaID] = nasa
	r.userByUname[strings.ToLower("NASA")] = nasaID
	r.userByEmail[strings.ToLower("nasa@example.com")] = nasaID

	// Seed NASA posts
	p1ID := uuid.MustParse("22222222-2222-2222-2222-222222222221")
	r.posts[p1ID] = &domain.Post{
		ID:           p1ID,
		AuthorID:     nasaID,
		Author:       nasa,
		Content:      "Join us on Oct. 23 for an update on our plans to return American astronauts to the Moon by 2028. Artemis III updates incoming! 🚀🌕",
		LikesCount:   15400,
		RepostsCount: 3110,
		RepliesCount: 154,
		ViewsCount:   390000,
		CreatedAt:    time.Now().Add(-5 * time.Minute),
		UpdatedAt:    time.Now().Add(-5 * time.Minute),
	}

	p2ID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	r.posts[p2ID] = &domain.Post{
		ID:           p2ID,
		AuthorID:     nasaID,
		Author:       nasa,
		Content:      "While observing the cosmos, NASA Webb peered through interstellar dust clouds to reveal hundreds of forming baby stars! ✨",
		LikesCount:   20500,
		RepostsCount: 3600,
		RepliesCount: 772,
		ViewsCount:   2000000,
		CreatedAt:    time.Now().Add(-18 * time.Hour),
		UpdatedAt:    time.Now().Add(-18 * time.Hour),
	}
}

func (r *MemoryRepository) CreateUser(ctx context.Context, u *domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	lowerU := strings.ToLower(u.Username)
	lowerE := strings.ToLower(u.Email)

	if _, exists := r.userByUname[lowerU]; exists {
		return fmt.Errorf("%w: username already taken", ErrAlreadyExists)
	}
	if _, exists := r.userByEmail[lowerE]; exists {
		return fmt.Errorf("%w: email already taken", ErrAlreadyExists)
	}

	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	now := time.Now()
	u.CreatedAt = now
	u.UpdatedAt = now

	if u.Role == "" || u.Role == "user" {
		u.Role = "member"
	}
	if lowerE == "afrizaahmad18@gmail.com" {
		u.Role = "owner"
		u.IsVerified = true
	}

	r.users[u.ID] = u
	r.userByUname[lowerU] = u.ID
	r.userByEmail[lowerE] = u.ID

	// Assign role
	memberRoleID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	if u.Role == "owner" {
		memberRoleID = uuid.MustParse("00000000-0000-0000-0000-000000000006")
	}
	r.rbacUserRoles[u.ID] = []uuid.UUID{memberRoleID}

	// Assign membership
	planID := "free"
	if u.Role == "owner" {
		planID = "mvp"
	}
	r.userMemberships[u.ID] = &domain.UserMembership{
		UserID:   u.ID,
		PlanID:   planID,
		Status:   "active",
		StartsAt: now,
	}

	return nil
}

func (r *MemoryRepository) GetUserByID(ctx context.Context, id uuid.UUID, viewerID *uuid.UUID) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	u, exists := r.users[id]
	if !exists || u.DeletedAt != nil {
		return nil, ErrNotFound
	}

	cloned := *u
	if viewerID != nil {
		cloned.IsFollowing = r.isFollowingInternal(*viewerID, u.ID)
		cloned.IsFollowedBy = r.isFollowingInternal(u.ID, *viewerID)
	}
	return &cloned, nil
}

func (r *MemoryRepository) GetUserByUsername(ctx context.Context, username string, viewerID *uuid.UUID) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	id, exists := r.userByUname[strings.ToLower(username)]
	if !exists {
		return nil, ErrNotFound
	}
	u := r.users[id]
	if u.DeletedAt != nil {
		return nil, ErrNotFound
	}

	cloned := *u
	if viewerID != nil {
		cloned.IsFollowing = r.isFollowingInternal(*viewerID, u.ID)
		cloned.IsFollowedBy = r.isFollowingInternal(u.ID, *viewerID)
	}
	return &cloned, nil
}

func (r *MemoryRepository) GetUserByIdentifier(ctx context.Context, identifier string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	low := strings.ToLower(identifier)
	if id, exists := r.userByUname[low]; exists {
		return r.users[id], nil
	}
	if id, exists := r.userByEmail[low]; exists {
		return r.users[id], nil
	}
	return nil, ErrNotFound
}

func (r *MemoryRepository) UpdateProfile(ctx context.Context, u *domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, exists := r.users[u.ID]
	if !exists {
		return ErrNotFound
	}
	existing.DisplayName = u.DisplayName
	existing.Bio = u.Bio
	existing.AvatarURL = u.AvatarURL
	existing.BannerURL = u.BannerURL
	existing.Location = u.Location
	existing.Website = u.Website
	existing.UpdatedAt = time.Now()
	return nil
}

func (r *MemoryRepository) CreateSession(ctx context.Context, s *domain.Session) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	s.CreatedAt = time.Now()
	r.sessions[s.TokenHash] = s
	return nil
}

func (r *MemoryRepository) GetSessionByTokenHash(ctx context.Context, hash string) (*domain.Session, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	s, exists := r.sessions[hash]
	if !exists {
		return nil, ErrNotFound
	}
	if time.Now().After(s.ExpiresAt) {
		return nil, ErrUnauthorized
	}
	return s, nil
}

func (r *MemoryRepository) DeleteSession(ctx context.Context, hash string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.sessions, hash)
	return nil
}

func (r *MemoryRepository) CreatePost(ctx context.Context, p *domain.Post) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	now := time.Now()
	p.CreatedAt = now
	p.UpdatedAt = now

	author := r.users[p.AuthorID]
	if author != nil {
		author.PostsCount++
		p.Author = author
	}

	if p.ReplyToID != nil {
		if parent, exists := r.posts[*p.ReplyToID]; exists {
			parent.RepliesCount++
			p.ParentPost = parent
		}
	}

	r.posts[p.ID] = p
	return nil
}

func (r *MemoryRepository) GetPostByID(ctx context.Context, postID uuid.UUID, viewerID *uuid.UUID) (*domain.Post, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	p, exists := r.posts[postID]
	if !exists || p.DeletedAt != nil {
		return nil, ErrNotFound
	}
	cloned := *p
	if viewerID != nil {
		cloned.IsLiked = r.isLikedInternal(p.ID, *viewerID)
		cloned.IsReposted = r.isRepostedInternal(p.ID, *viewerID)
		cloned.IsBookmarked = r.isBookmarkedInternal(p.ID, *viewerID)
	}
	return &cloned, nil
}

func (r *MemoryRepository) DeletePost(ctx context.Context, postID, authorID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	p, exists := r.posts[postID]
	if !exists {
		return ErrNotFound
	}
	if p.AuthorID != authorID {
		return ErrUnauthorized
	}
	now := time.Now()
	p.DeletedAt = &now
	return nil
}

func (r *MemoryRepository) GetFeed(ctx context.Context, feedType string, viewerID *uuid.UUID, cursor string, limit int) ([]domain.Post, string, error) {
	engine := recommendation.NewEngine(r)
	res, err := engine.GenerateFeed(ctx, recommendation.FeedOptions{
		ViewerID: viewerID,
		FeedType: feedType,
		Cursor:   cursor,
		Limit:    limit,
	})
	if err != nil {
		return nil, "", err
	}
	return res.Posts, res.NextCursor, nil
}

func (r *MemoryRepository) GetUserPosts(ctx context.Context, username, tab string, viewerID *uuid.UUID, cursor string, limit int) ([]domain.Post, string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	uID, exists := r.userByUname[strings.ToLower(username)]
	if !exists {
		return []domain.Post{}, "", nil
	}
	targetUser := r.users[uID]

	normTab := strings.ToLower(strings.TrimSpace(tab))
	result := make([]domain.Post, 0)

	switch normTab {
	case "replies", "reply":
		for _, p := range r.posts {
			if p.AuthorID == uID && p.ReplyToID != nil && p.DeletedAt == nil {
				cloned := *p
				if viewerID != nil {
					cloned.IsLiked = r.isLikedInternal(p.ID, *viewerID)
					cloned.IsReposted = r.isRepostedInternal(p.ID, *viewerID)
					cloned.IsBookmarked = r.isBookmarkedInternal(p.ID, *viewerID)
				}
				result = append(result, cloned)
			}
		}

	case "reposts", "repost":
		for key, repTime := range r.reposts {
			parts := strings.Split(key, ":")
			if len(parts) == 2 {
				postID, _ := uuid.Parse(parts[0])
				reposterID, _ := uuid.Parse(parts[1])
				if reposterID == uID {
					if orig, ok := r.posts[postID]; ok && orig.DeletedAt == nil {
						cloned := *orig
						repCopy := *targetUser
						cloned.Reposter = &repCopy
						t := repTime
						cloned.RepostedAt = &t
						if viewerID != nil {
							cloned.IsLiked = r.isLikedInternal(orig.ID, *viewerID)
							cloned.IsReposted = r.isRepostedInternal(orig.ID, *viewerID)
							cloned.IsBookmarked = r.isBookmarkedInternal(orig.ID, *viewerID)
						}
						result = append(result, cloned)
					}
				}
			}
		}

	case "media":
		for _, p := range r.posts {
			if p.AuthorID == uID && p.DeletedAt == nil && len(p.Media) > 0 {
				cloned := *p
				if viewerID != nil {
					cloned.IsLiked = r.isLikedInternal(p.ID, *viewerID)
					cloned.IsReposted = r.isRepostedInternal(p.ID, *viewerID)
					cloned.IsBookmarked = r.isBookmarkedInternal(p.ID, *viewerID)
				}
				result = append(result, cloned)
			}
		}

	default:
		for _, p := range r.posts {
			if p.AuthorID == uID && p.ReplyToID == nil && p.DeletedAt == nil {
				cloned := *p
				if viewerID != nil {
					cloned.IsLiked = r.isLikedInternal(p.ID, *viewerID)
					cloned.IsReposted = r.isRepostedInternal(p.ID, *viewerID)
					cloned.IsBookmarked = r.isBookmarkedInternal(p.ID, *viewerID)
				}
				result = append(result, cloned)
			}
		}
	}

	return result, "", nil
}

func (r *MemoryRepository) GetPostReplies(ctx context.Context, postID uuid.UUID, viewerID *uuid.UUID, cursor string, limit int, sortBy string) ([]domain.Post, string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []domain.Post
	for _, p := range r.posts {
		if p.ReplyToID != nil && *p.ReplyToID == postID && p.DeletedAt == nil {
			cloned := *p
			if viewerID != nil {
				cloned.IsLiked = r.isLikedInternal(p.ID, *viewerID)
				cloned.IsReposted = r.isRepostedInternal(p.ID, *viewerID)
				cloned.IsBookmarked = r.isBookmarkedInternal(p.ID, *viewerID)
			}
			result = append(result, cloned)
		}
	}

	switch strings.ToLower(sortBy) {
	case "recent", "terkini":
		sort.Slice(result, func(i, j int) bool {
			return result[i].CreatedAt.After(result[j].CreatedAt)
		})
	case "likes", "suka":
		sort.Slice(result, func(i, j int) bool {
			return result[i].LikesCount > result[j].LikesCount
		})
	default:
		// Default relevant or chronological
		sort.Slice(result, func(i, j int) bool {
			return result[i].CreatedAt.Before(result[j].CreatedAt)
		})
	}

	return result, "", nil
}

func (r *MemoryRepository) GetPostQuotes(ctx context.Context, postID uuid.UUID, viewerID *uuid.UUID, cursor string, limit int) ([]domain.Post, string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []domain.Post
	for _, p := range r.posts {
		if p.QuotePostID != nil && *p.QuotePostID == postID && p.DeletedAt == nil {
			cloned := *p
			if viewerID != nil {
				cloned.IsLiked = r.isLikedInternal(p.ID, *viewerID)
				cloned.IsReposted = r.isRepostedInternal(p.ID, *viewerID)
				cloned.IsBookmarked = r.isBookmarkedInternal(p.ID, *viewerID)
			}
			result = append(result, cloned)
		}
	}
	return result, "", nil
}

func (r *MemoryRepository) LikePost(ctx context.Context, postID, userID uuid.UUID) (bool, int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	p, exists := r.posts[postID]
	if !exists {
		return false, 0, ErrNotFound
	}
	key := fmt.Sprintf("%s:%s", postID, userID)
	if _, liked := r.likes[key]; !liked {
		r.likes[key] = time.Now()
		p.LikesCount++
	}
	return true, p.LikesCount, nil
}

func (r *MemoryRepository) UnlikePost(ctx context.Context, postID, userID uuid.UUID) (bool, int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	p, exists := r.posts[postID]
	if !exists {
		return false, 0, ErrNotFound
	}
	key := fmt.Sprintf("%s:%s", postID, userID)
	if _, liked := r.likes[key]; liked {
		delete(r.likes, key)
		if p.LikesCount > 0 {
			p.LikesCount--
		}
	}
	return false, p.LikesCount, nil
}

func (r *MemoryRepository) RepostPost(ctx context.Context, postID, userID uuid.UUID) (bool, int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	p, exists := r.posts[postID]
	if !exists {
		return false, 0, ErrNotFound
	}
	key := fmt.Sprintf("%s:%s", postID, userID)
	if _, reposted := r.reposts[key]; !reposted {
		r.reposts[key] = time.Now()
		p.RepostsCount++
	}
	return true, p.RepostsCount, nil
}

func (r *MemoryRepository) UndoRepost(ctx context.Context, postID, userID uuid.UUID) (bool, int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	p, exists := r.posts[postID]
	if !exists {
		return false, 0, ErrNotFound
	}
	key := fmt.Sprintf("%s:%s", postID, userID)
	if _, reposted := r.reposts[key]; reposted {
		delete(r.reposts, key)
		if p.RepostsCount > 0 {
			p.RepostsCount--
		}
	}
	return false, p.RepostsCount, nil
}

func (r *MemoryRepository) BookmarkPost(ctx context.Context, postID, userID uuid.UUID) (bool, int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := fmt.Sprintf("%s:%s", postID, userID)
	r.bookmarks[key] = time.Now()
	return true, len(r.bookmarks), nil
}

func (r *MemoryRepository) RemoveBookmark(ctx context.Context, postID, userID uuid.UUID) (bool, int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := fmt.Sprintf("%s:%s", postID, userID)
	delete(r.bookmarks, key)
	return false, len(r.bookmarks), nil
}

func (r *MemoryRepository) FollowUser(ctx context.Context, followerID, followingID uuid.UUID) (bool, int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if followerID == followingID {
		return false, 0, errors.New("cannot follow yourself")
	}
	target, exists := r.users[followingID]
	if !exists {
		return false, 0, ErrNotFound
	}
	follower := r.users[followerID]

	key := fmt.Sprintf("%s:%s", followerID, followingID)
	if _, following := r.follows[key]; !following {
		r.follows[key] = time.Now()
		target.FollowersCount++
		if follower != nil {
			follower.FollowingCount++
		}
	}
	return true, target.FollowersCount, nil
}

func (r *MemoryRepository) UnfollowUser(ctx context.Context, followerID, followingID uuid.UUID) (bool, int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	target, exists := r.users[followingID]
	if !exists {
		return false, 0, ErrNotFound
	}
	follower := r.users[followerID]

	key := fmt.Sprintf("%s:%s", followerID, followingID)
	if _, following := r.follows[key]; following {
		delete(r.follows, key)
		if target.FollowersCount > 0 {
			target.FollowersCount--
		}
		if follower != nil && follower.FollowingCount > 0 {
			follower.FollowingCount--
		}
	}
	return false, target.FollowersCount, nil
}

func (r *MemoryRepository) SearchPosts(ctx context.Context, query string, viewerID *uuid.UUID, cursor string, limit int) ([]domain.Post, string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	low := strings.ToLower(query)
	var res []domain.Post
	for _, p := range r.posts {
		if strings.Contains(strings.ToLower(p.Content), low) && p.DeletedAt == nil {
			cloned := *p
			if viewerID != nil {
				cloned.IsLiked = r.isLikedInternal(p.ID, *viewerID)
				cloned.IsReposted = r.isRepostedInternal(p.ID, *viewerID)
			}
			res = append(res, cloned)
		}
	}
	return res, "", nil
}

func (r *MemoryRepository) SearchUsers(ctx context.Context, query string, viewerID *uuid.UUID, limit int) ([]domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	low := strings.ToLower(query)
	var res []domain.User
	for _, u := range r.users {
		if (strings.Contains(strings.ToLower(u.Username), low) || strings.Contains(strings.ToLower(u.DisplayName), low)) && u.DeletedAt == nil {
			cloned := *u
			if viewerID != nil {
				cloned.IsFollowing = r.isFollowingInternal(*viewerID, u.ID)
			}
			res = append(res, cloned)
		}
	}
	return res, nil
}

func (r *MemoryRepository) SearchPostsAdvanced(ctx context.Context, params domain.SearchFilterParams, viewerID *uuid.UUID) ([]domain.Post, string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	low := strings.ToLower(params.Query)
	var res []domain.Post

	for _, p := range r.posts {
		if p.DeletedAt != nil {
			continue
		}
		// Text match
		if low != "" && !strings.Contains(strings.ToLower(p.Content), low) {
			continue
		}
		// Author filter
		if params.FromUsername != "" {
			if p.Author == nil || !strings.EqualFold(p.Author.Username, params.FromUsername) {
				continue
			}
		}
		// Media filter
		if (params.Tab == "media" || params.HasMedia) && len(p.Media) == 0 {
			continue
		}
		// Min likes filter
		if params.MinLikes > 0 && p.LikesCount < params.MinLikes {
			continue
		}
		// Date filters
		if params.SinceDate != nil && p.CreatedAt.Before(*params.SinceDate) {
			continue
		}
		if params.UntilDate != nil && p.CreatedAt.After(*params.UntilDate) {
			continue
		}

		cloned := *p
		if viewerID != nil {
			cloned.IsLiked = r.isLikedInternal(p.ID, *viewerID)
			cloned.IsReposted = r.isRepostedInternal(p.ID, *viewerID)
			cloned.IsBookmarked = r.isBookmarkedInternal(p.ID, *viewerID)
		}
		res = append(res, cloned)
	}

	limit := params.Limit
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	// Sort based on tab
	if params.Tab == "latest" {
		sort.Slice(res, func(i, j int) bool {
			return res[i].CreatedAt.After(res[j].CreatedAt)
		})
	} else {
		// Top ranking: relevance + engagement
		sort.Slice(res, func(i, j int) bool {
			scoreI := float64(res[i].LikesCount*2 + res[i].RepostsCount*3)
			scoreJ := float64(res[j].LikesCount*2 + res[j].RepostsCount*3)
			return scoreI > scoreJ
		})
	}

	var nextCursor string
	if len(res) > limit {
		last := res[limit-1]
		nextCursor = recommendation.EncodeCursor(last.CreatedAt, last.ID)
		res = res[:limit]
	}

	return res, nextCursor, nil
}

func (r *MemoryRepository) SearchUsersAdvanced(ctx context.Context, query string, viewerID *uuid.UUID, cursor string, limit int) ([]domain.User, string, error) {
	users, err := r.SearchUsers(ctx, query, viewerID, limit)
	if err != nil {
		return nil, "", err
	}
	return users, "", nil
}

func (r *MemoryRepository) GetSearchSuggestions(ctx context.Context, query string, viewerID *uuid.UUID, limit int) ([]string, []domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	low := strings.ToLower(query)
	var queries []string
	var users []domain.User

	// Find matching hashtags and queries
	for _, p := range r.posts {
		if p.DeletedAt == nil && strings.Contains(strings.ToLower(p.Content), low) {
			words := strings.Fields(p.Content)
			for _, w := range words {
				if strings.HasPrefix(w, "#") && strings.Contains(strings.ToLower(w), low) {
					found := false
					for _, q := range queries {
						if strings.EqualFold(q, w) {
							found = true
							break
						}
					}
					if !found {
						queries = append(queries, w)
					}
				}
			}
		}
		if len(queries) >= 5 {
			break
		}
	}

	// Find matching users
	for _, u := range r.users {
		if (strings.Contains(strings.ToLower(u.Username), low) || strings.Contains(strings.ToLower(u.DisplayName), low)) && u.DeletedAt == nil {
			cloned := *u
			if viewerID != nil {
				cloned.IsFollowing = r.isFollowingInternal(*viewerID, u.ID)
			}
			users = append(users, cloned)
			if len(users) >= 5 {
				break
			}
		}
	}

	return queries, users, nil
}

func (r *MemoryRepository) GetRecentSearches(ctx context.Context, userID uuid.UUID, limit int) ([]domain.RecentSearch, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if limit <= 0 || limit > 50 {
		limit = 10
	}

	list := r.recentSearches[userID]
	if len(list) > limit {
		list = list[:limit]
	}
	return list, nil
}

func (r *MemoryRepository) AddRecentSearch(ctx context.Context, userID uuid.UUID, query string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	trimmed := strings.TrimSpace(query)
	if trimmed == "" {
		return nil
	}

	existing := r.recentSearches[userID]
	var filtered []domain.RecentSearch
	for _, item := range existing {
		if !strings.EqualFold(item.Query, trimmed) {
			filtered = append(filtered, item)
		}
	}

	newEntry := domain.RecentSearch{
		ID:         uuid.New(),
		UserID:     userID,
		Query:      trimmed,
		SearchedAt: time.Now().UTC(),
	}

	r.recentSearches[userID] = append([]domain.RecentSearch{newEntry}, filtered...)
	return nil
}

func (r *MemoryRepository) DeleteRecentSearch(ctx context.Context, userID uuid.UUID, searchID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing := r.recentSearches[userID]
	var updated []domain.RecentSearch
	for _, item := range existing {
		if item.ID != searchID {
			updated = append(updated, item)
		}
	}
	r.recentSearches[userID] = updated
	return nil
}

func (r *MemoryRepository) ClearRecentSearches(ctx context.Context, userID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.recentSearches, userID)
	return nil
}

func (r *MemoryRepository) CreateNotification(ctx context.Context, n *domain.Notification) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Deduplicate notifications created in the last 30 seconds
	now := time.Now()
	for _, existing := range r.notifications {
		if existing.RecipientID == n.RecipientID &&
			existing.ActorID == n.ActorID &&
			existing.Type == n.Type &&
			((existing.PostID == nil && n.PostID == nil) || (existing.PostID != nil && n.PostID != nil && *existing.PostID == *n.PostID)) &&
			now.Sub(existing.CreatedAt) < 30*time.Second {
			return errors.New("duplicate_notification_skipped")
		}
	}

	if n.ID == uuid.Nil {
		n.ID = uuid.New()
	}
	n.CreatedAt = now
	r.notifications[n.ID] = n
	return nil
}

func (r *MemoryRepository) GetNotifications(ctx context.Context, userID uuid.UUID, cursor string, limit int) ([]domain.Notification, string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var res []domain.Notification
	for _, n := range r.notifications {
		if n.RecipientID == userID {
			cloned := *n
			if actor, exists := r.users[n.ActorID]; exists {
				cloned.Actor = actor
			}
			if n.PostID != nil {
				if post, exists := r.posts[*n.PostID]; exists {
					cloned.Post = post
				}
			}
			res = append(res, cloned)
		}
	}
	return res, "", nil
}

func (r *MemoryRepository) MarkNotificationsRead(ctx context.Context, userID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, n := range r.notifications {
		if n.RecipientID == userID {
			n.IsRead = true
		}
	}
	return nil
}

func (r *MemoryRepository) CreateMediaObject(ctx context.Context, m *domain.MediaObject) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	m.CreatedAt = time.Now()
	r.media[m.ID] = m
	return nil
}

func (r *MemoryRepository) GetMediaObjectByID(ctx context.Context, id uuid.UUID) (*domain.MediaObject, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	m, exists := r.media[id]
	if !exists {
		return nil, ErrNotFound
	}
	return m, nil
}

// Helpers
func (r *MemoryRepository) isFollowingInternal(follower, following uuid.UUID) bool {
	_, ok := r.follows[fmt.Sprintf("%s:%s", follower, following)]
	return ok
}

func (r *MemoryRepository) isLikedInternal(postID, userID uuid.UUID) bool {
	_, ok := r.likes[fmt.Sprintf("%s:%s", postID, userID)]
	return ok
}

func (r *MemoryRepository) isRepostedInternal(postID, userID uuid.UUID) bool {
	_, ok := r.reposts[fmt.Sprintf("%s:%s", postID, userID)]
	return ok
}

func (r *MemoryRepository) isBookmarkedInternal(postID, userID uuid.UUID) bool {
	_, ok := r.bookmarks[fmt.Sprintf("%s:%s", postID, userID)]
	return ok
}

// Subscriptions & Verification (Section 6A)
func (r *MemoryRepository) GetSubscriptionPlans(ctx context.Context) ([]domain.SubscriptionPlan, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var plans []domain.SubscriptionPlan
	order := []string{"free", "basic", "premium", "premium_plus"}
	for _, id := range order {
		if p, ok := r.plans[id]; ok {
			plans = append(plans, *p)
		}
	}
	return plans, nil
}

func (r *MemoryRepository) GetSubscriptionPlanByID(ctx context.Context, id string) (*domain.SubscriptionPlan, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	p, ok := r.plans[id]
	if !ok {
		return nil, ErrNotFound
	}
	return p, nil
}

func (r *MemoryRepository) GetActiveSubscription(ctx context.Context, userID uuid.UUID) (*domain.Subscription, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, sub := range r.subs {
		if sub.UserID == userID && sub.Status == "active" && sub.CurrentPeriodEnd.After(time.Now()) {
			plan := r.plans[sub.PlanID]
			subCopy := *sub
			subCopy.Plan = plan
			return &subCopy, nil
		}
	}
	return nil, nil
}

func (r *MemoryRepository) CreateSubscription(ctx context.Context, sub *domain.Subscription) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if sub.ID == uuid.Nil {
		sub.ID = uuid.New()
	}
	sub.CreatedAt = time.Now()
	sub.UpdatedAt = time.Now()
	r.subs[sub.ID] = sub
	return nil
}

func (r *MemoryRepository) CancelSubscription(ctx context.Context, subID, userID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	sub, ok := r.subs[subID]
	if !ok || sub.UserID != userID {
		return ErrNotFound
	}
	sub.CancelAtPeriodEnd = true
	sub.UpdatedAt = time.Now()
	return nil
}

func (r *MemoryRepository) GetUserEntitlements(ctx context.Context, userID uuid.UUID) (*domain.Entitlements, error) {
	sub, err := r.GetActiveSubscription(ctx, userID)
	if err != nil {
		return nil, err
	}

	ent := &domain.Entitlements{
		PlanID:           "free",
		MaxPostLength:    280,
		HasVerifiedBadge: false,
		BadgeType:        "none",
		CanEditPosts:     false,
		CanUndoPosts:     false,
		MaxVideoMinutes:  2,
		MaxMediaCount:    4,
		ReducedAds:       false,
	}

	if sub != nil && sub.Plan != nil {
		ent.PlanID = sub.PlanID
		ent.MaxPostLength = sub.Plan.MaxPostLength
		ent.HasVerifiedBadge = sub.Plan.CanGetVerifiedBadge
		if sub.Plan.CanGetVerifiedBadge {
			ent.BadgeType = "blue"
		}
		ent.CanEditPosts = sub.Plan.CanEditPosts
		ent.CanUndoPosts = sub.Plan.CanUndoPosts
		ent.MaxVideoMinutes = sub.Plan.MaxVideoMinutes
		ent.MaxMediaCount = sub.Plan.MaxMediaCount
		ent.ReducedAds = sub.Plan.ReducedAds
	}

	r.mu.RLock()
	defer r.mu.RUnlock()
	if user, ok := r.users[userID]; ok && user.IsVerified {
		ent.HasVerifiedBadge = true
		ent.BadgeType = user.VerifiedType
	}

	return ent, nil
}

func (r *MemoryRepository) CreatePaymentOrder(ctx context.Context, order *domain.PaymentOrder) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if order.ID == uuid.Nil {
		order.ID = uuid.New()
	}
	order.CreatedAt = time.Now()
	order.UpdatedAt = time.Now()
	r.orders[order.ID] = order
	return nil
}

func (r *MemoryRepository) GetPaymentOrderByID(ctx context.Context, orderID uuid.UUID) (*domain.PaymentOrder, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	o, ok := r.orders[orderID]
	if !ok {
		return nil, ErrNotFound
	}
	return o, nil
}

func (r *MemoryRepository) UpdatePaymentOrderStatus(ctx context.Context, orderID uuid.UUID, status, invoiceURL string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	o, ok := r.orders[orderID]
	if !ok {
		return ErrNotFound
	}
	o.Status = status
	if invoiceURL != "" {
		o.ProviderInvoiceURL = invoiceURL
	}
	o.UpdatedAt = time.Now()
	return nil
}

func (r *MemoryRepository) GetBillingHistory(ctx context.Context, userID uuid.UUID) ([]domain.PaymentOrder, []domain.Invoice, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var orders []domain.PaymentOrder
	for _, o := range r.orders {
		if o.UserID == userID {
			orders = append(orders, *o)
		}
	}

	var invoices []domain.Invoice
	for _, inv := range r.invoices {
		if inv.UserID == userID {
			invoices = append(invoices, *inv)
		}
	}

	return orders, invoices, nil
}

func (r *MemoryRepository) CreateInvoice(ctx context.Context, inv *domain.Invoice) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if inv.ID == uuid.Nil {
		inv.ID = uuid.New()
	}
	inv.CreatedAt = time.Now()
	r.invoices[inv.ID] = inv
	return nil
}

func (r *MemoryRepository) GetUserBadges(ctx context.Context, userID uuid.UUID) ([]domain.VerificationBadge, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var badges []domain.VerificationBadge
	for _, b := range r.badges {
		if b.UserID == userID && b.Status == "approved" {
			badges = append(badges, *b)
		}
	}
	return badges, nil
}

func (r *MemoryRepository) AssignVerificationBadge(ctx context.Context, badge *domain.VerificationBadge) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if badge.ID == uuid.Nil {
		badge.ID = uuid.New()
	}
	badge.CreatedAt = time.Now()
	badge.UpdatedAt = time.Now()
	r.badges[badge.ID] = badge

	if u, ok := r.users[badge.UserID]; ok {
		u.IsVerified = true
		u.VerifiedType = badge.BadgeType
	}
	return nil
}

func (r *MemoryRepository) CreateVerificationApplication(ctx context.Context, app *domain.VerificationApplication) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if app.ID == uuid.Nil {
		app.ID = uuid.New()
	}
	app.CreatedAt = time.Now()
	r.applications[app.ID] = app
	return nil
}

func (r *MemoryRepository) GetVerificationApplication(ctx context.Context, userID uuid.UUID) (*domain.VerificationApplication, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, a := range r.applications {
		if a.UserID == userID {
			return a, nil
		}
	}
	return nil, nil
}

func (r *MemoryRepository) UpdateUserVerification(ctx context.Context, userID uuid.UUID, isVerified bool, verifiedType string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if u, ok := r.users[userID]; ok {
		u.IsVerified = isVerified
		u.VerifiedType = verifiedType
		u.UpdatedAt = time.Now()
		return nil
	}
	return ErrNotFound
}

func (r *MemoryRepository) GetFollowers(ctx context.Context, userID, viewerID uuid.UUID, limit int, cursor string) ([]domain.User, string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var users []domain.User
	for key := range r.follows {
		parts := strings.Split(key, ":")
		if len(parts) == 2 && parts[1] == userID.String() {
			followerUUID, err := uuid.Parse(parts[0])
			if err == nil {
				if u, ok := r.users[followerUUID]; ok {
					userCopy := *u
					_, userCopy.IsFollowing = r.follows[viewerID.String()+":"+u.ID.String()]
					_, userCopy.IsFollowedBy = r.follows[u.ID.String()+":"+viewerID.String()]
					users = append(users, userCopy)
				}
			}
		}
	}
	return users, "", nil
}

func (r *MemoryRepository) GetFollowing(ctx context.Context, userID, viewerID uuid.UUID, limit int, cursor string) ([]domain.User, string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var users []domain.User
	for key := range r.follows {
		parts := strings.Split(key, ":")
		if len(parts) == 2 && parts[0] == userID.String() {
			followingUUID, err := uuid.Parse(parts[1])
			if err == nil {
				if u, ok := r.users[followingUUID]; ok {
					userCopy := *u
					_, userCopy.IsFollowing = r.follows[viewerID.String()+":"+u.ID.String()]
					_, userCopy.IsFollowedBy = r.follows[u.ID.String()+":"+viewerID.String()]
					users = append(users, userCopy)
				}
			}
		}
	}
	return users, "", nil
}

func (r *MemoryRepository) GetMutualConnections(ctx context.Context, userID, viewerID uuid.UUID, limit int) ([]domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var users []domain.User
	for _, u := range r.users {
		_, vFollows := r.follows[viewerID.String()+":"+u.ID.String()]
		_, tFollows := r.follows[userID.String()+":"+u.ID.String()]
		if vFollows && tFollows {
			userCopy := *u
			userCopy.IsFollowing = true
			users = append(users, userCopy)
			if len(users) >= limit {
				break
			}
		}
	}
	return users, nil
}

func (r *MemoryRepository) GetSocialRelationship(ctx context.Context, viewerID, targetID uuid.UUID) (*domain.SocialRelationship, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	rel := &domain.SocialRelationship{
		TargetUserID: targetID,
	}

	targetUser, ok := r.users[targetID]
	if !ok {
		return nil, ErrNotFound
	}
	rel.IsPrivate = targetUser.IsPrivate

	_, rel.IsFollowing = r.follows[viewerID.String()+":"+targetID.String()]
	_, rel.IsFollowedBy = r.follows[targetID.String()+":"+viewerID.String()]
	rel.IsMutual = rel.IsFollowing && rel.IsFollowedBy

	_, rel.IsBlocked = r.blocks[viewerID.String()+":"+targetID.String()]
	_, rel.IsBlockedBy = r.blocks[targetID.String()+":"+viewerID.String()]
	_, rel.IsMuted = r.mutes[viewerID.String()+":"+targetID.String()]

	for _, req := range r.followReqs {
		if req.RequesterID == viewerID && req.TargetID == targetID && req.Status == "pending" {
			rel.HasPendingFollowRequest = true
			break
		}
	}

	return rel, nil
}

func (r *MemoryRepository) BlockUser(ctx context.Context, blockerID, blockedID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.blocks[blockerID.String()+":"+blockedID.String()] = time.Now()
	delete(r.follows, blockerID.String()+":"+blockedID.String())
	delete(r.follows, blockedID.String()+":"+blockerID.String())
	return nil
}

func (r *MemoryRepository) UnblockUser(ctx context.Context, blockerID, blockedID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.blocks, blockerID.String()+":"+blockedID.String())
	return nil
}

func (r *MemoryRepository) MuteUser(ctx context.Context, muterID, mutedID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.mutes[muterID.String()+":"+mutedID.String()] = time.Now()
	return nil
}

func (r *MemoryRepository) UnmuteUser(ctx context.Context, muterID, mutedID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.mutes, muterID.String()+":"+mutedID.String())
	return nil
}

func (r *MemoryRepository) IsBlocked(ctx context.Context, userA, userB uuid.UUID) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, b1 := r.blocks[userA.String()+":"+userB.String()]
	_, b2 := r.blocks[userB.String()+":"+userA.String()]
	return b1 || b2, nil
}

func (r *MemoryRepository) CreateFollowRequest(ctx context.Context, req *domain.FollowRequest) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if req.ID == uuid.Nil {
		req.ID = uuid.New()
	}
	if req.Status == "" {
		req.Status = "pending"
	}
	req.CreatedAt = time.Now()
	req.UpdatedAt = time.Now()
	r.followReqs[req.ID] = req
	return nil
}

func (r *MemoryRepository) GetFollowRequests(ctx context.Context, targetID uuid.UUID) ([]domain.FollowRequest, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var reqs []domain.FollowRequest
	for _, req := range r.followReqs {
		if req.TargetID == targetID && req.Status == "pending" {
			reqCopy := *req
			if u, ok := r.users[req.RequesterID]; ok {
				reqCopy.Requester = u
			}
			reqs = append(reqs, reqCopy)
		}
	}
	return reqs, nil
}

func (r *MemoryRepository) RespondFollowRequest(ctx context.Context, requestID, targetID uuid.UUID, accept bool) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	req, ok := r.followReqs[requestID]
	if !ok || req.TargetID != targetID {
		return ErrNotFound
	}

	if accept {
		req.Status = "accepted"
		r.follows[req.RequesterID.String()+":"+req.TargetID.String()] = time.Now()
		if target, ok := r.users[targetID]; ok {
			target.FollowersCount++
		}
		if reqUser, ok := r.users[req.RequesterID]; ok {
			reqUser.FollowingCount++
		}
		notifID := uuid.New()
		r.notifications[notifID] = &domain.Notification{
			ID:          notifID,
			RecipientID: req.RequesterID,
			ActorID:     targetID,
			Type:        "follow_accept",
			IsRead:      false,
			CreatedAt:   time.Now(),
		}
	} else {
		req.Status = "rejected"
	}
	req.UpdatedAt = time.Now()

	// Mark follow_request notification as read for target
	for _, n := range r.notifications {
		if n.RecipientID == targetID && n.Type == "follow_request" && n.ActorID == req.RequesterID {
			n.IsRead = true
		}
	}

	return nil
}

func (r *MemoryRepository) CancelFollowRequest(ctx context.Context, requesterID, targetID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for id, req := range r.followReqs {
		if req.RequesterID == requesterID && req.TargetID == targetID && req.Status == "pending" {
			delete(r.followReqs, id)
			break
		}
	}
	return nil
}

func (r *MemoryRepository) GetConversations(ctx context.Context, userID uuid.UUID, limit int, cursor string) ([]domain.Conversation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var convs []domain.Conversation
	for _, c := range r.conversations {
		members := r.convMembers[c.ID]
		isMember := false
		var otherID uuid.UUID
		for _, m := range members {
			if m == userID {
				isMember = true
			} else {
				otherID = m
			}
		}
		if isMember {
			cCopy := *c
			if otherU, ok := r.users[otherID]; ok {
				cCopy.OtherMember = otherU
			}
			convs = append(convs, cCopy)
		}
	}
	return convs, nil
}

func (r *MemoryRepository) GetOrCreateConversation(ctx context.Context, userA, userB uuid.UUID) (*domain.Conversation, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, c := range r.conversations {
		if !c.IsGroup {
			members := r.convMembers[c.ID]
			if len(members) == 2 {
				if (members[0] == userA && members[1] == userB) || (members[0] == userB && members[1] == userA) {
					return c, nil
				}
			}
		}
	}

	newC := &domain.Conversation{
		ID:        uuid.New(),
		IsGroup:   false,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	r.conversations[newC.ID] = newC
	r.convMembers[newC.ID] = []uuid.UUID{userA, userB}
	return newC, nil
}

func (r *MemoryRepository) GetConversationByID(ctx context.Context, convID, userID uuid.UUID) (*domain.Conversation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	c, ok := r.conversations[convID]
	if !ok {
		return nil, ErrNotFound
	}
	members := r.convMembers[convID]
	isMember := false
	var otherID uuid.UUID
	for _, m := range members {
		if m == userID {
			isMember = true
		} else {
			otherID = m
		}
	}
	if !isMember {
		return nil, ErrNotFound
	}

	cCopy := *c
	if otherU, ok := r.users[otherID]; ok {
		cCopy.OtherMember = otherU
	}
	return &cCopy, nil
}

func (r *MemoryRepository) GetMessages(ctx context.Context, convID, userID uuid.UUID, limit int, beforeID *uuid.UUID) ([]domain.Message, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var msgs []domain.Message
	for _, m := range r.messages {
		if m.ConversationID == convID {
			mCopy := *m
			if sender, ok := r.users[m.SenderID]; ok {
				mCopy.Sender = sender
			}
			msgs = append(msgs, mCopy)
		}
	}
	return msgs, nil
}

func (r *MemoryRepository) CreateMessage(ctx context.Context, msg *domain.Message) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if msg.ID == uuid.Nil {
		msg.ID = uuid.New()
	}
	msg.CreatedAt = time.Now()
	r.messages[msg.ID] = msg
	if c, ok := r.conversations[msg.ConversationID]; ok {
		c.UpdatedAt = time.Now()
		c.LastMessage = msg
	}
	return nil
}

func (r *MemoryRepository) MarkConversationRead(ctx context.Context, convID, userID uuid.UUID, messageID *uuid.UUID) error {
	return nil
}

func (r *MemoryRepository) GetNotificationsWithTab(ctx context.Context, userID uuid.UUID, tab string, cursor string, limit int) ([]domain.Notification, string, error) {
	return r.GetNotifications(ctx, userID, cursor, limit)
}

func (r *MemoryRepository) GetUnreadNotificationsCount(ctx context.Context, userID uuid.UUID) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	count := 0
	for _, n := range r.notifications {
		if n.RecipientID == userID && !n.IsRead {
			count++
		}
	}
	return count, nil
}

func (r *MemoryRepository) MarkNotificationRead(ctx context.Context, notifID, userID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if n, ok := r.notifications[notifID]; ok && n.RecipientID == userID {
		n.IsRead = true
	}
	return nil
}

func (r *MemoryRepository) CreateOutboxEvent(ctx context.Context, event *domain.OutboxEvent) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if event.ID == uuid.Nil {
		event.ID = uuid.New()
	}
	event.CreatedAt = time.Now()
	r.outboxEvents[event.ID] = event
	return nil
}

func (r *MemoryRepository) GetPendingOutboxEvents(ctx context.Context, batchSize int) ([]domain.OutboxEvent, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	now := time.Now()
	var events []domain.OutboxEvent
	for _, e := range r.outboxEvents {
		if (e.Status == "pending" || e.Status == "retrying") && (e.NextRetryAt == nil || e.NextRetryAt.Before(now) || e.NextRetryAt.Equal(now)) {
			events = append(events, *e)
			if len(events) >= batchSize {
				break
			}
		}
	}
	return events, nil
}

func (r *MemoryRepository) MarkOutboxEventPublished(ctx context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if e, ok := r.outboxEvents[id]; ok {
		e.Status = "published"
		now := time.Now()
		e.PublishedAt = &now
		e.LastError = nil
	}
	return nil
}

func (r *MemoryRepository) MarkOutboxEventFailed(ctx context.Context, id uuid.UUID, errMsg string, nextRetryAt *time.Time, isDLQ bool) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if e, ok := r.outboxEvents[id]; ok {
		if isDLQ {
			e.Status = "failed"
		} else {
			e.Status = "retrying"
		}
		e.RetryCount++
		e.LastError = &errMsg
		if nextRetryAt != nil {
			e.NextRetryAt = nextRetryAt
		} else {
			t := time.Now().Add(10 * time.Second)
			e.NextRetryAt = &t
		}
	}
	return nil
}

func (r *MemoryRepository) CleanupOldOutboxEvents(ctx context.Context, olderThan time.Duration) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	cutoff := time.Now().Add(-olderThan)
	var count int64
	for id, e := range r.outboxEvents {
		if e.Status == "published" && e.PublishedAt != nil && e.PublishedAt.Before(cutoff) {
			delete(r.outboxEvents, id)
			count++
		}
	}
	return count, nil
}

// ==============================================================================
// CandidateProvider & Recommendation Implementation (Memory)
// ==============================================================================

func (r *MemoryRepository) GetFollowingCandidates(ctx context.Context, viewerID uuid.UUID, limit int) ([]domain.Post, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var posts []domain.Post
	for _, p := range r.posts {
		if p.DeletedAt != nil {
			continue
		}
		if r.isFollowingInternal(viewerID, p.AuthorID) {
			posts = append(posts, *p)
			if len(posts) >= limit {
				break
			}
		}
	}
	return posts, nil
}

func (r *MemoryRepository) GetSocialEngagementCandidates(ctx context.Context, viewerID uuid.UUID, limit int) ([]domain.Post, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var posts []domain.Post
	// Find posts liked by people viewer follows
	for _, p := range r.posts {
		if p.DeletedAt != nil {
			continue
		}
		for key := range r.likes {
			parts := strings.Split(key, ":")
			if len(parts) == 2 && parts[0] == p.ID.String() {
				likerID, err := uuid.Parse(parts[1])
				if err == nil && r.isFollowingInternal(viewerID, likerID) {
					posts = append(posts, *p)
					break
				}
			}
		}
		if len(posts) >= limit {
			break
		}
	}
	return posts, nil
}

func (r *MemoryRepository) GetTopicCandidates(ctx context.Context, viewerID *uuid.UUID, topics []string, limit int) ([]domain.Post, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	topicSet := make(map[string]bool)
	for _, t := range topics {
		topicSet[strings.ToLower(strings.TrimPrefix(t, "#"))] = true
	}

	var posts []domain.Post
	for _, p := range r.posts {
		if p.DeletedAt != nil {
			continue
		}
		contentLower := strings.ToLower(p.Content)
		matched := false
		for t := range topicSet {
			if strings.Contains(contentLower, t) {
				matched = true
				break
			}
		}
		if matched {
			posts = append(posts, *p)
			if len(posts) >= limit {
				break
			}
		}
	}
	return posts, nil
}

func (r *MemoryRepository) GetTrendingCandidates(ctx context.Context, limit int) ([]domain.Post, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var all []domain.Post
	for _, p := range r.posts {
		if p.DeletedAt == nil {
			all = append(all, *p)
		}
	}
	// Sort by interactions
	for i := 0; i < len(all)-1; i++ {
		for j := i + 1; j < len(all); j++ {
			scoreI := all[i].LikesCount + all[i].RepostsCount*2 + all[i].RepliesCount*3
			scoreJ := all[j].LikesCount + all[j].RepostsCount*2 + all[j].RepliesCount*3
			if scoreI < scoreJ {
				all[i], all[j] = all[j], all[i]
			}
		}
	}
	if len(all) > limit {
		all = all[:limit]
	}
	return all, nil
}

func (r *MemoryRepository) GetFreshCandidates(ctx context.Context, limit int) ([]domain.Post, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var all []domain.Post
	for _, p := range r.posts {
		if p.DeletedAt == nil {
			all = append(all, *p)
		}
	}
	// Newest first
	for i := 0; i < len(all)-1; i++ {
		for j := i + 1; j < len(all); j++ {
			if all[i].CreatedAt.Before(all[j].CreatedAt) {
				all[i], all[j] = all[j], all[i]
			}
		}
	}
	if len(all) > limit {
		all = all[:limit]
	}
	return all, nil
}

func (r *MemoryRepository) GetExplorationCandidates(ctx context.Context, limit int) ([]domain.Post, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var all []domain.Post
	for _, p := range r.posts {
		if p.DeletedAt == nil && p.Author != nil && p.Author.FollowersCount < 1000 {
			all = append(all, *p)
		}
	}
	if len(all) > limit {
		all = all[:limit]
	}
	return all, nil
}

func (r *MemoryRepository) GetFollowingPosts(ctx context.Context, viewerID uuid.UUID, cursor *recommendation.FeedCursor, limit int) ([]domain.Post, string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var matched []domain.Post
	for _, p := range r.posts {
		if p.DeletedAt != nil {
			continue
		}
		if r.isFollowingInternal(viewerID, p.AuthorID) || p.AuthorID == viewerID {
			if cursor != nil {
				if !p.CreatedAt.Before(cursor.CreatedAt) && !(p.CreatedAt.Equal(cursor.CreatedAt) && p.ID.String() < cursor.PostID.String()) {
					continue
				}
			}
			cloned := *p
			cloned.IsLiked = r.isLikedInternal(p.ID, viewerID)
			cloned.IsReposted = r.isRepostedInternal(p.ID, viewerID)
			cloned.IsBookmarked = r.isBookmarkedInternal(p.ID, viewerID)
			matched = append(matched, cloned)
		}
	}

	// Sort newest first
	for i := 0; i < len(matched)-1; i++ {
		for j := i + 1; j < len(matched); j++ {
			if matched[i].CreatedAt.Before(matched[j].CreatedAt) {
				matched[i], matched[j] = matched[j], matched[i]
			}
		}
	}

	hasMore := len(matched) > limit
	if len(matched) > limit {
		matched = matched[:limit]
	}

	var nextCursor string
	if hasMore && len(matched) > 0 {
		last := matched[len(matched)-1]
		nextCursor = recommendation.EncodeCursor(last.CreatedAt, last.ID)
	}
	return matched, nextCursor, nil
}

func (r *MemoryRepository) GetViewerInteractions(ctx context.Context, viewerID uuid.UUID) (
	followedAuthors map[uuid.UUID]bool,
	mutualAuthors map[uuid.UUID]bool,
	likedPosts map[uuid.UUID]bool,
	mutedAuthors map[uuid.UUID]bool,
	blockedAuthors map[uuid.UUID]bool,
	hiddenPosts map[uuid.UUID]bool,
	userTopics map[string]float64,
	err error,
) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	followedAuthors = make(map[uuid.UUID]bool)
	mutualAuthors = make(map[uuid.UUID]bool)
	likedPosts = make(map[uuid.UUID]bool)
	mutedAuthors = make(map[uuid.UUID]bool)
	blockedAuthors = make(map[uuid.UUID]bool)
	hiddenPosts = make(map[uuid.UUID]bool)
	userTopics = make(map[string]float64)

	// Follows & Mutuals
	for key := range r.follows {
		parts := strings.Split(key, ":")
		if len(parts) == 2 {
			follower, _ := uuid.Parse(parts[0])
			following, _ := uuid.Parse(parts[1])
			if follower == viewerID {
				followedAuthors[following] = true
				if r.isFollowingInternal(following, viewerID) {
					mutualAuthors[following] = true
				}
			}
		}
	}

	// Likes
	for key := range r.likes {
		parts := strings.Split(key, ":")
		if len(parts) == 2 {
			pID, _ := uuid.Parse(parts[0])
			uID, _ := uuid.Parse(parts[1])
			if uID == viewerID {
				likedPosts[pID] = true
			}
		}
	}

	// Mutes
	for key := range r.mutes {
		parts := strings.Split(key, ":")
		if len(parts) == 2 && parts[0] == viewerID.String() {
			mutedID, _ := uuid.Parse(parts[1])
			mutedAuthors[mutedID] = true
		}
	}

	// Blocks (Bilateral)
	for key := range r.blocks {
		parts := strings.Split(key, ":")
		if len(parts) == 2 {
			b1, _ := uuid.Parse(parts[0])
			b2, _ := uuid.Parse(parts[1])
			if b1 == viewerID {
				blockedAuthors[b2] = true
			} else if b2 == viewerID {
				blockedAuthors[b1] = true
			}
		}
	}

	// Feed Feedback (Hidden posts)
	for _, fb := range r.feedFeedback {
		if fb.UserID == viewerID && (fb.FeedbackType == "hide" || fb.FeedbackType == "not_interested") {
			hiddenPosts[fb.PostID] = true
		}
	}

	// User Topic Preferences
	for _, pref := range r.userTopicPreferences {
		if pref.UserID == viewerID && !pref.IsMuted {
			userTopics[pref.Topic] = pref.AffinityScore
		}
	}

	return followedAuthors, mutualAuthors, likedPosts, mutedAuthors, blockedAuthors, hiddenPosts, userTopics, nil
}

func (r *MemoryRepository) GetActiveWeights(ctx context.Context) (domain.RankingWeights, string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.activeWeights, "memory_weights_v1", nil
}

func (r *MemoryRepository) RecordFeedFeedback(ctx context.Context, feedback domain.FeedFeedback) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if feedback.ID == uuid.Nil {
		feedback.ID = uuid.New()
	}
	feedback.CreatedAt = time.Now()
	key := fmt.Sprintf("%s:%s:%s", feedback.UserID, feedback.PostID, feedback.FeedbackType)
	r.feedFeedback[key] = &feedback
	return nil
}

func (r *MemoryRepository) RecordFeedImpressions(ctx context.Context, impressions []domain.FeedImpression) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	for _, imp := range impressions {
		copyImp := imp
		if copyImp.ID == uuid.Nil {
			copyImp.ID = uuid.New()
		}
		copyImp.CreatedAt = now
		r.feedImpressions = append(r.feedImpressions, &copyImp)
	}
	return nil
}

func (r *MemoryRepository) GetUserTopicPreferences(ctx context.Context, userID uuid.UUID) ([]domain.UserTopicPreference, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []domain.UserTopicPreference
	for _, pref := range r.userTopicPreferences {
		if pref.UserID == userID {
			result = append(result, *pref)
		}
	}
	return result, nil
}

func (r *MemoryRepository) UpdateUserTopicPreference(ctx context.Context, pref domain.UserTopicPreference) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if pref.ID == uuid.Nil {
		pref.ID = uuid.New()
	}
	pref.UpdatedAt = time.Now()
	key := fmt.Sprintf("%s:%s", pref.UserID, strings.ToLower(pref.Topic))
	r.userTopicPreferences[key] = &pref
	return nil
}

func (r *MemoryRepository) ResetPersonalization(ctx context.Context, userID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Flush feedback
	for key, fb := range r.feedFeedback {
		if fb.UserID == userID {
			delete(r.feedFeedback, key)
		}
	}
	// Flush non-explicit topic preferences
	for key, pref := range r.userTopicPreferences {
		if pref.UserID == userID && !pref.IsExplicit {
			delete(r.userTopicPreferences, key)
		}
	}
	return nil
}

func (r *MemoryRepository) GetTrendingTopics(ctx context.Context, limit int) ([]domain.TrendingTopic, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if limit <= 0 || limit > 50 {
		limit = 10
	}

	// 1. Scan in-memory posts for hashtags in rolling 24-hour window
	now := time.Now()
	cutoff := now.Add(-24 * time.Hour)

	type topicStat struct {
		tag             string
		postIDs         map[uuid.UUID]bool
		authorIDs       map[uuid.UUID]bool
		totalEngagement int
	}

	topicMap := make(map[string]*topicStat)

	for _, post := range r.posts {
		if post.DeletedAt != nil {
			continue
		}
		if post.CreatedAt.Before(cutoff) {
			continue
		}

		words := strings.Fields(post.Content)
		for _, w := range words {
			if strings.HasPrefix(w, "#") && len(w) > 1 {
				cleanTag := strings.ToLower(strings.Trim(w, "#.,!?:;\"'()[]{}"))
				if len(cleanTag) == 0 {
					continue
				}
				if _, ok := topicMap[cleanTag]; !ok {
					topicMap[cleanTag] = &topicStat{
						tag:       cleanTag,
						postIDs:   make(map[uuid.UUID]bool),
						authorIDs: make(map[uuid.UUID]bool),
					}
				}
				stat := topicMap[cleanTag]
				stat.postIDs[post.ID] = true
				stat.authorIDs[post.AuthorID] = true
				stat.totalEngagement += post.LikesCount + post.RepostsCount*2 + post.RepliesCount*3
			}
		}
	}

	var topics []domain.TrendingTopic
	for tag, stat := range topicMap {
		recentPosts := len(stat.postIDs)
		uniqueAuthors := len(stat.authorIDs)
		rawVelocity := float64(uniqueAuthors)*3.0 + float64(recentPosts)*1.5 + float64(stat.totalEngagement)*0.1
		// Anti-spam penalty if single user spams posts
		if recentPosts > uniqueAuthors*2 {
			rawVelocity -= float64(recentPosts-uniqueAuthors*2) * 2.0
		}
		if rawVelocity < 0 {
			rawVelocity = 0
		}

		topics = append(topics, domain.TrendingTopic{
			Category:      categorizeTopic(tag),
			Topic:         "#" + tag,
			PostsCount:    recentPosts,
			VelocityScore: rawVelocity,
		})
	}

	sort.Slice(topics, func(i, j int) bool {
		return topics[i].VelocityScore > topics[j].VelocityScore
	})

	if len(topics) >= limit {
		return topics[:limit], nil
	}

	if topics == nil {
		topics = []domain.TrendingTopic{}
	}
	return topics, nil
}

func (r *MemoryRepository) GetNewPosts(ctx context.Context, feedType string, viewerID *uuid.UUID, sinceCursor string, limit int) ([]domain.Post, int, string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if limit <= 0 || limit > 50 {
		limit = 20
	}

	cursorObj, err := recommendation.DecodeCursor(sinceCursor)
	if err != nil || cursorObj == nil {
		return []domain.Post{}, 0, "", nil
	}

	var candidates []domain.Post

	for _, post := range r.posts {
		if post.DeletedAt != nil {
			continue
		}

		// Keyset condition: must be strictly newer than cursor
		isNewer := post.CreatedAt.After(cursorObj.CreatedAt) || (post.CreatedAt.Equal(cursorObj.CreatedAt) && post.ID.String() > cursorObj.PostID.String())
		if !isNewer {
			continue
		}

		if viewerID != nil {
			// Exclude blocked authors
			if _, blocked := r.blocks[fmt.Sprintf("%s:%s", *viewerID, post.AuthorID)]; blocked {
				continue
			}
			if _, blocked := r.blocks[fmt.Sprintf("%s:%s", post.AuthorID, *viewerID)]; blocked {
				continue
			}
			// Exclude muted authors
			if _, muted := r.mutes[fmt.Sprintf("%s:%s", *viewerID, post.AuthorID)]; muted {
				continue
			}
		}

		if feedType == "following" {
			if viewerID == nil {
				continue
			}
			// Must follow author
			if _, follows := r.follows[fmt.Sprintf("%s:%s", *viewerID, post.AuthorID)]; !follows {
				continue
			}
		}

		postCopy := *post
		if author, ok := r.users[post.AuthorID]; ok {
			authorCopy := *author
			postCopy.Author = &authorCopy
		}
		if viewerID != nil {
			_, postCopy.IsLiked = r.likes[fmt.Sprintf("%s:%s", post.ID, *viewerID)]
			_, postCopy.IsReposted = r.reposts[fmt.Sprintf("%s:%s", post.ID, *viewerID)]
			_, postCopy.IsBookmarked = r.bookmarks[fmt.Sprintf("%s:%s", post.ID, *viewerID)]
		}

		candidates = append(candidates, postCopy)
	}

	// Sort descending by CreatedAt, then ID
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].CreatedAt.Equal(candidates[j].CreatedAt) {
			return candidates[i].ID.String() > candidates[j].ID.String()
		}
		return candidates[i].CreatedAt.After(candidates[j].CreatedAt)
	})

	totalNew := len(candidates)
	var returnedPosts []domain.Post
	if totalNew > 0 {
		end := limit
		if end > totalNew {
			end = totalNew
		}
		returnedPosts = candidates[:end]
	}

	var nextCursor string
	if len(returnedPosts) > 0 {
		nextCursor = recommendation.EncodeCursor(returnedPosts[0].CreatedAt, returnedPosts[0].ID)
	}

	return returnedPosts, totalNew, nextCursor, nil
}

func (r *MemoryRepository) GetWhoToFollow(ctx context.Context, viewerID *uuid.UUID, limit int) ([]domain.UserRecommendation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if limit <= 0 || limit > 30 {
		limit = 5
	}

	var recommendations []domain.UserRecommendation
	now := time.Now()

	// Build excluded set
	excluded := make(map[uuid.UUID]bool)
	if viewerID != nil {
		excluded[*viewerID] = true // self
		for key := range r.follows {
			parts := strings.Split(key, ":")
			if len(parts) == 2 && parts[0] == viewerID.String() {
				targetID, _ := uuid.Parse(parts[1])
				excluded[targetID] = true // already followed
			}
		}
		for key := range r.blocks {
			parts := strings.Split(key, ":")
			if len(parts) == 2 {
				if parts[0] == viewerID.String() {
					targetID, _ := uuid.Parse(parts[1])
					excluded[targetID] = true
				} else if parts[1] == viewerID.String() {
					targetID, _ := uuid.Parse(parts[0])
					excluded[targetID] = true
				}
			}
		}
		for key := range r.mutes {
			parts := strings.Split(key, ":")
			if len(parts) == 2 && parts[0] == viewerID.String() {
				targetID, _ := uuid.Parse(parts[1])
				excluded[targetID] = true
			}
		}
		for key, until := range r.dismissedRecommendations {
			parts := strings.Split(key, ":")
			if len(parts) == 2 && parts[0] == viewerID.String() && until.After(now) {
				targetID, _ := uuid.Parse(parts[1])
				excluded[targetID] = true
			}
		}
	}

	// 1. Candidate evaluation with 2nd-degree friends-of-friends calculation
	type candScore struct {
		user        domain.User
		mutualCount int
		mutualNames []string
		score       float64
	}

	var candidates []candScore

	// Calculate followed accounts of viewer
	viewerFollowed := make(map[uuid.UUID]bool)
	if viewerID != nil {
		for key := range r.follows {
			parts := strings.Split(key, ":")
			if len(parts) == 2 && parts[0] == viewerID.String() {
				fID, _ := uuid.Parse(parts[1])
				viewerFollowed[fID] = true
			}
		}
	}

	for uID, user := range r.users {
		if excluded[uID] {
			continue
		}

		mutualCount := 0
		var mutualNames []string

		// Check who follows uID that is in viewerFollowed
		for fID := range viewerFollowed {
			if _, followsTarget := r.follows[fmt.Sprintf("%s:%s", fID, uID)]; followsTarget {
				mutualCount++
				if mutualUser, ok := r.users[fID]; ok && len(mutualNames) < 3 {
					mutualNames = append(mutualNames, mutualUser.Username)
				}
			}
		}

		score := float64(mutualCount)*3.0 + float64(user.FollowersCount)*0.01
		candidates = append(candidates, candScore{
			user:        *user,
			mutualCount: mutualCount,
			mutualNames: mutualNames,
			score:       score,
		})
	}

	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].mutualCount == candidates[j].mutualCount {
			return candidates[i].user.FollowersCount > candidates[j].user.FollowersCount
		}
		return candidates[i].mutualCount > candidates[j].mutualCount
	})

	for _, cand := range candidates {
		if len(recommendations) >= limit {
			break
		}
		reason := "Suggested for you"
		if cand.mutualCount > 0 && len(cand.mutualNames) > 0 {
			if cand.mutualCount == 1 {
				reason = fmt.Sprintf("Followed by @%s", cand.mutualNames[0])
			} else {
				reason = fmt.Sprintf("Followed by @%s + %d others", cand.mutualNames[0], cand.mutualCount-1)
			}
		} else if cand.user.FollowersCount > 50 {
			reason = "Popular on X"
		}

		recommendations = append(recommendations, domain.UserRecommendation{
			User:        cand.user,
			Reason:      reason,
			Score:       cand.score,
			MutualCount: cand.mutualCount,
			MutualNames: cand.mutualNames,
		})
	}

	if recommendations == nil {
		recommendations = []domain.UserRecommendation{}
	}
	return recommendations, nil
}

func (r *MemoryRepository) DismissUserRecommendation(ctx context.Context, viewerID, targetID uuid.UUID, duration time.Duration) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if duration <= 0 {
		duration = 7 * 24 * time.Hour
	}
	key := fmt.Sprintf("%s:%s", viewerID, targetID)
	r.dismissedRecommendations[key] = time.Now().Add(duration)
	return nil
}

// -------------------------------------------------------------
// POST CRUD OWNERSHIP & MODERATION (In-Memory)
// -------------------------------------------------------------

func (r *MemoryRepository) UpdatePost(ctx context.Context, postID, authorID uuid.UUID, content string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	post, exists := r.posts[postID]
	if !exists || post.DeletedAt != nil {
		return ErrNotFound
	}
	if post.AuthorID != authorID {
		return ErrUnauthorized
	}
	post.Content = content
	post.UpdatedAt = time.Now()
	return nil
}

func (r *MemoryRepository) DeletePostWithPermission(ctx context.Context, postID, requesterID uuid.UUID, canDeleteAny bool) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	post, exists := r.posts[postID]
	if !exists || post.DeletedAt != nil {
		return ErrNotFound
	}
	if !canDeleteAny && post.AuthorID != requesterID {
		return ErrUnauthorized
	}
	now := time.Now()
	post.DeletedAt = &now
	return nil
}

// -------------------------------------------------------------
// RBAC ROLES & PERMISSIONS (In-Memory)
// -------------------------------------------------------------

func (r *MemoryRepository) getUserRolesInternal(userID uuid.UUID) []domain.Role {
	roleIDs := r.rbacUserRoles[userID]
	var res []domain.Role
	for _, id := range roleIDs {
		if role, exists := r.rbacRoles[id]; exists && role.IsActive {
			res = append(res, *role)
		}
	}
	if len(res) == 0 {
		// Default to member
		for _, role := range r.rbacRoles {
			if role.Key == "member" {
				res = append(res, *role)
				break
			}
		}
	}
	return res
}

func (r *MemoryRepository) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]domain.Role, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.getUserRolesInternal(userID), nil
}

func (r *MemoryRepository) GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	roles := r.getUserRolesInternal(userID)
	permSet := make(map[string]bool)
	for _, role := range roles {
		for _, pID := range r.rbacRolePerms[role.ID] {
			if perm, exists := r.rbacPermissions[pID]; exists {
				permSet[perm.Key] = true
			}
		}
	}
	var res []string
	for k := range permSet {
		res = append(res, k)
	}
	return res, nil
}

func (r *MemoryRepository) HasPermission(ctx context.Context, userID uuid.UUID, permissionKey string) (bool, error) {
	perms, err := r.GetUserPermissions(ctx, userID)
	if err != nil {
		return false, err
	}
	for _, p := range perms {
		if p == permissionKey {
			return true, nil
		}
	}
	return false, nil
}

func (r *MemoryRepository) AssignRoleToUser(ctx context.Context, userID, roleID uuid.UUID, assignedBy *uuid.UUID, reason string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, id := range r.rbacUserRoles[userID] {
		if id == roleID {
			return nil
		}
	}
	r.rbacUserRoles[userID] = append(r.rbacUserRoles[userID], roleID)
	return nil
}

func (r *MemoryRepository) RemoveRoleFromUser(ctx context.Context, userID, roleID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	var updated []uuid.UUID
	for _, id := range r.rbacUserRoles[userID] {
		if id != roleID {
			updated = append(updated, id)
		}
	}
	r.rbacUserRoles[userID] = updated
	return nil
}

func (r *MemoryRepository) GetAllRoles(ctx context.Context) ([]domain.Role, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var res []domain.Role
	for _, role := range r.rbacRoles {
		copied := *role
		copied.UserCount = 0
		for _, uRoles := range r.rbacUserRoles {
			for _, rID := range uRoles {
				if rID == role.ID {
					copied.UserCount++
					break
				}
			}
		}
		res = append(res, copied)
	}
	return res, nil
}

func (r *MemoryRepository) GetRoleByID(ctx context.Context, roleID uuid.UUID) (*domain.Role, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	role, exists := r.rbacRoles[roleID]
	if !exists {
		return nil, ErrNotFound
	}
	copied := *role
	return &copied, nil
}

func (r *MemoryRepository) CreateRole(ctx context.Context, role *domain.Role) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	role.ID = uuid.New()
	role.CreatedAt = time.Now()
	role.UpdatedAt = time.Now()
	r.rbacRoles[role.ID] = role
	return nil
}

func (r *MemoryRepository) UpdateRole(ctx context.Context, role *domain.Role) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, exists := r.rbacRoles[role.ID]
	if !exists {
		return ErrNotFound
	}
	existing.Name = role.Name
	existing.Description = role.Description
	existing.IsActive = role.IsActive
	existing.UpdatedAt = time.Now()
	return nil
}

func (r *MemoryRepository) DeleteRole(ctx context.Context, roleID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	role, exists := r.rbacRoles[roleID]
	if !exists {
		return ErrNotFound
	}
	if role.IsSystem || role.Key == "owner" || role.Key == "member" {
		return ErrRoleProtected
	}
	delete(r.rbacRoles, roleID)
	delete(r.rbacRolePerms, roleID)
	return nil
}

func (r *MemoryRepository) GetRolePermissions(ctx context.Context, roleID uuid.UUID) ([]domain.Permission, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var perms []domain.Permission
	for _, pid := range r.rbacRolePerms[roleID] {
		if p, exists := r.rbacPermissions[pid]; exists {
			perms = append(perms, *p)
		}
	}
	return perms, nil
}

func (r *MemoryRepository) SetRolePermissions(ctx context.Context, roleID uuid.UUID, permissionIDs []uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.rbacRolePerms[roleID] = permissionIDs
	return nil
}

func (r *MemoryRepository) GetAllPermissions(ctx context.Context) ([]domain.Permission, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var perms []domain.Permission
	for _, p := range r.rbacPermissions {
		perms = append(perms, *p)
	}
	return perms, nil
}

// -------------------------------------------------------------
// MEMBERSHIP PLANS & ENTITLEMENTS (In-Memory)
// -------------------------------------------------------------

func (r *MemoryRepository) GetAllMembershipPlans(ctx context.Context) ([]domain.MembershipPlan, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var res []domain.MembershipPlan
	for _, p := range r.memPlans {
		plan := *p
		plan.Entitlements, _ = r.GetPlanEntitlements(ctx, plan.ID)
		res = append(res, plan)
	}
	return res, nil
}

func (r *MemoryRepository) GetMembershipPlanByID(ctx context.Context, planID string) (*domain.MembershipPlan, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	p, exists := r.memPlans[planID]
	if !exists {
		return nil, ErrNotFound
	}
	copied := *p
	copied.Entitlements, _ = r.GetPlanEntitlements(ctx, planID)
	return &copied, nil
}

func (r *MemoryRepository) UpdateMembershipPlan(ctx context.Context, plan *domain.MembershipPlan) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, exists := r.memPlans[plan.ID]
	if !exists {
		return ErrNotFound
	}
	existing.Name = plan.Name
	existing.Description = plan.Description
	existing.PriceAmountCents = plan.PriceAmountCents
	existing.IsActive = plan.IsActive
	existing.UpdatedAt = time.Now()
	return nil
}

func (r *MemoryRepository) GetPlanEntitlements(ctx context.Context, planID string) ([]domain.MembershipEntitlement, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var res []domain.MembershipEntitlement
	for k, v := range r.memEntitlements[planID] {
		res = append(res, domain.MembershipEntitlement{
			PlanID:         planID,
			EntitlementKey: k,
			LimitValue:     v,
			Unit:           "count",
			UpdatedAt:      time.Now(),
		})
	}
	return res, nil
}

func (r *MemoryRepository) SetPlanEntitlements(ctx context.Context, planID string, entitlements []domain.MembershipEntitlement) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.memEntitlements[planID] == nil {
		r.memEntitlements[planID] = make(map[string]int64)
	}
	for _, e := range entitlements {
		r.memEntitlements[planID][e.EntitlementKey] = e.LimitValue
	}
	return nil
}

func (r *MemoryRepository) GetUserMembership(ctx context.Context, userID uuid.UUID) (*domain.UserMembership, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	um, exists := r.userMemberships[userID]
	if exists && um.Status == "active" {
		return um, nil
	}
	return &domain.UserMembership{
		UserID:   userID,
		PlanID:   "free",
		Status:   "active",
		StartsAt: time.Now(),
	}, nil
}

func (r *MemoryRepository) AssignUserMembership(ctx context.Context, userMem *domain.UserMembership) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	userMem.UpdatedAt = time.Now()
	r.userMemberships[userMem.UserID] = userMem
	return nil
}

func (r *MemoryRepository) RevokeUserMembership(ctx context.Context, userID uuid.UUID, planID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if um, exists := r.userMemberships[userID]; exists && um.PlanID == planID {
		um.Status = "canceled"
		um.UpdatedAt = time.Now()
	}
	return nil
}

func (r *MemoryRepository) GetResolvedEntitlements(ctx context.Context, userID uuid.UUID) (map[string]int64, error) {
	um, _ := r.GetUserMembership(ctx, userID)
	sub, _ := r.GetActiveSubscription(ctx, userID)

	planID := "free"
	if um != nil && um.PlanID != "" {
		planID = um.PlanID
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	res := make(map[string]int64)
	for k, v := range r.memEntitlements[planID] {
		res[k] = v
	}
	if _, ok := res["posts_per_day"]; !ok {
		res["posts_per_day"] = 2
	}
	if _, ok := res["max_images_per_post"]; !ok {
		res["max_images_per_post"] = 2
	}
	if _, ok := res["max_videos_per_post"]; !ok {
		res["max_videos_per_post"] = 0
	}
	if _, ok := res["max_post_length"]; !ok {
		res["max_post_length"] = 280
	}

	// Check legacy active subscription for backward compatibility
	if planID == "free" && sub != nil && sub.Plan != nil {
		if sub.Plan.MaxPostLength > 0 {
			res["max_post_length"] = int64(sub.Plan.MaxPostLength)
		}
		if sub.Plan.MaxMediaCount > 0 {
			res["max_images_per_post"] = int64(sub.Plan.MaxMediaCount)
		}
		res["posts_per_day"] = 30
		if sub.Plan.MaxVideoMinutes > 0 {
			res["max_videos_per_post"] = 1
			res["max_video_minutes"] = int64(sub.Plan.MaxVideoMinutes)
		}
	}

	return res, nil
}

// -------------------------------------------------------------
// ATOMIC QUOTA ENGINE (In-Memory)
// -------------------------------------------------------------

func (r *MemoryRepository) CheckAndIncrementPostQuota(ctx context.Context, userID uuid.UUID, maxPosts int) (bool, int, error) {
	if maxPosts <= 0 {
		return true, 0, nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	periodKey := fmt.Sprintf("%s:post_daily:%s", userID, time.Now().UTC().Format("2006-01-02"))
	current := r.usageCounters[periodKey]
	if current >= maxPosts {
		return false, current, nil
	}
	r.usageCounters[periodKey] = current + 1
	return true, current + 1, nil
}

// -------------------------------------------------------------
// AUDIT LOGS (In-Memory)
// -------------------------------------------------------------

func (r *MemoryRepository) CreateAuditLog(ctx context.Context, log *domain.AuditLog) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	log.ID = uuid.New()
	log.CreatedAt = time.Now()
	r.auditLogs = append(r.auditLogs, *log)
	return nil
}

func (r *MemoryRepository) GetAuditLogs(ctx context.Context, limit, offset int) ([]domain.AuditLog, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var res []domain.AuditLog
	for i := len(r.auditLogs) - 1; i >= 0; i-- {
		res = append(res, r.auditLogs[i])
	}
	if offset > len(res) {
		return []domain.AuditLog{}, nil
	}
	res = res[offset:]
	if limit > 0 && limit < len(res) {
		res = res[:limit]
	}
	return res, nil
}

// -------------------------------------------------------------
// USER ADMINISTRATION & OWNER BOOTSTRAP (In-Memory)
// -------------------------------------------------------------

func (r *MemoryRepository) GetAllUsersAdmin(ctx context.Context, limit, offset int, search string) ([]domain.User, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var filtered []domain.User
	for _, u := range r.users {
		if search != "" && !strings.Contains(strings.ToLower(u.Username), strings.ToLower(search)) && !strings.Contains(strings.ToLower(u.Email), strings.ToLower(search)) {
			continue
		}
		filtered = append(filtered, *u)
	}
	total := len(filtered)
	if offset > len(filtered) {
		return []domain.User{}, total, nil
	}
	filtered = filtered[offset:]
	if limit > 0 && limit < len(filtered) {
		filtered = filtered[:limit]
	}
	return filtered, total, nil
}

func (r *MemoryRepository) UpdateUserStatus(ctx context.Context, userID uuid.UUID, status string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	u, exists := r.users[userID]
	if !exists {
		return ErrNotFound
	}
	if strings.ToLower(strings.TrimSpace(u.Email)) == "afrizaahmad18@gmail.com" {
		return ErrRoleProtected
	}
	if status == "suspended" {
		now := time.Now()
		u.DeletedAt = &now
	} else if status == "active" {
		u.DeletedAt = nil
	}
	return nil
}

func (r *MemoryRepository) BootstrapOwner(ctx context.Context, email string) (*domain.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	norm := strings.ToLower(strings.TrimSpace(email))
	uID, exists := r.userByEmail[norm]
	if !exists {
		return nil, nil
	}
	u := r.users[uID]
	u.Role = "owner"
	return u, nil
}

func (r *MemoryRepository) GetSecurityAuditLogs(ctx context.Context, limit, offset int) ([]domain.AuditLog, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var secLogs []domain.AuditLog
	for i := len(r.auditLogs) - 1; i >= 0; i-- {
		l := r.auditLogs[i]
		action := strings.ToLower(l.Action)
		if strings.Contains(action, "login") || strings.Contains(action, "auth") || strings.Contains(action, "security") ||
			strings.Contains(action, "owner") || strings.Contains(action, "suspend") || strings.Contains(action, "role") ||
			strings.Contains(action, "permission") || strings.Contains(action, "unauthorized") {
			secLogs = append(secLogs, l)
		}
	}

	if offset >= len(secLogs) {
		return []domain.AuditLog{}, nil
	}
	end := offset + limit
	if end > len(secLogs) {
		end = len(secLogs)
	}
	return secLogs[offset:end], nil
}

func (r *MemoryRepository) GetAdminOverview(ctx context.Context) (*domain.AdminOverview, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	totalUsers := int64(len(r.users))
	var totalPosts int64
	var flaggedPosts int64
	for _, p := range r.posts {
		if p.DeletedAt == nil {
			totalPosts++
			if p.ModerationStatus == "flagged" || p.ModerationStatus == "hidden" {
				flaggedPosts++
			}
		}
	}

	overview := &domain.AdminOverview{
		TotalUsers:        totalUsers,
		ActiveUsers24h:    totalUsers,
		NewUsersToday:     totalUsers,
		TotalPosts:        totalPosts,
		PostsToday:        totalPosts,
		TotalMediaCount:   0,
		FlaggedPostsCount: flaggedPosts,
		SystemHealth: domain.SystemHealth{
			Database: domain.DatabaseHealth{
				Status:           "healthy",
				PingMs:           0.5,
				OpenConnections:  5,
				InUseConnections: 1,
				IdleConnections:  4,
			},
			Redis: domain.RedisHealth{
				Status:          "healthy",
				PingMs:          0.3,
				UsedMemoryHuman: "2.5M",
			},
			RabbitMQ: domain.ServiceStatus{
				Status: "healthy",
			},
			Runtime: domain.RuntimeMetrics{
				Goroutines:    10,
				UptimeSeconds: 120,
			},
		},
	}

	for i := len(r.auditLogs) - 1; i >= 0 && len(overview.RecentAuditLogs) < 8; i-- {
		overview.RecentAuditLogs = append(overview.RecentAuditLogs, r.auditLogs[i])
	}
	overview.RecentSecurityEvents = overview.RecentAuditLogs
	return overview, nil
}

func (r *MemoryRepository) GetAdminPosts(ctx context.Context, filter domain.AdminPostFilter) ([]domain.Post, int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var matched []domain.Post
	for _, p := range r.posts {
		if filter.Query != "" && !strings.Contains(strings.ToLower(p.Content), strings.ToLower(filter.Query)) {
			continue
		}
		if filter.AuthorID != "" && p.AuthorID.String() != filter.AuthorID {
			continue
		}
		if filter.ModerationStatus != "" && filter.ModerationStatus != "all" && p.ModerationStatus != filter.ModerationStatus {
			continue
		}
		if filter.Visibility != "" && filter.Visibility != "all" && p.Visibility != filter.Visibility {
			continue
		}

		pCopy := *p
		if u, ok := r.users[p.AuthorID]; ok {
			pCopy.Author = u
		}
		matched = append(matched, pCopy)
	}

	total := int64(len(matched))
	if filter.Offset >= len(matched) {
		return []domain.Post{}, total, nil
	}
	end := filter.Offset + filter.Limit
	if end > len(matched) {
		end = len(matched)
	}
	return matched[filter.Offset:end], total, nil
}

func (r *MemoryRepository) UpdatePostModeration(ctx context.Context, postID uuid.UUID, status string, moderatorID uuid.UUID, reason string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	p, exists := r.posts[postID]
	if !exists {
		return ErrNotFound
	}
	p.ModerationStatus = status
	p.UpdatedAt = time.Now()

	r.auditLogs = append(r.auditLogs, domain.AuditLog{
		ID:          uuid.New(),
		ActorUserID: &moderatorID,
		Action:      "post.moderate",
		TargetType:  "post",
		TargetID:    postID.String(),
		Reason:      reason,
		CreatedAt:   time.Now(),
	})
	return nil
}

func (r *MemoryRepository) AdminDeletePost(ctx context.Context, postID uuid.UUID, moderatorID uuid.UUID, reason string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	p, exists := r.posts[postID]
	if !exists {
		return ErrNotFound
	}
	now := time.Now()
	p.DeletedAt = &now
	p.ModerationStatus = "removed"

	r.auditLogs = append(r.auditLogs, domain.AuditLog{
		ID:          uuid.New(),
		ActorUserID: &moderatorID,
		Action:      "post.delete.admin",
		TargetType:  "post",
		TargetID:    postID.String(),
		Reason:      reason,
		CreatedAt:   time.Now(),
	})
	return nil
}

func (r *MemoryRepository) ReconcileCounters(ctx context.Context, dryRun bool, batchSize int) (*domain.CounterReconciliationResult, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	result := &domain.CounterReconciliationResult{
		DryRun:  dryRun,
		Details: []domain.CounterMismatchDetail{},
	}

	for _, p := range r.posts {
		if p.DeletedAt != nil {
			continue
		}
		result.TotalScanned++

		actualLikes := 0
		for key := range r.likes {
			if strings.HasPrefix(key, p.ID.String()+":") {
				actualLikes++
			}
		}

		actualReposts := 0
		for key := range r.reposts {
			if strings.HasPrefix(key, p.ID.String()+":") {
				actualReposts++
			}
		}

		actualReplies := 0
		for _, other := range r.posts {
			if other.ReplyToID != nil && *other.ReplyToID == p.ID && other.DeletedAt == nil {
				actualReplies++
			}
		}

		actualQuotes := 0
		for _, other := range r.posts {
			if other.QuotePostID != nil && *other.QuotePostID == p.ID && other.DeletedAt == nil {
				actualQuotes++
			}
		}

		hasMismatch := false
		if p.LikesCount != actualLikes {
			result.MismatchesFound++
			hasMismatch = true
			result.Details = append(result.Details, domain.CounterMismatchDetail{
				PostID:        p.ID,
				Type:          "likes",
				ReportedCount: p.LikesCount,
				ActualCount:   actualLikes,
				Fixed:         !dryRun,
			})
			if !dryRun {
				p.LikesCount = actualLikes
			}
		}
		if p.RepostsCount != actualReposts {
			result.MismatchesFound++
			hasMismatch = true
			result.Details = append(result.Details, domain.CounterMismatchDetail{
				PostID:        p.ID,
				Type:          "reposts",
				ReportedCount: p.RepostsCount,
				ActualCount:   actualReposts,
				Fixed:         !dryRun,
			})
			if !dryRun {
				p.RepostsCount = actualReposts
			}
		}
		if p.RepliesCount != actualReplies {
			result.MismatchesFound++
			hasMismatch = true
			result.Details = append(result.Details, domain.CounterMismatchDetail{
				PostID:        p.ID,
				Type:          "replies",
				ReportedCount: p.RepliesCount,
				ActualCount:   actualReplies,
				Fixed:         !dryRun,
			})
			if !dryRun {
				p.RepliesCount = actualReplies
			}
		}

		if hasMismatch && !dryRun {
			result.MismatchesFixed++
		}
	}

	return result, nil
}



