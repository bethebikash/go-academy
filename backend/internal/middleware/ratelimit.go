package middleware

import (
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"

	"go-academy/backend/internal/platform/response"
)

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// IPRateLimiter limits incoming requests per IP address.
type IPRateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
	rate     rate.Limit
	burst    int
}

// NewIPRateLimiter constructs a new IP-based rate limiter.
func NewIPRateLimiter(r rate.Limit, burst int) *IPRateLimiter {
	limiter := &IPRateLimiter{
		visitors: make(map[string]*visitor),
		rate:     r,
		burst:    burst,
	}

	// Periodically remove stale entries (older than 10 minutes)
	go limiter.cleanupStaleVisitors()

	return limiter
}

func (i *IPRateLimiter) getVisitor(ip string) *rate.Limiter {
	i.mu.Lock()
	defer i.mu.Unlock()

	v, exists := i.visitors[ip]
	if !exists {
		limiter := rate.NewLimiter(i.rate, i.burst)
		i.visitors[ip] = &visitor{limiter: limiter, lastSeen: time.Now()}
		return limiter
	}

	v.lastSeen = time.Now()
	return v.limiter
}

func (i *IPRateLimiter) cleanupStaleVisitors() {
	ticker := time.NewTicker(3 * time.Minute)
	for range ticker.C {
		i.mu.Lock()
		for ip, v := range i.visitors {
			if time.Since(v.lastSeen) > 10*time.Minute {
				delete(i.visitors, ip)
			}
		}
		i.mu.Unlock()
	}
}

// Limit returns a middleware that limits requests based on client IP.
func (i *IPRateLimiter) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		limiter := i.getVisitor(ip)

		if !limiter.Allow() {
			response.Error(
				w,
				http.StatusTooManyRequests,
				"RATE_LIMIT_EXCEEDED",
				"Too many requests. Please try again later.",
				nil,
			)
			return
		}

		next.ServeHTTP(w, r)
	})
}
