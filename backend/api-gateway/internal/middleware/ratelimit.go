package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// IPRateLimiter manages rate limiting per IP
type IPRateLimiter struct {
	ips map[string]*rate.Limiter
	mu  sync.Mutex
	r   rate.Limit
	b   int
}

// NewIPRateLimiter creates a new IP-based rate limiter
func NewIPRateLimiter(r rate.Limit, b int) *IPRateLimiter {
	return &IPRateLimiter{
		ips: make(map[string]*rate.Limiter),
		r:   r,
		b:   b,
	}
}

// GetLimiter returns the rate limiter for the given IP
func (rl *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if limiter, exists := rl.ips[ip]; exists {
		return limiter
	}

	limiter := rate.NewLimiter(rl.r, rl.b)
	rl.ips[ip] = limiter
	return limiter
}

// RateLimit implements IP-based rate limiting using token bucket algorithm
// 100 requests per minute with burst of 10
func RateLimit(next http.Handler) http.Handler {
	limiter := NewIPRateLimiter(
		rate.Every(time.Minute/100), // 100 req/min
		10,                           // burst
	)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		if !limiter.GetLimiter(ip).Allow() {
			w.WriteHeader(http.StatusTooManyRequests)
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, `{"error":{"code":"RATE_LIMIT_EXCEEDED","message":"Too many requests"},"requestId":"%s"}`, generateRequestID())
			return
		}
		next.ServeHTTP(w, r)
	})
}
