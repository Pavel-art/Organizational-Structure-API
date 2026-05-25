package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/Pavel-art/Organizational-Structure-API/configs"
	handler2 "github.com/Pavel-art/Organizational-Structure-API/internal/api/handlers/department"
	"github.com/Pavel-art/Organizational-Structure-API/internal/api/handlers/health"
	"github.com/Pavel-art/Organizational-Structure-API/internal/api/middleware"
	"github.com/Pavel-art/Organizational-Structure-API/internal/api/swagger"
	"github.com/Pavel-art/Organizational-Structure-API/internal/application/services/impl"
	"github.com/Pavel-art/Organizational-Structure-API/internal/core/closer"
	"github.com/Pavel-art/Organizational-Structure-API/internal/core/logger"
	persistdb "github.com/Pavel-art/Organizational-Structure-API/internal/persistence/db"
	"github.com/Pavel-art/Organizational-Structure-API/internal/persistence/migrate"
	repoimpl "github.com/Pavel-art/Organizational-Structure-API/internal/persistence/repositories/impl"
	"github.com/rs/zerolog/log"
	"golang.org/x/time/rate"
)

const (
	serverReadHeaderTimeout = 5 * time.Second
	serverReadTimeout       = 15 * time.Second
	serverWriteTimeout      = 15 * time.Second
	serverIdleTimeout       = 60 * time.Second

	shutdownTimeout = 10 * time.Second
)

func main() {
	configs.Init()

	dbCfg := configs.NewDataBaseConfig()
	logCfg := configs.NewLogConfig()
	srvCfg := configs.NewServerConfig()

	logger.Init(logCfg.Level, logCfg.Format)

	if dbCfg.Url == "" {
		log.Fatal().Msg("DB_URL is required")
	}

	gormDB, err := persistdb.NewPostgres(dbCfg.Url)
	if err != nil {
		log.Fatal().Err(err).Msg("db connect failed")
	}
	sqlDB, err := gormDB.DB()
	if err != nil {
		log.Fatal().Err(err).Msg("db handle failed")
	}

	if err := migrate.Up(sqlDB, "migrations"); err != nil {
		log.Fatal().Err(err).Msg("migrations failed")
	}

	deptRepo := repoimpl.NewDepartmentRepository(gormDB)
	empRepo := repoimpl.NewEmployeeRepository(gormDB)

	deptService := impl.NewDepartmentService(gormDB, deptRepo, empRepo)
	empService := impl.NewEmployeeService(deptRepo, empRepo)

	deptHandler := handler2.NewDepartmentHandler(deptService, empService)
	healthHandler := health.NewHealthHandler(sqlDB)

	mux := http.NewServeMux()
	mux.Handle("/departments", deptHandler)
	mux.Handle("/departments/", deptHandler)
	mux.Handle("/healthz", healthHandler)
	mux.Handle("/swagger/", swagger.SwaggerUIHandler())
	mux.Handle("/openapi.json", swagger.OpenAPIHandler())

	var handler http.Handler = mux
	handler = middleware.RequestLog(handler)

	// Same behavior as internal/api/router: enabled, env overrides defaults.
	enableRateLimit := true
	enableCORS := true

	if enableRateLimit {
		rl := middleware.DefaultRateLimitConfig()
		if rl.RPS <= 0 {
			rl.RPS = rate.Limit(getIntEnv("RATE_LIMIT_RPS", int(rl.RPS)))
		}
		if rl.Burst <= 0 {
			rl.Burst = getIntEnv("RATE_LIMIT_BURST", rl.Burst)
		}
		handler = middleware.RateLimit(rl)(handler)
	}
	if enableCORS {
		handler = middleware.CORS(middleware.DefaultCORSConfig())(handler)
	}

	server := &http.Server{
		Addr:              ":" + srvCfg.Port,
		Handler:           handler,
		ReadHeaderTimeout: serverReadHeaderTimeout,
		ReadTimeout:       serverReadTimeout,
		WriteTimeout:      serverWriteTimeout,
		IdleTimeout:       serverIdleTimeout,
	}

	go func() {
		log.Info().Str("port", srvCfg.Port).Msg("http server started")
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).Msg("http server failed")
		}
	}()

	sigCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	<-sigCtx.Done()
	log.Info().Msg("shutdown signal received")

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	log.Info().Dur("timeout", shutdownTimeout).Msg("graceful shutdown started")

	cl := closer.New()
	cl.Add("db", sqlDB.Close)
	cl.Add("http server", func() error { return server.Shutdown(ctx) })

	if err := cl.Close(); err != nil {
		log.Error().Err(err).Msg("graceful shutdown finished with errors")
		return
	}

	log.Info().Msg("graceful shutdown completed")
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
