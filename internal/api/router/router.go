package router

import (
	"net/http"
	"os"
	"strconv"

	"github.com/Pavel-art/Organizational-Structure-API/internal/api/middleware"
	"golang.org/x/time/rate"
)

type RouterConfig struct {
	EnableCORS      bool
	EnableRateLimit bool

	RateLimitRPS   rate.Limit
	RateLimitBurst int
}

func NewRouter(deptHandler http.Handler, healthHandler http.Handler, swaggerHandler http.Handler, openAPIHandler http.Handler, cfg RouterConfig) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/departments", deptHandler)
	mux.Handle("/departments/", deptHandler)
	mux.Handle("/healthz", healthHandler)
	mux.Handle("/swagger/", swaggerHandler)
	mux.Handle("/openapi.json", openAPIHandler)

	var handler http.Handler = mux
	handler = middleware.RequestLog(handler)
	if cfg.EnableRateLimit {
		rl := middleware.DefaultRateLimitConfig()
		rl.RPS = cfg.RateLimitRPS
		rl.Burst = cfg.RateLimitBurst
		if rl.RPS <= 0 {
			rl.RPS = rate.Limit(getIntEnv("RATE_LIMIT_RPS", int(rl.RPS)))
		}
		if rl.Burst <= 0 {
			rl.Burst = getIntEnv("RATE_LIMIT_BURST", rl.Burst)
		}
		handler = middleware.RateLimit(rl)(handler)
	}
	if cfg.EnableCORS {
		handler = middleware.CORS(middleware.DefaultCORSConfig())(handler)
	}
	return handler
}

func getIntEnv(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return i
}
