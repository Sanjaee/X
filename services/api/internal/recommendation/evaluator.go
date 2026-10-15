package recommendation

import (
	"time"

	"github.com/google/uuid"
	"github.com/x-reconstruction/api/internal/domain"
)

// FeedMetrics represents computed statistical metrics for offline or online evaluation
type FeedMetrics struct {
	TotalPosts           int     `json:"total_posts"`
	UniqueAuthors        int     `json:"unique_authors"`
	AuthorDiversityRatio float64 `json:"author_diversity_ratio"`
	AverageAgeHours      float64 `json:"average_age_hours"`
	AverageLikes         float64 `json:"average_likes"`
	MediaPostRatio       float64 `json:"media_post_ratio"`
	VerifiedAuthorRatio  float64 `json:"verified_author_ratio"`
	RepetitionViolations int     `json:"repetition_violations"`
}

// EvaluateFeed computes statistical health metrics across a generated timeline slice
func EvaluateFeed(posts []domain.Post, now time.Time) FeedMetrics {
	if len(posts) == 0 {
		return FeedMetrics{}
	}

	authorSet := make(map[uuid.UUID]bool)
	var totalAgeHours float64
	var totalLikes float64
	var mediaCount int
	var verifiedCount int
	var repetitionViolations int

	// Sliding window for checking repetition violations (max 2 per 10 posts)
	windowSize := 10
	for i, p := range posts {
		authorSet[p.AuthorID] = true
		age := now.Sub(p.CreatedAt).Hours()
		if age < 0 {
			age = 0
		}
		totalAgeHours += age
		totalLikes += float64(p.LikesCount)

		if len(p.Media) > 0 {
			mediaCount++
		}
		if p.Author != nil && p.Author.IsVerified {
			verifiedCount++
		}

		// Check window violations
		start := i - windowSize + 1
		if start < 0 {
			start = 0
		}
		authorOccurrences := 0
		for j := start; j <= i; j++ {
			if posts[j].AuthorID == p.AuthorID {
				authorOccurrences++
			}
		}
		if authorOccurrences > 2 {
			repetitionViolations++
		}
	}

	total := float64(len(posts))
	return FeedMetrics{
		TotalPosts:           len(posts),
		UniqueAuthors:        len(authorSet),
		AuthorDiversityRatio: float64(len(authorSet)) / total,
		AverageAgeHours:      totalAgeHours / total,
		AverageLikes:         totalLikes / total,
		MediaPostRatio:       float64(mediaCount) / total,
		VerifiedAuthorRatio:  float64(verifiedCount) / total,
		RepetitionViolations: repetitionViolations,
	}
}
