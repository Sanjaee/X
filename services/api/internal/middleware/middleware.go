package middleware

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/x-reconstruction/api/internal/domain"
	"github.com/x-reconstruction/api/internal/repository"
)

type contextKey string

const (
	UserContextKey contextKey = "user_context"
)

// AuthMiddleware inspects session_token with optional Redis caching
func AuthMiddleware(repo repository.Repository, optional bool) func(http.Handler) http.Handler {
	return AuthMiddlewareWithCache(repo, nil, optional)
}

// AuthMiddlewareWithCache uses Redis cache-aside to eliminate DB roundtrips on authenticated requests
func AuthMiddlewareWithCache(repo repository.Repository, rdb *redis.Client, optional bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := ""

			// 1. Check cookie
			if cookie, err := r.Cookie("session_token"); err == nil && cookie.Value != "" {
				token = cookie.Value
			}

			// 2. Check Authorization header
			if token == "" {
				authHeader := r.Header.Get("Authorization")
				if strings.HasPrefix(authHeader, "Bearer ") {
					token = strings.TrimPrefix(authHeader, "Bearer ")
				}
			}

			if token == "" {
				if optional {
					next.ServeHTTP(w, r)
					return
				}
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}

			// Hash token with SHA-256
			h := sha256.Sum256([]byte(token))
			tokenHash := hex.EncodeToString(h[:])

			// 3. Fast-path: Check Redis session cache
			cacheKey := fmt.Sprintf("session:token:%s", tokenHash)
			if rdb != nil {
				if cachedUserJSON, err := rdb.Get(r.Context(), cacheKey).Result(); err == nil && cachedUserJSON != "" {
					var user domain.User
					if err := json.Unmarshal([]byte(cachedUserJSON), &user); err == nil {
						ctx := context.WithValue(r.Context(), UserContextKey, &user)
						next.ServeHTTP(w, r.WithContext(ctx))
						return
					}
				}
			}

			// 4. Slow-path: Query database
			session, err := repo.GetSessionByTokenHash(r.Context(), tokenHash)
			if err != nil {
				if optional {
					next.ServeHTTP(w, r)
					return
				}
				http.Error(w, `{"error":"invalid or expired session"}`, http.StatusUnauthorized)
				return
			}

			user, err := repo.GetUserByID(r.Context(), session.UserID, nil)
			if err != nil {
				if optional {
					next.ServeHTTP(w, r)
					return
				}
				http.Error(w, `{"error":"user not found"}`, http.StatusUnauthorized)
				return
			}

			// Populate Redis session cache with TTL
			if rdb != nil {
				ttl := 15 * time.Minute
				if rem := time.Until(session.ExpiresAt); rem > 0 && rem < ttl {
					ttl = rem
				}
				if userBytes, err := json.Marshal(user); err == nil {
					_ = rdb.Set(r.Context(), cacheKey, userBytes, ttl).Err()
				}
			}

			// Inject user into context
			ctx := context.WithValue(r.Context(), UserContextKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserFromContext extracts authenticated user or nil
func GetUserFromContext(ctx context.Context) *domain.User {
	val := ctx.Value(UserContextKey)
	if u, ok := val.(*domain.User); ok {
		return u
	}
	return nil
}

// GetViewerID helper
func GetViewerID(ctx context.Context) *uuid.UUID {
	u := GetUserFromContext(ctx)
	if u != nil {
		return &u.ID
	}
	return nil
}

// LoggerMiddleware logs every incoming request using slog
func LoggerMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := &responseWriterWrapper{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(ww, r)

			logger.Info("http_request",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", ww.statusCode),
				slog.Duration("duration", time.Since(start)),
				slog.String("ip", r.RemoteAddr),
			)
		})
	}
}

// RecoveryMiddleware recovers from panics and returns 500 JSON
func RecoveryMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					logger.Error("panic_recovered", slog.Any("error", err))
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(`{"error":"internal server error"}`))
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriterWrapper) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriterWrapper) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hj, ok := rw.ResponseWriter.(http.Hijacker); ok {
		return hj.Hijack()
	}
	return nil, nil, errors.New("underlying ResponseWriter does not implement http.Hijacker")
}

func (rw *responseWriterWrapper) Flush() {
	if fl, ok := rw.ResponseWriter.(http.Flusher); ok {
		fl.Flush()
	}
}

func (rw *responseWriterWrapper) Unwrap() http.ResponseWriter {
	return rw.ResponseWriter
}
