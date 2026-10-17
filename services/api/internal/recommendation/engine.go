package recommendation

import (
	"context"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/x-reconstruction/api/internal/domain"
)

// CandidateProvider defines the storage contract for candidate generation and relational lookups
type CandidateProvider interface {
	GetFollowingCandidates(ctx context.Context, viewerID uuid.UUID, limit int) ([]domain.Post, error)
	GetSocialEngagementCandidates(ctx context.Context, viewerID uuid.UUID, limit int) ([]domain.Post, error)
	GetTopicCandidates(ctx context.Context, viewerID *uuid.UUID, topics []string, limit int) ([]domain.Post, error)
	GetTrendingCandidates(ctx context.Context, limit int) ([]domain.Post, error)
	GetFreshCandidates(ctx context.Context, limit int) ([]domain.Post, error)
	GetExplorationCandidates(ctx context.Context, limit int) ([]domain.Post, error)
	GetFollowingPosts(ctx context.Context, viewerID uuid.UUID, cursor *FeedCursor, limit int) ([]domain.Post, string, error)
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
}

// Engine encapsulates the timeline assembly and ranking pipeline
type Engine struct {
	provider CandidateProvider
	reranker *DiversityReranker
}

// NewEngine instantiates a recommendation engine with default configurations
func NewEngine(provider CandidateProvider) *Engine {
	return &Engine{
		provider: provider,
		reranker: NewDiversityReranker(),
	}
}

// GenerateFeed orchestrates the end-to-end multi-stage pipeline
func (e *Engine) GenerateFeed(ctx context.Context, opts FeedOptions) (*FeedResult, error) {
	start := time.Now()
	if opts.Limit <= 0 || opts.Limit > 50 {
		opts.Limit = 20
	}

	// Route 1: Pure Chronological Following Feed
	if opts.FeedType == "following" {
		if opts.ViewerID == nil {
			return &FeedResult{
				Posts:          []domain.Post{},
				ElapsedMs:      time.Since(start).Milliseconds(),
				WeightsVersion: "following_unauthenticated_empty",
			}, nil
		}

		decodedCursor, _ := DecodeCursor(opts.Cursor)
		posts, nextCursor, err := e.provider.GetFollowingPosts(ctx, *opts.ViewerID, decodedCursor, opts.Limit)
		if err != nil {
			return nil, err
		}
		return &FeedResult{
			Posts:          posts,
			NextCursor:     nextCursor,
			HasMore:        nextCursor != "",
			CandidateCount: len(posts),
			ElapsedMs:      time.Since(start).Milliseconds(),
			WeightsVersion: "chronological_following_v1",
		}, nil
	}

	// Route 2 & 3: For You & Explore Feeds
	weights, weightsVersion, err := e.provider.GetActiveWeights(ctx)
	if err != nil {
		weights = domain.DefaultRankingWeights()
		weightsVersion = "default_fallback"
	}

	var followedAuthors, mutualAuthors, likedPosts, mutedAuthors, blockedAuthors, hiddenPosts map[uuid.UUID]bool
	var userTopics map[string]float64

	if opts.ViewerID != nil {
		followedAuthors, mutualAuthors, likedPosts, mutedAuthors, blockedAuthors, hiddenPosts, userTopics, _ =
			e.provider.GetViewerInteractions(ctx, *opts.ViewerID)
	}

	// -------------------------------------------------------------
	// Phase 1: Multi-Source Candidate Retrieval
	// -------------------------------------------------------------
	candidateMap := make(map[uuid.UUID]*Candidate)

	addCandidates := func(posts []domain.Post, sourceFlag CandidateSource) {
		for _, p := range posts {
			if existing, ok := candidateMap[p.ID]; ok {
				existing.Sources |= sourceFlag
			} else {
				postCopy := p
				candidateMap[p.ID] = &Candidate{
					Post:    postCopy,
					Sources: sourceFlag,
				}
			}
		}
	}

	// Explore Feed uses trending, fresh, and topic discovery
	if opts.FeedType == "explore" {
		trending, _ := e.provider.GetTrendingCandidates(ctx, 40)
		addCandidates(trending, SourceTrending)

		fresh, _ := e.provider.GetFreshCandidates(ctx, 30)
		addCandidates(fresh, SourceFresh)

		if opts.Topic != "" {
			topicPosts, _ := e.provider.GetTopicCandidates(ctx, opts.ViewerID, []string{opts.Topic}, 40)
			addCandidates(topicPosts, SourceTopicSimilarity)
		}
	} else {
		// For You Feed: Mix all sources
		if opts.ViewerID != nil {
			followingPosts, _ := e.provider.GetFollowingCandidates(ctx, *opts.ViewerID, 80)
			addCandidates(followingPosts, SourceFollowing)

			socialPosts, _ := e.provider.GetSocialEngagementCandidates(ctx, *opts.ViewerID, 40)
			addCandidates(socialPosts, SourceSocialEngagement)

			var topics []string
			for t := range userTopics {
				topics = append(topics, t)
			}
			if len(topics) > 0 {
				topicPosts, _ := e.provider.GetTopicCandidates(ctx, opts.ViewerID, topics, 30)
				addCandidates(topicPosts, SourceTopicSimilarity)
			}
		}

		trending, _ := e.provider.GetTrendingCandidates(ctx, 30)
		addCandidates(trending, SourceTrending)

		fresh, _ := e.provider.GetFreshCandidates(ctx, 25)
		addCandidates(fresh, SourceFresh)

		explore, _ := e.provider.GetExplorationCandidates(ctx, 15)
		addCandidates(explore, SourceExploration)
	}

	// -------------------------------------------------------------
	// Phase 2: Eligibility & Safety Filtering
	// -------------------------------------------------------------
	eligible := make([]Candidate, 0, len(candidateMap))
	now := time.Now().UTC()

	// Parse cursor if provided for pagination filtering
	cursorObj, _ := DecodeCursor(opts.Cursor)

	for _, cand := range candidateMap {
		p := cand.Post

		// Filter A: Soft-deleted posts
		if p.DeletedAt != nil {
			continue
		}

		// Filter B: Cursor keyset check (must be older than cursor if paginating)
		if cursorObj != nil {
			if !p.CreatedAt.Before(cursorObj.CreatedAt) && !(p.CreatedAt.Equal(cursorObj.CreatedAt) && p.ID.String() < cursorObj.PostID.String()) {
				continue
			}
		}

		// Filter C: Bilateral blocks
		if blockedAuthors != nil && blockedAuthors[p.AuthorID] {
			continue
		}

		// Filter D: Muted authors
		if mutedAuthors != nil && mutedAuthors[p.AuthorID] {
			continue
		}

		// Filter E: Hidden or reported posts
		if hiddenPosts != nil && hiddenPosts[p.ID] {
			continue
		}

		// Filter F: Private account visibility
		if p.Author != nil && p.Author.IsPrivate {
			isViewer := opts.ViewerID != nil && *opts.ViewerID == p.AuthorID
			isFollowing := followedAuthors != nil && followedAuthors[p.AuthorID]
			if !isViewer && !isFollowing {
				continue // Do not leak private account posts into out-of-network feed
			}
		}

		// -------------------------------------------------------------
		// Phase 3: Feature Extraction & Linear Scoring
		// -------------------------------------------------------------
		f := Features{}

		// 1. Relationship Affinity
		if mutualAuthors != nil && mutualAuthors[p.AuthorID] {
			f.RelationshipScore = 1.0
		} else if followedAuthors != nil && followedAuthors[p.AuthorID] {
			f.RelationshipScore = 0.75
		} else if cand.Sources.Has(SourceSocialEngagement) {
			f.RelationshipScore = 0.50
		} else {
			f.RelationshipScore = 0.10 // discovery baseline
		}

		// 2. Freshness Decay
		f.FreshnessScore = CalculateFreshness(p.CreatedAt, now, weights.FreshnessHalfLifeHours)

		// 3. Engagement Velocity
		ageHours := now.Sub(p.CreatedAt).Hours()
		f.EngagementVelocityScore = CalculateEngagementVelocity(p.LikesCount, p.RepostsCount, p.RepliesCount, ageHours)

		// 4. Content Quality
		hasMedia := len(p.Media) > 0
		isVerified := p.Author != nil && p.Author.IsVerified
		f.QualityScore = CalculateQualityScore(hasMedia, isVerified, len(p.Content))

		// 5. Emerging Creator Exploration Boost
		followers := 1000
		if p.Author != nil {
			followers = p.Author.FollowersCount
		}
		f.ExplorationScore = CalculateExplorationScore(followers, ageHours)

		// 6. Topic Relevance & Interest Affinity
		if userTopics != nil && len(userTopics) > 0 {
			lowerContent := strings.ToLower(p.Content)
			for topic, aff := range userTopics {
				cleanTopic := strings.ToLower(strings.TrimPrefix(topic, "#"))
				if cleanTopic != "" && strings.Contains(lowerContent, cleanTopic) {
					if aff > f.TopicRelevanceScore {
						f.TopicRelevanceScore = math.Min(1.0, aff)
					}
				}
			}
		}

		// Compute final score
		cand.Features = f
		cand.FinalScore = CalculateLinearScore(weights, f)

		eligible = append(eligible, *cand)
	}

	// -------------------------------------------------------------
	// Phase 4: Diversity Reranker & Anti-Repetition
	// -------------------------------------------------------------
	reranked := e.reranker.Rerank(eligible, opts.Limit+1)

	// Determine pagination
	hasMore := len(reranked) > opts.Limit
	var nextCursor string
	finalPosts := make([]domain.Post, 0, opts.Limit)

	limitCount := opts.Limit
	if len(reranked) < limitCount {
		limitCount = len(reranked)
	}

	for i := 0; i < limitCount; i++ {
		p := reranked[i].Post
		// Annotate viewer context if available
		if likedPosts != nil && likedPosts[p.ID] {
			p.IsLiked = true
		}
		finalPosts = append(finalPosts, p)
	}

	if hasMore && len(finalPosts) > 0 {
		lastPost := finalPosts[len(finalPosts)-1]
		nextCursor = EncodeCursor(lastPost.CreatedAt, lastPost.ID)
	}

	return &FeedResult{
		Posts:          finalPosts,
		NextCursor:     nextCursor,
		HasMore:        hasMore,
		CandidateCount: len(candidateMap),
		ElapsedMs:      time.Since(start).Milliseconds(),
		WeightsVersion: weightsVersion,
	}, nil
}
