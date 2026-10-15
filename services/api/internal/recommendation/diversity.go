package recommendation

import (
	"sort"

	"github.com/google/uuid"
)

// DiversityReranker applies author capping, topic interleaving, and anti-repetition rules
type DiversityReranker struct {
	MaxAuthorPerWindow int
	WindowSize         int
}

// NewDiversityReranker initializes default production diversity constraints
func NewDiversityReranker() *DiversityReranker {
	return &DiversityReranker{
		MaxAuthorPerWindow: 2,
		WindowSize:         10,
	}
}

// Rerank reorders scored candidates to enforce diversity and anti-repetition
func (dr *DiversityReranker) Rerank(candidates []Candidate, limit int) []Candidate {
	if len(candidates) <= 1 {
		return candidates
	}

	// 1. Sort by initial FinalScore descending
	sorted := make([]Candidate, len(candidates))
	copy(sorted, candidates)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].FinalScore > sorted[j].FinalScore
	})

	result := make([]Candidate, 0, limit)
	authorCountsInWindow := make(map[uuid.UUID]int)
	deferred := make([]Candidate, 0, len(sorted))

	for _, cand := range sorted {
		if len(result) >= limit {
			break
		}

		authorID := cand.Post.AuthorID
		currentCount := authorCountsInWindow[authorID]

		if currentCount >= dr.MaxAuthorPerWindow {
			// Author quota exceeded in current window, defer candidate
			cand.Features.RepetitionPenalty = 0.5
			deferred = append(deferred, cand)
			continue
		}

		result = append(result, cand)
		authorCountsInWindow[authorID]++

		// Slide window if window size is reached
		if len(result) > dr.WindowSize {
			oldestAuthor := result[len(result)-dr.WindowSize-1].Post.AuthorID
			if authorCountsInWindow[oldestAuthor] > 0 {
				authorCountsInWindow[oldestAuthor]--
			}
		}
	}

	// Fill remaining slots from deferred pool if limit not reached
	if len(result) < limit && len(deferred) > 0 {
		for _, cand := range deferred {
			if len(result) >= limit {
				break
			}
			result = append(result, cand)
		}
	}

	return result
}
