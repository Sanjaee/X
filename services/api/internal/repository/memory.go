package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/x-reconstruction/api/internal/domain"
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
	notifications map[uuid.UUID]*domain.Notification

	// Subscriptions & Verification (Section 6A)
	plans        map[string]*domain.SubscriptionPlan
	subs         map[uuid.UUID]*domain.Subscription
	orders       map[uuid.UUID]*domain.PaymentOrder
	invoices     map[uuid.UUID]*domain.Invoice
	badges       map[uuid.UUID]*domain.VerificationBadge
	applications map[uuid.UUID]*domain.VerificationApplication
}

func NewMemoryRepository() *MemoryRepository {
	repo := &MemoryRepository{
		users:         make(map[uuid.UUID]*domain.User),
		userByUname:   make(map[string]uuid.UUID),
		userByEmail:   make(map[string]uuid.UUID),
		sessions:      make(map[string]*domain.Session),
		posts:         make(map[uuid.UUID]*domain.Post),
		media:         make(map[uuid.UUID]*domain.MediaObject),
		likes:         make(map[string]time.Time),
		reposts:       make(map[string]time.Time),
		bookmarks:     make(map[string]time.Time),
		follows:       make(map[string]time.Time),
		notifications: make(map[uuid.UUID]*domain.Notification),
		plans:         make(map[string]*domain.SubscriptionPlan),
		subs:          make(map[uuid.UUID]*domain.Subscription),
		orders:        make(map[uuid.UUID]*domain.PaymentOrder),
		invoices:      make(map[uuid.UUID]*domain.Invoice),
		badges:        make(map[uuid.UUID]*domain.VerificationBadge),
		applications:  make(map[uuid.UUID]*domain.VerificationApplication),
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

	r.users[u.ID] = u
	r.userByUname[lowerU] = u.ID
	r.userByEmail[lowerE] = u.ID
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
	r.mu.RLock()
	defer r.mu.RUnlock()

	if limit <= 0 || limit > 50 {
		limit = 20
	}

	var result []domain.Post
	for _, p := range r.posts {
		if p.DeletedAt != nil {
			continue
		}
		if feedType == "following" && viewerID != nil {
			if !r.isFollowingInternal(*viewerID, p.AuthorID) && p.AuthorID != *viewerID {
				continue
			}
		}

		cloned := *p
		if viewerID != nil {
			cloned.IsLiked = r.isLikedInternal(p.ID, *viewerID)
			cloned.IsReposted = r.isRepostedInternal(p.ID, *viewerID)
			cloned.IsBookmarked = r.isBookmarkedInternal(p.ID, *viewerID)
		}
		result = append(result, cloned)
	}

	// Sort newest first
	for i := 0; i < len(result)-1; i++ {
		for j := i + 1; j < len(result); j++ {
			if result[i].CreatedAt.Before(result[j].CreatedAt) {
				result[i], result[j] = result[j], result[i]
			}
		}
	}

	if len(result) > limit {
		result = result[:limit]
	}
	return result, "", nil
}

func (r *MemoryRepository) GetUserPosts(ctx context.Context, username, tab string, viewerID *uuid.UUID, cursor string, limit int) ([]domain.Post, string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	uID, exists := r.userByUname[strings.ToLower(username)]
	if !exists {
		return nil, "", ErrNotFound
	}

	var result []domain.Post
	for _, p := range r.posts {
		if p.AuthorID == uID && p.DeletedAt == nil {
			if tab == "replies" && p.ReplyToID == nil {
				continue
			}
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

func (r *MemoryRepository) GetPostReplies(ctx context.Context, postID uuid.UUID, viewerID *uuid.UUID, cursor string, limit int) ([]domain.Post, string, error) {
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

func (r *MemoryRepository) CreateNotification(ctx context.Context, n *domain.Notification) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if n.ID == uuid.Nil {
		n.ID = uuid.New()
	}
	n.CreatedAt = time.Now()
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

