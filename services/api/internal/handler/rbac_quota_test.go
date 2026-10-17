package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/x-reconstruction/api/internal/domain"
	"github.com/x-reconstruction/api/internal/handler"
	"github.com/x-reconstruction/api/internal/middleware"
	"github.com/x-reconstruction/api/internal/repository"
)

func setupTestRBACRouter(repo repository.Repository) http.Handler {
	h := handler.NewHandler(repo, "./test_uploads", nil, nil, nil, nil)
	r := chi.NewRouter()

	r.Route("/api/v1", func(v1 chi.Router) {
		v1.Post("/auth/register", h.Register)
		v1.Post("/auth/login", h.Login)

		v1.Group(func(auth chi.Router) {
			auth.Use(middleware.AuthMiddleware(repo, false))

			auth.Get("/me", h.GetMe)
			auth.Post("/posts", h.CreatePost)
			auth.Patch("/posts/{postId}", h.PatchPost)
			auth.Delete("/posts/{postId}", h.DeletePost)

			// Admin & Governance Endpoints
			adminPerms := middleware.RequireAnyPermission(repo, "role.read", "role.create", "role.update", "permission.manage", "user.read", "audit.read", "membership.plan.manage")
			auth.Group(func(admin chi.Router) {
				admin.Use(adminPerms)

				admin.Get("/admin/roles", h.AdminGetRoles)
				admin.Post("/admin/roles", h.AdminCreateRole)
				admin.Get("/admin/roles/{id}", h.AdminGetRoleByID)
				admin.Patch("/admin/roles/{id}", h.AdminUpdateRole)
				admin.Delete("/admin/roles/{id}", h.AdminDeleteRole)
				admin.Get("/admin/permissions", h.AdminGetPermissions)
				admin.Get("/admin/roles/{id}/permissions", h.AdminGetRolePermissions)
				admin.Put("/admin/roles/{id}/permissions", h.AdminSetRolePermissions)

				admin.Get("/admin/memberships", h.AdminGetMembershipPlans)
				admin.Patch("/admin/memberships/{id}", h.AdminUpdateMembershipPlan)
				admin.Put("/admin/plans/{id}/entitlements", h.AdminSetPlanEntitlements)
				admin.Post("/admin/users/{id}/membership", h.AdminAssignUserMembership)
				admin.Delete("/admin/users/{id}/membership/{planId}", h.AdminRevokeUserMembership)

				admin.Get("/admin/users", h.AdminGetUsers)
				admin.Get("/admin/users/{id}", h.AdminGetUserByID)
				admin.Patch("/admin/users/{id}/roles", h.AdminUpdateUserRoles)
				admin.Patch("/admin/users/{id}/status", h.AdminUpdateUserStatus)

				admin.Get("/admin/audit-logs", h.AdminGetAuditLogs)
			})
		})
	})

	return r
}

func registerAndLoginUser(t *testing.T, r http.Handler, username, email, password string) (string, *domain.User) {
	t.Helper()
	regBody, _ := json.Marshal(map[string]string{
		"username": username,
		"email":    email,
		"password": password,
	})
	reqReg := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(regBody))
	reqReg.Header.Set("Content-Type", "application/json")
	wReg := httptest.NewRecorder()
	r.ServeHTTP(wReg, reqReg)

	if wReg.Code != http.StatusCreated {
		t.Fatalf("failed to register %s: %s", username, wReg.Body.String())
	}

	var res struct {
		User         *domain.User `json:"user"`
		SessionToken string       `json:"session_token"`
	}
	_ = json.Unmarshal(wReg.Body.Bytes(), &res)
	return res.SessionToken, res.User
}

func TestNewUserDefaultMemberRoleAndAdminProtection(t *testing.T) {
	repo := repository.NewMemoryRepository()
	router := setupTestRBACRouter(repo)

	// 1. Register regular user
	tokenAlice, userAlice := registerAndLoginUser(t, router, "alice", "alice@example.com", "Password123!")

	// 2. Verify Alice has role 'member' by querying /api/v1/me
	reqMe := httptest.NewRequest("GET", "/api/v1/me", nil)
	reqMe.Header.Set("Authorization", "Bearer "+tokenAlice)
	wMe := httptest.NewRecorder()
	router.ServeHTTP(wMe, reqMe)

	if wMe.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for /me, got %d", wMe.Code)
	}

	var meRes struct {
		Role        string        `json:"role"`
		Roles       []domain.Role `json:"roles"`
		Permissions []string      `json:"permissions"`
		IsOwner     bool          `json:"is_owner"`
	}
	_ = json.Unmarshal(wMe.Body.Bytes(), &meRes)

	if meRes.Role != "member" {
		t.Errorf("expected default role 'member', got %s", meRes.Role)
	}
	if meRes.IsOwner {
		t.Errorf("expected alice to not be owner")
	}

	// 3. Regular user Alice attempts to access Admin endpoints -> Must receive 403 Forbidden
	reqAdmin := httptest.NewRequest("GET", "/api/v1/admin/roles", nil)
	reqAdmin.Header.Set("Authorization", "Bearer "+tokenAlice)
	wAdmin := httptest.NewRecorder()
	router.ServeHTTP(wAdmin, reqAdmin)

	if wAdmin.Code != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden for non-admin accessing /admin/roles, got %d", wAdmin.Code)
	}

	// 4. Regular user Alice attempts to self-elevate roles via admin API -> Must receive 403 Forbidden
	elevateBody, _ := json.Marshal(map[string]interface{}{
		"role_ids": []string{"00000000-0000-0000-0000-000000000006"}, // owner role
	})
	reqElevate := httptest.NewRequest("PATCH", fmt.Sprintf("/api/v1/admin/users/%s/roles", userAlice.ID), bytes.NewBuffer(elevateBody))
	reqElevate.Header.Set("Authorization", "Bearer "+tokenAlice)
	reqElevate.Header.Set("Content-Type", "application/json")
	wElevate := httptest.NewRecorder()
	router.ServeHTTP(wElevate, reqElevate)

	if wElevate.Code != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden for self-role elevation, got %d", wElevate.Code)
	}
}

func TestOwnerBootstrapAndNonDemotionRule(t *testing.T) {
	repo := repository.NewMemoryRepository()
	router := setupTestRBACRouter(repo)

	// 1. Register candidate owner account afrizaahmad18@gmail.com
	ownerToken, ownerUser := registerAndLoginUser(t, router, "afriza", "afrizaahmad18@gmail.com", "SuperSecureOwnerPass123!")

	// 2. Check /me -> Must have role 'owner' and is_owner == true
	reqMe := httptest.NewRequest("GET", "/api/v1/me", nil)
	reqMe.Header.Set("Authorization", "Bearer "+ownerToken)
	wMe := httptest.NewRecorder()
	router.ServeHTTP(wMe, reqMe)

	var meRes struct {
		Role    string `json:"role"`
		IsOwner bool   `json:"is_owner"`
	}
	_ = json.Unmarshal(wMe.Body.Bytes(), &meRes)

	if meRes.Role != "owner" || !meRes.IsOwner {
		t.Fatalf("expected owner account to have role 'owner' and is_owner=true, got role=%s, is_owner=%v", meRes.Role, meRes.IsOwner)
	}

	// 3. Owner CAN access admin endpoints
	reqRoles := httptest.NewRequest("GET", "/api/v1/admin/roles", nil)
	reqRoles.Header.Set("Authorization", "Bearer "+ownerToken)
	wRoles := httptest.NewRecorder()
	router.ServeHTTP(wRoles, reqRoles)

	if wRoles.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for Owner accessing /admin/roles, got %d", wRoles.Code)
	}

	// 4. Create another admin to test demotion protection
	_, anotherAdmin := registerAndLoginUser(t, router, "anotheradmin", "anotheradmin@example.com", "Password123!")
	godRoleID := uuid.MustParse("00000000-0000-0000-0000-000000000005")
	_ = repo.AssignRoleToUser(context.Background(), anotherAdmin.ID, godRoleID, &ownerUser.ID, "Test elevation")

	// Another admin attempts to demote or suspend the Owner -> Must be rejected (403 Forbidden)
	demoteBody, _ := json.Marshal(map[string]interface{}{
		"role_ids": []string{"00000000-0000-0000-0000-000000000001"}, // demote to member
	})
	adminToken, adminCaller := registerAndLoginUser(t, router, "admincaller", "caller@example.com", "Password123!")
	_ = repo.AssignRoleToUser(context.Background(), adminCaller.ID, godRoleID, nil, "Elevate caller")

	reqDemoteOwner := httptest.NewRequest("PATCH", fmt.Sprintf("/api/v1/admin/users/%s/roles", ownerUser.ID), bytes.NewBuffer(demoteBody))
	reqDemoteOwner.Header.Set("Authorization", "Bearer "+adminToken)
	reqDemoteOwner.Header.Set("Content-Type", "application/json")
	wDemoteOwner := httptest.NewRecorder()
	router.ServeHTTP(wDemoteOwner, reqDemoteOwner)

	// Owner account cannot be demoted or changed through admin endpoint
	if wDemoteOwner.Code != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden when trying to demote owner account, got %d", wDemoteOwner.Code)
	}

	// 5. Deleting Owner role from system roles via CRUD must be forbidden
	ownerRoleUUID := uuid.MustParse("00000000-0000-0000-0000-000000000006")
	reqDeleteRole := httptest.NewRequest("DELETE", fmt.Sprintf("/api/v1/admin/roles/%s", ownerRoleUUID), nil)
	reqDeleteRole.Header.Set("Authorization", "Bearer "+ownerToken)
	wDeleteRole := httptest.NewRecorder()
	router.ServeHTTP(wDeleteRole, reqDeleteRole)

	if wDeleteRole.Code != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden when deleting system Owner role, got %d", wDeleteRole.Code)
	}
}

func TestPostingQuotaEnforcementAndConcurrency(t *testing.T) {
	repo := repository.NewMemoryRepository()
	router := setupTestRBACRouter(repo)

	// Member plan: 2 posts/day
	tokenBob, _ := registerAndLoginUser(t, router, "bob", "bob@example.com", "Password123!")

	// Post 1 -> Must succeed (201)
	postBody1, _ := json.Marshal(map[string]string{"content": "First post of the day!"})
	req1 := httptest.NewRequest("POST", "/api/v1/posts", bytes.NewBuffer(postBody1))
	req1.Header.Set("Authorization", "Bearer "+tokenBob)
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	if w1.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created for post 1, got %d: %s", w1.Code, w1.Body.String())
	}

	// Post 2 -> Must succeed (201)
	postBody2, _ := json.Marshal(map[string]string{"content": "Second post of the day!"})
	req2 := httptest.NewRequest("POST", "/api/v1/posts", bytes.NewBuffer(postBody2))
	req2.Header.Set("Authorization", "Bearer "+tokenBob)
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	if w2.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created for post 2, got %d: %s", w2.Code, w2.Body.String())
	}

	// Post 3 -> Must be rejected (429 Too Many Requests or 403)
	postBody3, _ := json.Marshal(map[string]string{"content": "Third post exceeding quota!"})
	req3 := httptest.NewRequest("POST", "/api/v1/posts", bytes.NewBuffer(postBody3))
	req3.Header.Set("Authorization", "Bearer "+tokenBob)
	req3.Header.Set("Content-Type", "application/json")
	w3 := httptest.NewRecorder()
	router.ServeHTTP(w3, req3)

	if w3.Code != http.StatusTooManyRequests && w3.Code != http.StatusForbidden {
		t.Fatalf("expected quota rejection (429 or 403) for post 3, got %d: %s", w3.Code, w3.Body.String())
	}

	// Re-login does NOT reset counter
	loginBody, _ := json.Marshal(map[string]string{"identifier": "bob", "password": "Password123!"})
	reqLogin := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(loginBody))
	reqLogin.Header.Set("Content-Type", "application/json")
	wLogin := httptest.NewRecorder()
	router.ServeHTTP(wLogin, reqLogin)

	var loginRes struct {
		SessionToken string `json:"session_token"`
	}
	_ = json.Unmarshal(wLogin.Body.Bytes(), &loginRes)

	req4 := httptest.NewRequest("POST", "/api/v1/posts", bytes.NewBuffer(postBody3))
	req4.Header.Set("Authorization", "Bearer "+loginRes.SessionToken)
	req4.Header.Set("Content-Type", "application/json")
	w4 := httptest.NewRecorder()
	router.ServeHTTP(w4, req4)

	if w4.Code != http.StatusTooManyRequests && w4.Code != http.StatusForbidden {
		t.Fatalf("re-login must NOT reset quota counter, got status %d", w4.Code)
	}
}

func TestConcurrentPostingCannotBypassQuota(t *testing.T) {
	repo := repository.NewMemoryRepository()
	router := setupTestRBACRouter(repo)

	tokenCharlie, _ := registerAndLoginUser(t, router, "charlie", "charlie@example.com", "Password123!")

	// Fire 10 concurrent requests simultaneously; only 2 may succeed
	var wg sync.WaitGroup
	successCount := 0
	var mu sync.Mutex

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			b, _ := json.Marshal(map[string]string{"content": fmt.Sprintf("Concurrent post %d", idx)})
			req := httptest.NewRequest("POST", "/api/v1/posts", bytes.NewBuffer(b))
			req.Header.Set("Authorization", "Bearer "+tokenCharlie)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code == http.StatusCreated {
				mu.Lock()
				successCount++
				mu.Unlock()
			}
		}(i)
	}

	wg.Wait()

	if successCount > 2 {
		t.Fatalf("race condition detected: expected at most 2 posts to succeed, got %d", successCount)
	}
	if successCount != 2 {
		t.Fatalf("expected exactly 2 posts to succeed, got %d", successCount)
	}
}

func TestPostEditDoesNotConsumeQuotaAndValidatesOwnership(t *testing.T) {
	repo := repository.NewMemoryRepository()
	router := setupTestRBACRouter(repo)

	tokenDave, userDave := registerAndLoginUser(t, router, "dave", "dave@example.com", "Password123!")
	tokenEve, _ := registerAndLoginUser(t, router, "eve", "eve@example.com", "Password123!")

	// Dave creates 2 posts (reaches quota)
	p1Body, _ := json.Marshal(map[string]string{"content": "Dave initial post 1"})
	req1 := httptest.NewRequest("POST", "/api/v1/posts", bytes.NewBuffer(p1Body))
	req1.Header.Set("Authorization", "Bearer "+tokenDave)
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	var createdPost domain.Post
	_ = json.Unmarshal(w1.Body.Bytes(), &createdPost)

	p2Body, _ := json.Marshal(map[string]string{"content": "Dave initial post 2"})
	req2 := httptest.NewRequest("POST", "/api/v1/posts", bytes.NewBuffer(p2Body))
	req2.Header.Set("Authorization", "Bearer "+tokenDave)
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	// Dave edits Post 1 -> Must succeed (200 OK) even though daily post creation quota is exhausted!
	editBody, _ := json.Marshal(map[string]string{"content": "Dave edited content"})
	reqEdit := httptest.NewRequest("PATCH", fmt.Sprintf("/api/v1/posts/%s", createdPost.ID), bytes.NewBuffer(editBody))
	reqEdit.Header.Set("Authorization", "Bearer "+tokenDave)
	reqEdit.Header.Set("Content-Type", "application/json")
	wEdit := httptest.NewRecorder()
	router.ServeHTTP(wEdit, reqEdit)

	if wEdit.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for post edit by owner without consuming quota, got %d: %s", wEdit.Code, wEdit.Body.String())
	}

	// Eve attempts to edit Dave's post -> Must be rejected (403 Forbidden)
	reqEveEdit := httptest.NewRequest("PATCH", fmt.Sprintf("/api/v1/posts/%s", createdPost.ID), bytes.NewBuffer(editBody))
	reqEveEdit.Header.Set("Authorization", "Bearer "+tokenEve)
	reqEveEdit.Header.Set("Content-Type", "application/json")
	wEveEdit := httptest.NewRecorder()
	router.ServeHTTP(wEveEdit, reqEveEdit)

	if wEveEdit.Code != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden when Eve tries to edit Dave's post, got %d", wEveEdit.Code)
	}

	// Eve attempts to delete Dave's post -> Must be rejected (403 Forbidden)
	reqEveDelete := httptest.NewRequest("DELETE", fmt.Sprintf("/api/v1/posts/%s", createdPost.ID), nil)
	reqEveDelete.Header.Set("Authorization", "Bearer "+tokenEve)
	wEveDelete := httptest.NewRecorder()
	router.ServeHTTP(wEveDelete, reqEveDelete)

	if wEveDelete.Code != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden when Eve tries to delete Dave's post, got %d", wEveDelete.Code)
	}

	// Dave deletes own post -> Must succeed (204 No Content)
	reqDaveDelete := httptest.NewRequest("DELETE", fmt.Sprintf("/api/v1/posts/%s", createdPost.ID), nil)
	reqDaveDelete.Header.Set("Authorization", "Bearer "+tokenDave)
	wDaveDelete := httptest.NewRecorder()
	router.ServeHTTP(wDaveDelete, reqDaveDelete)

	if wDaveDelete.Code != http.StatusNoContent {
		t.Fatalf("expected 204 No Content when Dave deletes own post, got %d", wDaveDelete.Code)
	}

	// Deleting a post does NOT restore quota; Dave still cannot post a 3rd post
	p3Body, _ := json.Marshal(map[string]string{"content": "Dave attempting 3rd post after delete"})
	req3 := httptest.NewRequest("POST", "/api/v1/posts", bytes.NewBuffer(p3Body))
	req3.Header.Set("Authorization", "Bearer "+tokenDave)
	req3.Header.Set("Content-Type", "application/json")
	w3 := httptest.NewRecorder()
	router.ServeHTTP(w3, req3)

	if w3.Code != http.StatusTooManyRequests && w3.Code != http.StatusForbidden {
		t.Fatalf("deleting post must NOT restore quota, got %d", w3.Code)
	}

	_ = userDave
}

func TestMediaUploadLimitsByTier(t *testing.T) {
	repo := repository.NewMemoryRepository()
	router := setupTestRBACRouter(repo)

	tokenFrank, _ := registerAndLoginUser(t, router, "frank", "frank@example.com", "Password123!")

	// Member tier allows 2 images, 0 videos
	// Attach 3 images -> Must be rejected (403 Forbidden)
	pOverImages, _ := json.Marshal(map[string]interface{}{
		"content": "Too many images",
		"media": []map[string]interface{}{
			{"media_type": "image", "url": "https://example.com/1.jpg"},
			{"media_type": "image", "url": "https://example.com/2.jpg"},
			{"media_type": "image", "url": "https://example.com/3.jpg"},
		},
	})
	reqImages := httptest.NewRequest("POST", "/api/v1/posts", bytes.NewBuffer(pOverImages))
	reqImages.Header.Set("Authorization", "Bearer "+tokenFrank)
	reqImages.Header.Set("Content-Type", "application/json")
	wImages := httptest.NewRecorder()
	router.ServeHTTP(wImages, reqImages)

	if wImages.Code != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden for exceeding image limit, got %d: %s", wImages.Code, wImages.Body.String())
	}

	// Attach 1 video for free member -> Must be rejected (403 Forbidden)
	pVideo, _ := json.Marshal(map[string]interface{}{
		"content": "Video on free member",
		"media": []map[string]interface{}{
			{"media_type": "video", "url": "https://example.com/1.mp4"},
		},
	})
	reqVideo := httptest.NewRequest("POST", "/api/v1/posts", bytes.NewBuffer(pVideo))
	reqVideo.Header.Set("Authorization", "Bearer "+tokenFrank)
	reqVideo.Header.Set("Content-Type", "application/json")
	wVideo := httptest.NewRecorder()
	router.ServeHTTP(wVideo, reqVideo)

	if wVideo.Code != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden for video on free tier, got %d: %s", wVideo.Code, wVideo.Body.String())
	}
}
