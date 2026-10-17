package middleware

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/x-reconstruction/api/internal/domain"
	"github.com/x-reconstruction/api/internal/repository"
)

// RequirePermission verifies the authenticated user possesses the specific permission
func RequirePermission(repo repository.Repository, permissionKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := GetUserFromContext(r.Context())
			if user == nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				_ = json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
				return
			}

			// Suspended user protection
			if user.DeletedAt != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				_ = json.NewEncoder(w).Encode(map[string]string{"error": "account suspended"})
				return
			}

			// Owner bypass / absolute platform governance
			if IsOwnerAccount(user) {
				next.ServeHTTP(w, r)
				return
			}

			hasPerm, err := repo.HasPermission(r.Context(), user.ID, permissionKey)
			if err != nil || !hasPerm {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				_ = json.NewEncoder(w).Encode(map[string]string{
					"error": "forbidden: insufficient permissions",
					"required_permission": permissionKey,
				})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireAnyPermission allows access if the user has any of the specified permissions
func RequireAnyPermission(repo repository.Repository, permissionKeys ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := GetUserFromContext(r.Context())
			if user == nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				_ = json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
				return
			}

			if user.DeletedAt != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				_ = json.NewEncoder(w).Encode(map[string]string{"error": "account suspended"})
				return
			}

			if IsOwnerAccount(user) {
				next.ServeHTTP(w, r)
				return
			}

			for _, permKey := range permissionKeys {
				if hasPerm, _ := repo.HasPermission(r.Context(), user.ID, permKey); hasPerm {
					next.ServeHTTP(w, r)
					return
				}
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error": "forbidden: insufficient administrative permissions",
			})
		})
	}
}

// IsOwnerAccount checks if the user is verified platform owner
func IsOwnerAccount(user *domain.User) bool {
	if user == nil {
		return false
	}
	if user.Role == "owner" {
		return true
	}
	if strings.ToLower(strings.TrimSpace(user.Email)) == "afrizaahmad18@gmail.com" {
		return true
	}
	return false
}

// RequireOwner strictly ensures only platform owners can access the route
func RequireOwner(repo repository.Repository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := GetUserFromContext(r.Context())
			if user == nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				_ = json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
				return
			}

			if user.DeletedAt != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				_ = json.NewEncoder(w).Encode(map[string]string{"error": "account suspended"})
				return
			}

			if IsOwnerAccount(user) {
				next.ServeHTTP(w, r)
				return
			}

			// Check roles in repository
			roles, err := repo.GetUserRoles(r.Context(), user.ID)
			if err == nil {
				for _, rItem := range roles {
					if rItem.Key == "owner" {
						next.ServeHTTP(w, r)
						return
					}
				}
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error": "forbidden: platform owner access required",
			})
		})
	}
}

// AuthorizeResource verifies resource ownership or administrative overrides
func AuthorizeResource(user *domain.User, resourceAuthorID uuid.UUID, hasAdminPermission bool) bool {
	if user == nil {
		return false
	}
	if IsOwnerAccount(user) {
		return true
	}
	if user.ID == resourceAuthorID {
		return true
	}
	return hasAdminPermission
}
