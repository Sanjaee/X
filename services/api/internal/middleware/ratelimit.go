package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

type memoryLimiter struct {
	sync.Mutex
	hits      map[string][]int64
	lastClean time.Time
}

var memLimiter = &memoryLimiter{
	hits:      make(map[string][]int64),
	lastClean: time.Now(),
}

func (m *memoryLimiter) isAllowed(key string, limit int, window time.Duration) (bool, int, int64) {
	m.Lock()
	defer m.Unlock()

	now := time.Now().Unix()
	cutoff := now - int64(window.Seconds())

	// Periodic clean
	if time.Since(m.lastClean) > 5*time.Minute {
		for k, timestamps := range m.hits {
			var valid []int64
			for _, ts := range timestamps {
				if ts > cutoff {
					valid = append(valid, ts)
				}
			}
			if len(valid) == 0 {
				delete(m.hits, k)
			} else {
				m.hits[k] = valid
			}
		}
		m.lastClean = time.Now()
	}

	timestamps := m.hits[key]
	var current []int64
	for _, ts := range timestamps {
		if ts > cutoff {
			current = append(current, ts)
		}
	}

	if len(current) >= limit {
		m.hits[key] = current
		resetSec := cutoff + int64(window.Seconds()) - now
		if resetSec <= 0 {
			resetSec = 1
		}
		return false, 0, resetSec
	}

	current = append(current, now)
	m.hits[key] = current
	remaining := limit - len(current)
	return true, remaining, 60
}

// RateLimiter returns an admission-control rate limiting middleware
func RateLimiter(rdb *redis.Client, requestsPerMinute int) func(http.Handler) http.Handler {
	if requestsPerMinute <= 0 {
		requestsPerMinute = 120 // 120 requests/min default
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip rate limiting for internal health and metrics
			if r.URL.Path == "/health/live" || r.URL.Path == "/health/ready" || r.URL.Path == "/health/metrics" {
				next.ServeHTTP(w, r)
				return
			}

			// Determine identifier: authenticated user ID or Remote IP
			identifier := ""
			if u := GetUserFromContext(r.Context()); u != nil {
				identifier = "usr:" + u.ID.String()
			} else {
				host, _, err := net.SplitHostPort(r.RemoteAddr)
				if err != nil {
					host = r.RemoteAddr
				}
				identifier = "ip:" + host
			}

			key := fmt.Sprintf("ratelimit:%s:%s", r.URL.Path, identifier)

			// 1. If Redis is available, execute atomic sliding window
			if rdb != nil {
				ctx, cancel := context.WithTimeout(r.Context(), 500*time.Millisecond)
				defer cancel()

				pipe := rdb.Pipeline()
				now := time.Now().Unix()
				windowStart := now - 60

				pipe.ZRemRangeByScore(ctx, key, "-inf", strconv.FormatInt(windowStart, 10))
				pipe.ZAdd(ctx, key, redis.Z{Score: float64(now), Member: fmt.Sprintf("%d_%d", now, time.Now().UnixNano())})
				cardCmd := pipe.ZCard(ctx, key)
				pipe.Expire(ctx, key, 70*time.Second)

				_, err := pipe.Exec(ctx)
				if err == nil {
					count := cardCmd.Val()
					remaining := int64(requestsPerMinute) - count
					if remaining < 0 {
						remaining = 0
					}

					w.Header().Set("X-RateLimit-Limit", strconv.Itoa(requestsPerMinute))
					w.Header().Set("X-RateLimit-Remaining", strconv.FormatInt(remaining, 10))

					if count > int64(requestsPerMinute) {
						w.Header().Set("Retry-After", "60")
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusTooManyRequests)
						_ = json.NewEncoder(w).Encode(map[string]interface{}{
							"error":       "rate_limit_exceeded",
							"message":     "Too many requests. Please slow down.",
							"retry_after": 60,
						})
						return
					}

					next.ServeHTTP(w, r)
					return
				}
				// If Redis command fails, fall through to in-memory limiter
			}

			// 2. In-memory limiter fallback
			allowed, remaining, resetSec := memLimiter.isAllowed(key, requestsPerMinute, 60*time.Second)
			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(requestsPerMinute))
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))

			if !allowed {
				w.Header().Set("Retry-After", strconv.FormatInt(resetSec, 10))
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"error":       "rate_limit_exceeded",
					"message":     "Too many requests. Please slow down.",
					"retry_after": resetSec,
				})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
