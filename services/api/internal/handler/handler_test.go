package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/x-reconstruction/api/internal/handler"
	"github.com/x-reconstruction/api/internal/middleware"
	"github.com/x-reconstruction/api/internal/repository"
)

func setupTestRouter() (*chi.Mux, repository.Repository) {
	repo := repository.NewMemoryRepository()
	tmpDir, _ := os.MkdirTemp("", "media_test")
	h := handler.NewHandler(repo, tmpDir, nil, nil, nil, nil)

	r := chi.NewRouter()

	r.Get("/health/live", h.HealthLive)
	r.Get("/health/ready", h.HealthReady)

	r.Route("/api/v1", func(v1 chi.Router) {
		v1.Post("/auth/register", h.Register)
		v1.Post("/auth/login", h.Login)

		v1.Group(func(pub chi.Router) {
			pub.Use(middleware.AuthMiddleware(repo, true))
			pub.Get("/feed", h.GetFeed)
			pub.Get("/users/{username}", h.GetUserByUsername)
			pub.Get("/posts/{postId}", h.GetPostByID)
			pub.Get("/subscription-plans", h.GetSubscriptionPlans)
		})

		v1.Group(func(auth chi.Router) {
			auth.Use(middleware.AuthMiddleware(repo, false))
			auth.Get("/me", h.GetMe)
			auth.Post("/posts", h.CreatePost)
			auth.Post("/posts/{postId}/likes", h.LikePost)
			auth.Delete("/posts/{postId}/likes", h.UnlikePost)
			auth.Post("/users/{username}/follow", h.FollowUser)
			auth.Delete("/users/{username}/follow", h.UnfollowUser)

			auth.Get("/subscriptions/me", h.GetMySubscription)
			auth.Post("/subscriptions/checkout", h.CheckoutSubscription)
			auth.Get("/subscriptions/orders/{orderId}", h.GetSubscriptionOrder)
			auth.Post("/subscriptions/simulate-payment", h.SimulatePayment)
			auth.Get("/billing/history", h.GetBillingHistory)
		})
	})

	return r, repo
}


func TestHealthCheck(t *testing.T) {
	r, _ := setupTestRouter()

	req := httptest.NewRequest("GET", "/health/ready", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", w.Code)
	}
}

func TestAuthAndPostingFlow(t *testing.T) {
	r, _ := setupTestRouter()

	// 1. Register new user
	regPayload := `{"username":"alice","email":"alice@example.com","password":"Password123!","display_name":"Alice"}`
	req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewBufferString(regPayload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %d: %s", w.Code, w.Body.String())
	}

	var authRes struct {
		SessionToken string `json:"session_token"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &authRes); err != nil || authRes.SessionToken == "" {
		t.Fatalf("failed to parse auth response or missing token")
	}

	// 2. Fetch /me
	reqMe := httptest.NewRequest("GET", "/api/v1/me", nil)
	reqMe.Header.Set("Authorization", "Bearer "+authRes.SessionToken)
	wMe := httptest.NewRecorder()
	r.ServeHTTP(wMe, reqMe)

	if wMe.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for /me, got %d: %s", wMe.Code, wMe.Body.String())
	}

	// 3. Create a post
	postPayload := `{"content":"Hello world from test suite!"}`
	reqPost := httptest.NewRequest("POST", "/api/v1/posts", bytes.NewBufferString(postPayload))
	reqPost.Header.Set("Content-Type", "application/json")
	reqPost.Header.Set("Authorization", "Bearer "+authRes.SessionToken)
	wPost := httptest.NewRecorder()
	r.ServeHTTP(wPost, reqPost)

	if wPost.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created for post, got %d: %s", wPost.Code, wPost.Body.String())
	}

	var postCreated struct {
		ID string `json:"id"`
	}
	_ = json.Unmarshal(wPost.Body.Bytes(), &postCreated)

	// 4. Like post
	reqLike := httptest.NewRequest("POST", "/api/v1/posts/"+postCreated.ID+"/likes", nil)
	reqLike.Header.Set("Authorization", "Bearer "+authRes.SessionToken)
	wLike := httptest.NewRecorder()
	r.ServeHTTP(wLike, reqLike)

	if wLike.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for like, got %d", wLike.Code)
	}

	// 5. Follow NASA
	reqFollow := httptest.NewRequest("POST", "/api/v1/users/NASA/follow", nil)
	reqFollow.Header.Set("Authorization", "Bearer "+authRes.SessionToken)
	wFollow := httptest.NewRecorder()
	r.ServeHTTP(wFollow, reqFollow)

	if wFollow.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for follow, got %d: %s", wFollow.Code, wFollow.Body.String())
	}

	// 6. Read feed
	reqFeed := httptest.NewRequest("GET", "/api/v1/feed?type=for_you", nil)
	wFeed := httptest.NewRecorder()
	r.ServeHTTP(wFeed, reqFeed)

	if wFeed.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for feed, got %d", wFeed.Code)
	}
}

func TestSubscriptionsAndBadges(t *testing.T) {
	r, _ := setupTestRouter()

	// 1. Register a test creator
	regPayload := []byte(`{"username":"elonmusk","email":"elon@x.com","password":"Password123!","display_name":"Elon Musk"}`)
	reqReg := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(regPayload))
	reqReg.Header.Set("Content-Type", "application/json")
	wReg := httptest.NewRecorder()
	r.ServeHTTP(wReg, reqReg)

	if wReg.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created for register, got %d: %s", wReg.Code, wReg.Body.String())
	}

	var authRes struct {
		SessionToken string `json:"session_token"`
	}
	_ = json.Unmarshal(wReg.Body.Bytes(), &authRes)

	// 2. Fetch Subscription Catalog
	reqPlans := httptest.NewRequest("GET", "/api/v1/subscription-plans", nil)
	wPlans := httptest.NewRecorder()
	r.ServeHTTP(wPlans, reqPlans)

	if wPlans.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for plans, got %d", wPlans.Code)
	}

	// 3. Free user attempts 350-character post (Free limit is 280) -> Should fail with 403 Forbidden
	longPost := strings.Repeat("A", 350)
	postBody, _ := json.Marshal(map[string]string{"content": longPost})
	reqLongPost := httptest.NewRequest("POST", "/api/v1/posts", bytes.NewBuffer(postBody))
	reqLongPost.Header.Set("Authorization", "Bearer "+authRes.SessionToken)
	reqLongPost.Header.Set("Content-Type", "application/json")
	wLongPost := httptest.NewRecorder()
	r.ServeHTTP(wLongPost, reqLongPost)

	if wLongPost.Code != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden for free user posting >280 chars, got %d", wLongPost.Code)
	}

	// 4. Initiate Checkout for Premium Plan
	checkoutBody := []byte(`{"plan_id":"premium","billing_cycle":"monthly"}`)
	reqCheckout := httptest.NewRequest("POST", "/api/v1/subscriptions/checkout", bytes.NewBuffer(checkoutBody))
	reqCheckout.Header.Set("Authorization", "Bearer "+authRes.SessionToken)
	reqCheckout.Header.Set("Content-Type", "application/json")
	wCheckout := httptest.NewRecorder()
	r.ServeHTTP(wCheckout, reqCheckout)

	if wCheckout.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created for checkout, got %d: %s", wCheckout.Code, wCheckout.Body.String())
	}

	var checkoutRes struct {
		OrderID string `json:"order_id"`
	}
	_ = json.Unmarshal(wCheckout.Body.Bytes(), &checkoutRes)

	// 5. Simulate Plisio payment event (Section 6K.12)
	simBody, _ := json.Marshal(map[string]string{"order_id": checkoutRes.OrderID})
	reqSim := httptest.NewRequest("POST", "/api/v1/subscriptions/simulate-payment", bytes.NewBuffer(simBody))
	reqSim.Header.Set("Authorization", "Bearer "+authRes.SessionToken)
	reqSim.Header.Set("Content-Type", "application/json")
	wSim := httptest.NewRecorder()
	r.ServeHTTP(wSim, reqSim)

	if wSim.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for payment simulation, got %d: %s", wSim.Code, wSim.Body.String())
	}

	var simRes struct {
		Status     string `json:"status"`
		IsVerified bool   `json:"is_verified"`
	}
	_ = json.Unmarshal(wSim.Body.Bytes(), &simRes)
	if !simRes.IsVerified {
		t.Fatalf("expected user to be verified after Premium activation")
	}

	// 6. Post long content again -> Now entitled up to 4,000 characters!
	reqLongPost2 := httptest.NewRequest("POST", "/api/v1/posts", bytes.NewBuffer(postBody))
	reqLongPost2.Header.Set("Authorization", "Bearer "+authRes.SessionToken)
	reqLongPost2.Header.Set("Content-Type", "application/json")
	wLongPost2 := httptest.NewRecorder()
	r.ServeHTTP(wLongPost2, reqLongPost2)

	if wLongPost2.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created for Premium user posting >280 chars, got %d: %s", wLongPost2.Code, wLongPost2.Body.String())
	}

	// 7. Check /me/subscription entitlements
	reqSubMe := httptest.NewRequest("GET", "/api/v1/subscriptions/me", nil)
	reqSubMe.Header.Set("Authorization", "Bearer "+authRes.SessionToken)
	wSubMe := httptest.NewRecorder()
	r.ServeHTTP(wSubMe, reqSubMe)

	if wSubMe.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for subscription/me, got %d", wSubMe.Code)
	}
}

