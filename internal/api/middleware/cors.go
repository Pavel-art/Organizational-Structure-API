package middleware

import "net/http"

type CORSConfig struct {
	AllowOrigin  string
	AllowMethods string
	AllowHeaders string
}

func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowOrigin:  "*",
		AllowMethods: "GET,POST,PATCH,DELETE,OPTIONS",
		AllowHeaders: "Content-Type, Authorization",
	}
}

func CORS(cfg CORSConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", cfg.AllowOrigin)
			w.Header().Set("Access-Control-Allow-Methods", cfg.AllowMethods)
			w.Header().Set("Access-Control-Allow-Headers", cfg.AllowHeaders)
			w.Header().Set("Vary", "Origin")

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
