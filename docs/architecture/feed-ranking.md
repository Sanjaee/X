# Feed Ranking & Scoring Architecture

## 1. Architectural Philosophy
Social media timeline ranking on modern platforms must balance **relevance**, **freshness**, **author diversity**, and **creator discovery**. Relying merely on raw like counts results in stale feeds dominated by celebrity accounts, while purely chronological sorting fails to highlight conversations the user genuinely cares about.

Our ranking engine is implemented as a modular, testable pipeline written in pure Go without unverified external ML blackboxes.

---

## 2. Mathematical Scoring Model

The ranking score $S(u, p)$ for post $p$ given viewing user $u$ is calculated as:

$$S(u, p) = \sum_{i} (w_i \cdot F_i(u, p)) - \sum_{j} (w_j \cdot P_j(u, p))$$

where:
- $F_i(u, p) \in [0.0, 1.0]$ are normalized positive features.
- $w_i \ge 0$ are configurable feature weights.
- $P_j(u, p) \in [0.0, 1.0]$ are normalized penalty signals.
- $w_j \ge 0$ are configurable penalty weights.

---

## 3. Normalized Feature Definitions

All features are normalized into the standard range $[0.0, 1.0]$ prior to weighting:

### Positive Features ($F_i$)
1. **Relationship Affinity (`relationship_score`)**:
   - $1.0$ if mutual follow ($u \leftrightarrow \text{author}$).
   - $0.75$ if viewer follows author ($u \rightarrow \text{author}$).
   - $0.50$ if author follows viewer ($\text{author} \rightarrow u$).
   - $0.25$ if viewer has engaged (liked/replied) with author in the past 7 days.
   - $0.05$ baseline for out-of-network discovery.

2. **Freshness Decay (`freshness_score`)**:
   Computed using an exponential half-life decay model:
   $$\text{freshness\_score} = \exp\left(-\frac{\text{age\_hours}}{\lambda_{half}}\right)$$
   Where default $\lambda_{half} = 12.0\text{ hours}$.
   - Post age 1 hour: $\exp(-1/12) \approx 0.92$
   - Post age 6 hours: $\exp(-6/12) \approx 0.60$
   - Post age 24 hours: $\exp(-24/12) \approx 0.13$
   - Post age 48 hours: $\exp(-48/12) \approx 0.018$

3. **Engagement Velocity (`engagement_velocity_score`)**:
   Measures recent acceleration rather than cumulative all-time metrics:
   $$\text{velocity} = \frac{1.0 \times \text{likes} + 2.5 \times \text{reposts} + 3.0 \times \text{replies}}{\max(0.5, \text{age\_hours})^{1.1}}$$
   $$\text{normalized\_velocity} = \frac{\text{velocity}}{\text{velocity} + 50.0}$$

4. **Topic Relevance (`topic_relevance_score`)**:
   Cosine/Jaccard similarity between post hashtags/keywords and the viewer's active topic interest vector $\vec{T}_u$:
   $$\text{topic\_score} = \sum_{t \in p.\text{topics}} \text{weight}(u, t)$$

5. **Quality Score (`quality_score`)**:
   Signals content richness and author authenticity:
   - Contains high-resolution media attachment: $+0.25$
   - Author is verified: $+0.20$
   - Optimal content length ($> 50$ chars, not spam text): $+0.15$
   - Low report ratio: $+0.40$

6. **Exploration & Novelty (`exploration_score`)**:
   Gives a calibrated boost to fresh content from accounts with low follower counts ($< 500$) to guarantee new creator visibility.

---

### Negative Penalties ($P_j$)
1. **Negative Feedback Penalty (`negative_feedback_penalty`)**:
   - $1.0$ if viewer clicked "Not interested" or "Hide post" on this specific post or topic.
   - $0.6$ if author was recently reported.

2. **Repetition Penalty (`repetition_penalty`)**:
   Calculated dynamically during the post-ranking reranking pass:
   - Increments for every post from the same author already selected in the current feed page.

3. **Spam & Excess Engagement Penalty (`spam_penalty`)**:
   Penalizes repetitive hashtags, excessive external link density, or accounts showing bot-like activity.

---

## 4. Centralized Configuration Structure

All weights are centralized in Go config structs (`FeedConfig`), versioned, and swappable at runtime without recompiling SQL or backend services:

```go
type FeedWeights struct {
    RelationshipWeight    float64 `json:"relationship_weight"`     // default: 2.5
    FreshnessWeight       float64 `json:"freshness_weight"`        // default: 3.0
    EngagementWeight      float64 `json:"engagement_weight"`       // default: 2.0
    TopicWeight           float64 `json:"topic_weight"`            // default: 1.8
    QualityWeight         float64 `json:"quality_weight"`          // default: 1.2
    ExplorationWeight     float64 `json:"exploration_weight"`      // default: 1.0
    NegativePenaltyWeight float64 `json:"negative_penalty_weight"` // default: 5.0
    RepetitionWeight      float64 `json:"repetition_weight"`       // default: 3.0
}
```

---

## 5. Diversity Reranking (Sliding Window Algorithm)

After calculating preliminary scores, the top $K$ candidates undergo a diversity reranker:
1. **Author Capping**: Maximum 2 posts per author within any sliding window of 10 posts.
2. **Topic Interleaving**: No more than 3 consecutive posts sharing the exact same primary topic.
3. **Composition Ratio**: Target mix in For You feed:
   - $50\%$ Following Graph
   - $30\%$ High-Relevance Out-of-Network (Topic & Social Engagement)
   - $10\%$ Trending Velocity
   - $10\%$ Cold-Start / Exploration
