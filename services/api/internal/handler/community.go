package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/x-reconstruction/api/internal/domain"
	"github.com/x-reconstruction/api/internal/middleware"
)

// Community Request Payloads
type CreateCommunityRequest struct {
	Name        string                 `json:"name"`
	Slug        string                 `json:"slug"`
	Description string                 `json:"description"`
	AvatarURL   string                 `json:"avatar_url"`
	BannerURL   string                 `json:"banner_url"`
	AccessModel string                 `json:"access_model"` // "open", "restricted"
	Visibility  string                 `json:"visibility"`   // "public", "unlisted"
	Rules       []domain.CommunityRule `json:"rules"`
}

type UpdateCommunityRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	AvatarURL   string `json:"avatar_url"`
	BannerURL   string `json:"banner_url"`
	AccessModel string `json:"access_model"`
	Visibility  string `json:"visibility"`
}

type ReviewJoinRequestPayload struct {
	Status string `json:"status"` // "approved", "rejected"
	Note   string `json:"note"`
}

type UpdateMemberRolePayload struct {
	Role string `json:"role"` // "moderator", "member"
}

// GetCommunities lists communities for discovery or search
func (h *Handler) GetCommunities(w http.ResponseWriter, r *http.Request) {
	searchQuery := r.URL.Query().Get("q")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	if limit <= 0 || limit > 50 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	vID := middleware.GetViewerID(r.Context())

	communities, total, err := h.repo.GetCommunities(r.Context(), searchQuery, limit, offset, vID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to fetch communities")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"communities": communities,
		"total":       total,
		"limit":       limit,
		"offset":      offset,
	})
}

// GetUserCommunities returns communities joined by the current user
func (h *Handler) GetUserCommunities(w http.ResponseWriter, r *http.Request) {
	u := middleware.GetUserFromContext(r.Context())
	if u == nil {
		writeError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	communities, err := h.repo.GetUserCommunities(r.Context(), u.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to fetch user communities")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"communities": communities,
	})
}

// GetCommunityBySlug fetches a single community by slug
func (h *Handler) GetCommunityBySlug(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	if slug == "" {
		writeError(w, http.StatusBadRequest, "Community slug required")
		return
	}

	vID := middleware.GetViewerID(r.Context())

	comm, err := h.repo.GetCommunityBySlug(r.Context(), slug, vID)
	if err != nil {
		writeError(w, http.StatusNotFound, "Community not found")
		return
	}

	rules, _ := h.repo.GetCommunityRules(r.Context(), comm.ID)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"community": comm,
		"rules":     rules,
	})
}

// CreateCommunity creates a new community
func (h *Handler) CreateCommunity(w http.ResponseWriter, r *http.Request) {
	u := middleware.GetUserFromContext(r.Context())
	if u == nil {
		writeError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	var req CreateCommunityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	name := strings.TrimSpace(req.Name)
	slug := strings.TrimSpace(strings.ToLower(req.Slug))
	if name == "" || slug == "" {
		writeError(w, http.StatusBadRequest, "Community name and slug are required")
		return
	}

	accessModel := req.AccessModel
	if accessModel != "restricted" {
		accessModel = "open"
	}

	visibility := req.Visibility
	if visibility != "unlisted" {
		visibility = "public"
	}

	comm := &domain.Community{
		ID:          uuid.New(),
		Slug:        slug,
		Name:        name,
		Description: req.Description,
		AvatarURL:   req.AvatarURL,
		BannerURL:   req.BannerURL,
		AccessModel: accessModel,
		Visibility:  visibility,
		OwnerID:     u.ID,
	}

	if err := h.repo.CreateCommunity(r.Context(), comm, req.Rules); err != nil {
		writeError(w, http.StatusConflict, "Community slug already exists or creation failed")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"community": comm,
		"message":   "Community created successfully",
	})
}

// UpdateCommunity updates community metadata (owner or moderator only)
func (h *Handler) UpdateCommunity(w http.ResponseWriter, r *http.Request) {
	u := middleware.GetUserFromContext(r.Context())
	if u == nil {
		writeError(w, http.StatusUnauthorized, "Authentication required")
		return
	}
	slug := chi.URLParam(r, "slug")

	comm, err := h.repo.GetCommunityBySlug(r.Context(), slug, &u.ID)
	if err != nil {
		writeError(w, http.StatusNotFound, "Community not found")
		return
	}

	if comm.ViewerRole != "owner" && comm.ViewerRole != "admin" && comm.ViewerRole != "moderator" {
		writeError(w, http.StatusForbidden, "Only moderators or owner can edit community")
		return
	}

	var req UpdateCommunityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if strings.TrimSpace(req.Name) != "" {
		comm.Name = strings.TrimSpace(req.Name)
	}
	comm.Description = req.Description
	comm.AvatarURL = req.AvatarURL
	comm.BannerURL = req.BannerURL
	if req.AccessModel == "open" || req.AccessModel == "restricted" {
		comm.AccessModel = req.AccessModel
	}
	if req.Visibility == "public" || req.Visibility == "unlisted" {
		comm.Visibility = req.Visibility
	}

	if err := h.repo.UpdateCommunity(r.Context(), comm); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to update community")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"community": comm,
		"message":   "Community updated successfully",
	})
}

// JoinCommunity handles joining an open community or requesting to join a restricted community
func (h *Handler) JoinCommunity(w http.ResponseWriter, r *http.Request) {
	u := middleware.GetUserFromContext(r.Context())
	if u == nil {
		writeError(w, http.StatusUnauthorized, "Authentication required")
		return
	}
	slug := chi.URLParam(r, "slug")

	comm, err := h.repo.GetCommunityBySlug(r.Context(), slug, &u.ID)
	if err != nil {
		writeError(w, http.StatusNotFound, "Community not found")
		return
	}

	if comm.ViewerRole != "" && comm.ViewerStatus == "active" {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"message": "Already a member of this community",
			"status":  "active",
		})
		return
	}

	// Restricted access model requires approval
	if comm.AccessModel == "restricted" {
		joinReq := &domain.CommunityJoinRequest{
			CommunityID: comm.ID,
			UserID:      u.ID,
		}
		if err := h.repo.CreateCommunityJoinRequest(r.Context(), joinReq); err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to submit join request")
			return
		}

		writeJSON(w, http.StatusAccepted, map[string]interface{}{
			"message": "Membership request submitted for review",
			"status":  "pending",
		})
		return
	}

	// Open community -> immediate join
	member := &domain.CommunityMember{
		CommunityID: comm.ID,
		UserID:      u.ID,
		Role:        "member",
		Status:      "active",
	}
	if err := h.repo.AddCommunityMember(r.Context(), member); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to join community")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Joined community successfully",
		"status":  "active",
	})
}

// LeaveCommunity leaves a community
func (h *Handler) LeaveCommunity(w http.ResponseWriter, r *http.Request) {
	u := middleware.GetUserFromContext(r.Context())
	if u == nil {
		writeError(w, http.StatusUnauthorized, "Authentication required")
		return
	}
	slug := chi.URLParam(r, "slug")

	comm, err := h.repo.GetCommunityBySlug(r.Context(), slug, &u.ID)
	if err != nil {
		writeError(w, http.StatusNotFound, "Community not found")
		return
	}

	if comm.OwnerID == u.ID {
		writeError(w, http.StatusBadRequest, "Owner cannot leave community directly. Transfer ownership first.")
		return
	}

	if err := h.repo.RemoveCommunityMember(r.Context(), comm.ID, u.ID); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to leave community")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Successfully left the community",
	})
}

// GetCommunityMembers lists members of the community
func (h *Handler) GetCommunityMembers(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	if limit <= 0 || limit > 50 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	comm, err := h.repo.GetCommunityBySlug(r.Context(), slug, nil)
	if err != nil {
		writeError(w, http.StatusNotFound, "Community not found")
		return
	}

	members, total, err := h.repo.GetCommunityMembers(r.Context(), comm.ID, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to fetch members")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"members": members,
		"total":   total,
		"limit":   limit,
		"offset":  offset,
	})
}

// GetCommunityPosts returns timeline posts linked to this community
func (h *Handler) GetCommunityPosts(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	if limit <= 0 || limit > 50 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	vID := middleware.GetViewerID(r.Context())

	comm, err := h.repo.GetCommunityBySlug(r.Context(), slug, vID)
	if err != nil {
		writeError(w, http.StatusNotFound, "Community not found")
		return
	}

	posts, err := h.repo.GetCommunityPosts(r.Context(), comm.ID, limit, offset, vID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to fetch community posts")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"posts":  posts,
		"limit":  limit,
		"offset": offset,
	})
}

// GetCommunityJoinRequests returns pending membership requests for moderators
func (h *Handler) GetCommunityJoinRequests(w http.ResponseWriter, r *http.Request) {
	u := middleware.GetUserFromContext(r.Context())
	if u == nil {
		writeError(w, http.StatusUnauthorized, "Authentication required")
		return
	}
	slug := chi.URLParam(r, "slug")

	comm, err := h.repo.GetCommunityBySlug(r.Context(), slug, &u.ID)
	if err != nil {
		writeError(w, http.StatusNotFound, "Community not found")
		return
	}

	if comm.ViewerRole != "owner" && comm.ViewerRole != "admin" && comm.ViewerRole != "moderator" {
		writeError(w, http.StatusForbidden, "Moderator access required")
		return
	}

	reqs, err := h.repo.GetCommunityJoinRequests(r.Context(), comm.ID, "pending", 50, 0)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to fetch join requests")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"requests": reqs,
	})
}

// ReviewCommunityJoinRequest approves or rejects a join request
func (h *Handler) ReviewCommunityJoinRequest(w http.ResponseWriter, r *http.Request) {
	u := middleware.GetUserFromContext(r.Context())
	if u == nil {
		writeError(w, http.StatusUnauthorized, "Authentication required")
		return
	}
	slug := chi.URLParam(r, "slug")
	requestIDStr := chi.URLParam(r, "requestId")

	reqID, err := uuid.Parse(requestIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request ID")
		return
	}

	comm, err := h.repo.GetCommunityBySlug(r.Context(), slug, &u.ID)
	if err != nil {
		writeError(w, http.StatusNotFound, "Community not found")
		return
	}

	if comm.ViewerRole != "owner" && comm.ViewerRole != "admin" && comm.ViewerRole != "moderator" {
		writeError(w, http.StatusForbidden, "Moderator access required")
		return
	}

	var payload ReviewJoinRequestPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if payload.Status != "approved" && payload.Status != "rejected" {
		writeError(w, http.StatusBadRequest, "Status must be approved or rejected")
		return
	}

	if err := h.repo.ReviewCommunityJoinRequest(r.Context(), reqID, u.ID, payload.Status, payload.Note); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to review join request")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Join request reviewed successfully",
		"status":  payload.Status,
	})
}
