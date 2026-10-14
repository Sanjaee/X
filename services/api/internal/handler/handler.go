package handler

import (
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
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/x-reconstruction/api/internal/domain"
	"github.com/x-reconstruction/api/internal/media"
	"github.com/x-reconstruction/api/internal/middleware"
	"github.com/x-reconstruction/api/internal/payment"
	"github.com/x-reconstruction/api/internal/repository"
)

type Handler struct {
	repo       repository.Repository
	storageDir string
	mediaSvc   *media.MediaService
	plisioSvc  *payment.PlisioService
}

func NewHandler(repo repository.Repository, storageDir string, mediaSvc *media.MediaService, plisioSvc *payment.PlisioService) *Handler {
	_ = os.MkdirAll(storageDir, 0755)
	return &Handler{
		repo:       repo,
		storageDir: storageDir,
		mediaSvc:   mediaSvc,
		plisioSvc:  plisioSvc,
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
	if err := h.repo.Ping(r.Context()); err != nil {
		writeError(w, http.StatusServiceUnavailable, "database unavailable")
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":   "healthy",
		"database": "connected",
		"time":     time.Now().UTC().Format(time.RFC3339),
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
	writeJSON(w, http.StatusOK, user)
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
	user.Website = req.Website

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
		SameSite: http.SameSiteLaxMode,
	})
}

// ==========================================
// POSTS & FEED
// ==========================================
type CreatePostRequest struct {
	Content     string      `json:"content"`
	MediaIDs    []uuid.UUID `json:"media_ids"`
	ReplyToID   *uuid.UUID  `json:"reply_to_id"`
	QuotePostID *uuid.UUID  `json:"quote_post_id"`
}

func (h *Handler) CreatePost(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req CreatePostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}

	req.Content = strings.TrimSpace(req.Content)
	if req.Content == "" && len(req.MediaIDs) == 0 {
		writeError(w, http.StatusBadRequest, "content or media required")
		return
	}

	// Server-side entitlement check (Section 6A.B)
	entitlements, err := h.repo.GetUserEntitlements(r.Context(), user.ID)
	maxLen := 280
	if err == nil && entitlements != nil && entitlements.MaxPostLength > 0 {
		maxLen = entitlements.MaxPostLength
	}
	if len([]rune(req.Content)) > maxLen {
		writeError(w, http.StatusForbidden, fmt.Sprintf("content exceeds maximum limit of %d characters for your plan. Upgrade to Premium for longer posts.", maxLen))
		return
	}

	post := &domain.Post{

		AuthorID:    user.ID,
		Content:     req.Content,
		ReplyToID:   req.ReplyToID,
		QuotePostID: req.QuotePostID,
	}

	for _, mID := range req.MediaIDs {
		if media, err := h.repo.GetMediaObjectByID(r.Context(), mID); err == nil {
			post.Media = append(post.Media, *media)
		}
	}

	if err := h.repo.CreatePost(r.Context(), post); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create post")
		return
	}

	// If this was a reply, notify author of parent
	if post.ReplyToID != nil {
		if parent, err := h.repo.GetPostByID(r.Context(), *post.ReplyToID, nil); err == nil && parent.AuthorID != user.ID {
			_ = h.repo.CreateNotification(r.Context(), &domain.Notification{
				RecipientID: parent.AuthorID,
				ActorID:     user.ID,
				Type:        "reply",
				PostID:      &post.ID,
			})
		}
	}

	writeJSON(w, http.StatusCreated, post)
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

	if err := h.repo.DeletePost(r.Context(), postID, user.ID); err != nil {
		writeError(w, http.StatusForbidden, "cannot delete post")
		return
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
		_ = h.repo.CreateNotification(r.Context(), &domain.Notification{
			RecipientID: p.AuthorID,
			ActorID:     user.ID,
			Type:        "like",
			PostID:      &postID,
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
		_ = h.repo.CreateNotification(r.Context(), &domain.Notification{
			RecipientID: p.AuthorID,
			ActorID:     user.ID,
			Type:        "repost",
			PostID:      &postID,
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

	status, followersCount, err := h.repo.FollowUser(r.Context(), user.ID, target.ID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	_ = h.repo.CreateNotification(r.Context(), &domain.Notification{
		RecipientID: target.ID,
		ActorID:     user.ID,
		Type:        "follow",
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
// SEARCH
// ==========================================
func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	searchType := r.URL.Query().Get("type")
	if searchType == "" {
		searchType = "top"
	}
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

	// Validate cryptographic signature with PLISIO_API_KEY
	if h.plisioSvc != nil && payload.VerifyHash != "" {
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

