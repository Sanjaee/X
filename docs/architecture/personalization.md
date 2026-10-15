# Personalization Engine Architecture

## 1. Overview
The Personalization Engine builds and maintains a dynamic interest profile for each user, allowing the timeline to adapt gracefully to evolving tastes without overreacting to isolated clicks or random browsing anomalies.

---

## 2. Dual-Horizon Profile Model

User interests are split into two complementary temporal horizons:

```mermaid
graph TD
    Interactions[User Actions: Likes, Reposts, Replies, Bookmarks, Views] --> Buffer[Event Stream]
    Buffer --> ShortTerm[Short-Term Horizon\n(Half-life: 24 Hours)\nHigh Agility, Instant Adaptation]
    Buffer --> LongTerm[Long-Term Horizon\n(Half-life: 14 Days)\nCore Interests, Stability Anchor]

    ShortTerm --> BlendedProfile[Blended User Interest Vector]
    LongTerm --> BlendedProfile
    Explicit[Explicit Topic Preferences\n(User Settings)] --> BlendedProfile
```

### Short-Term Interests ($I_{short}$)
- **Time Horizon**: Rolling 24–48 hours.
- **Objective**: Capture immediate topical interests (e.g. following a live breaking tech announcement or sporting event).
- **Decay**: Rapid exponential decay ($\lambda = 24\text{ hours}$).

### Long-Term Interests ($I_{long}$)
- **Time Horizon**: 30 days.
- **Objective**: Anchor the user's persistent preferences (e.g., software engineering, web development, photography).
- **Decay**: Slow decay ($\lambda = 14\text{ days}$).

### Blended Affinity Score
For any topic $T$:
$$\text{Affinity}(u, T) = 0.6 \cdot I_{short}(u, T) + 0.3 \cdot I_{long}(u, T) + 0.5 \cdot \text{ExplicitFollow}(u, T)$$

---

## 3. Explicit Preferences vs. Inferred Signals

1. **Explicit Preferences**:
   - Stored in `user_topic_preferences`.
   - Modifiable directly by the user in settings (e.g. "Follow topic: #golang", "Mute topic: #crypto").
   - Explicit mutes act as hard negative filters (drop matching candidates immediately).

2. **Inferred Signals**:
   - Calculated from user engagements:
     - Like: $+1.0$ point to associated hashtags/topics.
     - Repost: $+2.5$ points.
     - Reply: $+3.0$ points.
     - Bookmark: $+2.0$ points.
     - Post Click / Detail View: $+0.5$ points.
     - Hide / "Not Interested": $-4.0$ penalty points.

---

## 4. Privacy, Control & History Reset

1. **History Reset API**:
   Users can reset their recommendation profile at any time via `POST /api/v1/me/personalization/reset`. This flushes cached interaction embeddings in Redis and resets inferred topic weights, returning the user to the default onboarding discovery state.

2. **Sensitive Attribute Protection**:
   The engine strictly restricts topic extraction to publicly declared hashtags, user-curated categories, and keyword tokens. The system never infers or stores sensitive demographics (health, political leanings, religion, or orientation) in personalization models.
