# Architecture: Search Ranking Algorithms

## 1. Multi-Tab Ranking Philosophy
Search ranking in this architecture avoids "one size fits all" algorithms. Each search tab serves distinct user intents:

```
[User Intent]
   |
   +--> "Show me the best, most authoritative discussion" ---> [Top Tab]
   +--> "Show me what happened seconds ago"               ---> [Latest Tab]
   +--> "Help me find this creator or friend"            ---> [People Tab]
   +--> "Show me visual evidence or photo/video"          ---> [Media Tab]
```

---

## 2. Top Tab Ranking Formula `[PROPOSED]`

For the **Top** tab, posts are scored using a linear combination of normalized signal values:

$$\text{FinalScore} = W_{\text{text}} \cdot S_{\text{text}} + W_{\text{eng}} \cdot S_{\text{eng}} + W_{\text{fresh}} \cdot S_{\text{fresh}} + W_{\text{author}} \cdot S_{\text{author}} - W_{\text{penalty}} \cdot S_{\text{penalty}}$$

### Component Formulas:
1. **Text Relevance ($S_{\text{text}}$)**:
   Calculated using PostgreSQL Cover Density ranking:
   $$S_{\text{text}} = \text{ts\_rank\_cd}(p.\text{search\_vector}, \text{query}, 32)$$
   Normalized using a logarithmic scale: $\ln(1 + S_{\text{text}})$.

2. **Engagement Quality ($S_{\text{eng}}$)**:
   Log-scaled to prevent viral runaway posts from completely suppressing quality niche content:
   $$S_{\text{eng}} = 1.5 \cdot \ln(1 + \text{likes}) + 2.0 \cdot \ln(1 + \text{reposts}) + 1.0 \cdot \ln(1 + \text{replies}) + 0.8 \cdot \ln(1 + \text{bookmarks})$$

3. **Freshness Recency Decay ($S_{\text{fresh}}$)**:
   Exponential half-life decay function (half-life $\lambda = 24\text{ hours} = 86,400\text{ seconds}$):
   $$S_{\text{fresh}} = \exp\left(-\frac{\Delta t}{86400}\right)$$

4. **Author Authority ($S_{\text{author}}$)**:
   - Verified creator bonus: $+2.0$ (Government/Business: $+2.5$, Blue: $+1.5$).
   - Follower count signal: $\ln(1 + \text{followers\_count}) \times 0.2$.

5. **Spam & Negative Penalties ($S_{\text{penalty}}$)**:
   - Reported post or negative feedback: $-5.0$.
   - Excessive hashtag stuffing (>4 hashtags): $-2.0$.

---

## 3. Latest Tab Ranking

The **Latest** tab adheres strictly to reverse-chronological order:
- **Order clause**: `ORDER BY p.created_at DESC, p.id DESC`
- **Keyset cursor**: `(p.created_at < $cursor_time OR (p.created_at = $cursor_time AND p.id < $cursor_id))`
- **No popularity boost**: Guarantees zero bias toward high-engagement posts.

---

## 4. People Tab Ranking Formula

Accounts matching the query are ranked by trigram similarity and authority:

$$\text{UserScore} = 5.0 \cdot \text{ExactMatch} + 3.0 \cdot \text{sim}(\text{username}, q) + 2.0 \cdot \text{sim}(\text{display\_name}, q) + 0.5 \cdot \ln(1 + \text{followers}) + \text{VerifiedBonus}$$

- Trigram similarity is computed via `pg_trgm`'s `similarity(target, query)` function.
- Exact username matches receive a guaranteed top-slot anchor.

---

## 5. Media Tab Ranking

Posts containing valid media assets (`has_media = TRUE`) are filtered and ordered by composite Top score or Latest timestamp depending on the sub-filter toggle.

---

## 6. Language & Multilingual Search Optimization
- Default PostgreSQL full-text parser utilizes `'simple'` configuration to avoid aggressive English stemming on Indonesian words (e.g., preventing Indonesian "menggoreng" or "terbang" from being corrupted by English stemmers).
- Search token matching preserves punctuation for hashtags (`#technology`) and mentions (`@username`).
