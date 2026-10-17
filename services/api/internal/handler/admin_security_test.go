package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/x-reconstruction/api/internal/domain"
	"github.com/x-reconstruction/api/internal/handler"
	"github.com/x-reconstruction/api/internal/middleware"
	"github.com/x-reconstruction/api/internal/repository"
)

func setupTestFullAdminRouter(repo repository.Repository) http.Handler {
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
			adminPerms := middleware.RequireAnyPermission(repo, "role.read", "role.create", "role.update", "permission.manage", "user.read", "audit.read", "membership.plan.manage", "posts.moderate", "posts.delete.any", "monitoring.read")
			auth.Group(func(admin chi.Router) {
				admin.Use(adminPerms)

				// Overview & Monitoring
				admin.Get("/admin/overview", h.AdminGetOverview)
				admin.Get("/admin/monitoring/overview", h.AdminGetMonitoringOverview)
				admin.Get("/admin/security/events", h.AdminGetSecurityEvents)

				// Roles & Permissions
				admin.Get("/admin/roles", h.AdminGetRoles)
				admin.Post("/admin/roles", h.AdminCreateRole)
				admin.Get("/admin/roles/{id}", h.AdminGetRoleByID)
				admin.Patch("/admin/roles/{id}", h.AdminUpdateRole)
				admin.Delete("/admin/roles/{id}", h.AdminDeleteRole)
				admin.Get("/admin/permissions", h.AdminGetPermissions)
				admin.Get("/admin/roles/{id}/permissions", h.AdminGetRolePermissions)
				admin.Put("/admin/roles/{id}/permissions", h.AdminSetRolePermissions)

				// Membership Plans & Entitlements
				admin.Get("/admin/memberships", h.AdminGetMembershipPlans)
				admin.Patch("/admin/memberships/{id}", h.AdminUpdateMembershipPlan)
				admin.Put("/admin/plans/{id}/entitlements", h.AdminSetPlanEntitlements)
				admin.Post("/admin/users/{id}/membership", h.AdminAssignUserMembership)
				admin.Delete("/admin/users/{id}/membership/{planId}", h.AdminRevokeUserMembership)

				// User Administration
				admin.Get("/admin/users", h.AdminGetUsers)
				admin.Get("/admin/users/{id}", h.AdminGetUserByID)
				admin.Patch("/admin/users/{id}/roles", h.AdminUpdateUserRoles)
				admin.Patch("/admin/users/{id}/status", h.AdminUpdateUserStatus)

				// Post Moderation & Management
				admin.Get("/admin/posts", h.AdminGetPosts)
				admin.Patch("/admin/posts/{id}/moderation", h.AdminUpdatePostModeration)
				admin.Delete("/admin/posts/{id}", h.AdminDeletePost)

				// Audit Logs
				admin.Get("/admin/audit-logs", h.AdminGetAuditLogs)
			})
		})
	})

	return r
}

func TestOwnerAdminOverviewSecurity(t *testing.T) {
	repo := repository.NewMemoryRepository()
	router := setupTestFullAdminRouter(repo)

	// 1. Register regular user
	tokenJoe, _ := registerAndLoginUser(t, router, "regular_joe", "joe@example.com", "Password123!")

	// 2. Register Owner candidate user
	tokenOwner, _ := registerAndLoginUser(t, router, "afriza", "afrizaahmad18@gmail.com", "Password123!")

	// Test A: Unauthenticated request to /admin/overview -> 401
	{
		req := httptest.NewRequest("GET", "/api/v1/admin/overview", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401 Unauthorized, got %d", w.Code)
		}
	}

	// Test B: Regular member requesting /admin/overview -> 403 Forbidden
	{
		req := httptest.NewRequest("GET", "/api/v1/admin/overview", nil)
		req.Header.Set("Authorization", "Bearer "+tokenJoe)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusForbidden {
			t.Fatalf("expected 403 Forbidden for regular member, got %d", w.Code)
		}
	}

	// Test C: Owner requesting /admin/overview -> 200 OK with valid live stats
	{
		req := httptest.NewRequest("GET", "/api/v1/admin/overview", nil)
		req.Header.Set("Authorization", "Bearer "+tokenOwner)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 OK for owner, got %d: %s", w.Code, w.Body.String())
		}

		var overview domain.AdminOverview
		if err := json.Unmarshal(w.Body.Bytes(), &overview); err != nil {
			t.Fatalf("failed to decode overview JSON: %v", err)
		}
		if overview.TotalUsers < 2 {
			t.Fatalf("expected at least 2 users, got %d", overview.TotalUsers)
		}
		if overview.SystemHealth.Database.Status != "healthy" {
			t.Fatalf("expected healthy DB status, got %s", overview.SystemHealth.Database.Status)
		}
	}

	// Test D: Owner requesting /admin/monitoring/overview -> 200 OK
	{
		req := httptest.NewRequest("GET", "/api/v1/admin/monitoring/overview", nil)
		req.Header.Set("Authorization", "Bearer "+tokenOwner)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 OK for monitoring overview, got %d", w.Code)
		}
	}

	// Test E: Owner requesting /admin/security/events -> 200 OK
	{
		req := httptest.NewRequest("GET", "/api/v1/admin/security/events", nil)
		req.Header.Set("Authorization", "Bearer "+tokenOwner)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 OK for security events, got %d", w.Code)
		}
	}
}

func TestAdminPostModerationAndAuditLogging(t *testing.T) {
	repo := repository.NewMemoryRepository()
	router := setupTestFullAdminRouter(repo)

	// 1. Author and regular member
	tokenBob, _ := registerAndLoginUser(t, router, "author_bob", "bob@example.com", "Password123!")
	tokenCharlie, _ := registerAndLoginUser(t, router, "charlie", "charlie@example.com", "Password123!")

	// 2. Owner
	tokenOwner, _ := registerAndLoginUser(t, router, "afriza", "afrizaahmad18@gmail.com", "Password123!")

	// 3. Bob creates a post
	postBody, _ := json.Marshal(map[string]interface{}{
		"content": "A post to be moderated",
	})
	reqPost := httptest.NewRequest("POST", "/api/v1/posts", bytes.NewReader(postBody))
	reqPost.Header.Set("Authorization", "Bearer "+tokenBob)
	reqPost.Header.Set("Content-Type", "application/json")
	wPost := httptest.NewRecorder()
	router.ServeHTTP(wPost, reqPost)

	if wPost.Code != http.StatusCreated {
		t.Fatalf("failed to create post: %s", wPost.Body.String())
	}

	var createdPost domain.Post
	if err := json.Unmarshal(wPost.Body.Bytes(), &createdPost); err != nil {
		t.Fatalf("failed to parse created post: %v", err)
	}

	// Test A: Charlie (regular user) attempts moderation -> 403 Forbidden
	{
		modBody, _ := json.Marshal(map[string]string{
			"moderation_status": "hidden",
			"reason":            "Spam",
		})
		reqMod := httptest.NewRequest("PATCH", "/api/v1/admin/posts/"+createdPost.ID.String()+"/moderation", bytes.NewReader(modBody))
		reqMod.Header.Set("Authorization", "Bearer "+tokenCharlie)
		reqMod.Header.Set("Content-Type", "application/json")
		wMod := httptest.NewRecorder()
		router.ServeHTTP(wMod, reqMod)

		if wMod.Code != http.StatusForbidden {
			t.Fatalf("expected 403 Forbidden for non-admin moderation, got %d", wMod.Code)
		}
	}

	// Test B: Owner moderates the post to 'hidden' -> 200 OK
	{
		modBody, _ := json.Marshal(map[string]string{
			"moderation_status": "hidden",
			"reason":            "Policy violation",
		})
		reqMod := httptest.NewRequest("PATCH", "/api/v1/admin/posts/"+createdPost.ID.String()+"/moderation", bytes.NewReader(modBody))
		reqMod.Header.Set("Authorization", "Bearer "+tokenOwner)
		reqMod.Header.Set("Content-Type", "application/json")
		wMod := httptest.NewRecorder()
		router.ServeHTTP(wMod, reqMod)

		if wMod.Code != http.StatusOK {
			t.Fatalf("expected 200 OK for owner post moderation, got %d: %s", wMod.Code, wMod.Body.String())
		}
	}

	// Test C: Owner lists admin posts -> 200 OK and includes moderated post
	{
		reqList := httptest.NewRequest("GET", "/api/v1/admin/posts?moderation_status=hidden", nil)
		reqList.Header.Set("Authorization", "Bearer "+tokenOwner)
		wList := httptest.NewRecorder()
		router.ServeHTTP(wList, reqList)

		if wList.Code != http.StatusOK {
			t.Fatalf("expected 200 OK for admin posts list, got %d: %s", wList.Code, wList.Body.String())
		}

		var res struct {
			Posts []domain.Post `json:"posts"`
			Total int64         `json:"total"`
		}
		if err := json.Unmarshal(wList.Body.Bytes(), &res); err != nil {
			t.Fatalf("failed to decode admin posts: %v", err)
		}
		if res.Total < 1 {
			t.Fatalf("expected at least 1 hidden post, got %d", res.Total)
		}
	}

	// Test D: Owner deletes post administratively -> 200 OK
	{
		delBody, _ := json.Marshal(map[string]string{
			"reason": "Severe violation removal",
		})
		reqDel := httptest.NewRequest("DELETE", "/api/v1/admin/posts/"+createdPost.ID.String(), bytes.NewReader(delBody))
		reqDel.Header.Set("Authorization", "Bearer "+tokenOwner)
		reqDel.Header.Set("Content-Type", "application/json")
		wDel := httptest.NewRecorder()
		router.ServeHTTP(wDel, reqDel)

		if wDel.Code != http.StatusOK {
			t.Fatalf("expected 200 OK for admin post deletion, got %d: %s", wDel.Code, wDel.Body.String())
		}
	}

	// Test E: Verify audit log recorded the moderation actions
	{
		reqAudit := httptest.NewRequest("GET", "/api/v1/admin/audit-logs", nil)
		reqAudit.Header.Set("Authorization", "Bearer "+tokenOwner)
		wAudit := httptest.NewRecorder()
		router.ServeHTTP(wAudit, reqAudit)

		if wAudit.Code != http.StatusOK {
			t.Fatalf("expected 200 OK for audit logs, got %d", wAudit.Code)
		}

		var auditRes struct {
			AuditLogs []domain.AuditLog `json:"audit_logs"`
		}
		_ = json.Unmarshal(wAudit.Body.Bytes(), &auditRes)

		hasModAudit := false
		hasDelAudit := false
		for _, l := range auditRes.AuditLogs {
			if l.Action == "post.moderate" {
				hasModAudit = true
			}
			if l.Action == "post.delete.admin" {
				hasDelAudit = true
			}
		}

		if !hasModAudit {
			t.Fatalf("expected audit log entry for post.moderate")
		}
		if !hasDelAudit {
			t.Fatalf("expected audit log entry for post.delete.admin")
		}
	}
}
