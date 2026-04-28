package middleware

import (
	"net/http"
	"sync"
	"time"

	"stackyard/pkg/response"

	"github.com/labstack/echo/v4"
)

// RateLimiterConfig holds rate limiter configuration
type RateLimiterConfig struct {
	// Requests per time window
	Requests int
	// Time window duration
	Window time.Duration
	// Key function to identify clients (default: IP address)
	KeyFunc func(c echo.Context) string
}

// DefaultRateLimiterConfig returns default rate limiter configuration
func DefaultRateLimiterConfig() RateLimiterConfig {
	return RateLimiterConfig{
		Requests: 60,
		Window:   time.Minute,
		KeyFunc:  DefaultKeyFunc,
	}
}

// DefaultKeyFunc uses client IP address as the key
func DefaultKeyFunc(c echo.Context) string {
	return c.RealIP()
}

// rateLimitEntry tracks requests for a client
type rateLimitEntry struct {
	count   int
	resetAt time.Time
}

// RateLimiter implements a token bucket rate limiter
type RateLimiter struct {
	config  RateLimiterConfig
	clients map[string]*rateLimitEntry
	mu      sync.RWMutex
	cleanup *time.Ticker
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(config RateLimiterConfig) *RateLimiter {
	rl := &RateLimiter{
		config:  config,
		clients: make(map[string]*rateLimitEntry),
		cleanup: time.NewTicker(time.Minute),
	}

	// Start cleanup goroutine
	go rl.cleanupExpired()

	return rl
}

// cleanupExpired removes expired entries
func (rl *RateLimiter) cleanupExpired() {
	for range rl.cleanup.C {
		rl.mu.Lock()
		now := time.Now()
		for key, entry := range rl.clients {
			if now.After(entry.resetAt) {
				delete(rl.clients, key)
			}
		}
		rl.mu.Unlock()
	}
}

// Allow checks if a request is allowed
func (rl *RateLimiter) Allow(key string) (bool, int, time.Time) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	entry, exists := rl.clients[key]

	if !exists || now.After(entry.resetAt) {
		// Create new entry
		rl.clients[key] = &rateLimitEntry{
			count:   1,
			resetAt: now.Add(rl.config.Window),
		}
		return true, rl.config.Requests - 1, now.Add(rl.config.Window)
	}

	if entry.count >= rl.config.Requests {
		// Rate limit exceeded
		return false, 0, entry.resetAt
	}

	// Increment counter
	entry.count++
	return true, rl.config.Requests - entry.count, entry.resetAt
}

// Stop stops the cleanup goroutine
func (rl *RateLimiter) Stop() {
	rl.cleanup.Stop()
}

// RateLimit returns rate limiting middleware
func RateLimit(config ...RateLimiterConfig) echo.MiddlewareFunc {
	var cfg RateLimiterConfig
	if len(config) > 0 {
		cfg = config[0]
	} else {
		cfg = DefaultRateLimiterConfig()
	}

	if cfg.KeyFunc == nil {
		cfg.KeyFunc = DefaultKeyFunc
	}

	limiter := NewRateLimiter(cfg)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			key := cfg.KeyFunc(c)
			allowed, remaining, resetAt := limiter.Allow(key)

			// Set rate limit headers
			c.Response().Header().Set("X-RateLimit-Limit", string(rune(cfg.Requests+'0')))
			c.Response().Header().Set("X-RateLimit-Remaining", string(rune(remaining+'0')))
			c.Response().Header().Set("X-RateLimit-Reset", resetAt.Format(time.RFC3339))

			if !allowed {
				return response.Error(c, http.StatusTooManyRequests, "RATE_LIMIT_EXCEEDED", "Rate limit exceeded. Please try again later.", map[string]interface{}{
					"retry_after": resetAt.Unix(),
				})
			}

			return next(c)
		}
	}
}

// RateLimitWithConfig returns rate limiting middleware with custom config
func RateLimitWithConfig(requests int, window time.Duration) echo.MiddlewareFunc {
	return RateLimit(RateLimiterConfig{
		Requests: requests,
		Window:   window,
		KeyFunc:  DefaultKeyFunc,
	})
}

// RateLimitPerIP returns rate limiting middleware per IP
func RateLimitPerIP(requests int, window time.Duration) echo.MiddlewareFunc {
	return RateLimit(RateLimiterConfig{
		Requests: requests,
		Window:   window,
		KeyFunc:  DefaultKeyFunc,
	})
}

// RateLimitPerUser returns rate limiting middleware per user (requires auth)
func RateLimitPerUser(requests int, window time.Duration) echo.MiddlewareFunc {
	return RateLimit(RateLimiterConfig{
		Requests: requests,
		Window:   window,
		KeyFunc: func(c echo.Context) string {
			// Try to get user ID from context (set by auth middleware)
			if userID := c.Get("user_id"); userID != nil {
				return "user:" + userID.(string)
			}
			// Fallback to IP
			return c.RealIP()
		},
	})
}
