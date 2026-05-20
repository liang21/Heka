package middleware

import (
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// RateLimiter manages rate limits for specific endpoints
type RateLimiter struct {
	limiter *rate.Limiter
	mu      sync.Mutex
}

// Standard rate limiters for different endpoints
var (
	registerLimiter = &RateLimiter{limiter: rate.NewLimiter(rate.Every(time.Minute), 10)} // 10 req/min
	loginLimiter    = &RateLimiter{limiter: rate.NewLimiter(rate.Every(time.Minute), 20)}    // 20 req/min
	uploadLimiter   = &RateLimiter{limiter: rate.NewLimiter(rate.Every(time.Minute), 5)}     // 5 req/min
)

// RegisterRateLimit applies rate limiting to user registration endpoint
// spec.md §11: Registration rate limited to prevent abuse
func RegisterRateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !registerLimiter.Allow() {
			http.Error(w, "Too many registration attempts. Please try again later.", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// LoginRateLimit applies rate limiting to login endpoint
// spec.md §11: Login rate limited to prevent brute force attacks
func LoginRateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !loginLimiter.Allow() {
			http.Error(w, "Too many login attempts. Please try again later.", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// UploadRateLimit applies rate limiting to file upload endpoint
// spec.md §11: Upload rate limited to prevent resource abuse
func UploadRateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !uploadLimiter.Allow() {
			http.Error(w, "Too many upload attempts. Please try again later.", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// Allow checks if the request is allowed under the rate limit
func (rl *RateLimiter) Allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	return rl.limiter.Allow()
}
