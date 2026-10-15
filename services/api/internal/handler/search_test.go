package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/x-reconstruction/api/internal/domain"
	"github.com/x-reconstruction/api/internal/repository"
)

func TestSearchAPIEndpoints(t *testing.T) {
	repo := repository.NewMemoryRepository()
	h := NewHandler(repo, "./test_uploads", nil, nil, nil, nil)

	// Seed user and post
	testUser := &domain.User{
		ID:          uuid.New(),
		Username:    "technologist",
		DisplayName: "Tech Pioneer",
		Role:        "user",
		IsVerified:  true,
	}
	_ = repo.CreateUser(nil, testUser)

	post1 := &domain.Post{
		ID:         uuid.New(),
		AuthorID:   testUser.ID,
		Author:     testUser,
		Content:    "Exploring Golang and high performance systems #golang",
		LikesCount: 15,
	}
	_ = repo.CreatePost(nil, post1)

	r := chi.NewRouter()
	r.Get("/api/v1/search/posts", h.SearchPosts)
	r.Get("/api/v1/search/users", h.SearchUsers)
	r.Get("/api/v1/search/suggestions", h.GetSearchSuggestions)

	// 1. Test Search Posts
	t.Run("SearchPosts top tab", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/search/posts?q=golang&tab=top", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}

		var resp struct {
			Data     []domain.Post   `json:"data"`
			PageInfo domain.PageInfo `json:"page_info"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if len(resp.Data) != 1 {
			t.Fatalf("expected 1 post, got %d", len(resp.Data))
		}
	})

	// 2. Test Search Users
	t.Run("SearchUsers", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/search/users?q=techno", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}

		var resp struct {
			Data []domain.User `json:"data"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if len(resp.Data) == 0 {
			t.Fatalf("expected to find user, got 0")
		}
		if resp.Data[0].Username != "technologist" {
			t.Errorf("expected technologist, got %s", resp.Data[0].Username)
		}
	})

	// 3. Test Search Suggestions
	t.Run("SearchSuggestions", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/search/suggestions?q=techno", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}

		var resp struct {
			Queries []string      `json:"queries"`
			Users   []domain.User `json:"users"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if len(resp.Users) == 0 {
			t.Errorf("expected matching user suggestion")
		}
	})
}
