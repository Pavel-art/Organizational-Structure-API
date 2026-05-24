package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Pavel-art/Organizational-Structure-API/configs"
	handler2 "github.com/Pavel-art/Organizational-Structure-API/internal/api/handlers/department"
	"github.com/Pavel-art/Organizational-Structure-API/internal/api/handlers/health"
	httpapi "github.com/Pavel-art/Organizational-Structure-API/internal/api/router"
	"github.com/Pavel-art/Organizational-Structure-API/internal/api/swagger"
	"github.com/Pavel-art/Organizational-Structure-API/internal/application/services/impl"
	"github.com/Pavel-art/Organizational-Structure-API/internal/core/closer"
	"github.com/Pavel-art/Organizational-Structure-API/internal/core/logger"
	persistdb "github.com/Pavel-art/Organizational-Structure-API/internal/persistence/db"
	"github.com/Pavel-art/Organizational-Structure-API/internal/persistence/migrate"
	repoimpl "github.com/Pavel-art/Organizational-Structure-API/internal/persistence/repositories/impl"
	"github.com/rs/zerolog/log"
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

	rateLimitCfg := httpapi.RouterConfig{
		EnableCORS:      true,
		EnableRateLimit: true,
	}

	handler := httpapi.NewRouter(
		deptHandler,
		healthHandler,
		swagger.SwaggerUIHandler(),
		swagger.OpenAPIHandler(),
		rateLimitCfg,
	)

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
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("http server failed")
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	sig := <-stop
	log.Info().Str("signal", sig.String()).Msg("shutdown signal received")

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
