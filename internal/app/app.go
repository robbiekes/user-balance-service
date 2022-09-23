package app

import (
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"user-balance-api/config"
	"user-balance-api/internal/controller/http/v1"
	"user-balance-api/internal/service"
	"user-balance-api/internal/service/rediscache"
	"user-balance-api/internal/service/repo"
	"user-balance-api/internal/service/webapi"
	"user-balance-api/pkg/httpserver"
	"user-balance-api/pkg/postgres"
)

func Run(cfg *config.Config) {
	// Cache
	log.Info("Initializing Redis...")
	redisCache := rediscache.NewRedisLib(redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	}))

	// Repository
	log.Info("Initializing repository")
	pg, err := postgres.New(cfg.PG.URL, postgres.MaxPoolSize(cfg.PG.PoolMax))
	if err != nil {
		log.Fatal(fmt.Errorf("app - Run - postgres.New: %w", err))
	}
	defer pg.Close()

	// Web API
	log.Info("Initializing webapi...")
	converterWebApi := webapi.NewConverterAPI(http.DefaultClient)

	// Service
	log.Info("Initializing service...")
	services := service.New(
		repo.New(pg, redisCache),
		converterWebApi,
	)

	// HTTP Server
	log.Info("Initializing http server...")
	handler := echo.New()
	v1.NewRouter(handler, services)
	httpServer := httpserver.New(handler, httpserver.Port(cfg.HTTP.Port))

	// Waiting signal
	log.Info("Configuring graceful shutdown...")
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	select {
	case s := <-interrupt:
		log.Info("app - Run - signal: " + s.String())
	case err = <-httpServer.Notify():
		log.Error(fmt.Errorf("app - Run - httpServer.Notify: %w", err))
	}

	// Graceful shutdown
	log.Info("Shutting down...")
	err = httpServer.Shutdown()
	if err != nil {
		log.Error(fmt.Errorf("app - Run - httpServer.Shutdown: %w", err))
	}
}
