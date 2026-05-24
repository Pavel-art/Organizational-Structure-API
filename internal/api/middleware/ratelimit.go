package middleware

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type RateLimitConfig struct {
	RPS   rate.Limit
	Burst int

	TTL time.Duration
}

func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		RPS:   10,
		Burst: 20,
		TTL:   5 * time.Minute,
	}
}

type clientLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

func RateLimit(cfg RateLimitConfig) func(http.Handler) http.Handler {
	var mu sync.Mutex
	clients := map[string]*clientLimiter{}

	cleanup := time.NewTicker(time.Minute)
	go func() {
		for range cleanup.C {
			mu.Lock()
			now := time.Now()
			for k, v := range clients {
				if now.Sub(v.lastSeen) > cfg.TTL {
					delete(clients, k)
				}
			}
			mu.Unlock()
		}
	}()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := clientIP(r)

			mu.Lock()
			cl, ok := clients[ip]
			if !ok {
				cl = &clientLimiter{limiter: rate.NewLimiter(cfg.RPS, cfg.Burst)}
				clients[ip] = cl
			}
			cl.lastSeen = time.Now()
			lim := cl.limiter
			mu.Unlock()

			if !lim.Allow() {
				http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func clientIP(r *http.Request) string {
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			ip := strings.TrimSpace(parts[0])
			if ip != "" {
				return ip
			}
		}
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil && host != "" {
		return host
	}
	return r.RemoteAddr
}
