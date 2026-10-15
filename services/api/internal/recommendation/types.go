package recommendation

import (
	"time"

	"github.com/google/uuid"
	"github.com/x-reconstruction/api/internal/domain"
)

// CandidateSource represents the origin generator of a candidate
type CandidateSource uint8

const (
	SourceFollowing        CandidateSource = 1 << 0
	SourceSocialEngagement CandidateSource = 1 << 1
	SourceTopicSimilarity  CandidateSource = 1 << 2
	SourceTrending         CandidateSource = 1 << 3
	SourceFresh            CandidateSource = 1 << 4
	SourceExploration      CandidateSource = 1 << 5
)

func (s CandidateSource) Has(flag CandidateSource) bool {
	return s&flag != 0
}

// Features holds normalized values in [0.0, 1.0] for scoring
type Features struct {
	RelationshipScore       float64 `json:"relationship_score"`
	FreshnessScore          float64 `json:"freshness_score"`
	EngagementVelocityScore float64 `json:"engagement_velocity_score"`
	TopicRelevanceScore     float64 `json:"topic_relevance_score"`
	QualityScore            float64 `json:"quality_score"`
	ExplorationScore        float64 `json:"exploration_score"`

	// Penalties (subtracted)
	NegativeFeedbackPenalty float64 `json:"negative_feedback_penalty"`
	SpamPenalty             float64 `json:"spam_penalty"`
	RepetitionPenalty       float64 `json:"repetition_penalty"`
}

// Candidate represents a candidate post undergoing processing in the ranking pipeline
type Candidate struct {
	Post       domain.Post
	Sources    CandidateSource
	Features   Features
	FinalScore float64
}

// FeedOptions contains query options for timeline generation
type FeedOptions struct {
	ViewerID *uuid.UUID
	FeedType string // "for_you", "following", "explore"
	Cursor   string
	Limit    int
	Topic    string // optional filter for Explore
}

// FeedResult represents the assembled and paginated feed response
type FeedResult struct {
	Posts          []domain.Post  `json:"posts"`
	NextCursor     string         `json:"next_cursor,omitempty"`
	HasMore        bool           `json:"has_more"`
	CandidateCount int            `json:"candidate_count"`
	ElapsedMs      int64          `json:"elapsed_ms"`
	WeightsVersion string         `json:"weights_version,omitempty"`
}

// FeedCursor encodes composite cursor pagination metadata
type FeedCursor struct {
	CreatedAt time.Time `json:"created_at"`
	PostID    uuid.UUID `json:"post_id"`
}
