package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/x-reconstruction/api/internal/domain"
	"github.com/x-reconstruction/api/internal/middleware"
	"github.com/x-reconstruction/api/internal/repository"
)

// ==========================================
// ROLES & PERMISSIONS MANAGEMENT (ADMIN)
// ==========================================

func (h *Handler) AdminGetRoles(w http.ResponseWriter, r *http.Request) {
	roles, err := h.repo.GetAllRoles(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch roles")
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"roles": roles})
}

func (h *Handler) AdminGetRoleByID(w http.ResponseWriter, r *http.Request) {
	roleID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid role id")
		return
	}

	role, err := h.repo.GetRoleByID(r.Context(), roleID)
	if err != nil {
		writeError(w, http.StatusNotFound, "role not found")
		return
	}

	perms, _ := h.repo.GetRolePermissions(r.Context(), roleID)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"role":        role,
		"permissions": perms,
	})
}

type CreateRoleRequest struct {
	Key         string      `json:"key"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	IsActive    bool        `json:"is_active"`
	Permissions []uuid.UUID `json:"permissions"`
}

func (h *Handler) AdminCreateRole(w http.ResponseWriter, r *http.Request) {
	actor := middleware.GetUserFromContext(r.Context())
	var req CreateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	req.Key = strings.ToLower(strings.TrimSpace(req.Key))
	req.Name = strings.TrimSpace(req.Name)
	if req.Key == "" || req.Name == "" {
		writeError(w, http.StatusBadRequest, "role key and name are required")
		return
	}

	role := &domain.Role{
		Key:         req.Key,
		Name:        req.Name,
		Description: req.Description,
		IsActive:    req.IsActive,
	}

	if err := h.repo.CreateRole(r.Context(), role); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("failed to create role: %v", err))
		return
	}

	if len(req.Permissions) > 0 {
		_ = h.repo.SetRolePermissions(r.Context(), role.ID, req.Permissions)
	}

	// Audit Log
	_ = h.repo.CreateAuditLog(r.Context(), &domain.AuditLog{
		ActorUserID: &actor.ID,
		Action:      "role.create",
		TargetType:  "role",
		TargetID:    role.ID.String(),
		Reason:      fmt.Sprintf("Created role %s (%s)", role.Name, role.Key),
		MetadataRedacted: map[string]interface{}{
			"key":  role.Key,
			"name": role.Name,
		},
	})

	writeJSON(w, http.StatusCreated, role)
}

type UpdateRoleRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	IsActive    bool   `json:"is_active"`
}

func (h *Handler) AdminUpdateRole(w http.ResponseWriter, r *http.Request) {
	actor := middleware.GetUserFromContext(r.Context())
	roleID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid role id")
		return
	}

	existing, err := h.repo.GetRoleByID(r.Context(), roleID)
	if err != nil {
		writeError(w, http.StatusNotFound, "role not found")
		return
	}

	var req UpdateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	if req.Name != "" {
		existing.Name = strings.TrimSpace(req.Name)
	}
	existing.Description = req.Description
	existing.IsActive = req.IsActive

	if err := h.repo.UpdateRole(r.Context(), existing); err != nil {
		if err == repository.ErrRoleProtected {
			writeError(w, http.StatusForbidden, "system role cannot be deactivated or renamed")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to update role")
		return
	}

	// Audit Log
	_ = h.repo.CreateAuditLog(r.Context(), &domain.AuditLog{
		ActorUserID: &actor.ID,
		Action:      "role.update",
		TargetType:  "role",
		TargetID:    roleID.String(),
		Reason:      fmt.Sprintf("Updated role %s", existing.Key),
		MetadataRedacted: map[string]interface{}{
			"key":       existing.Key,
			"is_active": existing.IsActive,
		},
	})

	writeJSON(w, http.StatusOK, existing)
}

func (h *Handler) AdminDeleteRole(w http.ResponseWriter, r *http.Request) {
	actor := middleware.GetUserFromContext(r.Context())
	roleID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid role id")
		return
	}

	role, err := h.repo.GetRoleByID(r.Context(), roleID)
	if err != nil {
		writeError(w, http.StatusNotFound, "role not found")
		return
	}

	if err := h.repo.DeleteRole(r.Context(), roleID); err != nil {
		if err == repository.ErrRoleProtected || err == repository.ErrForbidden {
			writeError(w, http.StatusForbidden, "system roles cannot be deleted")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to delete role")
		return
	}

	// Audit Log
	_ = h.repo.CreateAuditLog(r.Context(), &domain.AuditLog{
		ActorUserID: &actor.ID,
		Action:      "role.delete",
		TargetType:  "role",
		TargetID:    roleID.String(),
		Reason:      fmt.Sprintf("Deleted custom role %s (%s)", role.Name, role.Key),
	})

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) AdminGetPermissions(w http.ResponseWriter, r *http.Request) {
	perms, err := h.repo.GetAllPermissions(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch permissions")
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"permissions": perms})
}

func (h *Handler) AdminGetRolePermissions(w http.ResponseWriter, r *http.Request) {
	roleID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid role id")
		return
	}

	perms, err := h.repo.GetRolePermissions(r.Context(), roleID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch role permissions")
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"permissions": perms})
}

type SetPermissionsRequest struct {
	PermissionIDs []uuid.UUID `json:"permission_ids"`
}

func (h *Handler) AdminSetRolePermissions(w http.ResponseWriter, r *http.Request) {
	actor := middleware.GetUserFromContext(r.Context())
	roleID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid role id")
		return
	}

	role, err := h.repo.GetRoleByID(r.Context(), roleID)
	if err != nil {
		writeError(w, http.StatusNotFound, "role not found")
		return
	}

	var req SetPermissionsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	if err := h.repo.SetRolePermissions(r.Context(), roleID, req.PermissionIDs); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update role permissions")
		return
	}

	// Audit Log
	_ = h.repo.CreateAuditLog(r.Context(), &domain.AuditLog{
		ActorUserID: &actor.ID,
		Action:      "role.permissions.update",
		TargetType:  "role",
		TargetID:    roleID.String(),
		Reason:      fmt.Sprintf("Updated permissions for role %s (%d permissions assigned)", role.Key, len(req.PermissionIDs)),
		MetadataRedacted: map[string]interface{}{
			"role_key":         role.Key,
			"permission_count": len(req.PermissionIDs),
		},
	})

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message":          "role permissions updated successfully",
		"permission_count": len(req.PermissionIDs),
	})
}

// ==========================================
// MEMBERSHIP PLANS & ENTITLEMENTS (ADMIN & PUBLIC)
// ==========================================

func (h *Handler) GetMembershipPlansCatalog(w http.ResponseWriter, r *http.Request) {
	plans, err := h.repo.GetAllMembershipPlans(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch membership plans")
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"plans": plans})
}

func (h *Handler) GetMyMembershipStatus(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	mem, err := h.repo.GetUserMembership(r.Context(), user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch membership status")
		return
	}

	entitlements, _ := h.repo.GetResolvedEntitlements(r.Context(), user.ID)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"membership":   mem,
		"entitlements": entitlements,
	})
}

func (h *Handler) AdminGetMembershipPlans(w http.ResponseWriter, r *http.Request) {
	plans, err := h.repo.GetAllMembershipPlans(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch membership plans")
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"plans": plans})
}

type UpdatePlanRequest struct {
	Name             string `json:"name"`
	Description      string `json:"description"`
	PriceAmountCents int    `json:"price_amount_cents"`
	Currency         string `json:"currency"`
	BillingInterval  string `json:"billing_interval"`
	IsActive         bool   `json:"is_active"`
}

func (h *Handler) AdminUpdateMembershipPlan(w http.ResponseWriter, r *http.Request) {
	actor := middleware.GetUserFromContext(r.Context())
	planID := chi.URLParam(r, "id")

	existing, err := h.repo.GetMembershipPlanByID(r.Context(), planID)
	if err != nil {
		writeError(w, http.StatusNotFound, "plan not found")
		return
	}

	var req UpdatePlanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	if req.Name != "" {
		existing.Name = req.Name
	}
	existing.Description = req.Description
	existing.PriceAmountCents = req.PriceAmountCents
	if req.Currency != "" {
		existing.Currency = req.Currency
	}
	if req.BillingInterval != "" {
		existing.BillingInterval = req.BillingInterval
	}
	existing.IsActive = req.IsActive

	if err := h.repo.UpdateMembershipPlan(r.Context(), existing); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update plan")
		return
	}

	// Audit Log
	_ = h.repo.CreateAuditLog(r.Context(), &domain.AuditLog{
		ActorUserID: &actor.ID,
		Action:      "membership.plan.update",
		TargetType:  "membership_plan",
		TargetID:    planID,
		Reason:      fmt.Sprintf("Updated membership plan %s (price: %d %s)", existing.Key, existing.PriceAmountCents, existing.Currency),
		MetadataRedacted: map[string]interface{}{
			"plan_id":            planID,
			"price_amount_cents": existing.PriceAmountCents,
			"is_active":          existing.IsActive,
		},
	})

	writeJSON(w, http.StatusOK, existing)
}

type SetPlanEntitlementsRequest struct {
	Entitlements []domain.MembershipEntitlement `json:"entitlements"`
}

func (h *Handler) AdminSetPlanEntitlements(w http.ResponseWriter, r *http.Request) {
	actor := middleware.GetUserFromContext(r.Context())
	planID := chi.URLParam(r, "id")

	_, err := h.repo.GetMembershipPlanByID(r.Context(), planID)
	if err != nil {
		writeError(w, http.StatusNotFound, "membership plan not found")
		return
	}

	var req SetPlanEntitlementsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	if err := h.repo.SetPlanEntitlements(r.Context(), planID, req.Entitlements); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to set plan entitlements")
		return
	}

	// Audit Log
	_ = h.repo.CreateAuditLog(r.Context(), &domain.AuditLog{
		ActorUserID: &actor.ID,
		Action:      "membership.entitlements.update",
		TargetType:  "membership_plan",
		TargetID:    planID,
		Reason:      fmt.Sprintf("Updated %d entitlements for plan %s", len(req.Entitlements), planID),
	})

	updated, _ := h.repo.GetPlanEntitlements(r.Context(), planID)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"plan_id":      planID,
		"entitlements": updated,
	})
}

// ==========================================
// USER ADMINISTRATION (ADMIN)
// ==========================================

func (h *Handler) AdminGetUsers(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	search := r.URL.Query().Get("search")

	users, total, err := h.repo.GetAllUsersAdmin(r.Context(), limit, offset, search)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch users")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"users":  users,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

func (h *Handler) AdminGetUserByID(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	user, err := h.repo.GetUserByID(r.Context(), userID, nil)
	if err != nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}

	roles, _ := h.repo.GetUserRoles(r.Context(), userID)
	perms, _ := h.repo.GetUserPermissions(r.Context(), userID)
	membership, _ := h.repo.GetUserMembership(r.Context(), userID)
	entitlements, _ := h.repo.GetResolvedEntitlements(r.Context(), userID)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"user":         user,
		"roles":        roles,
		"permissions":  perms,
		"membership":   membership,
		"entitlements": entitlements,
	})
}

type UpdateUserRolesRequest struct {
	RoleIDs []uuid.UUID `json:"role_ids"`
	Reason  string      `json:"reason"`
}

func (h *Handler) AdminUpdateUserRoles(w http.ResponseWriter, r *http.Request) {
	actor := middleware.GetUserFromContext(r.Context())
	targetUserID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	// Security Rule: Users cannot modify their own roles via admin endpoint
	if actor.ID == targetUserID {
		writeError(w, http.StatusForbidden, "cannot modify your own administrative roles")
		return
	}

	targetUser, err := h.repo.GetUserByID(r.Context(), targetUserID, nil)
	if err != nil {
		writeError(w, http.StatusNotFound, "target user not found")
		return
	}

	// Security Rule: Platform Owner cannot be demoted through admin CRUD
	if strings.ToLower(strings.TrimSpace(targetUser.Email)) == "afrizaahmad18@gmail.com" {
		writeError(w, http.StatusForbidden, "owner account roles cannot be modified through standard administrative operations")
		return
	}

	var req UpdateUserRolesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	// Fetch current roles to clean up or replace
	currentRoles, _ := h.repo.GetUserRoles(r.Context(), targetUserID)
	for _, cr := range currentRoles {
		_ = h.repo.RemoveRoleFromUser(r.Context(), targetUserID, cr.ID)
	}

	for _, newRoleID := range req.RoleIDs {
		_ = h.repo.AssignRoleToUser(r.Context(), targetUserID, newRoleID, &actor.ID, req.Reason)
	}

	// Audit Log
	_ = h.repo.CreateAuditLog(r.Context(), &domain.AuditLog{
		ActorUserID: &actor.ID,
		Action:      "user.roles.update",
		TargetType:  "user",
		TargetID:    targetUserID.String(),
		Reason:      req.Reason,
		MetadataRedacted: map[string]interface{}{
			"target_username": targetUser.Username,
			"target_email":    targetUser.Email,
			"assigned_roles":  req.RoleIDs,
		},
	})

	updatedRoles, _ := h.repo.GetUserRoles(r.Context(), targetUserID)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message": "user roles updated successfully",
		"roles":   updatedRoles,
	})
}

type UpdateUserStatusRequest struct {
	Status string `json:"status"` // 'active', 'suspended'
	Reason string `json:"reason"`
}

func (h *Handler) AdminUpdateUserStatus(w http.ResponseWriter, r *http.Request) {
	actor := middleware.GetUserFromContext(r.Context())
	targetUserID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	if actor.ID == targetUserID {
		writeError(w, http.StatusForbidden, "cannot change your own account status")
		return
	}

	var req UpdateUserStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	if req.Status != "active" && req.Status != "suspended" {
		writeError(w, http.StatusBadRequest, "status must be 'active' or 'suspended'")
		return
	}

	if err := h.repo.UpdateUserStatus(r.Context(), targetUserID, req.Status); err != nil {
		if err == repository.ErrRoleProtected {
			writeError(w, http.StatusForbidden, "platform owner account cannot be suspended or deleted")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to update user status")
		return
	}

	// Audit Log
	_ = h.repo.CreateAuditLog(r.Context(), &domain.AuditLog{
		ActorUserID: &actor.ID,
		Action:      "user.status.update",
		TargetType:  "user",
		TargetID:    targetUserID.String(),
		Reason:      req.Reason,
		MetadataRedacted: map[string]interface{}{
			"new_status": req.Status,
		},
	})

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message": fmt.Sprintf("user status set to %s", req.Status),
	})
}

type AssignMembershipRequest struct {
	PlanID    string     `json:"plan_id"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	Reason    string     `json:"reason"`
}

func (h *Handler) AdminAssignUserMembership(w http.ResponseWriter, r *http.Request) {
	actor := middleware.GetUserFromContext(r.Context())
	targetUserID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	var req AssignMembershipRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	plan, err := h.repo.GetMembershipPlanByID(r.Context(), req.PlanID)
	if err != nil {
		writeError(w, http.StatusNotFound, "membership plan not found")
		return
	}

	userMem := &domain.UserMembership{
		UserID:    targetUserID,
		PlanID:    plan.ID,
		Status:    "active",
		StartsAt:  time.Now(),
		ExpiresAt: req.ExpiresAt,
	}

	if err := h.repo.AssignUserMembership(r.Context(), userMem); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to assign membership")
		return
	}

	// Audit Log
	_ = h.repo.CreateAuditLog(r.Context(), &domain.AuditLog{
		ActorUserID: &actor.ID,
		Action:      "membership.assign.manual",
		TargetType:  "user",
		TargetID:    targetUserID.String(),
		Reason:      req.Reason,
		MetadataRedacted: map[string]interface{}{
			"plan_id": plan.ID,
		},
	})

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message":    fmt.Sprintf("Membership %s manually assigned", plan.Name),
		"membership": userMem,
	})
}

func (h *Handler) AdminRevokeUserMembership(w http.ResponseWriter, r *http.Request) {
	actor := middleware.GetUserFromContext(r.Context())
	targetUserID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}
	planID := chi.URLParam(r, "planId")

	if err := h.repo.RevokeUserMembership(r.Context(), targetUserID, planID); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to revoke membership")
		return
	}

	// Audit Log
	_ = h.repo.CreateAuditLog(r.Context(), &domain.AuditLog{
		ActorUserID: &actor.ID,
		Action:      "membership.revoke.manual",
		TargetType:  "user",
		TargetID:    targetUserID.String(),
		Reason:      "Administrative membership revocation",
		MetadataRedacted: map[string]interface{}{
			"plan_id": planID,
		},
	})

	writeJSON(w, http.StatusOK, map[string]string{"message": "membership revoked"})
}

// ==========================================
// AUDIT LOGS (ADMIN)
// ==========================================

func (h *Handler) AdminGetAuditLogs(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	logs, err := h.repo.GetAuditLogs(r.Context(), limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch audit logs")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"audit_logs": logs,
		"limit":      limit,
		"offset":     offset,
	})
}

// ==========================================
// OVERVIEW & MONITORING (ADMIN)
// ==========================================

func (h *Handler) AdminGetOverview(w http.ResponseWriter, r *http.Request) {
	overview, err := h.repo.GetAdminOverview(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to generate admin overview")
		return
	}
	writeJSON(w, http.StatusOK, overview)
}

func (h *Handler) AdminGetMonitoringOverview(w http.ResponseWriter, r *http.Request) {
	overview, err := h.repo.GetAdminOverview(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get monitoring data")
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"system_health": overview.SystemHealth,
	})
}

func (h *Handler) AdminGetSecurityEvents(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	events, err := h.repo.GetSecurityAuditLogs(r.Context(), limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch security events")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"security_events": events,
		"limit":           limit,
		"offset":          offset,
	})
}

// ==========================================
// POST MANAGEMENT & MODERATION (ADMIN)
// ==========================================

func (h *Handler) AdminGetPosts(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	query := r.URL.Query().Get("query")
	authorID := r.URL.Query().Get("author_id")
	modStatus := r.URL.Query().Get("moderation_status")
	visibility := r.URL.Query().Get("visibility")

	filter := domain.AdminPostFilter{
		Query:            query,
		AuthorID:         authorID,
		ModerationStatus: modStatus,
		Visibility:       visibility,
		Limit:            limit,
		Offset:           offset,
	}

	posts, total, err := h.repo.GetAdminPosts(r.Context(), filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch posts")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"posts":  posts,
		"total":  total,
		"limit":  filter.Limit,
		"offset": filter.Offset,
	})
}

type UpdatePostModerationRequest struct {
	ModerationStatus string `json:"moderation_status"`
	Reason           string `json:"reason"`
}

func (h *Handler) AdminUpdatePostModeration(w http.ResponseWriter, r *http.Request) {
	actor := middleware.GetUserFromContext(r.Context())
	postID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid post id")
		return
	}

	var req UpdatePostModerationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	req.ModerationStatus = strings.ToLower(strings.TrimSpace(req.ModerationStatus))
	if req.ModerationStatus != "normal" && req.ModerationStatus != "flagged" && req.ModerationStatus != "hidden" && req.ModerationStatus != "removed" {
		writeError(w, http.StatusBadRequest, "invalid moderation status (allowed: normal, flagged, hidden, removed)")
		return
	}

	if err := h.repo.UpdatePostModeration(r.Context(), postID, req.ModerationStatus, actor.ID, req.Reason); err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to update post moderation: %v", err))
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message":           "post moderation status updated",
		"moderation_status": req.ModerationStatus,
	})
}

type AdminDeletePostRequest struct {
	Reason string `json:"reason"`
}

func (h *Handler) AdminDeletePost(w http.ResponseWriter, r *http.Request) {
	actor := middleware.GetUserFromContext(r.Context())
	postID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid post id")
		return
	}

	var req AdminDeletePostRequest
	_ = json.NewDecoder(r.Body).Decode(&req)
	if req.Reason == "" {
		req.Reason = "Administrative deletion"
	}

	if err := h.repo.AdminDeletePost(r.Context(), postID, actor.ID, req.Reason); err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to delete post: %v", err))
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "post deleted by administrator",
	})
}
