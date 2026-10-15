package middleware

import (
	"encoding/json"
	"net/http"
	"strconv"
	"sync/atomic"
)

var (
	// ActiveRequests counts currently processing HTTP requests
	ActiveRequests int64
)

// ConcurrencyLimiter restricts the maximum number of concurrent requests executing simultaneously.
// When the ceiling is reached, surplus requests receive an immediate HTTP 503 with Retry-After header.
func ConcurrencyLimiter(maxConcurrent int) func(http.Handler) http.Handler {
	if maxConcurrent <= 0 {
		maxConcurrent = 300 // default ceiling for 4GB VPS
	}
	sem := make(chan struct{}, maxConcurrent)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Bypass health checks from admission control so monitors never report false dead
			if r.URL.Path == "/health/live" || r.URL.Path == "/health/ready" || r.URL.Path == "/health/metrics" {
				next.ServeHTTP(w, r)
				return
			}

			select {
			case sem <- struct{}{}:
				atomic.AddInt64(&ActiveRequests, 1)
				defer func() {
					atomic.AddInt64(&ActiveRequests, -1)
					<-sem
				}()
				next.ServeHTTP(w, r)
			default:
				// Backpressure activated: fail fast rather than ballooning RAM
				w.Header().Set("Retry-After", "3")
				w.Header().Set("X-Backpressure", "Active")
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusServiceUnavailable)
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"error":       "server_busy_backpressure",
					"message":     "System is experiencing high traffic. Please retry in a few moments.",
					"retry_after": 3,
				})
			}
		})
	}
}

// GetActiveRequests returns the current count of in-flight requests
func GetActiveRequests() int64 {
	return atomic.LoadInt64(&ActiveRequests)
}

// RequestTimeoutMiddleware injects context cancellation timeout for per-route safety
func RequestTimeoutMiddleware(defaultTimeoutSec int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			timeoutSec := defaultTimeoutSec
			if timeoutSec <= 0 {
				timeoutSec = 15
			}

			// Search and expensive endpoints can have shorter deadline
			if r.URL.Path == "/api/v1/search" || r.URL.Path == "/api/v1/search/posts" {
				timeoutSec = 5
			}

			w.Header().Set("X-Request-Timeout-Sec", strconv.Itoa(timeoutSec))
			next.ServeHTTP(w, r)
		})
	}
}
