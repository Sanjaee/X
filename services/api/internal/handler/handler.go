package handler

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/redis/go-redis/v9"
	"github.com/x-reconstruction/api/internal/domain"
	"github.com/x-reconstruction/api/internal/media"
	"github.com/x-reconstruction/api/internal/middleware"
	"github.com/x-reconstruction/api/internal/payment"
	"github.com/x-reconstruction/api/internal/realtime"
	"github.com/x-reconstruction/api/internal/repository"
	"github.com/x-reconstruction/api/internal/search"
)

type Handler struct {
	repo       repository.Repository
	storageDir string
	mediaSvc   *media.MediaService
	plisioSvc  *payment.PlisioService
	hub        *realtime.Hub
	rdb        *redis.Client
}

func NewHandler(repo repository.Repository, storageDir string, mediaSvc *media.MediaService, plisioSvc *payment.PlisioService, hub *realtime.Hub, rdb *redis.Client) *Handler {
	_ = os.MkdirAll(storageDir, 0755)
	return &Handler{
		repo:       repo,
		storageDir: storageDir,
		mediaSvc:   mediaSvc,
		plisioSvc:  plisioSvc,
		hub:        hub,
		rdb:        rdb,
	}
}

// Helpers
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		_ = json.NewEncoder(w).Encode(data)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

// ==========================================
// HEALTH
// ==========================================
func (h *Handler) HealthLive(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "alive"})
}

func (h *Handler) HealthReady(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	dbStatus := "connected"
	if err := h.repo.Ping(ctx); err != nil {
		writeError(w, http.StatusServiceUnavailable, fmt.Sprintf("database unavailable: %v", err))
		return
	}

	redisStatus := "disabled"
	if h.rdb != nil {
		if err := h.rdb.Ping(ctx).Err(); err != nil {
			redisStatus = fmt.Sprintf("degraded: %v", err)
		} else {
			redisStatus = "connected"
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":   "ready",
		"database": dbStatus,
		"redis":    redisStatus,
		"time":     time.Now().UTC().Format(time.RFC3339),
	})
}

func (h *Handler) HealthMetrics(w http.ResponseWriter, r *http.Request) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":           "ok",
		"active_requests":  middleware.GetActiveRequests(),
		"goroutines":       runtime.NumGoroutine(),
		"alloc_bytes":      m.Alloc,
		"alloc_mb":         fmt.Sprintf("%.2f MB", float64(m.Alloc)/1024/1024),
		"total_alloc_mb":   fmt.Sprintf("%.2f MB", float64(m.TotalAlloc)/1024/1024),
		"sys_mb":           fmt.Sprintf("%.2f MB", float64(m.Sys)/1024/1024),
		"num_gc":           m.NumGC,
		"gc_pause_total_ms": fmt.Sprintf("%.2f ms", float64(m.PauseTotalNs)/1e6),
		"timestamp":        time.Now().UTC().Format(time.RFC3339),
	})
}

// ==========================================
// AUTHENTICATION
// ==========================================
type RegisterRequest struct {
	Username    string `json:"username"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	req.Username = strings.TrimSpace(req.Username)
	req.Email = strings.TrimSpace(req.Email)
	if len(req.Username) < 3 || len(req.Password) < 8 || !strings.Contains(req.Email, "@") {
		writeError(w, http.StatusBadRequest, "username min 3 chars, password min 8 chars, valid email required")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to hash password")
		return
	}

	dispName := req.DisplayName
	if dispName == "" {
		dispName = req.Username
	}

	user := &domain.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(hash),
		DisplayName:  dispName,
		AvatarURL:    "https://abs.twimg.com/sticky/default_profile_images/default_profile_400x400.png",
	}

	if err := h.repo.CreateUser(r.Context(), user); err != nil {
		if errors.Is(err, repository.ErrAlreadyExists) {
			writeError(w, http.StatusConflict, "username or email already in use")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to create user")
		return
	}

	// Create session
	token, err := h.createSessionForUser(r, user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to generate session")
		return
	}

	h.setSessionCookie(w, token)
	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"user":          user,
		"session_token": token,
	})
}

type LoginRequest struct {
	Identifier string `json:"identifier"`
	Password   string `json:"password"`
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	user, err := h.repo.GetUserByIdentifier(r.Context(), strings.TrimSpace(req.Identifier))
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	// Secure, idempotent Owner bootstrap for afrizaahmad18@gmail.com
	if strings.ToLower(strings.TrimSpace(user.Email)) == "afrizaahmad18@gmail.com" {
		_, _ = h.repo.BootstrapOwner(r.Context(), user.Email)
		user.Role = "owner"
	}

	token, err := h.createSessionForUser(r, user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to generate session")
		return
	}

	h.setSessionCookie(w, token)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"user":          user,
		"session_token": token,
	})
}

type GoogleAuthRequest struct {
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
	AvatarURL   string `json:"avatar_url"`
	GoogleID    string `json:"google_id"`
}

func (h *Handler) GoogleAuth(w http.ResponseWriter, r *http.Request) {
	var req GoogleAuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	email := strings.TrimSpace(strings.ToLower(req.Email))
	if email == "" || !strings.Contains(email, "@") {
		writeError(w, http.StatusBadRequest, "valid email required from Google profile")
		return
	}

	// 1. Check if user already exists by email
	existingUser, err := h.repo.GetUserByIdentifier(r.Context(), email)
	if err == nil && existingUser != nil {
		// Update avatar if currently empty or default twitter egg
		if req.AvatarURL != "" && (existingUser.AvatarURL == "" || strings.Contains(existingUser.AvatarURL, "default_profile")) {
			existingUser.AvatarURL = req.AvatarURL
			_ = h.repo.UpdateProfile(r.Context(), existingUser)
		}

		// Secure, idempotent Owner bootstrap for afrizaahmad18@gmail.com
		if email == "afrizaahmad18@gmail.com" {
			_, _ = h.repo.BootstrapOwner(r.Context(), email)
			existingUser.Role = "owner"
		}

		token, err := h.createSessionForUser(r, existingUser.ID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to generate session")
			return
		}

		h.setSessionCookie(w, token)
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"user":          existingUser,
			"session_token": token,
		})
		return
	}

	// 2. User does not exist -> Create new user with Google details
	parts := strings.Split(email, "@")
	baseUsername := parts[0]
	var clean strings.Builder
	for _, ch := range baseUsername {
		if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '_' {
			clean.WriteRune(ch)
		}
	}
	uname := strings.ToLower(clean.String())
	if len(uname) < 3 {
		uname = "user"
	}
	if len(uname) > 15 {
		uname = uname[:15]
	}

	finalUsername := uname
	for i := 0; i < 5; i++ {
		_, err := h.repo.GetUserByUsername(r.Context(), finalUsername, nil)
		if err != nil {
			break
		}
		finalUsername = fmt.Sprintf("%s%d", uname, time.Now().UnixNano()%10000)
	}

	dispName := strings.TrimSpace(req.DisplayName)
	if dispName == "" {
		dispName = finalUsername
	}

	avatar := req.AvatarURL
	if avatar == "" {
		avatar = "https://abs.twimg.com/sticky/default_profile_images/default_profile_400x400.png"
	}

	randomPass := make([]byte, 16)
	_, _ = rand.Read(randomPass)
	hash, _ := bcrypt.GenerateFromPassword(randomPass, bcrypt.DefaultCost)

	newUser := &domain.User{
		Username:     finalUsername,
		Email:        email,
		PasswordHash: string(hash),
		DisplayName:  dispName,
		AvatarURL:    avatar,
	}

	if err := h.repo.CreateUser(r.Context(), newUser); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to register user from Google profile")
		return
	}

	// Secure, idempotent Owner bootstrap for afrizaahmad18@gmail.com
	if email == "afrizaahmad18@gmail.com" {
		_, _ = h.repo.BootstrapOwner(r.Context(), email)
		newUser.Role = "owner"
	}

	token, err := h.createSessionForUser(r, newUser.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to generate session")
		return
	}

	h.setSessionCookie(w, token)
	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"user":          newUser,
		"session_token": token,
	})
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie("session_token"); err == nil && cookie.Value != "" {
		hash := sha256.Sum256([]byte(cookie.Value))
		_ = h.repo.DeleteSession(r.Context(), hex.EncodeToString(hash[:]))
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})
	writeJSON(w, http.StatusOK, map[string]string{"message": "logged out"})
}

func (h *Handler) GetMe(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	// Idempotent owner bootstrap check for afrizaahmad18@gmail.com
	if strings.ToLower(strings.TrimSpace(user.Email)) == "afrizaahmad18@gmail.com" {
		_, _ = h.repo.BootstrapOwner(r.Context(), user.Email)
		user.Role = "owner"
	}

	roles, _ := h.repo.GetUserRoles(r.Context(), user.ID)
	perms, _ := h.repo.GetUserPermissions(r.Context(), user.ID)
	membership, _ := h.repo.GetUserMembership(r.Context(), user.ID)
	entitlements, _ := h.repo.GetResolvedEntitlements(r.Context(), user.ID)

	response := map[string]interface{}{
		"id":              user.ID,
		"username":        user.Username,
		"email":           user.Email,
		"role":            user.Role,
		"is_verified":     user.IsVerified,
		"verified_type":   user.VerifiedType,
		"is_private":      user.IsPrivate,
		"created_at":      user.CreatedAt,
		"updated_at":      user.UpdatedAt,
		"display_name":    user.DisplayName,
		"bio":             user.Bio,
		"avatar_url":      user.AvatarURL,
		"banner_url":      user.BannerURL,
		"location":        user.Location,
		"website":         user.Website,
		"followers_count": user.FollowersCount,
		"following_count": user.FollowingCount,
		"posts_count":     user.PostsCount,
		"roles":           roles,
		"permissions":     perms,
		"membership":      membership,
		"entitlements":    entitlements,
		"is_owner":        middleware.IsOwnerAccount(user),
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *Handler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req struct {
		DisplayName string `json:"display_name"`
		Bio         string `json:"bio"`
		AvatarURL   string `json:"avatar_url"`
		BannerURL   string `json:"banner_url"`
		Location    string `json:"location"`
		Website     string `json:"website"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid payload")
		return
	}

	if req.DisplayName != "" {
		user.DisplayName = req.DisplayName
	}
	user.Bio = req.Bio
	if req.AvatarURL != "" {
		user.AvatarURL = req.AvatarURL
	}
	if req.BannerURL != "" {
		user.BannerURL = req.BannerURL
	}
	user.Location = req.Location

	cleanWebsite := strings.TrimSpace(req.Website)
	if cleanWebsite != "" {
		lower := strings.ToLower(cleanWebsite)
		if strings.HasPrefix(lower, "javascript:") || strings.HasPrefix(lower, "data:") || strings.HasPrefix(lower, "vbscript:") {
			cleanWebsite = ""
		} else if !strings.HasPrefix(lower, "http://") && !strings.HasPrefix(lower, "https://") {
			cleanWebsite = "https://" + cleanWebsite
		}
	}
	user.Website = cleanWebsite

	if err := h.repo.UpdateProfile(r.Context(), user); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update profile")
		return
	}
	writeJSON(w, http.StatusOK, user)
}

func (h *Handler) GetUserByUsername(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")
	viewerID := middleware.GetViewerID(r.Context())

	user, err := h.repo.GetUserByUsername(r.Context(), username, viewerID)
	if err != nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}
	writeJSON(w, http.StatusOK, user)
}

// Session helpers
func (h *Handler) createSessionForUser(r *http.Request, userID uuid.UUID) (string, error) {
	rawBytes := make([]byte, 32)
	if _, err := rand.Read(rawBytes); err != nil {
		return "", err
	}
	token := hex.EncodeToString(rawBytes)

	hash := sha256.Sum256([]byte(token))
	tokenHash := hex.EncodeToString(hash[:])

	session := &domain.Session{
		UserID:    userID,
		TokenHash: tokenHash,
		IPAddress: r.RemoteAddr,
		UserAgent: r.UserAgent(),
		ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
	}
	return token, h.repo.CreateSession(r.Context(), session)
}

func (h *Handler) setSessionCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    token,
		Path:     "/",
		MaxAge:   int((30 * 24 * time.Hour).Seconds()),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
}

// ==========================================
// POSTS & FEED
// ==========================================
type CreatePostRequest struct {
	Content     string               `json:"content"`
	MediaIDs    []uuid.UUID          `json:"media_ids"`
	Media       []domain.MediaObject `json:"media"`
	ReplyToID   *uuid.UUID           `json:"reply_to_id"`
	QuotePostID *uuid.UUID           `json:"quote_post_id"`
}

func (h *Handler) CreatePost(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	// 1. Verify post.create permission
	hasCreatePerm, _ := h.repo.HasPermission(r.Context(), user.ID, "post.create")
	if !hasCreatePerm && !middleware.IsOwnerAccount(user) {
		writeError(w, http.StatusForbidden, "permission denied: post.create required")
		return
	}

	var req CreatePostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}

	req.Content = strings.TrimSpace(req.Content)
	if req.Content == "" && len(req.MediaIDs) == 0 && len(req.Media) == 0 {
		writeError(w, http.StatusBadRequest, "content or media required")
		return
	}

	// 2. Resolve entitlements & limits from database
	entitlements, _ := h.repo.GetResolvedEntitlements(r.Context(), user.ID)
	maxPosts := 2
	maxImages := 2
	maxVideos := 0
	maxLen := 280

	if v, ok := entitlements["posts_per_day"]; ok {
		maxPosts = int(v)
	}
	if v, ok := entitlements["max_images_per_post"]; ok {
		maxImages = int(v)
	}
	if v, ok := entitlements["max_videos_per_post"]; ok {
		maxVideos = int(v)
	}
	if v, ok := entitlements["max_post_length"]; ok && v > 0 {
		maxLen = int(v)
	}

	// Platform Owner has unlimited posting quota
	if middleware.IsOwnerAccount(user) {
		maxPosts = 0
		maxImages = 10
		maxVideos = 5
		maxLen = 10000
	}

	if len([]rune(req.Content)) > maxLen {
		writeError(w, http.StatusForbidden, fmt.Sprintf("content exceeds maximum limit of %d characters for your plan. Upgrade your membership for longer posts.", maxLen))
		return
	}

	post := &domain.Post{
		AuthorID:         user.ID,
		Author:           user,
		Content:          req.Content,
		Visibility:       "public",
		ModerationStatus: "active",
		ReplyToID:        req.ReplyToID,
		QuotePostID:      req.QuotePostID,
	}

	imageCount := 0
	videoCount := 0

	for _, mID := range req.MediaIDs {
		if media, err := h.repo.GetMediaObjectByID(r.Context(), mID); err == nil {
			if media.MediaType == "video" {
				videoCount++
			} else {
				imageCount++
			}
			post.Media = append(post.Media, *media)
		}
	}

	for _, m := range req.Media {
		if m.ID == uuid.Nil {
			m.ID = uuid.New()
		}
		m.UploaderID = user.ID
		if m.StorageKey == "" {
			m.StorageKey = fmt.Sprintf("media/%s/%s", user.ID, m.ID)
		}
		if m.MediaType == "" {
			m.MediaType = "image"
		}
		if m.MimeType == "" {
			if m.MediaType == "video" {
				m.MimeType = "video/mp4"
			} else {
				m.MimeType = "image/jpeg"
			}
		}
		if m.FileSizeBytes == 0 {
			m.FileSizeBytes = 1024
		}
		if m.CreatedAt.IsZero() {
			m.CreatedAt = time.Now()
		}
		if m.MediaType == "video" {
			videoCount++
		} else {
			imageCount++
		}
		if err := h.repo.CreateMediaObject(r.Context(), &m); err != nil {
			slog.Error("failed to create inline media object", "error", err)
			continue
		}
		post.Media = append(post.Media, m)
	}

	// 3. Validate media count limits against database tier entitlements
	if imageCount > maxImages {
		writeError(w, http.StatusForbidden, fmt.Sprintf("Image limit exceeded (%d images attached, maximum allowed is %d for your membership tier).", imageCount, maxImages))
		return
	}
	if videoCount > maxVideos {
		writeError(w, http.StatusForbidden, fmt.Sprintf("Video upload limit exceeded (%d videos attached, maximum allowed is %d for your membership tier). Upgrade to VIP or MVP for video capability.", videoCount, maxVideos))
		return
	}

	// 4. Atomic Quota Engine: Concurrency-safe limit enforcement
	if maxPosts > 0 {
		allowed, currentCount, err := h.repo.CheckAndIncrementPostQuota(r.Context(), user.ID, maxPosts)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to verify posting quota")
			return
		}
		if !allowed {
			writeError(w, http.StatusTooManyRequests, fmt.Sprintf("Daily posting limit reached (%d/%d posts today). Upgrade your membership tier for more posts.", currentCount, maxPosts))
			return
		}
	}

	if err := h.repo.CreatePost(r.Context(), post); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create post")
		return
	}

	// If this was a reply, notify author of parent
	if post.ReplyToID != nil {
		if parent, err := h.repo.GetPostByID(r.Context(), *post.ReplyToID, nil); err == nil && parent.AuthorID != user.ID {
			h.notifyUser(r.Context(), parent.AuthorID, user.ID, "reply", "post", &post.ID, &post.ID, map[string]interface{}{
				"post_id": post.ID.String(),
				"content": post.Content,
			})
		}
	}

	writeJSON(w, http.StatusCreated, post)
}

type UpdatePostRequest struct {
	Content string `json:"content"`
}

func (h *Handler) PatchPost(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	postID, err := uuid.Parse(chi.URLParam(r, "postId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid post id")
		return
	}

	var req UpdatePostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	req.Content = strings.TrimSpace(req.Content)
	if req.Content == "" {
		writeError(w, http.StatusBadRequest, "content cannot be empty")
		return
	}

	existingPost, err := h.repo.GetPostByID(r.Context(), postID, nil)
	if err != nil || existingPost == nil {
		writeError(w, http.StatusNotFound, "post not found")
		return
	}

	canUpdateAny, _ := h.repo.HasPermission(r.Context(), user.ID, "post.update.any")
	isOwner := middleware.IsOwnerAccount(user)

	// Ownership check: users can only edit their own posts unless administrative override
	if existingPost.AuthorID != user.ID && !canUpdateAny && !isOwner {
		writeError(w, http.StatusForbidden, "cannot edit another user's post")
		return
	}

	entitlements, _ := h.repo.GetResolvedEntitlements(r.Context(), user.ID)
	maxLen := 280
	if v, ok := entitlements["max_post_length"]; ok && v > 0 {
		maxLen = int(v)
	}
	if isOwner {
		maxLen = 10000
	}
	if len([]rune(req.Content)) > maxLen {
		writeError(w, http.StatusForbidden, fmt.Sprintf("content exceeds maximum limit of %d characters for your plan", maxLen))
		return
	}

	// Update content without consuming posting quota
	if err := h.repo.UpdatePost(r.Context(), postID, existingPost.AuthorID, req.Content); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update post")
		return
	}

	if existingPost.AuthorID != user.ID {
		_ = h.repo.CreateAuditLog(r.Context(), &domain.AuditLog{
			ActorUserID: &user.ID,
			Action:      "post.update.admin",
			TargetType:  "post",
			TargetID:    postID.String(),
			Reason:      "Administrative post edit",
			MetadataRedacted: map[string]interface{}{
				"author_id": existingPost.AuthorID.String(),
			},
		})
	}

	existingPost.Content = req.Content
	existingPost.UpdatedAt = time.Now()
	writeJSON(w, http.StatusOK, existingPost)
}

func (h *Handler) GetPostByID(w http.ResponseWriter, r *http.Request) {
	postIDStr := chi.URLParam(r, "postId")
	postID, err := uuid.Parse(postIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid post id")
		return
	}

	viewerID := middleware.GetViewerID(r.Context())
	post, err := h.repo.GetPostByID(r.Context(), postID, viewerID)
	if err != nil {
		writeError(w, http.StatusNotFound, "post not found")
		return
	}

	writeJSON(w, http.StatusOK, post)
}

func (h *Handler) DeletePost(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	postID, err := uuid.Parse(chi.URLParam(r, "postId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid post id")
		return
	}

	existingPost, err := h.repo.GetPostByID(r.Context(), postID, nil)
	if err != nil || existingPost == nil {
		writeError(w, http.StatusNotFound, "post not found")
		return
	}

	canDeleteAny, _ := h.repo.HasPermission(r.Context(), user.ID, "post.delete.any")
	isOwner := middleware.IsOwnerAccount(user)

	// Ownership check: regular users can only delete their own posts
	if existingPost.AuthorID != user.ID && !canDeleteAny && !isOwner {
		writeError(w, http.StatusForbidden, "cannot delete another user's post")
		return
	}

	if err := h.repo.DeletePostWithPermission(r.Context(), postID, user.ID, canDeleteAny || isOwner); err != nil {
		writeError(w, http.StatusForbidden, "cannot delete post")
		return
	}

	if existingPost.AuthorID != user.ID {
		_ = h.repo.CreateAuditLog(r.Context(), &domain.AuditLog{
			ActorUserID: &user.ID,
			Action:      "post.delete.moderation",
			TargetType:  "post",
			TargetID:    postID.String(),
			Reason:      "Moderation or administrative post deletion",
			MetadataRedacted: map[string]interface{}{
				"author_id": existingPost.AuthorID.String(),
			},
		})
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) GetFeed(w http.ResponseWriter, r *http.Request) {
	feedType := r.URL.Query().Get("type")
	if feedType == "" {
		feedType = "for_you"
	}
	cursor := r.URL.Query().Get("cursor")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 20
	}

	viewerID := middleware.GetViewerID(r.Context())
	posts, nextCursor, err := h.repo.GetFeed(r.Context(), feedType, viewerID, cursor, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch feed")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"posts":       posts,
		"next_cursor": nextCursor,
	})
}

// ==========================================
// RECOMMENDATION, FEEDBACK & TRENDING
// ==========================================

type FeedFeedbackRequest struct {
	PostID       uuid.UUID `json:"post_id"`
	FeedbackType string    `json:"feedback_type"` // 'hide', 'not_interested', 'show_less', 'report'
	Topic        string    `json:"topic,omitempty"`
}

func (h *Handler) RecordFeedFeedback(w http.ResponseWriter, r *http.Request) {
	viewerID := middleware.GetViewerID(r.Context())
	if viewerID == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req FeedFeedbackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.PostID == uuid.Nil || req.FeedbackType == "" {
		writeError(w, http.StatusBadRequest, "post_id and feedback_type are required")
		return
	}

	fb := domain.FeedFeedback{
		ID:           uuid.New(),
		UserID:       *viewerID,
		PostID:       req.PostID,
		FeedbackType: req.FeedbackType,
		Topic:        req.Topic,
		CreatedAt:    time.Now(),
	}

	if err := h.repo.RecordFeedFeedback(r.Context(), fb); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to record feedback")
		return
	}

	// Emit transactional outbox event
	_ = h.repo.CreateOutboxEvent(r.Context(), &domain.OutboxEvent{
		EventType:     "feed.feedback.created",
		AggregateType: "feed_feedback",
		AggregateID:   fb.ID,
		Payload: map[string]interface{}{
			"user_id":       viewerID.String(),
			"post_id":       req.PostID.String(),
			"feedback_type": req.FeedbackType,
			"topic":         req.Topic,
		},
		Status: "pending",
	})

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":  "recorded",
		"post_id": req.PostID,
	})
}

type FeedImpressionsRequest struct {
	Impressions []struct {
		PostID      uuid.UUID `json:"post_id"`
		FeedType    string    `json:"feed_type"`
		DwellTimeMs int       `json:"dwell_time_ms"`
	} `json:"impressions"`
}

func (h *Handler) RecordFeedImpressions(w http.ResponseWriter, r *http.Request) {
	viewerID := middleware.GetViewerID(r.Context())

	var req FeedImpressionsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	domainImpressions := make([]domain.FeedImpression, 0, len(req.Impressions))
	for _, imp := range req.Impressions {
		if imp.PostID == uuid.Nil {
			continue
		}
		fType := imp.FeedType
		if fType == "" {
			fType = "for_you"
		}
		domainImpressions = append(domainImpressions, domain.FeedImpression{
			ID:          uuid.New(),
			UserID:      viewerID,
			PostID:      imp.PostID,
			FeedType:    fType,
			DwellTimeMs: imp.DwellTimeMs,
			CreatedAt:   time.Now(),
		})
	}

	if err := h.repo.RecordFeedImpressions(r.Context(), domainImpressions); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to record impressions")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status": "recorded",
		"count":  len(domainImpressions),
	})
}

func (h *Handler) GetTrendingTopics(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 10
	}

	topics, err := h.repo.GetTrendingTopics(r.Context(), limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch trending topics")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"trending": topics,
	})
}

func (h *Handler) GetTopicPreferences(w http.ResponseWriter, r *http.Request) {
	viewerID := middleware.GetViewerID(r.Context())
	if viewerID == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	prefs, err := h.repo.GetUserTopicPreferences(r.Context(), *viewerID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get topic preferences")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"preferences": prefs,
	})
}

func (h *Handler) UpdateTopicPreference(w http.ResponseWriter, r *http.Request) {
	viewerID := middleware.GetViewerID(r.Context())
	if viewerID == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req domain.UserTopicPreference
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.UserID = *viewerID
	if err := h.repo.UpdateUserTopicPreference(r.Context(), req); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update topic preference")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status": "updated",
	})
}

func (h *Handler) ResetPersonalization(w http.ResponseWriter, r *http.Request) {
	viewerID := middleware.GetViewerID(r.Context())
	if viewerID == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	if err := h.repo.ResetPersonalization(r.Context(), *viewerID); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to reset personalization")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":  "reset_successful",
		"message": "personalization history and feedback successfully cleared",
	})
}

func (h *Handler) GetUserPosts(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")
	tab := r.URL.Query().Get("tab")
	if tab == "" {
		tab = "posts"
	}
	cursor := r.URL.Query().Get("cursor")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 20
	}

	viewerID := middleware.GetViewerID(r.Context())
	posts, nextCursor, err := h.repo.GetUserPosts(r.Context(), username, tab, viewerID, cursor, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch user posts")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"posts":       posts,
		"next_cursor": nextCursor,
	})
}

func (h *Handler) GetPostReplies(w http.ResponseWriter, r *http.Request) {
	postID, err := uuid.Parse(chi.URLParam(r, "postId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid post id")
		return
	}

	cursor := r.URL.Query().Get("cursor")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 20
	}

	viewerID := middleware.GetViewerID(r.Context())
	posts, nextCursor, err := h.repo.GetPostReplies(r.Context(), postID, viewerID, cursor, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch replies")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"posts":       posts,
		"next_cursor": nextCursor,
	})
}

// ==========================================
// INTERACTIONS (LIKE, REPOST, BOOKMARK, FOLLOW)
// ==========================================
func (h *Handler) LikePost(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	postID, err := uuid.Parse(chi.URLParam(r, "postId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid post id")
		return
	}

	status, count, err := h.repo.LikePost(r.Context(), postID, user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to like post")
		return
	}

	// Trigger notification
	if p, err := h.repo.GetPostByID(r.Context(), postID, nil); err == nil && p.AuthorID != user.ID {
		h.notifyUser(r.Context(), p.AuthorID, user.ID, "like", "post", &postID, &postID, map[string]interface{}{
			"post_id": postID.String(),
		})
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"status": status, "count": count})
}

func (h *Handler) UnlikePost(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	postID, err := uuid.Parse(chi.URLParam(r, "postId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid post id")
		return
	}

	status, count, err := h.repo.UnlikePost(r.Context(), postID, user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to unlike post")
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"status": status, "count": count})
}

func (h *Handler) RepostPost(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	postID, err := uuid.Parse(chi.URLParam(r, "postId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid post id")
		return
	}

	status, count, err := h.repo.RepostPost(r.Context(), postID, user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to repost")
		return
	}

	if p, err := h.repo.GetPostByID(r.Context(), postID, nil); err == nil && p.AuthorID != user.ID {
		h.notifyUser(r.Context(), p.AuthorID, user.ID, "repost", "post", &postID, &postID, map[string]interface{}{
			"post_id": postID.String(),
		})
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"status": status, "count": count})
}

func (h *Handler) UndoRepost(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	postID, err := uuid.Parse(chi.URLParam(r, "postId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid post id")
		return
	}

	status, count, err := h.repo.UndoRepost(r.Context(), postID, user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to undo repost")
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"status": status, "count": count})
}

func (h *Handler) BookmarkPost(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	postID, err := uuid.Parse(chi.URLParam(r, "postId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid post id")
		return
	}

	status, count, err := h.repo.BookmarkPost(r.Context(), postID, user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to bookmark")
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"status": status, "count": count})
}

func (h *Handler) RemoveBookmark(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	postID, err := uuid.Parse(chi.URLParam(r, "postId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid post id")
		return
	}

	status, count, err := h.repo.RemoveBookmark(r.Context(), postID, user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to remove bookmark")
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"status": status, "count": count})
}

// notifyUser creates a persisted notification and broadcasts real-time events over WebSocket
func (h *Handler) notifyUser(ctx context.Context, recipientID, actorID uuid.UUID, notifType, entityType string, entityID *uuid.UUID, postID *uuid.UUID, extra map[string]interface{}) {
	if recipientID == actorID {
		return
	}

	_ = h.repo.CreateNotification(ctx, &domain.Notification{
		RecipientID: recipientID,
		ActorID:     actorID,
		Type:        notifType,
		EntityType:  entityType,
		EntityID:    entityID,
		PostID:      postID,
	})

	if h.hub != nil {
		payload := map[string]interface{}{
			"recipient_id": recipientID.String(),
			"actor_id":     actorID.String(),
			"type":         notifType,
		}
		if entityType != "" {
			payload["entity_type"] = entityType
		}
		if entityID != nil {
			payload["entity_id"] = entityID.String()
		}
		if postID != nil {
			payload["post_id"] = postID.String()
		}
		for k, v := range extra {
			payload[k] = v
		}

		h.hub.BroadcastToUser(recipientID, "notification.created", payload)
		count, _ := h.repo.GetUnreadNotificationsCount(ctx, recipientID)
		h.hub.BroadcastToUser(recipientID, "notification.count.updated", map[string]interface{}{
			"unread_count": count,
		})
	}
}

func (h *Handler) FollowUser(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	targetUsername := chi.URLParam(r, "username")
	target, err := h.repo.GetUserByUsername(r.Context(), targetUsername, nil)
	if err != nil {
		writeError(w, http.StatusNotFound, "target user not found")
		return
	}

	if user.ID == target.ID {
		writeError(w, http.StatusBadRequest, "cannot follow yourself")
		return
	}

	// If target account is private, create a follow request instead of immediately following
	if target.IsPrivate {
		req := &domain.FollowRequest{
			RequesterID: user.ID,
			TargetID:    target.ID,
		}
		if err := h.repo.CreateFollowRequest(r.Context(), req); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to submit follow request")
			return
		}

		h.notifyUser(r.Context(), target.ID, user.ID, "follow_request", "follow_request", &req.ID, nil, map[string]interface{}{
			"requester_username": user.Username,
			"requester_name":     user.DisplayName,
			"requester_avatar":   user.AvatarURL,
		})

		writeJSON(w, http.StatusOK, map[string]interface{}{
			"is_following":               false,
			"has_pending_follow_request": true,
			"followers_count":            target.FollowersCount,
			"status":                     "requested",
		})
		return
	}

	status, followersCount, err := h.repo.FollowUser(r.Context(), user.ID, target.ID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.notifyUser(r.Context(), target.ID, user.ID, "follow", "user", &user.ID, nil, map[string]interface{}{
		"follower_username": user.Username,
		"follower_name":     user.DisplayName,
		"follower_avatar":   user.AvatarURL,
	})

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"is_following":    status,
		"followers_count": followersCount,
	})
}

func (h *Handler) UnfollowUser(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	targetUsername := chi.URLParam(r, "username")
	target, err := h.repo.GetUserByUsername(r.Context(), targetUsername, nil)
	if err != nil {
		writeError(w, http.StatusNotFound, "target user not found")
		return
	}

	status, followersCount, err := h.repo.UnfollowUser(r.Context(), user.ID, target.ID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"is_following":    status,
		"followers_count": followersCount,
	})
}

// ==========================================
// SEARCH & AUTOCOMPLETE
// ==========================================

func (h *Handler) SearchPosts(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	tab := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("tab")))
	if tab == "" {
		tab = "top"
	}
	cursor := r.URL.Query().Get("cursor")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	viewerID := middleware.GetViewerID(r.Context())
	params := search.ParseQuery(q)
	params.Tab = tab
	params.Cursor = cursor
	params.Limit = limit

	if user := middleware.GetUserFromContext(r.Context()); user != nil && q != "" && cursor == "" {
		go func(uid uuid.UUID, searchQ string) {
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			_ = h.repo.AddRecentSearch(ctx, uid, searchQ)
		}(user.ID, q)
	}

	cacheKey := fmt.Sprintf("search:posts:%s:%x", tab, sha256.Sum256([]byte(q+cursor)))
	if h.rdb != nil && viewerID == nil {
		if val, err := h.rdb.Get(r.Context(), cacheKey).Result(); err == nil && val != "" {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("X-Cache", "HIT")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(val))
			return
		}
	}

	posts, nextCursor, err := h.repo.SearchPostsAdvanced(r.Context(), params, viewerID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to search posts")
		return
	}
	if posts == nil {
		posts = []domain.Post{}
	}

	resp := map[string]interface{}{
		"data": posts,
		"page_info": map[string]interface{}{
			"next_cursor":   nextCursor,
			"has_next_page": nextCursor != "",
		},
	}

	if h.rdb != nil && viewerID == nil && cursor == "" {
		if dataBytes, err := json.Marshal(resp); err == nil {
			_ = h.rdb.Set(r.Context(), cacheKey, dataBytes, 60*time.Second).Err()
		}
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) SearchUsers(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	cursor := r.URL.Query().Get("cursor")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	viewerID := middleware.GetViewerID(r.Context())
	users, nextCursor, err := h.repo.SearchUsersAdvanced(r.Context(), q, viewerID, cursor, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to search users")
		return
	}
	if users == nil {
		users = []domain.User{}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"data": users,
		"page_info": map[string]interface{}{
			"next_cursor":   nextCursor,
			"has_next_page": nextCursor != "",
		},
	})
}

func (h *Handler) GetSearchSuggestions(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if q == "" {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"queries": []string{},
			"users":   []domain.User{},
		})
		return
	}

	cacheKey := fmt.Sprintf("search:sugg:%s", strings.ToLower(q))
	if h.rdb != nil {
		if val, err := h.rdb.Get(r.Context(), cacheKey).Result(); err == nil && val != "" {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("X-Cache", "HIT")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(val))
			return
		}
	}

	viewerID := middleware.GetViewerID(r.Context())
	queries, users, err := h.repo.GetSearchSuggestions(r.Context(), q, viewerID, 5)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get suggestions")
		return
	}
	if queries == nil {
		queries = []string{}
	}
	if users == nil {
		users = []domain.User{}
	}

	resp := map[string]interface{}{
		"queries": queries,
		"users":   users,
	}

	if h.rdb != nil {
		if dataBytes, err := json.Marshal(resp); err == nil {
			_ = h.rdb.Set(r.Context(), cacheKey, dataBytes, 5*time.Minute).Err()
		}
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) GetRecentSearches(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	searches, err := h.repo.GetRecentSearches(r.Context(), user.ID, 10)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get recent searches")
		return
	}
	if searches == nil {
		searches = []domain.RecentSearch{}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"data": searches,
	})
}

func (h *Handler) DeleteRecentSearch(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	searchIDStr := chi.URLParam(r, "searchId")
	searchID, err := uuid.Parse(searchIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid search id")
		return
	}

	if err := h.repo.DeleteRecentSearch(r.Context(), user.ID, searchID); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete recent search")
		return
	}

	writeJSON(w, http.StatusOK, map[string]bool{"success": true})
}

func (h *Handler) ClearRecentSearches(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	if err := h.repo.ClearRecentSearches(r.Context(), user.ID); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to clear recent searches")
		return
	}

	writeJSON(w, http.StatusOK, map[string]bool{"success": true})
}

func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	viewerID := middleware.GetViewerID(r.Context())

	posts, _, _ := h.repo.SearchPosts(r.Context(), q, viewerID, "", 20)
	users, _ := h.repo.SearchUsers(r.Context(), q, viewerID, 20)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"posts": posts,
		"users": users,
	})
}

// ==========================================
// NOTIFICATIONS
// ==========================================
func (h *Handler) GetNotifications(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	notifs, nextCursor, err := h.repo.GetNotifications(r.Context(), user.ID, "", 20)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get notifications")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"notifications": notifs,
		"next_cursor":   nextCursor,
	})
}

func (h *Handler) MarkNotificationsRead(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	_ = h.repo.MarkNotificationsRead(r.Context(), user.ID)
	writeJSON(w, http.StatusOK, map[string]interface{}{"unread_count": 0})
}

// ==========================================
// MEDIA UPLOAD
// ==========================================
// ==========================================
// MEDIA UPLOAD (SECTION 6A & MEDIA PIPELINE)
// ==========================================
func (h *Handler) UploadMedia(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	// 100MB max payload
	if err := r.ParseMultipartForm(100 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "file size exceeds limit")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "file parameter is required")
		return
	}
	defer file.Close()

	if h.mediaSvc != nil {
		mediaObj, err := h.mediaSvc.ProcessAndUpload(r.Context(), user.ID, header.Filename, file)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to process media: "+err.Error())
			return
		}

		if err := h.repo.CreateMediaObject(r.Context(), mediaObj); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to record media")
			return
		}

		writeJSON(w, http.StatusCreated, mediaObj)
		return
	}

	// Local fallback if mediaSvc is nil
	ext := strings.ToLower(filepath.Ext(header.Filename))
	mediaID := uuid.New()
	filename := fmt.Sprintf("%s%s", mediaID.String(), ext)
	targetPath := filepath.Join(h.storageDir, filename)

	dst, err := os.Create(targetPath)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save media file")
		return
	}
	defer dst.Close()

	size, err := io.Copy(dst, file)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to write file")
		return
	}

	mediaType := "image"
	if ext == ".mp4" || ext == ".mov" {
		mediaType = "video"
	} else if ext == ".gif" {
		mediaType = "gif"
	}

	mediaObj := &domain.MediaObject{
		ID:            mediaID,
		UploaderID:    user.ID,
		StorageKey:    filename,
		URL:           "/uploads/" + filename,
		MediaType:     mediaType,
		MimeType:      header.Header.Get("Content-Type"),
		FileSizeBytes: size,
	}

	if err := h.repo.CreateMediaObject(r.Context(), mediaObj); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to record media")
		return
	}

	writeJSON(w, http.StatusCreated, mediaObj)
}

// ==========================================
// SUBSCRIPTIONS & VERIFICATION (SECTION 6A)
// ==========================================

func (h *Handler) GetSubscriptionPlans(w http.ResponseWriter, r *http.Request) {
	plans, err := h.repo.GetSubscriptionPlans(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch plans")
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"plans": plans,
	})
}

func (h *Handler) GetMySubscription(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	sub, err := h.repo.GetActiveSubscription(r.Context(), user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch subscription")
		return
	}

	entitlements, err := h.repo.GetUserEntitlements(r.Context(), user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to resolve entitlements")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"subscription": sub,
		"entitlements": entitlements,
	})
}

type CheckoutRequest struct {
	PlanID         string `json:"plan_id"`
	BillingCycle   string `json:"billing_cycle"` // 'monthly' or 'annual'
	CryptoCurrency string `json:"currency"`      // e.g. 'USDT_TRX', 'BTC', 'ETH'
}

func (h *Handler) CheckoutSubscription(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req CheckoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.PlanID == "" {
		req.PlanID = "premium"
	}
	if req.BillingCycle == "" {
		req.BillingCycle = "monthly"
	}
	if req.CryptoCurrency == "" {
		req.CryptoCurrency = "USDT_TRX"
	}

	plan, err := h.repo.GetSubscriptionPlanByID(r.Context(), req.PlanID)
	if err != nil {
		writeError(w, http.StatusNotFound, "subscription plan not found")
		return
	}

	amountCents := plan.MonthlyPriceCents
	if req.BillingCycle == "annual" {
		amountCents = plan.AnnualPriceCents
	}

	order := &domain.PaymentOrder{
		ID:           uuid.New(),
		UserID:       user.ID,
		PlanID:       plan.ID,
		BillingCycle: req.BillingCycle,
		AmountCents:  amountCents,
		Currency:     plan.Currency,
		Provider:     "plisio",
		Status:       "pending",
	}

	var wlInv *payment.PlisioWhiteLabelInvoice
	if h.plisioSvc != nil {
		inv, err := h.plisioSvc.CreateWhiteLabelInvoice(r.Context(), order.ID, plan.Name, amountCents, req.CryptoCurrency, user.Email)
		if err == nil && inv != nil {
			wlInv = inv
			order.ProviderInvoiceURL = inv.InvoiceURL
			order.ProviderOrderID = inv.TxnID

			// Store white label payload in metadata
			order.Metadata = map[string]interface{}{
				"white_label":            inv,
				"wallet_hash":            inv.WalletHash,
				"qr_code":                inv.QRCode,
				"amount":                 inv.Amount,
				"currency":               inv.Currency,
				"expire_utc":             inv.ExpireUTC,
				"expected_confirmations": inv.ExpectedConfirmations,
			}
		}
	}

	if err := h.repo.CreatePaymentOrder(r.Context(), order); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create payment order")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"order_id":     order.ID,
		"plan_id":      plan.ID,
		"plan_name":    plan.Name,
		"amount_cents": amountCents,
		"currency":     plan.Currency,
		"invoice_url":  order.ProviderInvoiceURL,
		"status":       order.Status,
		"white_label":  wlInv,
	})
}

func (h *Handler) CancelSubscription(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	subIDStr := chi.URLParam(r, "subscriptionId")
	subID, err := uuid.Parse(subIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid subscription id")
		return
	}

	if err := h.repo.CancelSubscription(r.Context(), subID, user.ID); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to cancel subscription")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "subscription will cancel at the end of the billing period"})
}

type SimulatePaymentRequest struct {
	OrderID string `json:"order_id"`
}

// SimulatePayment activates a subscription instantly for test/dev environments (Section 6K.12)
func (h *Handler) SimulatePayment(w http.ResponseWriter, r *http.Request) {
	// LM-SEC-001 Remediation: Strictly disable simulation in production unless explicitly permitted
	if os.Getenv("NODE_ENV") == "production" && os.Getenv("ALLOW_PAYMENT_SIMULATION") != "true" {
		writeError(w, http.StatusForbidden, "payment simulation is disabled in production")
		return
	}

	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req SimulatePaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}

	orderID, err := uuid.Parse(req.OrderID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid order id")
		return
	}

	order, err := h.repo.GetPaymentOrderByID(r.Context(), orderID)
	if err != nil {
		writeError(w, http.StatusNotFound, "order not found")
		return
	}

	// LM-SEC-001 Ownership validation
	if order.UserID != user.ID {
		writeError(w, http.StatusForbidden, "forbidden: cannot simulate payment for another user's order")
		return
	}

	_ = h.repo.UpdatePaymentOrderStatus(r.Context(), order.ID, "completed", "")

	plan, err := h.repo.GetSubscriptionPlanByID(r.Context(), order.PlanID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to find plan")
		return
	}

	periodDays := 30
	if order.BillingCycle == "annual" {
		periodDays = 365
	}

	sub := &domain.Subscription{
		ID:                 uuid.New(),
		UserID:             order.UserID,
		PlanID:             order.PlanID,
		BillingCycle:       order.BillingCycle,
		Status:             "active",
		CurrentPeriodStart: time.Now(),
		CurrentPeriodEnd:   time.Now().AddDate(0, 0, periodDays),
	}
	_ = h.repo.CreateSubscription(r.Context(), sub)

	if plan.CanGetVerifiedBadge {
		_ = h.repo.UpdateUserVerification(r.Context(), order.UserID, true, "blue")
		_ = h.repo.AssignVerificationBadge(r.Context(), &domain.VerificationBadge{
			UserID:       order.UserID,
			BadgeType:    "blue",
			DisplayLabel: "Verified",
			Description:  "This account is verified because it's subscribed to Premium and their identity was confirmed.",
			ColorToken:   "#1d9bf0",
			Status:       "approved",
		})
	}

	_ = h.repo.CreateInvoice(r.Context(), &domain.Invoice{
		OrderID:     order.ID,
		UserID:      order.UserID,
		AmountCents: order.AmountCents,
		Currency:    order.Currency,
		Status:      "paid",
	})

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":        "completed",
		"subscription":  sub,
		"is_verified":   plan.CanGetVerifiedBadge,
		"verified_type": "blue",
		"message":       "Payment simulation successful. Subscription and badge activated!",
	})
}

// GetSubscriptionOrder retrieves order status and white-label data, with optional live Plisio operations check
func (h *Handler) GetSubscriptionOrder(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	orderIDStr := chi.URLParam(r, "orderId")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid order id")
		return
	}

	order, err := h.repo.GetPaymentOrderByID(r.Context(), orderID)
	if err != nil {
		writeError(w, http.StatusNotFound, "order not found")
		return
	}

	if order.UserID != user.ID {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}

	// Active status verification against Plisio operations if still pending
	if order.Status == "pending" && h.plisioSvc != nil {
		opResp, err := h.plisioSvc.FetchOperations(r.Context(), order.ID.String())
		if err == nil && opResp != nil && opResp.Status == "success" {
			for _, op := range opResp.Data.Operations {
				if op.Type == "invoice" && (op.Params.OrderNumber == order.ID.String() || op.ID == order.ProviderOrderID) {
					opStatus := strings.ToLower(op.Status)
					if opStatus == "completed" || opStatus == "mismatch" {
						_ = h.repo.UpdatePaymentOrderStatus(r.Context(), order.ID, "completed", "")
						order.Status = "completed"

						plan, _ := h.repo.GetSubscriptionPlanByID(r.Context(), order.PlanID)
						periodDays := 30
						if order.BillingCycle == "annual" {
							periodDays = 365
						}

						sub := &domain.Subscription{
							ID:                 uuid.New(),
							UserID:             order.UserID,
							PlanID:             order.PlanID,
							BillingCycle:       order.BillingCycle,
							Status:             "active",
							CurrentPeriodStart: time.Now(),
							CurrentPeriodEnd:   time.Now().AddDate(0, 0, periodDays),
						}
						_ = h.repo.CreateSubscription(r.Context(), sub)

						if plan != nil && plan.CanGetVerifiedBadge {
							_ = h.repo.UpdateUserVerification(r.Context(), order.UserID, true, "blue")
							_ = h.repo.AssignVerificationBadge(r.Context(), &domain.VerificationBadge{
								UserID:       order.UserID,
								BadgeType:    "blue",
								DisplayLabel: "Verified",
								Description:  "This account is verified because it's subscribed to Premium.",
								ColorToken:   "#1d9bf0",
								Status:       "approved",
							})
						}

						_ = h.repo.CreateInvoice(r.Context(), &domain.Invoice{
							OrderID:     order.ID,
							UserID:      order.UserID,
							AmountCents: order.AmountCents,
							Currency:    order.Currency,
							Status:      "paid",
						})
					}
					break
				}
			}
		}
	}

	plan, _ := h.repo.GetSubscriptionPlanByID(r.Context(), order.PlanID)
	planName := "Premium"
	if plan != nil {
		planName = plan.Name
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"order":       order,
		"plan_name":   planName,
		"white_label": order.Metadata["white_label"],
	})
}

// GetCryptoCurrencies returns the list of supported cryptocurrencies
func (h *Handler) GetCryptoCurrencies(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"cryptos": payment.GetSupportedCryptos(),
	})
}

// HandlePlisioWebhook handles real IPN callbacks from Plisio cryptocurrency gateway (Section 6F)
func (h *Handler) HandlePlisioWebhook(w http.ResponseWriter, r *http.Request) {
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "failed to read body")
		return
	}

	var data map[string]interface{}
	_ = json.Unmarshal(bodyBytes, &data)

	var payload payment.PlisioWebhookPayload
	_ = json.Unmarshal(bodyBytes, &payload)

	// Fallback to form parsing if JSON empty
	if payload.OrderNumber == "" {
		_ = r.ParseForm()
		payload.Status = r.FormValue("status")
		payload.OrderNumber = r.FormValue("order_number")
		payload.TxnID = r.FormValue("txn_id")
		payload.VerifyHash = r.FormValue("verify_hash")
		if data == nil {
			data = make(map[string]interface{})
			for k, v := range r.Form {
				if len(v) > 0 {
					data[k] = v[0]
				}
			}
		}
	}

	if payload.OrderNumber == "" {
		writeError(w, http.StatusBadRequest, "order_number is required")
		return
	}

	// LM-SEC-002 Remediation: Validate cryptographic signature with PLISIO_API_KEY
	if h.plisioSvc != nil {
		if payload.VerifyHash == "" {
			slog.Warn("Plisio webhook rejected: verify_hash is missing")
			writeError(w, http.StatusUnauthorized, "verify_hash signature is required")
			return
		}
		if !h.plisioSvc.VerifyWebhook(data, payload.VerifyHash, bodyBytes) {
			slog.Warn("Plisio webhook signature verification failed", "received_hash", payload.VerifyHash)
			writeError(w, http.StatusUnprocessableEntity, "signature verification failed")
			return
		}
	}

	orderID, err := uuid.Parse(payload.OrderNumber)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid order_number format")
		return
	}

	order, err := h.repo.GetPaymentOrderByID(r.Context(), orderID)
	if err != nil {
		writeError(w, http.StatusNotFound, "order not found")
		return
	}

	// Idempotency: skip if already completed
	if order.Status == "completed" {
		writeJSON(w, http.StatusOK, map[string]string{"status": "already_processed"})
		return
	}

	status := strings.ToLower(payload.Status)
	if status == "completed" || status == "mismatch" {
		_ = h.repo.UpdatePaymentOrderStatus(r.Context(), order.ID, "completed", "")

		plan, _ := h.repo.GetSubscriptionPlanByID(r.Context(), order.PlanID)
		periodDays := 30
		if order.BillingCycle == "annual" {
			periodDays = 365
		}

		sub := &domain.Subscription{
			ID:                 uuid.New(),
			UserID:             order.UserID,
			PlanID:             order.PlanID,
			BillingCycle:       order.BillingCycle,
			Status:             "active",
			CurrentPeriodStart: time.Now(),
			CurrentPeriodEnd:   time.Now().AddDate(0, 0, periodDays),
		}
		_ = h.repo.CreateSubscription(r.Context(), sub)

		if plan != nil && plan.CanGetVerifiedBadge {
			_ = h.repo.UpdateUserVerification(r.Context(), order.UserID, true, "blue")
			_ = h.repo.AssignVerificationBadge(r.Context(), &domain.VerificationBadge{
				UserID:       order.UserID,
				BadgeType:    "blue",
				DisplayLabel: "Verified",
				Description:  "This account is verified because it's subscribed to Premium.",
				ColorToken:   "#1d9bf0",
				Status:       "approved",
			})
		}

		_ = h.repo.CreateInvoice(r.Context(), &domain.Invoice{
			OrderID:     order.ID,
			UserID:      order.UserID,
			AmountCents: order.AmountCents,
			Currency:    order.Currency,
			Status:      "paid",
		})
	} else if status == "expired" || status == "cancelled" {
		_ = h.repo.UpdatePaymentOrderStatus(r.Context(), order.ID, status, "")
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) GetBillingHistory(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	orders, invoices, err := h.repo.GetBillingHistory(r.Context(), user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch billing history")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"orders":   orders,
		"invoices": invoices,
	})
}

func (h *Handler) GetMyVerification(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	badges, err := h.repo.GetUserBadges(r.Context(), user.ID)
	if err != nil {
		badges = []domain.VerificationBadge{}
	}

	app, _ := h.repo.GetVerificationApplication(r.Context(), user.ID)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"is_verified":   user.IsVerified,
		"verified_type": user.VerifiedType,
		"badges":        badges,
		"application":   app,
	})
}

type VerificationAppRequest struct {
	BadgeType   string `json:"badge_type"`
	Category    string `json:"category"`
	DocumentURL string `json:"document_url"`
}

func (h *Handler) ApplyVerification(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req VerificationAppRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}

	if req.BadgeType == "" {
		req.BadgeType = "blue"
	}

	app := &domain.VerificationApplication{
		UserID:      user.ID,
		BadgeType:   req.BadgeType,
		Category:    req.Category,
		DocumentURL: req.DocumentURL,
		Status:      "pending",
	}

	if err := h.repo.CreateVerificationApplication(r.Context(), app); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to submit application")
		return
	}

	writeJSON(w, http.StatusCreated, app)
}

func (h *Handler) GetUserBadges(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")
	targetUser, err := h.repo.GetUserByUsername(r.Context(), username, nil)
	if err != nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}

	badges, err := h.repo.GetUserBadges(r.Context(), targetUser.ID)
	if err != nil {
		badges = []domain.VerificationBadge{}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"username":      targetUser.Username,
		"is_verified":   targetUser.IsVerified,
		"verified_type": targetUser.VerifiedType,
		"badges":        badges,
	})
}

// ==========================================
// REALTIME WEBSOCKET & OUTBOX WORKER
// ==========================================

func (h *Handler) ServeWebSocket(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		token := r.URL.Query().Get("token")
		if token != "" {
			hash := sha256.Sum256([]byte(token))
			tokenHash := hex.EncodeToString(hash[:])
			session, err := h.repo.GetSessionByTokenHash(r.Context(), tokenHash)
			if err == nil && session != nil && session.ExpiresAt.After(time.Now()) {
				u, err := h.repo.GetUserByID(r.Context(), session.UserID, nil)
				if err == nil && u != nil {
					user = u
				}
			}
		}
	}

	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	if h.hub != nil {
		realtime.ServeWS(h.hub, w, r, user.ID)
	}
}

func (h *Handler) StartOutboxWorker(ctx context.Context) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			events, err := h.repo.GetPendingOutboxEvents(ctx, 20)
			if err != nil || len(events) == 0 {
				continue
			}

			for _, e := range events {
				if h.hub != nil {
					switch e.EventType {
					case "notification.created":
						if recipientStr, ok := e.Payload["recipient_id"].(string); ok {
							if recID, err := uuid.Parse(recipientStr); err == nil {
								h.hub.BroadcastToUser(recID, "notification.created", e.Payload)
								count, _ := h.repo.GetUnreadNotificationsCount(ctx, recID)
								h.hub.BroadcastToUser(recID, "notification.count.updated", map[string]int{"unread_count": count})
							}
						}
					case "message.created":
						if convStr, ok := e.Payload["conversation_id"].(string); ok {
							if convID, err := uuid.Parse(convStr); err == nil {
								h.hub.BroadcastToConversation(convID, nil, "message.created", e.Payload)
							}
						}
						if recStr, ok := e.Payload["recipient_id"].(string); ok {
							if recID, err := uuid.Parse(recStr); err == nil {
								h.hub.BroadcastToUser(recID, "message.created", e.Payload)
							}
						}
					}
				}
				_ = h.repo.MarkOutboxEventPublished(ctx, e.ID)
			}
		}
	}
}

// ==========================================
// SOCIAL GRAPH: FOLLOWERS, FOLLOWING, MUTUALS, BLOCK, MUTE, REQUESTS
// ==========================================

func (h *Handler) GetFollowers(w http.ResponseWriter, r *http.Request) {
	targetUsername := chi.URLParam(r, "username")
	target, err := h.repo.GetUserByUsername(r.Context(), targetUsername, nil)
	if err != nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}

	viewerID := uuid.Nil
	if u := middleware.GetUserFromContext(r.Context()); u != nil {
		viewerID = u.ID
	}

	users, nextCursor, err := h.repo.GetFollowers(r.Context(), target.ID, viewerID, 30, "")
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get followers")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"users":       users,
		"next_cursor": nextCursor,
	})
}

func (h *Handler) GetFollowing(w http.ResponseWriter, r *http.Request) {
	targetUsername := chi.URLParam(r, "username")
	target, err := h.repo.GetUserByUsername(r.Context(), targetUsername, nil)
	if err != nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}

	viewerID := uuid.Nil
	if u := middleware.GetUserFromContext(r.Context()); u != nil {
		viewerID = u.ID
	}

	users, nextCursor, err := h.repo.GetFollowing(r.Context(), target.ID, viewerID, 30, "")
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get following")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"users":       users,
		"next_cursor": nextCursor,
	})
}

func (h *Handler) GetMutualConnections(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{"mutuals": []domain.User{}})
		return
	}

	targetUsername := chi.URLParam(r, "username")
	target, err := h.repo.GetUserByUsername(r.Context(), targetUsername, nil)
	if err != nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}

	mutuals, err := h.repo.GetMutualConnections(r.Context(), target.ID, user.ID, 5)
	if err != nil {
		mutuals = []domain.User{}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"mutuals": mutuals})
}

func (h *Handler) GetSocialRelationship(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	targetUsername := chi.URLParam(r, "username")
	target, err := h.repo.GetUserByUsername(r.Context(), targetUsername, nil)
	if err != nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}

	rel, err := h.repo.GetSocialRelationship(r.Context(), user.ID, target.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get relationship")
		return
	}

	writeJSON(w, http.StatusOK, rel)
}

func (h *Handler) BlockUser(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	targetUsername := chi.URLParam(r, "username")
	target, err := h.repo.GetUserByUsername(r.Context(), targetUsername, nil)
	if err != nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}

	if err := h.repo.BlockUser(r.Context(), user.ID, target.ID); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"status": "blocked", "is_blocked": true})
}

func (h *Handler) UnblockUser(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	targetUsername := chi.URLParam(r, "username")
	target, err := h.repo.GetUserByUsername(r.Context(), targetUsername, nil)
	if err != nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}

	if err := h.repo.UnblockUser(r.Context(), user.ID, target.ID); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"status": "unblocked", "is_blocked": false})
}

func (h *Handler) MuteUser(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	targetUsername := chi.URLParam(r, "username")
	target, err := h.repo.GetUserByUsername(r.Context(), targetUsername, nil)
	if err != nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}

	if err := h.repo.MuteUser(r.Context(), user.ID, target.ID); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"status": "muted", "is_muted": true})
}

func (h *Handler) UnmuteUser(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	targetUsername := chi.URLParam(r, "username")
	target, err := h.repo.GetUserByUsername(r.Context(), targetUsername, nil)
	if err != nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}

	if err := h.repo.UnmuteUser(r.Context(), user.ID, target.ID); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"status": "unmuted", "is_muted": false})
}

// Follow Requests
func (h *Handler) CreateFollowRequest(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	targetUsername := chi.URLParam(r, "username")
	target, err := h.repo.GetUserByUsername(r.Context(), targetUsername, nil)
	if err != nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}

	req := &domain.FollowRequest{
		RequesterID: user.ID,
		TargetID:    target.ID,
	}

	if err := h.repo.CreateFollowRequest(r.Context(), req); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create follow request")
		return
	}

	h.notifyUser(r.Context(), target.ID, user.ID, "follow_request", "follow_request", &req.ID, nil, map[string]interface{}{
		"requester_username": user.Username,
		"requester_name":     user.DisplayName,
		"requester_avatar":   user.AvatarURL,
	})

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"status":                     "requested",
		"has_pending_follow_request": true,
	})
}

func (h *Handler) GetFollowRequests(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	reqs, err := h.repo.GetFollowRequests(r.Context(), user.ID)
	if err != nil {
		reqs = []domain.FollowRequest{}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"requests": reqs})
}

func (h *Handler) RespondFollowRequest(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	requestID, err := uuid.Parse(chi.URLParam(r, "requestId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid request id")
		return
	}

	var body struct {
		Action string `json:"action"` // 'accept' or 'reject'
	}
	_ = json.NewDecoder(r.Body).Decode(&body)

	accept := body.Action == "accept"
	if err := h.repo.RespondFollowRequest(r.Context(), requestID, user.ID, accept); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to respond to request")
		return
	}

	// Update recipient's unread count
	if h.hub != nil {
		count, _ := h.repo.GetUnreadNotificationsCount(r.Context(), user.ID)
		h.hub.BroadcastToUser(user.ID, "notification.count.updated", map[string]interface{}{
			"unread_count": count,
		})
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": body.Action})
}

func (h *Handler) RespondFollowRequestByUsername(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	requesterUsername := chi.URLParam(r, "username")
	requester, err := h.repo.GetUserByUsername(r.Context(), requesterUsername, nil)
	if err != nil {
		writeError(w, http.StatusNotFound, "requester not found")
		return
	}

	var body struct {
		Action string `json:"action"` // 'accept' or 'reject'
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	if body.Action != "accept" && body.Action != "reject" {
		body.Action = "accept"
	}

	reqs, err := h.repo.GetFollowRequests(r.Context(), user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch follow requests")
		return
	}

	var targetReq *domain.FollowRequest
	for _, req := range reqs {
		if req.RequesterID == requester.ID {
			targetReq = &req
			break
		}
	}

	if targetReq == nil {
		writeError(w, http.StatusNotFound, "no pending follow request from this user")
		return
	}

	accept := body.Action == "accept"
	if err := h.repo.RespondFollowRequest(r.Context(), targetReq.ID, user.ID, accept); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to respond to follow request")
		return
	}

	if h.hub != nil {
		count, _ := h.repo.GetUnreadNotificationsCount(r.Context(), user.ID)
		h.hub.BroadcastToUser(user.ID, "notification.count.updated", map[string]interface{}{
			"unread_count": count,
		})
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":     body.Action,
		"request_id": targetReq.ID.String(),
	})
}

func (h *Handler) CancelFollowRequest(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	targetUsername := chi.URLParam(r, "username")
	target, err := h.repo.GetUserByUsername(r.Context(), targetUsername, nil)
	if err != nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}

	_ = h.repo.CancelFollowRequest(r.Context(), user.ID, target.ID)
	writeJSON(w, http.StatusOK, map[string]interface{}{"status": "cancelled", "has_pending_follow_request": false})
}

// ==========================================
// DIRECT MESSAGING
// ==========================================

func (h *Handler) GetConversations(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	convs, err := h.repo.GetConversations(r.Context(), user.ID, 30, "")
	if err != nil {
		convs = []domain.Conversation{}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"conversations": convs})
}

func (h *Handler) CreateConversation(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var body struct {
		RecipientID string `json:"recipient_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}

	recipientID, err := uuid.Parse(body.RecipientID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid recipient id")
		return
	}

	isBlocked, _ := h.repo.IsBlocked(r.Context(), user.ID, recipientID)
	if isBlocked {
		writeError(w, http.StatusForbidden, "cannot message this user")
		return
	}

	conv, err := h.repo.GetOrCreateConversation(r.Context(), user.ID, recipientID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create conversation")
		return
	}

	if otherU, err := h.repo.GetUserByID(r.Context(), recipientID, nil); err == nil {
		conv.OtherMember = otherU
	}

	writeJSON(w, http.StatusCreated, conv)
}

func (h *Handler) GetConversationByID(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	convID, err := uuid.Parse(chi.URLParam(r, "conversationId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid conversation id")
		return
	}

	conv, err := h.repo.GetConversationByID(r.Context(), convID, user.ID)
	if err != nil {
		writeError(w, http.StatusNotFound, "conversation not found")
		return
	}

	writeJSON(w, http.StatusOK, conv)
}

func (h *Handler) GetMessages(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	convID, err := uuid.Parse(chi.URLParam(r, "conversationId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid conversation id")
		return
	}

	var beforeID *uuid.UUID
	if bStr := r.URL.Query().Get("before"); bStr != "" {
		if bUUID, err := uuid.Parse(bStr); err == nil {
			beforeID = &bUUID
		}
	}

	messages, err := h.repo.GetMessages(r.Context(), convID, user.ID, 50, beforeID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"messages": messages})
}

func (h *Handler) SendMessage(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	convID, err := uuid.Parse(chi.URLParam(r, "conversationId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid conversation id")
		return
	}

	conv, err := h.repo.GetConversationByID(r.Context(), convID, user.ID)
	if err != nil {
		writeError(w, http.StatusNotFound, "conversation not found or access denied")
		return
	}

	if conv.OtherMember != nil {
		isBlocked, _ := h.repo.IsBlocked(r.Context(), conv.OtherMember.ID, user.ID)
		if isBlocked {
			writeError(w, http.StatusForbidden, "cannot message this user")
			return
		}
		isBlockedByMe, _ := h.repo.IsBlocked(r.Context(), user.ID, conv.OtherMember.ID)
		if isBlockedByMe {
			writeError(w, http.StatusForbidden, "cannot message a blocked user")
			return
		}
	}

	var body struct {
		Content         string     `json:"content"`
		ClientMessageID string     `json:"client_message_id"`
		MediaID         *uuid.UUID `json:"media_id"`
		ReplyToID       *uuid.UUID `json:"reply_to_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid payload")
		return
	}

	if strings.TrimSpace(body.Content) == "" && body.MediaID == nil {
		writeError(w, http.StatusBadRequest, "message cannot be empty")
		return
	}

	msg := &domain.Message{
		ID:              uuid.New(),
		ConversationID:  convID,
		SenderID:        user.ID,
		Sender:          user,
		Content:         body.Content,
		ClientMessageID: body.ClientMessageID,
		MediaID:         body.MediaID,
		ReplyToID:       body.ReplyToID,
		MessageType:     "text",
	}
	if body.MediaID != nil {
		msg.MessageType = "media"
	}

	if err := h.repo.CreateMessage(r.Context(), msg); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save message")
		return
	}

	if h.hub != nil {
		var recipientID *uuid.UUID
		if conv.OtherMember != nil {
			recipientID = &conv.OtherMember.ID
		}
		h.hub.BroadcastMessage(convID, recipientID, &user.ID, msg)
	}

	if conv.OtherMember != nil {
		h.notifyUser(r.Context(), conv.OtherMember.ID, user.ID, "dm", "conversation", &convID, nil, map[string]interface{}{
			"conversation_id": convID.String(),
			"message_id":      msg.ID.String(),
			"sender_username": user.Username,
			"sender_name":     user.DisplayName,
			"sender_avatar":   user.AvatarURL,
			"content":         msg.Content,
		})
	}

	outboxPayload := map[string]interface{}{
		"message_id":      msg.ID.String(),
		"conversation_id": convID.String(),
		"sender_id":       user.ID.String(),
		"content":         msg.Content,
	}
	if conv.OtherMember != nil {
		outboxPayload["recipient_id"] = conv.OtherMember.ID.String()
	}

	_ = h.repo.CreateOutboxEvent(r.Context(), &domain.OutboxEvent{
		EventType:     "message.created",
		AggregateType: "conversation",
		AggregateID:   convID,
		Payload:       outboxPayload,
	})

	writeJSON(w, http.StatusCreated, msg)
}

func (h *Handler) MarkConversationRead(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	convID, err := uuid.Parse(chi.URLParam(r, "conversationId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid conversation id")
		return
	}

	var body struct {
		MessageID *uuid.UUID `json:"message_id"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)

	_ = h.repo.MarkConversationRead(r.Context(), convID, user.ID, body.MessageID)

	if h.hub != nil {
		h.hub.BroadcastToConversation(convID, &user.ID, "message.read", map[string]interface{}{
			"conversation_id": convID.String(),
			"reader_id":       user.ID.String(),
			"read_at":         time.Now().UTC().Format(time.RFC3339),
		})
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// Extra Notification Handlers
func (h *Handler) GetUnreadNotificationsCount(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	count, _ := h.repo.GetUnreadNotificationsCount(r.Context(), user.ID)
	writeJSON(w, http.StatusOK, map[string]interface{}{"unread_count": count})
}

func (h *Handler) MarkNotificationRead(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	notifID, err := uuid.Parse(chi.URLParam(r, "notificationId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid notification id")
		return
	}

	_ = h.repo.MarkNotificationRead(r.Context(), notifID, user.ID)
	count, _ := h.repo.GetUnreadNotificationsCount(r.Context(), user.ID)
	writeJSON(w, http.StatusOK, map[string]interface{}{"status": "ok", "unread_count": count})
}

