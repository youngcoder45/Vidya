package middleware

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"github.com/schoolos/backend/internal/pkg/ctxuser"
	"github.com/schoolos/backend/internal/pkg/httpx"
)

// RateLimiter is a fixed-window limiter. Redis is the source of truth when
// available; an in-memory fallback keeps the API safe when Redis is down.
type RateLimiter struct {
	rdb    *redis.Client
	limit  int
	window time.Duration
	log    *slog.Logger

	mu    sync.Mutex
	local map[string]int
}

// NewRateLimiter creates a limiter (limit requests per window).
func NewRateLimiter(rdb *redis.Client, limit int, window time.Duration, log *slog.Logger) *RateLimiter {
	return &RateLimiter{rdb: rdb, limit: limit, window: window, log: log, local: map[string]int{}}
}

// Middleware returns a gin handler that limits per (route, user-or-IP).
func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := rl.key(c)
		allowed := rl.allow(c.Request.Context(), key)
		if !allowed {
			httpx.WriteError(c, httpx.ErrRateLimited)
			c.Abort()
			return
		}
		c.Next()
	}
}

func (rl *RateLimiter) key(c *gin.Context) string {
	uid := ctxuser.MustUserID(c.Request.Context())
	if uid.String() != "00000000-0000-0000-0000-000000000000" {
		return fmt.Sprintf("rl:%s:%s", c.FullPath(), uid)
	}
	return fmt.Sprintf("rl:%s:%s", c.FullPath(), c.ClientIP())
}

func (rl *RateLimiter) allow(ctx context.Context, key string) bool {
	if rl.rdb != nil {
		return rl.allowRedis(ctx, key)
	}
	return rl.allowLocal(key)
}

func (rl *RateLimiter) allowRedis(ctx context.Context, key string) bool {
	window := time.Now().UTC().Truncate(rl.window).Unix()
	k := fmt.Sprintf("%s:%d", key, window)
	n, err := rl.rdb.Incr(ctx, k).Result()
	if err != nil {
		rl.log.Warn("rate limit redis error, falling back to local", "err", err)
		return rl.allowLocal(key)
	}
	if n == 1 {
		rl.rdb.Expire(ctx, k, rl.window+time.Minute)
	}
	return n <= int64(rl.limit)
}

func (rl *RateLimiter) allowLocal(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.local[key]++
	if rl.local[key] > rl.limit {
		return false
	}
	// coarse sweep: reset counters every window
	if len(rl.local) > 10_000 {
		rl.local = map[string]int{}
	}
	return true
}

// AuthRateLimit is a stricter limiter for authentication endpoints keyed by
// identifier + IP.
func AuthRateLimit(rdb *redis.Client, limit int, window time.Duration, log *slog.Logger) gin.HandlerFunc {
	rl := NewRateLimiter(rdb, limit, window, log)
	return func(c *gin.Context) {
		key := "rl:auth:" + c.ClientIP() + ":" + c.GetHeader("X-Identifier")
		if !rl.allow(c.Request.Context(), key) {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": httpx.ErrRateLimited})
			c.Abort()
			return
		}
		c.Next()
	}
}
