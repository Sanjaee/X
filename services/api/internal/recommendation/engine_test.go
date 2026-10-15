package recommendation

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/x-reconstruction/api/internal/domain"
)

// MockCandidateProvider implements CandidateProvider for hermetic unit testing
type MockCandidateProvider struct {
	FollowingPosts []domain.Post
	TrendingPosts  []domain.Post
	FreshPosts     []domain.Post
	Weights        domain.RankingWeights
}

func (m *MockCandidateProvider) GetFollowingCandidates(ctx context.Context, viewerID uuid.UUID, limit int) ([]domain.Post, error) {
	return m.FollowingPosts, nil
}
func (m *MockCandidateProvider) GetSocialEngagementCandidates(ctx context.Context, viewerID uuid.UUID, limit int) ([]domain.Post, error) {
	return nil, nil
}
func (m *MockCandidateProvider) GetTopicCandidates(ctx context.Context, viewerID *uuid.UUID, topics []string, limit int) ([]domain.Post, error) {
	return nil, nil
}
func (m *MockCandidateProvider) GetTrendingCandidates(ctx context.Context, limit int) ([]domain.Post, error) {
	return m.TrendingPosts, nil
}
func (m *MockCandidateProvider) GetFreshCandidates(ctx context.Context, limit int) ([]domain.Post, error) {
	return m.FreshPosts, nil
}
func (m *MockCandidateProvider) GetExplorationCandidates(ctx context.Context, limit int) ([]domain.Post, error) {
	return nil, nil
}
func (m *MockCandidateProvider) GetFollowingPosts(ctx context.Context, viewerID uuid.UUID, cursor *FeedCursor, limit int) ([]domain.Post, string, error) {
	return m.FollowingPosts, "", nil
}
func (m *MockCandidateProvider) GetViewerInteractions(ctx context.Context, viewerID uuid.UUID) (
	map[uuid.UUID]bool, map[uuid.UUID]bool, map[uuid.UUID]bool, map[uuid.UUID]bool, map[uuid.UUID]bool, map[uuid.UUID]bool, map[string]float64, error,
) {
	return nil, nil, nil, nil, nil, nil, nil, nil
}
func (m *MockCandidateProvider) GetActiveWeights(ctx context.Context) (domain.RankingWeights, string, error) {
	return m.Weights, "mock_v1", nil
}

func TestFreshnessDecay(t *testing.T) {
	now := time.Now().UTC()
	halfLife := 12.0

	score0h := CalculateFreshness(now, now, halfLife)
	score6h := CalculateFreshness(now.Add(-6*time.Hour), now, halfLife)
	score12h := CalculateFreshness(now.Add(-12*time.Hour), now, halfLife)
	score24h := CalculateFreshness(now.Add(-24*time.Hour), now, halfLife)

	if score0h < 0.99 {
		t.Errorf("expected fresh score close to 1.0, got %f", score0h)
	}
	if score6h <= score12h || score12h <= score24h {
		t.Errorf("freshness decay must be strictly monotonic: 6h=%f, 12h=%f, 24h=%f", score6h, score12h, score24h)
	}
	// At half life (12h), exp(-1) is approximately 0.3678
	if score12h < 0.35 || score12h > 0.38 {
		t.Errorf("expected score at 12h half life ~0.368, got %f", score12h)
	}
}

func TestDiversityReranking(t *testing.T) {
	reranker := NewDiversityReranker()
	authorA := uuid.New()
	authorB := uuid.New()
	now := time.Now().UTC()

	var candidates []Candidate
	// Author A has 5 high scoring posts
	for i := 0; i < 5; i++ {
		candidates = append(candidates, Candidate{
			Post: domain.Post{
				ID:        uuid.New(),
				AuthorID:  authorA,
				CreatedAt: now,
			},
			FinalScore: 10.0 - float64(i),
		})
	}
	// Author B has 5 lower scoring posts
	for i := 0; i < 5; i++ {
		candidates = append(candidates, Candidate{
			Post: domain.Post{
				ID:        uuid.New(),
				AuthorID:  authorB,
				CreatedAt: now,
			},
			FinalScore: 4.0 - float64(i),
		})
	}

	reranked := reranker.Rerank(candidates, 10)

	// Author A should not have more than 2 posts in the first 2 slots
	authorACountInFirst3 := 0
	for i := 0; i < 3 && i < len(reranked); i++ {
		if reranked[i].Post.AuthorID == authorA {
			authorACountInFirst3++
		}
	}

	if authorACountInFirst3 > 2 {
		t.Errorf("author A exceeded max window cap of 2, found %d", authorACountInFirst3)
	}
}

func TestCursorRoundtrip(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Microsecond)
	pID := uuid.New()

	encoded := EncodeCursor(now, pID)
	if encoded == "" {
		t.Fatal("encoded cursor should not be empty")
	}

	decoded, err := DecodeCursor(encoded)
	if err != nil {
		t.Fatalf("failed to decode cursor: %v", err)
	}

	if decoded.PostID != pID {
		t.Errorf("expected postID %s, got %s", pID, decoded.PostID)
	}
	if !decoded.CreatedAt.Equal(now) {
		t.Errorf("expected createdAt %v, got %v", now, decoded.CreatedAt)
	}
}

func TestFeedEvaluator(t *testing.T) {
	now := time.Now().UTC()
	authorA := uuid.New()
	authorB := uuid.New()

	posts := []domain.Post{
		{ID: uuid.New(), AuthorID: authorA, CreatedAt: now.Add(-1 * time.Hour), LikesCount: 100},
		{ID: uuid.New(), AuthorID: authorB, CreatedAt: now.Add(-2 * time.Hour), LikesCount: 50},
	}

	metrics := EvaluateFeed(posts, now)
	if metrics.TotalPosts != 2 {
		t.Errorf("expected 2 total posts, got %d", metrics.TotalPosts)
	}
	if metrics.UniqueAuthors != 2 {
		t.Errorf("expected 2 unique authors, got %d", metrics.UniqueAuthors)
	}
	if metrics.AuthorDiversityRatio != 1.0 {
		t.Errorf("expected diversity ratio 1.0, got %f", metrics.AuthorDiversityRatio)
	}
}

func BenchmarkScoringEngine(b *testing.B) {
	weights := domain.DefaultRankingWeights()
	features := Features{
		RelationshipScore:       0.75,
		FreshnessScore:          0.85,
		EngagementVelocityScore: 0.60,
		TopicRelevanceScore:     0.50,
		QualityScore:            0.70,
		ExplorationScore:        0.20,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = CalculateLinearScore(weights, features)
	}
}

func BenchmarkDiversityReranker(b *testing.B) {
	reranker := NewDiversityReranker()
	now := time.Now().UTC()
	var candidates []Candidate

	for i := 0; i < 100; i++ {
		candidates = append(candidates, Candidate{
			Post: domain.Post{
				ID:        uuid.New(),
				AuthorID:  uuid.New(),
				CreatedAt: now,
			},
			FinalScore: float64(100 - i),
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = reranker.Rerank(candidates, 20)
	}
}
