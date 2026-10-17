package handler_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/x-reconstruction/api/internal/domain"
	"github.com/x-reconstruction/api/internal/handler"
	"github.com/x-reconstruction/api/internal/middleware"
	"github.com/x-reconstruction/api/internal/recommendation"
	"github.com/x-reconstruction/api/internal/repository"
)

func setupRecommendationRouter() (*chi.Mux, repository.Repository) {
	repo := repository.NewMemoryRepository()
	tmpDir, _ := os.MkdirTemp("", "rec_test")
	h := handler.NewHandler(repo, tmpDir, nil, nil, nil, nil)

	r := chi.NewRouter()
	r.Route("/api/v1", func(v1 chi.Router) {
		v1.Post("/auth/register", h.Register)
		v1.Post("/auth/login", h.Login)

		v1.Group(func(pub chi.Router) {
			pub.Use(middleware.AuthMiddleware(repo, true))
			pub.Get("/feed", h.GetFeed)
			pub.Get("/feed/new-posts", h.GetNewPosts)
			pub.Get("/explore/trending", h.GetTrendingTopics)
			pub.Get("/users/recommendations", h.GetWhoToFollow)
			pub.Post("/feed/impressions", h.RecordFeedImpressions)
		})

		v1.Group(func(auth chi.Router) {
			auth.Use(middleware.AuthMiddleware(repo, false))
			auth.Post("/posts", h.CreatePost)
			auth.Post("/feed/feedback", h.RecordFeedFeedback)
			auth.Post("/users/recommendations/{userId}/dismiss", h.DismissUserRecommendation)
			auth.Get("/me/topic-preferences", h.GetTopicPreferences)
			auth.Put("/me/topic-preferences", h.UpdateTopicPreference)
			auth.Post("/me/personalization/reset", h.ResetPersonalization)
			auth.Post("/users/{username}/follow", h.FollowUser)
			auth.Post("/users/{username}/block", h.BlockUser)
			auth.Post("/users/{username}/mute", h.MuteUser)
		})
	})

	return r, repo
}

func TestRecommendationAndFeedAPIs(t *testing.T) {
	r, _ := setupRecommendationRouter()

	// 1. Register User A (Author) and User B (Viewer)
	authorToken := registerTestUser(t, r, "author_alice", "alice@example.com", "Password123!")
	viewerToken := registerTestUser(t, r, "viewer_bob", "bob@example.com", "Password123!")

	// 2. Author creates post
	createBody := map[string]interface{}{"content": "Hello #React19 world with amazing new features!"}
	b, _ := json.Marshal(createBody)
	req := httptest.NewRequest("POST", "/api/v1/posts", bytes.NewBuffer(b))
	req.Header.Set("Authorization", "Bearer "+authorToken)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created for post, got %d", w.Code)
	}

	var postResp struct {
		ID string `json:"id"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &postResp)
	postID := postResp.ID

	// 3. Viewer reads 'For You' Feed
	reqFeed := httptest.NewRequest("GET", "/api/v1/feed?type=for_you&limit=10", nil)
	reqFeed.Header.Set("Authorization", "Bearer "+viewerToken)
	wFeed := httptest.NewRecorder()
	r.ServeHTTP(wFeed, reqFeed)
	if wFeed.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for for_you feed, got %d", wFeed.Code)
	}

	var feedResult struct {
		Posts []struct {
			ID string `json:"id"`
		} `json:"posts"`
		NextCursor string `json:"next_cursor"`
	}
	_ = json.Unmarshal(wFeed.Body.Bytes(), &feedResult)
	if len(feedResult.Posts) == 0 {
		t.Errorf("expected at least 1 post in for_you feed")
	}

	// 4. Viewer reads Trending Topics
	reqTrending := httptest.NewRequest("GET", "/api/v1/explore/trending", nil)
	wTrending := httptest.NewRecorder()
	r.ServeHTTP(wTrending, reqTrending)
	if wTrending.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for trending topics, got %d", wTrending.Code)
	}

	// 5. Viewer records Feed Impression
	impBody := map[string]interface{}{
		"impressions": []map[string]interface{}{
			{
				"post_id":       postID,
				"feed_type":     "for_you",
				"dwell_time_ms": 2500,
			},
		},
	}
	bImp, _ := json.Marshal(impBody)
	reqImp := httptest.NewRequest("POST", "/api/v1/feed/impressions", bytes.NewBuffer(bImp))
	reqImp.Header.Set("Authorization", "Bearer "+viewerToken)
	reqImp.Header.Set("Content-Type", "application/json")
	wImp := httptest.NewRecorder()
	r.ServeHTTP(wImp, reqImp)
	if wImp.Code != http.StatusOK {
		t.Errorf("expected 200 OK for recording impression, got %d", wImp.Code)
	}

	// 6. Viewer marks post as 'not_interested'
	feedBackBody := map[string]interface{}{
		"post_id":       postID,
		"feedback_type": "not_interested",
		"topic":         "react19",
	}
	bFB, _ := json.Marshal(feedBackBody)
	reqFB := httptest.NewRequest("POST", "/api/v1/feed/feedback", bytes.NewBuffer(bFB))
	reqFB.Header.Set("Authorization", "Bearer "+viewerToken)
	reqFB.Header.Set("Content-Type", "application/json")
	wFB := httptest.NewRecorder()
	r.ServeHTTP(wFB, reqFB)
	if wFB.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for feed feedback, got %d", wFB.Code)
	}

	// 7. Verify post is now suppressed for viewer
	reqFeedAfter := httptest.NewRequest("GET", "/api/v1/feed?type=for_you&limit=10", nil)
	reqFeedAfter.Header.Set("Authorization", "Bearer "+viewerToken)
	wFeedAfter := httptest.NewRecorder()
	r.ServeHTTP(wFeedAfter, reqFeedAfter)
	var feedResultAfter struct {
		Posts []struct {
			ID string `json:"id"`
		} `json:"posts"`
	}
	_ = json.Unmarshal(wFeedAfter.Body.Bytes(), &feedResultAfter)
	for _, p := range feedResultAfter.Posts {
		if p.ID == postID {
			t.Errorf("post %s was marked as not_interested but was still returned in feed", postID)
		}
	}

	// 8. Viewer resets personalization
	reqReset := httptest.NewRequest("POST", "/api/v1/me/personalization/reset", nil)
	reqReset.Header.Set("Authorization", "Bearer "+viewerToken)
	wReset := httptest.NewRecorder()
	r.ServeHTTP(wReset, reqReset)
	if wReset.Code != http.StatusOK {
		t.Errorf("expected 200 OK for personalization reset, got %d", wReset.Code)
	}
}

func TestRealtimeNewPostsIndicator(t *testing.T) {
	r, repo := setupRecommendationRouter()

	authorToken := registerTestUser(t, r, "stream_author", "stream@example.com", "Password123!")
	followerToken := registerTestUser(t, r, "stream_follower", "follower@example.com", "Password123!")

	// Follow author
	reqFollow := httptest.NewRequest("POST", "/api/v1/users/stream_author/follow", nil)
	reqFollow.Header.Set("Authorization", "Bearer "+followerToken)
	wFollow := httptest.NewRecorder()
	r.ServeHTTP(wFollow, reqFollow)
	if wFollow.Code != http.StatusOK && wFollow.Code != http.StatusCreated {
		t.Fatalf("expected follow success, got %d", wFollow.Code)
	}

	// 1. Author publishes Post 1
	p1Body := map[string]interface{}{"content": "Initial baseline post #stream"}
	bP1, _ := json.Marshal(p1Body)
	reqP1 := httptest.NewRequest("POST", "/api/v1/posts", bytes.NewBuffer(bP1))
	reqP1.Header.Set("Authorization", "Bearer "+authorToken)
	reqP1.Header.Set("Content-Type", "application/json")
	wP1 := httptest.NewRecorder()
	r.ServeHTTP(wP1, reqP1)
	var p1Resp struct {
		ID        string    `json:"id"`
		CreatedAt time.Time `json:"created_at"`
	}
	_ = json.Unmarshal(wP1.Body.Bytes(), &p1Resp)

	// Follower reads feed and obtains cursor for Post 1
	reqFeed := httptest.NewRequest("GET", "/api/v1/feed?type=following&limit=10", nil)
	reqFeed.Header.Set("Authorization", "Bearer "+followerToken)
	wFeed := httptest.NewRecorder()
	r.ServeHTTP(wFeed, reqFeed)

	var feedResult struct {
		Posts []domain.Post `json:"posts"`
	}
	_ = json.Unmarshal(wFeed.Body.Bytes(), &feedResult)
	if len(feedResult.Posts) == 0 {
		t.Fatalf("expected follower feed to contain initial post")
	}
	cursorPost1 := recommendation.EncodeCursor(feedResult.Posts[0].CreatedAt, feedResult.Posts[0].ID)

	// 2. Check for new posts with cursor of Post 1: should be 0
	reqCheck0 := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/feed/new-posts?tab=following&since=%s", cursorPost1), nil)
	reqCheck0.Header.Set("Authorization", "Bearer "+followerToken)
	wCheck0 := httptest.NewRecorder()
	r.ServeHTTP(wCheck0, reqCheck0)

	var check0Resp domain.NewPostsCheckResponse
	_ = json.Unmarshal(wCheck0.Body.Bytes(), &check0Resp)
	if check0Resp.Count != 0 || check0Resp.HasNew {
		t.Errorf("expected 0 new posts before any new publication, got %d", check0Resp.Count)
	}

	// 3. Author publishes Post 2
	time.Sleep(15 * time.Millisecond)
	p2Body := map[string]interface{}{"content": "Breaking announcement post #stream"}
	bP2, _ := json.Marshal(p2Body)
	reqP2 := httptest.NewRequest("POST", "/api/v1/posts", bytes.NewBuffer(bP2))
	reqP2.Header.Set("Authorization", "Bearer "+authorToken)
	reqP2.Header.Set("Content-Type", "application/json")
	wP2 := httptest.NewRecorder()
	r.ServeHTTP(wP2, reqP2)

	// 4. Follower polls new-posts endpoint with cursorPost1: should detect exactly 1 new post
	reqCheck1 := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/feed/new-posts?tab=following&since=%s", cursorPost1), nil)
	reqCheck1.Header.Set("Authorization", "Bearer "+followerToken)
	wCheck1 := httptest.NewRecorder()
	r.ServeHTTP(wCheck1, reqCheck1)

	var check1Resp domain.NewPostsCheckResponse
	_ = json.Unmarshal(wCheck1.Body.Bytes(), &check1Resp)
	if check1Resp.Count != 1 || !check1Resp.HasNew {
		t.Fatalf("expected exactly 1 new post detected, got %d, has_new=%v", check1Resp.Count, check1Resp.HasNew)
	}
	if len(check1Resp.Posts) != 1 {
		t.Fatalf("expected 1 post payload returned in new posts response")
	}

	// 5. Bilateral Block filter: Follower blocks Author -> new posts indicator drops to 0
	reqBlock := httptest.NewRequest("POST", "/api/v1/users/stream_author/block", nil)
	reqBlock.Header.Set("Authorization", "Bearer "+followerToken)
	wBlock := httptest.NewRecorder()
	r.ServeHTTP(wBlock, reqBlock)

	reqCheckBlocked := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/feed/new-posts?tab=following&since=%s", cursorPost1), nil)
	reqCheckBlocked.Header.Set("Authorization", "Bearer "+followerToken)
	wCheckBlocked := httptest.NewRecorder()
	r.ServeHTTP(wCheckBlocked, reqCheckBlocked)

	var checkBlockedResp domain.NewPostsCheckResponse
	_ = json.Unmarshal(wCheckBlocked.Body.Bytes(), &checkBlockedResp)
	if checkBlockedResp.Count != 0 || checkBlockedResp.HasNew {
		t.Errorf("blocked author posts must not leak into new posts indicator")
	}
	_ = repo
}

func TestTrendingAnalyticsAndSpamResistance(t *testing.T) {
	r, _ := setupRecommendationRouter()

	spammerToken := registerTestUser(t, r, "spammer_user", "spammer@example.com", "Password123!")
	user1Token := registerTestUser(t, r, "organic_user_1", "user1@example.com", "Password123!")
	user2Token := registerTestUser(t, r, "organic_user_2", "user2@example.com", "Password123!")
	user3Token := registerTestUser(t, r, "organic_user_3", "user3@example.com", "Password123!")

	// 1. Spammer publishes 4 posts with #CryptoSpam
	for i := 1; i <= 4; i++ {
		pBody := map[string]interface{}{"content": fmt.Sprintf("Spam repeat post %d #CryptoSpam buy now!", i)}
		b, _ := json.Marshal(pBody)
		req := httptest.NewRequest("POST", "/api/v1/posts", bytes.NewBuffer(b))
		req.Header.Set("Authorization", "Bearer "+spammerToken)
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(httptest.NewRecorder(), req)
	}

	// 2. Three distinct authors publish 1 post each with #AIInnovation
	for _, tok := range []string{user1Token, user2Token, user3Token} {
		pBody := map[string]interface{}{"content": "Deep research insights regarding #AIInnovation"}
		b, _ := json.Marshal(pBody)
		req := httptest.NewRequest("POST", "/api/v1/posts", bytes.NewBuffer(b))
		req.Header.Set("Authorization", "Bearer "+tok)
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(httptest.NewRecorder(), req)
	}

	// 3. Query trending topics
	reqTrending := httptest.NewRequest("GET", "/api/v1/explore/trending?limit=10", nil)
	wTrending := httptest.NewRecorder()
	r.ServeHTTP(wTrending, reqTrending)
	if wTrending.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for trending, got %d", wTrending.Code)
	}

	var res struct {
		Trending []domain.TrendingTopic `json:"trending"`
	}
	_ = json.Unmarshal(wTrending.Body.Bytes(), &res)

	// Verify #AIInnovation outranks #CryptoSpam due to unique author weight and anti-spam penalty
	var aiIndex = -1
	var cryptoIndex = -1
	for idx, tr := range res.Trending {
		if tr.Topic == "#aiinnovation" || tr.Topic == "#AIInnovation" {
			aiIndex = idx
		}
		if tr.Topic == "#cryptospam" || tr.Topic == "#CryptoSpam" {
			cryptoIndex = idx
		}
	}

	if aiIndex == -1 {
		t.Fatalf("expected #AIInnovation in trending topics, got: %+v", res.Trending)
	}
	if cryptoIndex != -1 && cryptoIndex < aiIndex {
		t.Errorf("spammed topic (#cryptospam) with 1 author should not rank higher than #aiinnovation with 3 distinct authors")
	}
}

func TestWhoToFollowAndDismissal(t *testing.T) {
	r, _ := setupRecommendationRouter()

	userAToken := registerTestUser(t, r, "network_a", "a@example.com", "Password123!")
	userBToken := registerTestUser(t, r, "network_b", "b@example.com", "Password123!")
	_ = registerTestUser(t, r, "network_c", "c@example.com", "Password123!")

	// User A follows User B
	reqAB := httptest.NewRequest("POST", "/api/v1/users/network_b/follow", nil)
	reqAB.Header.Set("Authorization", "Bearer "+userAToken)
	r.ServeHTTP(httptest.NewRecorder(), reqAB)

	// User B follows User C (making C a 2nd-degree friend of friend of A)
	reqBC := httptest.NewRequest("POST", "/api/v1/users/network_c/follow", nil)
	reqBC.Header.Set("Authorization", "Bearer "+userBToken)
	r.ServeHTTP(httptest.NewRecorder(), reqBC)

	// User A requests Who to Follow
	reqRec := httptest.NewRequest("GET", "/api/v1/users/recommendations?limit=10", nil)
	reqRec.Header.Set("Authorization", "Bearer "+userAToken)
	wRec := httptest.NewRecorder()
	r.ServeHTTP(wRec, reqRec)

	var recResp struct {
		Recommendations []domain.UserRecommendation `json:"recommendations"`
	}
	_ = json.Unmarshal(wRec.Body.Bytes(), &recResp)

	var foundC bool
	var userCID string
	for _, rec := range recResp.Recommendations {
		if rec.User.Username == "network_c" {
			foundC = true
			userCID = rec.User.ID.String()
			if rec.MutualCount < 1 {
				t.Errorf("expected mutual count >= 1 for 2nd-degree connection, got %d", rec.MutualCount)
			}
		}
		if rec.User.Username == "network_a" {
			t.Errorf("self must never be recommended in who to follow")
		}
		if rec.User.Username == "network_b" {
			t.Errorf("already followed accounts must not be recommended")
		}
	}
	if !foundC {
		t.Fatalf("expected network_c in recommendations, got: %+v", recResp.Recommendations)
	}

	// User A dismisses User C
	reqDismiss := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/users/recommendations/%s/dismiss", userCID), nil)
	reqDismiss.Header.Set("Authorization", "Bearer "+userAToken)
	wDismiss := httptest.NewRecorder()
	r.ServeHTTP(wDismiss, reqDismiss)
	if wDismiss.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for dismissing recommendation, got %d", wDismiss.Code)
	}

	// User A requests Who to Follow again -> User C must now be absent
	reqRecAfter := httptest.NewRequest("GET", "/api/v1/users/recommendations?limit=10", nil)
	reqRecAfter.Header.Set("Authorization", "Bearer "+userAToken)
	wRecAfter := httptest.NewRecorder()
	r.ServeHTTP(wRecAfter, reqRecAfter)

	var recRespAfter struct {
		Recommendations []domain.UserRecommendation `json:"recommendations"`
	}
	_ = json.Unmarshal(wRecAfter.Body.Bytes(), &recRespAfter)
	for _, rec := range recRespAfter.Recommendations {
		if rec.User.Username == "network_c" {
			t.Errorf("dismissed user network_c must not reappear in recommendations")
		}
	}
}

func registerTestUser(t *testing.T, r *chi.Mux, username, email, password string) string {
	t.Helper()
	regBody := map[string]string{
		"username":     username,
		"email":        email,
		"password":     password,
		"display_name": username,
	}
	b, _ := json.Marshal(regBody)
	req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("failed to register user %s, code=%d", username, w.Code)
	}

	var res struct {
		SessionToken string `json:"session_token"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &res)
	return res.SessionToken
}
