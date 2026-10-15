package recommendation

import (
	"math"
	"time"

	"github.com/x-reconstruction/api/internal/domain"
)

// CalculateFreshness computes exponential decay score based on half-life in hours
func CalculateFreshness(createdAt time.Time, now time.Time, halfLifeHours float64) float64 {
	if halfLifeHours <= 0 {
		halfLifeHours = 12.0
	}

	ageHours := now.Sub(createdAt).Hours()
	if ageHours < 0 {
		ageHours = 0
	}

	score := math.Exp(-ageHours / halfLifeHours)
	return math.Max(0.0, math.Min(1.0, score))
}

// CalculateEngagementVelocity computes normalized velocity considering time-decayed interactions
func CalculateEngagementVelocity(likes, reposts, replies int, ageHours float64) float64 {
	if ageHours < 0.5 {
		ageHours = 0.5
	}

	rawVelocity := (float64(likes)*1.0 + float64(reposts)*2.5 + float64(replies)*3.0) / math.Pow(ageHours, 1.1)

	// Sigmoidal/rational normalization: maps [0, inf) to [0.0, 1.0]
	normalized := rawVelocity / (rawVelocity + 50.0)
	return math.Max(0.0, math.Min(1.0, normalized))
}

// CalculateQualityScore evaluates content depth, attachments, and authenticity
func CalculateQualityScore(hasMedia bool, isVerified bool, contentLength int) float64 {
	score := 0.2 // baseline

	if hasMedia {
		score += 0.3
	}
	if isVerified {
		score += 0.2
	}
	if contentLength >= 60 && contentLength <= 280 {
		score += 0.2
	} else if contentLength > 280 {
		score += 0.3 // Long-form premium content
	}

	return math.Max(0.0, math.Min(1.0, score))
}

// CalculateExplorationScore boosts content from emerging creators
func CalculateExplorationScore(authorFollowers int, ageHours float64) float64 {
	if ageHours > 48.0 {
		return 0.0
	}

	// High boost if author has fewer than 500 followers
	if authorFollowers < 500 {
		return 0.8
	} else if authorFollowers < 2000 {
		return 0.4
	}
	return 0.1
}

// CalculateLinearScore applies the centralized weights across normalized features and penalties
func CalculateLinearScore(weights domain.RankingWeights, f Features) float64 {
	positiveScore := (weights.RelationshipWeight * f.RelationshipScore) +
		(weights.FreshnessWeight * f.FreshnessScore) +
		(weights.EngagementWeight * f.EngagementVelocityScore) +
		(weights.TopicWeight * f.TopicRelevanceScore) +
		(weights.QualityWeight * f.QualityScore) +
		(weights.ExplorationWeight * f.ExplorationScore)

	penalties := (weights.NegativePenaltyWeight * f.NegativeFeedbackPenalty) +
		(weights.NegativePenaltyWeight * f.SpamPenalty) +
		(weights.RepetitionWeight * f.RepetitionPenalty)

	finalScore := positiveScore - penalties
	if finalScore < 0 {
		finalScore = 0
	}
	return finalScore
}
