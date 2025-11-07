package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/JokerTrickster/joker_backend/shared/logger"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// RateLimiter implements IP-based rate limiting
type RateLimiter struct {
	visitors map[string]*visitor
	mu       sync.RWMutex
	rate     rate.Limit
	burst    int
	done     chan struct{}
}

// NewRateLimiter creates a new rate limiter
// rps: requests per second
// burst: maximum burst size
func NewRateLimiter(rps float64, burst int) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		rate:     rate.Limit(rps),
		burst:    burst,
		done:     make(chan struct{}),
	}

	// Clean up old visitors every 3 minutes
	go rl.cleanupVisitors()

	logger.Info("Rate limiter initialized",
		zap.Float64("rps", rps),
		zap.Int("burst", burst),
	)

	return rl
}

// Close stops the cleanup goroutine
func (rl *RateLimiter) Close() {
	close(rl.done)
	logger.Info("Rate limiter cleanup stopped")
}

func (rl *RateLimiter) getVisitor(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, exists := rl.visitors[ip]
	if !exists {
		limiter := rate.NewLimiter(rl.rate, rl.burst)
		rl.visitors[ip] = &visitor{
			limiter:  limiter,
			lastSeen: time.Now(),
		}
		return limiter
	}

	v.lastSeen = time.Now()
	return v.limiter
}

func (rl *RateLimiter) cleanupVisitors() {
	ticker := time.NewTicker(3 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rl.mu.Lock()
			before := len(rl.visitors)
			for ip, v := range rl.visitors {
				if time.Since(v.lastSeen) > 3*time.Minute {
					delete(rl.visitors, ip)
				}
			}
			after := len(rl.visitors)
			rl.mu.Unlock()

			if before != after {
				logger.Info("Rate limiter cleanup completed",
					zap.Int("visitors_before", before),
					zap.Int("visitors_after", after),
					zap.Int("removed", before-after),
				)
			}
		case <-rl.done:
			logger.Info("Rate limiter cleanup goroutine stopped")
			return
		}
	}
}

// Middleware returns the rate limiting middleware
func (rl *RateLimiter) Middleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ip := c.RealIP()
			limiter := rl.getVisitor(ip)

			if !limiter.Allow() {
				logger.Warn("Rate limit exceeded",
					zap.String("ip", ip),
					zap.String("method", c.Request().Method),
					zap.String("uri", c.Request().RequestURI),
				)

				return echo.NewHTTPError(
					http.StatusTooManyRequests,
					"Rate limit exceeded. Please try again later.",
				)
			}

			return next(c)
		}
	}
}
