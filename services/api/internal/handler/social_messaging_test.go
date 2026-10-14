package handler_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/x-reconstruction/api/internal/domain"
	"github.com/x-reconstruction/api/internal/handler"
	"github.com/x-reconstruction/api/internal/middleware"
	"github.com/x-reconstruction/api/internal/realtime"
	"github.com/x-reconstruction/api/internal/repository"
)

func setupTestApp() (*chi.Mux, repository.Repository, *handler.Handler) {
	repo := repository.NewMemoryRepository()
	hub := realtime.NewHub(nil)
	h := handler.NewHandler(repo, "./test_uploads", nil, nil, hub, nil)

	r := chi.NewRouter()
	r.Use(middleware.AuthMiddleware(repo, true))

	r.Route("/api/v1", func(v1 chi.Router) {
		v1.Get("/ws", h.ServeWebSocket)

		v1.Group(func(pub chi.Router) {
			pub.Use(middleware.AuthMiddleware(repo, true))
			pub.Get("/users/{username}/followers", h.GetFollowers)
			pub.Get("/users/{username}/following", h.GetFollowing)
		})

		v1.Group(func(auth chi.Router) {
			auth.Use(middleware.AuthMiddleware(repo, false))

			// Social Graph
			auth.Post("/users/{username}/follow", h.FollowUser)
			auth.Delete("/users/{username}/follow", h.UnfollowUser)
			auth.Get("/users/{username}/mutuals", h.GetMutualConnections)
			auth.Get("/users/{username}/relationship", h.GetSocialRelationship)
			auth.Post("/users/{username}/block", h.BlockUser)
			auth.Delete("/users/{username}/block", h.UnblockUser)
			auth.Post("/users/{username}/mute", h.MuteUser)
			auth.Delete("/users/{username}/mute", h.UnmuteUser)

			// Follow Requests
			auth.Post("/users/{username}/follow-request", h.CreateFollowRequest)
			auth.Delete("/users/{username}/follow-request", h.CancelFollowRequest)
			auth.Get("/follow-requests", h.GetFollowRequests)
			auth.Post("/follow-requests/{requestId}/respond", h.RespondFollowRequest)

			// Direct Messaging
			auth.Get("/conversations", h.GetConversations)
			auth.Post("/conversations", h.CreateConversation)
			auth.Get("/conversations/{conversationId}", h.GetConversationByID)
			auth.Get("/conversations/{conversationId}/messages", h.GetMessages)
			auth.Post("/conversations/{conversationId}/messages", h.SendMessage)
			auth.Post("/conversations/{conversationId}/read", h.MarkConversationRead)

			// Notifications
			auth.Get("/notifications", h.GetNotifications)
			auth.Get("/notifications/unread-count", h.GetUnreadNotificationsCount)
			auth.Post("/notifications/{notificationId}/read", h.MarkNotificationRead)
		})
	})

	return r, repo, h
}

func createTestUser(t *testing.T, repo repository.Repository, username, email string, isPrivate bool) (*domain.User, string) {
	ctx := context.Background()
	user := &domain.User{
		ID:           uuid.New(),
		Username:     username,
		Email:        email,
		PasswordHash: "hashed",
		DisplayName:  username + " display",
		IsPrivate:    isPrivate,
	}
	err := repo.CreateUser(ctx, user)
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	sessionToken := "token_" + username
	hash := sha256.Sum256([]byte(sessionToken))
	tokenHash := hex.EncodeToString(hash[:])

	session := &domain.Session{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
	}
	err = repo.CreateSession(ctx, session)
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	return user, sessionToken
}

func TestSocialGraph_FollowUnfollowBlockMute(t *testing.T) {
	router, repo, _ := setupTestApp()

	alice, aliceToken := createTestUser(t, repo, "alice", "alice@example.com", false)
	bob, _ := createTestUser(t, repo, "bob", "bob@example.com", false)

	// 1. Alice follows Bob
	req := httptest.NewRequest(http.MethodPost, "/api/v1/users/bob/follow", nil)
	req.Header.Set("Authorization", "Bearer "+aliceToken)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	// 2. Check relationship from Alice's perspective
	req = httptest.NewRequest(http.MethodGet, "/api/v1/users/bob/relationship", nil)
	req.Header.Set("Authorization", "Bearer "+aliceToken)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var rel domain.SocialRelationship
	if err := json.Unmarshal(w.Body.Bytes(), &rel); err != nil {
		t.Fatalf("failed to unmarshal rel: %v", err)
	}
	if !rel.IsFollowing {
		t.Errorf("expected is_following to be true")
	}
	if rel.IsFollowedBy {
		t.Errorf("expected is_followed_by to be false")
	}
	if rel.IsMutual {
		t.Errorf("expected is_mutual to be false")
	}

	// 3. Alice blocks Bob -> should dissolve follow relationship
	req = httptest.NewRequest(http.MethodPost, "/api/v1/users/bob/block", nil)
	req.Header.Set("Authorization", "Bearer "+aliceToken)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	// Verify blocked state and follow dropped
	relAfterBlock, err := repo.GetSocialRelationship(context.Background(), alice.ID, bob.ID)
	if err != nil {
		t.Fatalf("failed to check social relationship: %v", err)
	}
	if relAfterBlock.IsFollowing {
		t.Errorf("follow relationship must be dissolved upon blocking")
	}
	if !relAfterBlock.IsBlocked {
		t.Errorf("expected is_blocked to be true")
	}

	// 4. Alice unblocks Bob
	req = httptest.NewRequest(http.MethodDelete, "/api/v1/users/bob/block", nil)
	req.Header.Set("Authorization", "Bearer "+aliceToken)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	// 5. Alice mutes Bob
	req = httptest.NewRequest(http.MethodPost, "/api/v1/users/bob/mute", nil)
	req.Header.Set("Authorization", "Bearer "+aliceToken)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	relAfterMute, err := repo.GetSocialRelationship(context.Background(), alice.ID, bob.ID)
	if err != nil {
		t.Fatalf("failed to check mute status: %v", err)
	}
	if !relAfterMute.IsMuted {
		t.Errorf("expected is_muted to be true")
	}
}

func TestFollowRequest_PrivateAccount(t *testing.T) {
	router, repo, _ := setupTestApp()

	_, charlieToken := createTestUser(t, repo, "charlie", "charlie@example.com", false)
	_, danaToken := createTestUser(t, repo, "dana", "dana@example.com", true) // Private account

	// 1. Charlie requests to follow private Dana
	req := httptest.NewRequest(http.MethodPost, "/api/v1/users/dana/follow-request", nil)
	req.Header.Set("Authorization", "Bearer "+charlieToken)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d: %s", w.Code, w.Body.String())
	}

	// 2. Dana checks pending follow requests
	req = httptest.NewRequest(http.MethodGet, "/api/v1/follow-requests", nil)
	req.Header.Set("Authorization", "Bearer "+danaToken)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Requests []domain.FollowRequest `json:"requests"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal requests: %v", err)
	}
	if len(resp.Requests) != 1 {
		t.Fatalf("expected 1 request, got %d", len(resp.Requests))
	}
	if resp.Requests[0].Requester.Username != "charlie" {
		t.Errorf("expected requester charlie, got %s", resp.Requests[0].Requester.Username)
	}

	// 3. Dana accepts request
	requestID := resp.Requests[0].ID
	body, _ := json.Marshal(map[string]string{"action": "accept"})
	req = httptest.NewRequest(http.MethodPost, "/api/v1/follow-requests/"+requestID.String()+"/respond", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+danaToken)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	// Verify Charlie is now following Dana
	req = httptest.NewRequest(http.MethodGet, "/api/v1/users/dana/relationship", nil)
	req.Header.Set("Authorization", "Bearer "+charlieToken)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var rel domain.SocialRelationship
	_ = json.Unmarshal(w.Body.Bytes(), &rel)
	if !rel.IsFollowing {
		t.Errorf("expected charlie to be following dana after acceptance")
	}
}

func TestDirectMessaging_IdempotencyAndFlow(t *testing.T) {
	router, repo, _ := setupTestApp()

	eva, evaToken := createTestUser(t, repo, "eva", "eva@example.com", false)
	frank, frankToken := createTestUser(t, repo, "frank", "frank@example.com", false)

	// 1. Eva creates/gets conversation with Frank
	convBody, _ := json.Marshal(map[string]string{"recipient_id": frank.ID.String()})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/conversations", bytes.NewReader(convBody))
	req.Header.Set("Authorization", "Bearer "+evaToken)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d: %s", w.Code, w.Body.String())
	}

	var conv domain.Conversation
	if err := json.Unmarshal(w.Body.Bytes(), &conv); err != nil {
		t.Fatalf("failed to unmarshal conversation: %v", err)
	}
	if conv.ID == uuid.Nil {
		t.Fatalf("conversation ID cannot be empty")
	}

	// 2. Eva sends message with ClientMessageID (Idempotency)
	clientMsgID := "client_msg_uuid_12345"
	msgBody, _ := json.Marshal(map[string]string{
		"content":           "Hello Frank, realtime message!",
		"client_message_id": clientMsgID,
	})
	req = httptest.NewRequest(http.MethodPost, "/api/v1/conversations/"+conv.ID.String()+"/messages", bytes.NewReader(msgBody))
	req.Header.Set("Authorization", "Bearer "+evaToken)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d: %s", w.Code, w.Body.String())
	}

	var sentMsg domain.Message
	if err := json.Unmarshal(w.Body.Bytes(), &sentMsg); err != nil {
		t.Fatalf("failed to unmarshal sent msg: %v", err)
	}
	if sentMsg.Content != "Hello Frank, realtime message!" {
		t.Errorf("content mismatch")
	}
	if sentMsg.ClientMessageID != clientMsgID {
		t.Errorf("client message ID mismatch")
	}

	// 3. Frank lists messages
	req = httptest.NewRequest(http.MethodGet, "/api/v1/conversations/"+conv.ID.String()+"/messages", nil)
	req.Header.Set("Authorization", "Bearer "+frankToken)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var msgListResp struct {
		Messages []domain.Message `json:"messages"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &msgListResp); err != nil {
		t.Fatalf("failed to unmarshal msg list: %v", err)
	}
	if len(msgListResp.Messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(msgListResp.Messages))
	}

	// 4. Frank marks conversation as read
	readBody, _ := json.Marshal(map[string]string{"message_id": sentMsg.ID.String()})
	req = httptest.NewRequest(http.MethodPost, "/api/v1/conversations/"+conv.ID.String()+"/read", bytes.NewReader(readBody))
	req.Header.Set("Authorization", "Bearer "+frankToken)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	// 5. Block prevention: If Frank blocks Eva, Eva cannot message Frank
	_ = repo.BlockUser(context.Background(), frank.ID, eva.ID)
	req = httptest.NewRequest(http.MethodPost, "/api/v1/conversations/"+conv.ID.String()+"/messages", bytes.NewReader(msgBody))
	req.Header.Set("Authorization", "Bearer "+evaToken)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected status 403 Forbidden when messaging a user who blocked you, got %d", w.Code)
	}
}
