package handler_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/x-reconstruction/api/internal/domain"
	"github.com/x-reconstruction/api/internal/handler"
	"github.com/x-reconstruction/api/internal/middleware"
	"github.com/x-reconstruction/api/internal/repository"
)

func setupInteractionsTestRouter() (*chi.Mux, repository.Repository) {
	repo := repository.NewMemoryRepository()
	tmpDir, _ := os.MkdirTemp("", "interactions_test")
	h := handler.NewHandler(repo, tmpDir, nil, nil, nil, nil)

	r := chi.NewRouter()

	r.Route("/api/v1", func(v1 chi.Router) {
		v1.Post("/auth/register", h.Register)
		v1.Post("/auth/login", h.Login)

		v1.Group(func(pub chi.Router) {
			pub.Use(middleware.AuthMiddleware(repo, true))
			pub.Get("/feed", h.GetFeed)
			pub.Get("/users/{username}", h.GetUserByUsername)
			pub.Get("/users/{username}/posts", h.GetUserPosts)
			pub.Get("/posts/{postId}", h.GetPostByID)
			pub.Get("/posts/{postId}/replies", h.GetPostReplies)
			pub.Get("/posts/{postId}/quotes", h.GetPostQuotes)
		})

		v1.Group(func(auth chi.Router) {
			auth.Use(middleware.AuthMiddleware(repo, false))
			auth.Get("/me", h.GetMe)
			auth.Post("/posts", h.CreatePost)
			auth.Delete("/posts/{postId}", h.DeletePost)

			auth.Post("/posts/{postId}/likes", h.LikePost)
			auth.Put("/posts/{postId}/like", h.LikePost)
			auth.Delete("/posts/{postId}/likes", h.UnlikePost)
			auth.Delete("/posts/{postId}/like", h.UnlikePost)

			auth.Post("/posts/{postId}/reposts", h.RepostPost)
			auth.Post("/posts/{postId}/repost", h.RepostPost)
			auth.Delete("/posts/{postId}/reposts", h.UndoRepost)
			auth.Delete("/posts/{postId}/repost", h.UndoRepost)

			auth.Post("/admin/maintenance/reconcile-counters", h.ReconcileCounters)
		})
	})

	return r, repo
}

func registerTestUserInInteractions(t *testing.T, r *chi.Mux, username, email string) string {
	t.Helper()
	regBody, _ := json.Marshal(map[string]string{
		"username":     username,
		"email":        email,
		"password":     "SecretPass123!",
		"display_name": username,
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewReader(regBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated && w.Code != http.StatusOK {
		t.Fatalf("Failed to register user %s: code %d, body: %s", username, w.Code, w.Body.String())
	}

	var res struct {
		SessionToken string `json:"session_token"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &res)
	return res.SessionToken
}

func createTestPost(t *testing.T, r *chi.Mux, token, content string, replyToID, quotePostID *string) domain.Post {
	t.Helper()
	payload := map[string]interface{}{
		"content": content,
	}
	if replyToID != nil {
		payload["reply_to_id"] = *replyToID
	}
	if quotePostID != nil {
		payload["quote_post_id"] = *quotePostID
	}
	b, _ := json.Marshal(payload)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/posts", bytes.NewReader(b))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("Failed to create post: code %d, body: %s", w.Code, w.Body.String())
	}

	var p domain.Post
	_ = json.Unmarshal(w.Body.Bytes(), &p)
	return p
}

// Test 1: Like Idempotency and Duplicate Click Concurrency
func TestLikeIdempotencyAndConcurrency(t *testing.T) {
	r, _ := setupInteractionsTestRouter()
	userA := registerTestUserInInteractions(t, r, "alice", "alice@example.com")
	userB := registerTestUserInInteractions(t, r, "bob", "bob@example.com")

	post := createTestPost(t, r, userA, "Hello world", nil, nil)

	// Step 1: User B likes post
	w1 := httptest.NewRecorder()
	req1 := httptest.NewRequest("PUT", fmt.Sprintf("/api/v1/posts/%s/like", post.ID), nil)
	req1.Header.Set("Authorization", "Bearer "+userB)
	r.ServeHTTP(w1, req1)
	if w1.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK on like, got %d", w1.Code)
	}

	var likeRes struct {
		IsLiked   bool `json:"is_liked"`
		LikeCount int  `json:"like_count"`
	}
	_ = json.Unmarshal(w1.Body.Bytes(), &likeRes)
	if !likeRes.IsLiked || likeRes.LikeCount != 1 {
		t.Errorf("Expected is_liked=true and like_count=1, got %+v", likeRes)
	}

	// Step 2: User B repeats like request (Idempotency check)
	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest("PUT", fmt.Sprintf("/api/v1/posts/%s/like", post.ID), nil)
	req2.Header.Set("Authorization", "Bearer "+userB)
	r.ServeHTTP(w2, req2)
	_ = json.Unmarshal(w2.Body.Bytes(), &likeRes)
	if likeRes.LikeCount != 1 {
		t.Errorf("Duplicate like increased count! Expected 1, got %d", likeRes.LikeCount)
	}

	// Step 3: Concurrent like attempts from the same user must not duplicate
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			w := httptest.NewRecorder()
			req := httptest.NewRequest("PUT", fmt.Sprintf("/api/v1/posts/%s/like", post.ID), nil)
			req.Header.Set("Authorization", "Bearer "+userB)
			r.ServeHTTP(w, req)
		}()
	}
	wg.Wait()

	// Verify post like count is still exactly 1
	wPost := httptest.NewRecorder()
	reqPost := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/posts/%s", post.ID), nil)
	reqPost.Header.Set("Authorization", "Bearer "+userB)
	r.ServeHTTP(wPost, reqPost)
	var postChecked domain.Post
	_ = json.Unmarshal(wPost.Body.Bytes(), &postChecked)
	if postChecked.LikesCount != 1 {
		t.Errorf("Expected likes_count=1 after concurrency, got %d", postChecked.LikesCount)
	}
	if !postChecked.IsLiked {
		t.Errorf("Expected is_liked=true for user B")
	}

	// Step 4: Unlike
	wUnlike := httptest.NewRecorder()
	reqUnlike := httptest.NewRequest("DELETE", fmt.Sprintf("/api/v1/posts/%s/like", post.ID), nil)
	reqUnlike.Header.Set("Authorization", "Bearer "+userB)
	r.ServeHTTP(wUnlike, reqUnlike)
	if wUnlike.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK on unlike, got %d", wUnlike.Code)
	}

	// Step 5: Repeated Unlike does not drop below 0
	wUnlike2 := httptest.NewRecorder()
	reqUnlike2 := httptest.NewRequest("DELETE", fmt.Sprintf("/api/v1/posts/%s/like", post.ID), nil)
	reqUnlike2.Header.Set("Authorization", "Bearer "+userB)
	r.ServeHTTP(wUnlike2, reqUnlike2)

	wPostAfter := httptest.NewRecorder()
	reqPostAfter := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/posts/%s", post.ID), nil)
	reqPostAfter.Header.Set("Authorization", "Bearer "+userB)
	r.ServeHTTP(wPostAfter, reqPostAfter)
	_ = json.Unmarshal(wPostAfter.Body.Bytes(), &postChecked)
	if postChecked.LikesCount != 0 {
		t.Errorf("Expected likes_count=0 after repeated unlike, got %d", postChecked.LikesCount)
	}
	if postChecked.IsLiked {
		t.Errorf("Expected is_liked=false after unlike")
	}
}

// Test 2: Repost and Undo Repost Idempotency
func TestRepostIdempotency(t *testing.T) {
	r, _ := setupInteractionsTestRouter()
	userA := registerTestUserInInteractions(t, r, "charlie", "charlie@example.com")
	userB := registerTestUserInInteractions(t, r, "dave", "dave@example.com")

	post := createTestPost(t, r, userA, "Original post by Charlie", nil, nil)

	// Step 1: User B reposts
	w1 := httptest.NewRecorder()
	req1 := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/posts/%s/repost", post.ID), nil)
	req1.Header.Set("Authorization", "Bearer "+userB)
	r.ServeHTTP(w1, req1)
	if w1.Code != http.StatusOK {
		t.Fatalf("Expected 200 on repost, got %d", w1.Code)
	}

	var repostRes struct {
		IsReposted  bool `json:"is_reposted"`
		RepostCount int  `json:"repost_count"`
	}
	_ = json.Unmarshal(w1.Body.Bytes(), &repostRes)
	if !repostRes.IsReposted || repostRes.RepostCount != 1 {
		t.Errorf("Expected is_reposted=true and repost_count=1, got %+v", repostRes)
	}

	// Step 2: Repeated repost request is idempotent
	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/posts/%s/repost", post.ID), nil)
	req2.Header.Set("Authorization", "Bearer "+userB)
	r.ServeHTTP(w2, req2)
	_ = json.Unmarshal(w2.Body.Bytes(), &repostRes)
	if repostRes.RepostCount != 1 {
		t.Errorf("Duplicate repost increased count! Expected 1, got %d", repostRes.RepostCount)
	}

	// Step 3: Undo Repost
	wUndo := httptest.NewRecorder()
	reqUndo := httptest.NewRequest("DELETE", fmt.Sprintf("/api/v1/posts/%s/repost", post.ID), nil)
	reqUndo.Header.Set("Authorization", "Bearer "+userB)
	r.ServeHTTP(wUndo, reqUndo)
	if wUndo.Code != http.StatusOK {
		t.Fatalf("Expected 200 on undo repost, got %d", wUndo.Code)
	}

	// Step 4: Repeated Undo does not decrement below 0
	wUndo2 := httptest.NewRecorder()
	reqUndo2 := httptest.NewRequest("DELETE", fmt.Sprintf("/api/v1/posts/%s/repost", post.ID), nil)
	reqUndo2.Header.Set("Authorization", "Bearer "+userB)
	r.ServeHTTP(wUndo2, reqUndo2)

	wPost := httptest.NewRecorder()
	reqPost := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/posts/%s", post.ID), nil)
	reqPost.Header.Set("Authorization", "Bearer "+userB)
	r.ServeHTTP(wPost, reqPost)
	var postChecked domain.Post
	_ = json.Unmarshal(wPost.Body.Bytes(), &postChecked)
	if postChecked.RepostsCount != 0 {
		t.Errorf("Expected reposts_count=0, got %d", postChecked.RepostsCount)
	}
}

// Test 3: Quote Post Creation & Independent Counts
func TestQuotePostCreationAndLinkage(t *testing.T) {
	r, _ := setupInteractionsTestRouter()
	userA := registerTestUserInInteractions(t, r, "edward", "edward@example.com")
	userB := registerTestUserInInteractions(t, r, "fiona", "fiona@example.com")

	origPost := createTestPost(t, r, userA, "Original post for quoting", nil, nil)

	// User B creates quote post
	origIDStr := origPost.ID.String()
	quotePost := createTestPost(t, r, userB, "This is my commentary", nil, &origIDStr)

	if quotePost.QuotePostID == nil || *quotePost.QuotePostID != origPost.ID {
		t.Errorf("Expected quote_post_id=%s, got %v", origPost.ID, quotePost.QuotePostID)
	}

	// Quoting user likes the quote post, does not affect original post
	wLike := httptest.NewRecorder()
	reqLike := httptest.NewRequest("PUT", fmt.Sprintf("/api/v1/posts/%s/like", quotePost.ID), nil)
	reqLike.Header.Set("Authorization", "Bearer "+userB)
	r.ServeHTTP(wLike, reqLike)

	// Check original post counts are untainted
	wOrig := httptest.NewRecorder()
	reqOrig := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/posts/%s", origPost.ID), nil)
	reqOrig.Header.Set("Authorization", "Bearer "+userA)
	r.ServeHTTP(wOrig, reqOrig)
	var origChecked domain.Post
	_ = json.Unmarshal(wOrig.Body.Bytes(), &origChecked)
	if origChecked.LikesCount != 0 {
		t.Errorf("Original post likes modified by quote like! Expected 0, got %d", origChecked.LikesCount)
	}
}

// Test 4: Profile Tabs Filtering (Postingan, Reply, Repost, Media)
func TestProfileTabsFiltering(t *testing.T) {
	r, _ := setupInteractionsTestRouter()
	userA := registerTestUserInInteractions(t, r, "george", "george@example.com")
	userB := registerTestUserInInteractions(t, r, "hannah", "hannah@example.com")

	// 1. Original post
	origPost := createTestPost(t, r, userA, "George's original post", nil, nil)

	// 2. Reply to Hannah
	hannahPost := createTestPost(t, r, userB, "Hannah's status", nil, nil)
	hannahIDStr := hannahPost.ID.String()
	createTestPost(t, r, userA, "George's reply to Hannah", &hannahIDStr, nil)

	// 3. Repost Hannah's status
	wRep := httptest.NewRecorder()
	reqRep := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/posts/%s/repost", hannahPost.ID), nil)
	reqRep.Header.Set("Authorization", "Bearer "+userA)
	r.ServeHTTP(wRep, reqRep)

	// Tab: posts (Postingan) -> Must contain original post, must NOT contain reply
	wTabPosts := httptest.NewRecorder()
	reqTabPosts := httptest.NewRequest("GET", "/api/v1/users/george/posts?tab=posts", nil)
	r.ServeHTTP(wTabPosts, reqTabPosts)
	var postsRes struct {
		Posts []domain.Post `json:"posts"`
	}
	_ = json.Unmarshal(wTabPosts.Body.Bytes(), &postsRes)
	if len(postsRes.Posts) != 1 || postsRes.Posts[0].ID != origPost.ID {
		t.Errorf("Tab 'posts' returned unexpected posts count: %d (expected 1)", len(postsRes.Posts))
	}

	// Tab: replies (Reply) -> Must contain the reply only
	wTabReplies := httptest.NewRecorder()
	reqTabReplies := httptest.NewRequest("GET", "/api/v1/users/george/posts?tab=replies", nil)
	r.ServeHTTP(wTabReplies, reqTabReplies)
	var repliesRes struct {
		Posts []domain.Post `json:"posts"`
	}
	_ = json.Unmarshal(wTabReplies.Body.Bytes(), &repliesRes)
	if len(repliesRes.Posts) != 1 || repliesRes.Posts[0].ReplyToID == nil {
		t.Errorf("Tab 'replies' returned unexpected replies: %+v", repliesRes.Posts)
	}

	// Tab: reposts (Repost) -> Must contain Hannah's post with George as reposter
	wTabReposts := httptest.NewRecorder()
	reqTabReposts := httptest.NewRequest("GET", "/api/v1/users/george/posts?tab=reposts", nil)
	r.ServeHTTP(wTabReposts, reqTabReposts)
	var repostsRes struct {
		Posts []domain.Post `json:"posts"`
	}
	_ = json.Unmarshal(wTabReposts.Body.Bytes(), &repostsRes)
	if len(repostsRes.Posts) != 1 || repostsRes.Posts[0].ID != hannahPost.ID {
		t.Errorf("Tab 'reposts' expected 1 repost, got %d", len(repostsRes.Posts))
	}
	if len(repostsRes.Posts) > 0 && (repostsRes.Posts[0].Reposter == nil || repostsRes.Posts[0].Reposter.Username != "george") {
		t.Errorf("Repost attribution missing or incorrect: %+v", repostsRes.Posts[0].Reposter)
	}
}

// Test 5: Counter Reconciliation Maintenance Tooling
func TestCounterReconciliation(t *testing.T) {
	r, _ := setupInteractionsTestRouter()
	adminToken := registerTestUserInInteractions(t, r, "adminuser", "admin@example.com")

	wRec := httptest.NewRecorder()
	reqRec := httptest.NewRequest("POST", "/api/v1/admin/maintenance/reconcile-counters", bytes.NewReader([]byte(`{"dry_run": true, "limit": 50}`)))
	reqRec.Header.Set("Authorization", "Bearer "+adminToken)
	reqRec.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(wRec, reqRec)

	if wRec.Code != http.StatusOK {
		t.Fatalf("Expected 200 on reconcile counters, got %d: %s", wRec.Code, wRec.Body.String())
	}

	var recRes domain.CounterReconciliationResult
	_ = json.Unmarshal(wRec.Body.Bytes(), &recRes)
	if !recRes.DryRun {
		t.Errorf("Expected DryRun=true in response")
	}
}
