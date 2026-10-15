package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/x-reconstruction/api/internal/handler"
	"github.com/x-reconstruction/api/internal/middleware"
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
			pub.Get("/explore/trending", h.GetTrendingTopics)
			pub.Post("/feed/impressions", h.RecordFeedImpressions)
		})

		v1.Group(func(auth chi.Router) {
			auth.Use(middleware.AuthMiddleware(repo, false))
			auth.Post("/posts", h.CreatePost)
			auth.Post("/feed/feedback", h.RecordFeedFeedback)
			auth.Get("/me/topic-preferences", h.GetTopicPreferences)
			auth.Put("/me/topic-preferences", h.UpdateTopicPreference)
			auth.Post("/me/personalization/reset", h.ResetPersonalization)
			auth.Post("/users/{username}/follow", h.FollowUser)
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
